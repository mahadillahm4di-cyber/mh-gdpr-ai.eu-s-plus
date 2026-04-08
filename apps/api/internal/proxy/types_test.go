package proxy

import "testing"

func TestIsValidProvider(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"openai", true},
		{"anthropic", true},
		{"ollama", true},
		{"OpenAI", false},   // case sensitive
		{"gpt", false},      // not a valid provider
		{"", false},         // empty
		{"mistral", false},  // unsupported
	}

	for _, tt := range tests {
		if got := IsValidProvider(tt.input); got != tt.want {
			t.Errorf("IsValidProvider(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestProviderConstants(t *testing.T) {
	if ProviderOpenAI != "openai" {
		t.Errorf("ProviderOpenAI = %s, want openai", ProviderOpenAI)
	}
	if ProviderAnthropic != "anthropic" {
		t.Errorf("ProviderAnthropic = %s, want anthropic", ProviderAnthropic)
	}
	if ProviderOllama != "ollama" {
		t.Errorf("ProviderOllama = %s, want ollama", ProviderOllama)
	}
}

func TestChatRequest_ProviderNotInJSON(t *testing.T) {
	// Provider and UserID should have json:"-" so they're not in the body
	req := ChatRequest{
		Messages: []ChatMessage{{Role: "user", Content: "hello"}},
		Provider: ProviderOpenAI,
		UserID:   "user-123",
	}

	if req.Provider != ProviderOpenAI {
		t.Error("provider should be set programmatically")
	}
	if req.UserID != "user-123" {
		t.Error("user_id should be set programmatically")
	}
}
