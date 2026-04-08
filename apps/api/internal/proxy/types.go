// Package proxy provides a unified interface for routing AI requests to multiple providers.
// All providers are normalized to a common format internally.
package proxy

import "time"

// Provider identifies an AI provider.
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderOllama    Provider = "ollama"
)

// ValidProviders lists all supported providers for input validation.
// SECURITY: Only accept known providers, reject anything else.
var ValidProviders = map[Provider]bool{
	ProviderOpenAI:    true,
	ProviderAnthropic: true,
	ProviderOllama:    true,
}

// IsValidProvider checks if a provider string is supported.
func IsValidProvider(p string) bool {
	return ValidProviders[Provider(p)]
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string `json:"role" binding:"required,oneof=system user assistant"`
	Content string `json:"content" binding:"required,max=100000"`
}

// ChatRequest is the unified request format sent by clients.
// All providers are adapted from this format.
type ChatRequest struct {
	Messages []ChatMessage `json:"messages" binding:"required,min=1"`
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Provider Provider      `json:"-"` // Set from header, not body
	UserID   string        `json:"-"` // Set from JWT, not body
}

// ChatResponse is the unified response format returned to clients.
type ChatResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Role      string `json:"role"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	TokensIn  int    `json:"tokens_in"`
	TokensOut int    `json:"tokens_out"`
}

// StreamChunk represents a single chunk in a streaming response.
type StreamChunk struct {
	Content  string `json:"content"`
	Done     bool   `json:"done"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// ProviderAdapter is the interface each provider must implement.
type ProviderAdapter interface {
	// Send sends a chat request and returns the full response.
	Send(req *ChatRequest) (*ChatResponse, error)

	// SendStream sends a chat request and streams chunks via the channel.
	// The channel is closed when streaming is complete.
	SendStream(req *ChatRequest, chunks chan<- StreamChunk) error

	// Name returns the provider identifier.
	Name() Provider

	// DefaultModel returns the default model for this provider.
	DefaultModel() string

	// HealthCheck verifies the provider is reachable.
	HealthCheck() error
}

// UsageRecord logs a single API call for cost tracking.
// SECURITY: Never stores message content, only metadata.
type UsageRecord struct {
	Provider  Provider  `json:"provider"`
	Model     string    `json:"model"`
	TokensIn  int       `json:"tokens_in"`
	TokensOut int       `json:"tokens_out"`
	Latency   time.Duration `json:"latency_ms"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
}
