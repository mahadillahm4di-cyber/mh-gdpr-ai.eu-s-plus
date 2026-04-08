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

const openAIBaseURL = "https://api.openai.com/v1/chat/completions"

// OpenAIAdapter handles communication with the OpenAI API.
type OpenAIAdapter struct {
	apiKey string
	client *http.Client
}

// NewOpenAIAdapter creates a new OpenAI adapter.
// SECURITY: API key is passed in, never hardcoded.
func NewOpenAIAdapter(apiKey string) *OpenAIAdapter {
	return &OpenAIAdapter{
		apiKey: apiKey,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *OpenAIAdapter) Name() Provider    { return ProviderOpenAI }
func (a *OpenAIAdapter) DefaultModel() string { return "gpt-4o" }

func (a *OpenAIAdapter) HealthCheck() error {
	if a.apiKey == "" {
		return fmt.Errorf("openai: API key not configured")
	}
	return nil
}

// openAIRequest is the OpenAI API request format.
type openAIRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// openAIResponse is the OpenAI API response format.
type openAIResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// openAIStreamResponse is a single SSE chunk from OpenAI.
type openAIStreamResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Model string `json:"model"`
}

// Send sends a non-streaming request to OpenAI.
func (a *OpenAIAdapter) Send(req *ChatRequest) (*ChatResponse, error) {
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
		return nil, fmt.Errorf("openai: marshal error: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, openAIBaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai: request error: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var oaiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return nil, fmt.Errorf("openai: decode error: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices in response")
	}

	return &ChatResponse{
		ID:        oaiResp.ID,
		Content:   oaiResp.Choices[0].Message.Content,
		Role:      "assistant",
		Provider:  string(ProviderOpenAI),
		Model:     oaiResp.Model,
		TokensIn:  oaiResp.Usage.PromptTokens,
		TokensOut: oaiResp.Usage.CompletionTokens,
	}, nil
}

// SendStream sends a streaming request to OpenAI.
func (a *OpenAIAdapter) SendStream(req *ChatRequest, chunks chan<- StreamChunk) error {
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
		return fmt.Errorf("openai: marshal error: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, openAIBaseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("openai: request error: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai: API error %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			chunks <- StreamChunk{Done: true, Provider: string(ProviderOpenAI), Model: model}
			return nil
		}

		var chunk openAIStreamResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			slog.Warn("openai: stream decode error", "error", err)
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			chunks <- StreamChunk{
				Content:  chunk.Choices[0].Delta.Content,
				Provider: string(ProviderOpenAI),
				Model:    chunk.Model,
			}
		}
	}

	return scanner.Err()
}
