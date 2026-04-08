// Package middleware provides HTTP middleware for security, CORS, rate limiting, and logging.
// SECURITY: Every response gets hardened headers. CORS is explicit. Rate limiting is enforced.
package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// SecurityHeaders adds hardened HTTP security headers to every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-XSS-Protection", "1; mode=block")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

// CORS handles Cross-Origin Resource Sharing with explicit origins only.
// SECURITY: Never allows wildcard (*) origins.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[strings.TrimSpace(o)] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if originSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-MH-Provider")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ipLimiter stores per-IP rate limiters.
type ipLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rpm      int
}

var limiter *ipLimiter

// RateLimit enforces per-IP request rate limiting.
// SECURITY: Prevents abuse and DoS on public endpoints.
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	limiter = &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rpm:      requestsPerMinute,
	}

	// Cleanup old entries every 5 minutes to prevent memory leak
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			limiter.mu.Lock()
			limiter.limiters = make(map[string]*rate.Limiter)
			limiter.mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter.mu.RLock()
		lim, exists := limiter.limiters[ip]
		limiter.mu.RUnlock()

		if !exists {
			// Allow burst of 10, refill at RPM/60 per second
			rps := rate.Limit(float64(requestsPerMinute) / 60.0)
			lim = rate.NewLimiter(rps, 10)
			limiter.mu.Lock()
			limiter.limiters[ip] = lim
			limiter.mu.Unlock()
		}

		if !lim.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

// RequestSizeLimiter limits the maximum request body size.
// SECURITY: Prevents memory exhaustion from oversized payloads.
func RequestSizeLimiter(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "request body too large",
			})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
