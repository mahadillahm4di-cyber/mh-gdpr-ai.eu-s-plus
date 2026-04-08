package proxy

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AutoSaveStore is the interface for persisting conversations and messages.
// Defined here to avoid circular imports with the memory package.
type AutoSaveStore interface {
	SaveConversationRecord(ctx context.Context, id, userID, title, provider, model string) error
	SaveMessageRecord(ctx context.Context, id, conversationID, role, content, provider string, tokenCount int) error
}

// MemorySummarizer creates memories from conversations.
type MemorySummarizer interface {
	CheckAndSummarize(ctx context.Context, conversationID, userID string) error
}

// ContextInjector detects provider switches and injects context.
// Defined here to avoid circular imports with the injector package.
type ContextInjector interface {
	DetectSwitch(userID string, currentProvider Provider) bool
	InjectContext(ctx context.Context, userID string, messages []ChatMessage) ([]ChatMessage, error)
}

// UserKeyStore retrieves per-user API keys so personal keys take priority over the default.
type UserKeyStore interface {
	GetUserAPIKey(ctx context.Context, userID string, provider Provider) string
}

// ProxyHandler handles incoming chat requests and dispatches to the right provider.
type ProxyHandler struct {
	adapters   map[Provider]ProviderAdapter
	store      AutoSaveStore
	injector   ContextInjector
	summarizer MemorySummarizer
	userKeys   UserKeyStore
}

// NewProxyHandler creates a proxy handler with all configured adapters.
func NewProxyHandler(openaiKey, anthropicKey, groqKey, ollamaHost string) *ProxyHandler {
	h := &ProxyHandler{
		adapters: make(map[Provider]ProviderAdapter),
	}

	if openaiKey != "" {
		h.adapters[ProviderOpenAI] = NewOpenAIAdapter(openaiKey)
	}
	if anthropicKey != "" {
		h.adapters[ProviderAnthropic] = NewAnthropicAdapter(anthropicKey)
	}
	if groqKey != "" {
		h.adapters[ProviderGroq] = NewGroqAdapter(groqKey)
	}
	h.adapters[ProviderOllama] = NewOllamaAdapter(ollamaHost)

	return h
}

// SetMemoryStore sets the store for auto-saving conversations.
func (h *ProxyHandler) SetMemoryStore(store AutoSaveStore) {
	h.store = store
}

// SetInjector sets the context injector for provider switches.
func (h *ProxyHandler) SetInjector(inj ContextInjector) {
	h.injector = inj
}

// SetSummarizer sets the memory summarizer for auto-creating memories.
func (h *ProxyHandler) SetSummarizer(s MemorySummarizer) {
	h.summarizer = s
}

// SetUserKeyStore sets the store for per-user API keys.
func (h *ProxyHandler) SetUserKeyStore(s UserKeyStore) {
	h.userKeys = s
}

// ReloadAdapters updates the provider adapters with new API keys.
func (h *ProxyHandler) ReloadAdapters(openaiKey, anthropicKey, groqKey, ollamaHost string) {
	if openaiKey != "" {
		h.adapters[ProviderOpenAI] = NewOpenAIAdapter(openaiKey)
	} else {
		delete(h.adapters, ProviderOpenAI)
	}
	if anthropicKey != "" {
		h.adapters[ProviderAnthropic] = NewAnthropicAdapter(anthropicKey)
	} else {
		delete(h.adapters, ProviderAnthropic)
	}
	if groqKey != "" {
		h.adapters[ProviderGroq] = NewGroqAdapter(groqKey)
	} else {
		delete(h.adapters, ProviderGroq)
	}
	h.adapters[ProviderOllama] = NewOllamaAdapter(ollamaHost)
}

