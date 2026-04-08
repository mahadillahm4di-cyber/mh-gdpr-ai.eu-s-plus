package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const anthropicBaseURL = "https://api.anthropic.com/v1/messages"

// AnthropicAdapter handles communication with the Anthropic API.
type AnthropicAdapter struct {
	apiKey string
	client *http.Client
}

// NewAnthropicAdapter creates a new Anthropic adapter.
// SECURITY: API key is passed in, never hardcoded.
func NewAnthropicAdapter(apiKey string) *AnthropicAdapter {
	return &AnthropicAdapter{
		apiKey: apiKey,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *AnthropicAdapter) Name() Provider    { return ProviderAnthropic }
func (a *AnthropicAdapter) DefaultModel() string { return "claude-sonnet-4-20250514" }

func (a *AnthropicAdapter) HealthCheck() error {
	if a.apiKey == "" {
		return fmt.Errorf("anthropic: API key not configured")
	}
	return nil
}

// anthropicRequest is the Anthropic API request format.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse is the Anthropic API response format.
type anthropicResponse struct {
	ID      string `json:"id"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model string `json:"model"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// anthropicStreamEvent is a single SSE event from Anthropic.
type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta,omitempty"`
}

// convertToAnthropicFormat converts unified messages to Anthropic format.
// Anthropic requires system messages separate from the messages array.
func convertToAnthropicFormat(messages []ChatMessage) (string, []anthropicMessage) {
	var system string
	var msgs []anthropicMessage

	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}
		msgs = append(msgs, anthropicMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	return system, msgs
}

// Send sends a non-streaming request to Anthropic.
func (a *AnthropicAdapter) Send(req *ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = a.DefaultModel()
	}

	system, msgs := convertToAnthropicFormat(req.Messages)

	antReq := anthropicRequest{
		Model:     model,
		MaxTokens: 4096,
		System:    system,
		Messages:  msgs,
		Stream:    false,
	}

	body, err := json.Marshal(antReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal error: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, anthropicBaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic: request error: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var antResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&antResp); err != nil {
		return nil, fmt.Errorf("anthropic: decode error: %w", err)
	}

	content := ""
	if len(antResp.Content) > 0 {
		content = antResp.Content[0].Text
	}

	return &ChatResponse{
		ID:        antResp.ID,
		Content:   content,
		Role:      "assistant",
		Provider:  string(ProviderAnthropic),
		Model:     antResp.Model,
		TokensIn:  antResp.Usage.InputTokens,
		TokensOut: antResp.Usage.OutputTokens,
	}, nil
}

// SendStream sends a streaming request to Anthropic.
func (a *AnthropicAdapter) SendStream(req *ChatRequest, chunks chan<- StreamChunk) error {
	defer close(chunks)

	model := req.Model
	if model == "" {
		model = a.DefaultModel()
	}

	system, msgs := convertToAnthropicFormat(req.Messages)

	antReq := anthropicRequest{
		Model:     model,
		MaxTokens: 4096,
		System:    system,
		Messages:  msgs,
		Stream:    true,
	}

	body, err := json.Marshal(antReq)
	if err != nil {
		return fmt.Errorf("anthropic: marshal error: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, anthropicBaseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("anthropic: request error: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("anthropic: API error %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			slog.Warn("anthropic: stream decode error", "error", err)
			continue
		}

		switch event.Type {
		case "content_block_delta":
			if event.Delta != nil && event.Delta.Text != "" {
				chunks <- StreamChunk{
					Content:  event.Delta.Text,
					Provider: string(ProviderAnthropic),
					Model:    model,
				}
			}
		case "message_stop":
			chunks <- StreamChunk{Done: true, Provider: string(ProviderAnthropic), Model: model}
			return nil
		}
	}

	return scanner.Err()
}
