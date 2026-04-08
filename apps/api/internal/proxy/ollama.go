package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// OllamaAdapter handles communication with a local Ollama instance.
// SECURITY: No API key needed — Ollama runs locally. Data never leaves the machine.
type OllamaAdapter struct {
	host   string
	client *http.Client
}

// NewOllamaAdapter creates a new Ollama adapter.
func NewOllamaAdapter(host string) *OllamaAdapter {
	return &OllamaAdapter{
		host:   host,
		client: &http.Client{Timeout: 300 * time.Second}, // Local models can be slow
	}
}

func (a *OllamaAdapter) Name() Provider    { return ProviderOllama }
func (a *OllamaAdapter) DefaultModel() string { return "llama3" }

func (a *OllamaAdapter) HealthCheck() error {
	resp, err := a.client.Get(a.host + "/api/tags")
	if err != nil {
		return fmt.Errorf("ollama: not reachable at %s: %w", a.host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama: unhealthy, status %d", resp.StatusCode)
	}
	return nil
}

// ollamaRequest is the Ollama API request format.
type ollamaRequest struct {
	Model    string           `json:"model"`
	Messages []ollamaMessage  `json:"messages"`
	Stream   bool             `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaResponse is the Ollama non-streaming response.
type ollamaResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	PromptEvalCount int `json:"prompt_eval_count"`
	EvalCount       int `json:"eval_count"`
}

// ollamaStreamResponse is a single streaming chunk from Ollama.
type ollamaStreamResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// convertToOllamaFormat converts unified messages to Ollama format.
func convertToOllamaFormat(messages []ChatMessage) []ollamaMessage {
	out := make([]ollamaMessage, len(messages))
	for i, m := range messages {
		out[i] = ollamaMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

// Send sends a non-streaming request to Ollama.
func (a *OllamaAdapter) Send(req *ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = a.DefaultModel()
	}

	olReq := ollamaRequest{
		Model:    model,
		Messages: convertToOllamaFormat(req.Messages),
		Stream:   false,
	}

	body, err := json.Marshal(olReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal error: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, a.host+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama: error %d: %s", resp.StatusCode, string(respBody))
	}

	var olResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&olResp); err != nil {
		return nil, fmt.Errorf("ollama: decode error: %w", err)
	}

	return &ChatResponse{
		ID:        fmt.Sprintf("ollama-%d", time.Now().UnixNano()),
		Content:   olResp.Message.Content,
		Role:      "assistant",
		Provider:  string(ProviderOllama),
		Model:     olResp.Model,
		TokensIn:  olResp.PromptEvalCount,
		TokensOut: olResp.EvalCount,
	}, nil
}

// SendStream sends a streaming request to Ollama.
func (a *OllamaAdapter) SendStream(req *ChatRequest, chunks chan<- StreamChunk) error {
	defer close(chunks)

	model := req.Model
	if model == "" {
		model = a.DefaultModel()
	}

	olReq := ollamaRequest{
		Model:    model,
		Messages: convertToOllamaFormat(req.Messages),
		Stream:   true,
	}

	body, err := json.Marshal(olReq)
	if err != nil {
		return fmt.Errorf("ollama: marshal error: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, a.host+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ollama: request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ollama: request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama: error %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var chunk ollamaStreamResponse
		if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
			slog.Warn("ollama: stream decode error", "error", err)
			continue
		}

		if chunk.Done {
			chunks <- StreamChunk{Done: true, Provider: string(ProviderOllama), Model: model}
			return nil
		}

		if chunk.Message.Content != "" {
			chunks <- StreamChunk{
				Content:  chunk.Message.Content,
				Provider: string(ProviderOllama),
				Model:    model,
			}
		}
	}

	return scanner.Err()
}