// ChatCompletions handles POST /api/v1/chat/completions.
// SECURITY: Validates provider header, validates request body, requires auth.
func (h *ProxyHandler) ChatCompletions(c *gin.Context) {
	// 1. Get provider from header
	providerStr := strings.ToLower(c.GetHeader("X-MH-Provider"))
	if providerStr == "" {
		providerStr = "groq"
	}

	if !IsValidProvider(providerStr) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid provider: %s. Valid: openai, anthropic, ollama", providerStr),
		})
		return
	}

	provider := Provider(providerStr)

	// Check for per-user API key first, then fall back to global adapter
	var adapter ProviderAdapter
	userID := c.GetString("user_id")
	if h.userKeys != nil && userID != "" {
		if userKey := h.userKeys.GetUserAPIKey(c.Request.Context(), userID, provider); userKey != "" {
			switch provider {
			case ProviderGroq:
				adapter = NewGroqAdapter(userKey)
			case ProviderOpenAI:
				adapter = NewOpenAIAdapter(userKey)
			case ProviderAnthropic:
				adapter = NewAnthropicAdapter(userKey)
			}
		}
	}
	if adapter == nil {
		var ok bool
		adapter, ok = h.adapters[provider]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("provider %s is not configured (missing API key?)", provider),
			})
			return
		}
	}

	// 2. Parse and validate request body
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// SECURITY: Sanitize messages — reject empty content
	for i, msg := range req.Messages {
		if strings.TrimSpace(msg.Content) == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("message %d has empty content", i),
			})
			return
		}
	}

	req.Provider = provider
	req.UserID = c.GetString("user_id")

	// 3. Inject MH Assistant system prompt as first message
	hasSystemMsg := len(req.Messages) > 0 && req.Messages[0].Role == "system"
	if !hasSystemMsg {
		req.Messages = append([]ChatMessage{{Role: "system", Content: MHAssistantSystemPrompt}}, req.Messages...)
	}

	// 4. Context injection on provider switch
	if h.injector != nil && req.UserID != "" {
		if h.injector.DetectSwitch(req.UserID, provider) {
			injected, err := h.injector.InjectContext(c.Request.Context(), req.UserID, req.Messages)
			if err != nil {
				slog.Warn("context_injection_failed", "error", err)
			} else {
				req.Messages = injected
				slog.Info("context_injected", "user_id", req.UserID, "provider", provider)
			}
		}
	}

	slog.Info("chat_request",
		"provider", provider,
		"model", req.Model,
		"stream", req.Stream,
		"messages_count", len(req.Messages),
		"user_id", req.UserID,
	)

	// 4. Stream or non-stream
	if req.Stream {
		h.handleStream(c, adapter, &req)
	} else {
		h.handleSync(c, adapter, &req)
	}
}

// handleSync handles non-streaming requests.
func (h *ProxyHandler) handleSync(c *gin.Context, adapter ProviderAdapter, req *ChatRequest) {
	resp, err := adapter.Send(req)
	if err != nil {
		slog.Error("proxy_error", "provider", adapter.Name(), "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "provider error: " + err.Error()})
		return
	}

	// Auto-save conversation + messages
	h.autoSave(c.Request.Context(), req, resp.Content, resp.TokensIn, resp.TokensOut)

	c.JSON(http.StatusOK, resp)
}

// handleStream handles streaming SSE requests.
func (h *ProxyHandler) handleStream(c *gin.Context, adapter ProviderAdapter, req *ChatRequest) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	chunks := make(chan StreamChunk, 100)
	ctx := c.Request.Context()

	go func() {
		if err := adapter.SendStream(req, chunks); err != nil {
			slog.Error("stream_error", "provider", adapter.Name(), "error", err)
		}
	}()

	var fullContent strings.Builder

	c.Stream(func(w io.Writer) bool {
		chunk, ok := <-chunks
		if !ok {
			return false
		}
		fullContent.WriteString(chunk.Content)
		c.SSEvent("message", chunk)

		if chunk.Done {
			h.autoSave(ctx, req, fullContent.String(), 0, 0)
			return false
		}
		return true
	})
}

// autoSave persists the conversation and messages to the store.
func (h *ProxyHandler) autoSave(ctx context.Context, req *ChatRequest, assistantContent string, tokensIn, tokensOut int) {
	if h.store == nil || req.UserID == "" {
		return
	}

	convID := uuid.New().String()

	// Title from first user message
	title := ""
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			title = msg.Content
			if len(title) > 100 {
				title = title[:100]
			}
			break
		}
	}

	model := req.Model
	if model == "" {
		if adapter, ok := h.adapters[req.Provider]; ok {
			model = adapter.DefaultModel()
		}
	}

	if err := h.store.SaveConversationRecord(ctx, convID, req.UserID, title, string(req.Provider), model); err != nil {
		slog.Warn("auto_save_conversation_failed", "error", err)
		return
	}

	// Save last user message
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			msgID := uuid.New().String()
			if err := h.store.SaveMessageRecord(ctx, msgID, convID, "user", req.Messages[i].Content, string(req.Provider), tokensIn); err != nil {
				slog.Warn("auto_save_user_msg_failed", "error", err)
			}
			break
		}
	}

	// Save assistant response
	if assistantContent != "" {
		msgID := uuid.New().String()
		if err := h.store.SaveMessageRecord(ctx, msgID, convID, "assistant", assistantContent, string(req.Provider), tokensOut); err != nil {
			slog.Warn("auto_save_assistant_msg_failed", "error", err)
		}
	}

	// Create memory from this conversation
	if h.summarizer != nil {
		if err := h.summarizer.CheckAndSummarize(ctx, convID, req.UserID); err != nil {
			slog.Warn("auto_summarize_failed", "error", err)
		}
	}
}

// GetAvailableProviders returns which providers are configured and healthy.
func (h *ProxyHandler) GetAvailableProviders(c *gin.Context) {
	providers := make([]gin.H, 0, len(h.adapters))
	for name, adapter := range h.adapters {
		healthy := adapter.HealthCheck() == nil
		providers = append(providers, gin.H{
			"name":          name,
			"default_model": adapter.DefaultModel(),
			"healthy":       healthy,
		})
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}
