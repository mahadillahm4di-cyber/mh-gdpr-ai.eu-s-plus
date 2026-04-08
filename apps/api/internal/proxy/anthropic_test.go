package proxy

import "testing"

func TestAnthropicAdapter_Name(t *testing.T) {
	a := NewAnthropicAdapter("test-key")
	if a.Name() != ProviderAnthropic {
		t.Errorf("expected %s, got %s", ProviderAnthropic, a.Name())
	}
}

func TestAnthropicAdapter_DefaultModel(t *testing.T) {
	a := NewAnthropicAdapter("test-key")
	model := a.DefaultModel()
	if model == "" {
		t.Error("expected non-empty default model")
	}
}

func TestAnthropicAdapter_HealthCheck_NoKey(t *testing.T) {
	a := NewAnthropicAdapter("")
	if err := a.HealthCheck(); err == nil {
		t.Error("expected error for empty API key")
	}
}

func TestAnthropicAdapter_HealthCheck_WithKey(t *testing.T) {
	a := NewAnthropicAdapter("test-api-key-anthropic")
	if err := a.HealthCheck(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAnthropicAdapter_Interface(t *testing.T) {
	var _ ProviderAdapter = NewAnthropicAdapter("test-key")
}

func TestConvertToAnthropicFormat_SystemMessage(t *testing.T) {
	messages := []ChatMessage{
		{Role: "system", Content: "You are a helpful assistant"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	system, msgs := convertToAnthropicFormat(messages)

	if system != "You are a helpful assistant" {
		t.Errorf("expected system message, got %q", system)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "Hello" {
		t.Errorf("unexpected first message: %+v", msgs[0])
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "Hi there!" {
		t.Errorf("unexpected second message: %+v", msgs[1])
	}
}

func TestConvertToAnthropicFormat_NoSystem(t *testing.T) {
	messages := []ChatMessage{
		{Role: "user", Content: "Hello"},
	}

	system, msgs := convertToAnthropicFormat(messages)

	if system != "" {
		t.Errorf("expected empty system, got %q", system)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func TestConvertToAnthropicFormat_Empty(t *testing.T) {
	system, msgs := convertToAnthropicFormat([]ChatMessage{})

	if system != "" {
		t.Errorf("expected empty system, got %q", system)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}
