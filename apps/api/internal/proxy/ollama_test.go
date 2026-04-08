package proxy

import "testing"

func TestOllamaAdapter_Name(t *testing.T) {
	a := NewOllamaAdapter("http://localhost:11434")
	if a.Name() != ProviderOllama {
		t.Errorf("expected %s, got %s", ProviderOllama, a.Name())
	}
}

func TestOllamaAdapter_DefaultModel(t *testing.T) {
	a := NewOllamaAdapter("http://localhost:11434")
	if a.DefaultModel() != "tinyllama" {
		t.Errorf("expected tinyllama, got %s", a.DefaultModel())
	}
}

func TestOllamaAdapter_Interface(t *testing.T) {
	var _ ProviderAdapter = NewOllamaAdapter("http://localhost:11434")
}

func TestConvertToOllamaFormat(t *testing.T) {
	messages := []ChatMessage{
		{Role: "system", Content: "Be helpful"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi!"},
	}

	result := convertToOllamaFormat(messages)

	if len(result) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(result))
	}
	if result[0].Role != "system" || result[0].Content != "Be helpful" {
		t.Errorf("unexpected first message: %+v", result[0])
	}
	if result[1].Role != "user" || result[1].Content != "Hello" {
		t.Errorf("unexpected second message: %+v", result[1])
	}
	if result[2].Role != "assistant" || result[2].Content != "Hi!" {
		t.Errorf("unexpected third message: %+v", result[2])
	}
}

func TestConvertToOllamaFormat_Empty(t *testing.T) {
	result := convertToOllamaFormat([]ChatMessage{})
	if len(result) != 0 {
		t.Errorf("expected 0 messages, got %d", len(result))
	}
}
