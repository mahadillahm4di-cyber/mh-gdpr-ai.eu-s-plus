// mh-gdpr-ai.eu S+ — API Server entrypoint.
//
// This is the ONLY entrypoint. It initializes config, database, middleware,
// routes, and starts the HTTP server.
//
// SECURITY: All middleware is applied globally. No route escapes security headers,
// CORS, rate limiting, or request size limits.
package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/auth"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/config"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/injector"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/memory"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/middleware"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/proxy"
)

func main() {
	// Structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Load config from environment
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── Database ──
	store, err := memory.NewSQLiteStore("data/mh-gdpr.db", cfg.MemoryEncryptionKey)
	if err != nil {
		slog.Error("failed to init database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// ── Handlers ──
	// Dev secret fallback — in production, cfg.validate() enforces a real key
	secretKey := cfg.SecretKey
	if secretKey == "" {
		secretKey = "dev-secret-key-not-for-production-use"
	}

	userStore := auth.NewSQLiteUserStore(store.DB())
	authHandler := auth.NewHandler(userStore, secretKey)
	memoryHandler := memory.NewHandler(store)
	proxyHandler := proxy.NewProxyHandler(cfg.OpenAIAPIKey, cfg.AnthropicAPIKey, cfg.OllamaHost)
	contextInjector := injector.NewContextInjector(store)

	// Wire injector + memory store into proxy handler for auto-save and context injection
	proxyHandler.SetMemoryStore(store)
	proxyHandler.SetInjector(contextInjector)

	// Initialize router
	r := gin.New()

	// ── Global Middleware (applied to ALL routes) ──
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	r.Use(middleware.RateLimit(cfg.RateLimitRPM))
	r.Use(middleware.RequestSizeLimiter(1 * 1024 * 1024))

	// ── Public Routes ──
	v1 := r.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"version": "0.1.0",
			})
		})

		// Auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
		}

		// Available providers (public — lets the UI know what's configured)
		v1.GET("/providers", proxyHandler.GetAvailableProviders)
	}

	// ── Protected Routes (require JWT) ──
	protected := v1.Group("/")
	protected.Use(middleware.AuthRequired(secretKey))
	{
		// Chat proxy
		protected.POST("/chat/completions", proxyHandler.ChatCompletions)

		// Memories
		protected.GET("/memories", memoryHandler.ListMemories)
		protected.GET("/memories/search", memoryHandler.SearchMemories)
		protected.GET("/memories/:id", memoryHandler.GetMemory)
		protected.DELETE("/memories/:id", memoryHandler.DeleteMemory)

		// Conversations
		protected.GET("/conversations", memoryHandler.ListConversations)
		protected.GET("/conversations/:id", memoryHandler.GetConversation)
		protected.DELETE("/conversations/:id", memoryHandler.DeleteConversation)
	}

	// ── Start Server ──
	addr := cfg.Host + ":" + cfg.Port
	slog.Info("starting mh-gdpr-ai server",
		"address", addr,
		"env", cfg.Env,
	)

	if err := r.Run(addr); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
