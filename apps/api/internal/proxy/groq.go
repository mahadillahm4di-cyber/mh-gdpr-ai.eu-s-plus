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

const groqBaseURL = "https://api.groq.com/openai/v1/chat/completions"

// GroqAdapter handles communication with the Groq API.
// Groq uses OpenAI-compatible API format.
type GroqAdapter struct {
	apiKey string
	client *http.Client
}

// NewGroqAdapter creates a new Groq adapter.
// SECURITY: API key is passed in, never hardcoded.
func NewGroqAdapter(apiKey string) *GroqAdapter {
	return &GroqAdapter{
		apiKey: apiKey,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *GroqAdapter) Name() Provider      { return ProviderGroq }
func (a *GroqAdapter) DefaultModel() string { return "llama-3.3-70b-versatile" }

func (a *GroqAdapter) HealthCheck() error {
	if a.apiKey == "" {
		return fmt.Errorf("groq: API key not configured")
	}
	return nil
}

// Send sends a non-streaming request to Groq.
func (a *GroqAdapter) Send(req *ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = a.DefaultModel()
	}

	oaiReq := openAIRequest{
		Model:    model,
		Messages: req.Messages,
		Stream:   false,
	}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("groq: marshal error: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, groqBaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("groq: request error: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("groq: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("groq: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var oaiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return nil, fmt.Errorf("groq: decode error: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("groq: no choices in response")
	}

	return &ChatResponse{
		ID:        oaiResp.ID,
		Content:   oaiResp.Choices[0].Message.Content,
		Role:      "assistant",
		Provider:  string(ProviderGroq),
		Model:     oaiResp.Model,
		TokensIn:  oaiResp.Usage.PromptTokens,
		TokensOut: oaiResp.Usage.CompletionTokens,
	}, nil
}

// SendStream sends a streaming request to Groq.
func (a *GroqAdapter) SendStream(req *ChatRequest, chunks chan<- StreamChunk) error {
	defer close(chunks)

	model := req.Model
	if model == "" {
		model = a.DefaultModel()
	}

	oaiReq := openAIRequest{
		Model:    model,
		Messages: req.Messages,
		Stream:   true,
	}

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return fmt.Errorf("groq: marshal error: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, groqBaseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("groq: request error: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("groq: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("groq: API error %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			chunks <- StreamChunk{Done: true, Provider: string(ProviderGroq), Model: model}
			return nil
		}

		var chunk openAIStreamResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			slog.Warn("groq: stream decode error", "error", err)
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			chunks <- StreamChunk{
				Content:  chunk.Choices[0].Delta.Content,
				Provider: string(ProviderGroq),
				Model:    chunk.Model,
			}
		}
	}

	return scanner.Err()
}
