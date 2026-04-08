package injector

import (
	"context"
	"testing"
	"time"

	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/memory"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/proxy"
)

// mockStore implements memory.Store for testing.
type mockStore struct {
	memories []*memory.Memory
	messages []*memory.Message
}

func (m *mockStore) SaveConversation(ctx context.Context, conv *memory.Conversation) error {
	return nil
}
func (m *mockStore) GetConversation(ctx context.Context, id, userID string) (*memory.Conversation, error) {
	return nil, nil
}
func (m *mockStore) ListConversations(ctx context.Context, userID string, limit, offset int) ([]*memory.Conversation, error) {
	return nil, nil
}
func (m *mockStore) UpdateConversationTitle(ctx context.Context, id, userID, title string) error {
	return nil
}
func (m *mockStore) DeleteConversation(ctx context.Context, id, userID string) error {
	return nil
}
func (m *mockStore) SaveMessage(ctx context.Context, msg *memory.Message) error { return nil }
func (m *mockStore) GetMessages(ctx context.Context, conversationID, userID string) ([]*memory.Message, error) {
	return nil, nil
}
func (m *mockStore) GetRecentMessages(ctx context.Context, userID string, limit int) ([]*memory.Message, error) {
	if len(m.messages) > limit {
		return m.messages[:limit], nil
	}
	return m.messages, nil
}
func (m *mockStore) SaveMemory(ctx context.Context, mem *memory.Memory) error { return nil }
func (m *mockStore) GetMemory(ctx context.Context, id, userID string) (*memory.Memory, error) {
	return nil, nil
}
func (m *mockStore) GetMemories(ctx context.Context, userID string) ([]*memory.Memory, error) {
	return m.memories, nil
}
func (m *mockStore) SearchMemories(ctx context.Context, userID, query string) ([]*memory.Memory, error) {
	return nil, nil
}
func (m *mockStore) DeleteMemory(ctx context.Context, id, userID string) error { return nil }
func (m *mockStore) Close() error                                              { return nil }

func TestDetectSwitch_FirstRequest(t *testing.T) {
	ci := NewContextInjector(&mockStore{})

	switched := ci.DetectSwitch("user-1", proxy.ProviderOpenAI)
	if switched {
		t.Error("first request should not be a switch")
	}
}

func TestDetectSwitch_SameProvider(t *testing.T) {
	ci := NewContextInjector(&mockStore{})

	ci.DetectSwitch("user-1", proxy.ProviderOpenAI)
	switched := ci.DetectSwitch("user-1", proxy.ProviderOpenAI)
	if switched {
		t.Error("same provider should not be a switch")
	}
}

func TestDetectSwitch_DifferentProvider(t *testing.T) {
	ci := NewContextInjector(&mockStore{})

	ci.DetectSwitch("user-1", proxy.ProviderOpenAI)
	switched := ci.DetectSwitch("user-1", proxy.ProviderAnthropic)
	if !switched {
		t.Error("different provider should be a switch")
	}
}

func TestDetectSwitch_MultipleUsers(t *testing.T) {
	ci := NewContextInjector(&mockStore{})

	ci.DetectSwitch("user-1", proxy.ProviderOpenAI)
	ci.DetectSwitch("user-2", proxy.ProviderAnthropic)

	// User 1 switches
	if !ci.DetectSwitch("user-1", proxy.ProviderAnthropic) {
		t.Error("user-1 should detect switch")
	}
	// User 2 does not switch
	if ci.DetectSwitch("user-2", proxy.ProviderAnthropic) {
		t.Error("user-2 should not detect switch (same provider)")
	}
}

func TestInjectContext_WithMemories(t *testing.T) {
	store := &mockStore{
		memories: []*memory.Memory{
			{
				ID:        "mem-1",
				UserID:    "user-1",
				Summary:   "User likes Go programming",
				Theme:     "coding",
				CreatedAt: time.Now(),
			},
			{
				ID:        "mem-2",
				UserID:    "user-1",
				Summary:   "User is building an AI project",
				Theme:     "project",
				CreatedAt: time.Now(),
			},
		},
	}
	ci := NewContextInjector(store)
	ctx := context.Background()

	messages := []proxy.ChatMessage{
		{Role: "user", Content: "Hello, who am I?"},
	}

	injected, err := ci.InjectContext(ctx, "user-1", messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have system message prepended
	if len(injected) != 2 {
		t.Fatalf("expected 2 messages (system + user), got %d", len(injected))
	}
	if injected[0].Role != "system" {
		t.Errorf("expected system message first, got %q", injected[0].Role)
	}
	if injected[1].Content != "Hello, who am I?" {
		t.Errorf("user message should be preserved")
	}
}

func TestInjectContext_NoMemories_FallsBackToMessages(t *testing.T) {
	store := &mockStore{
		memories: []*memory.Memory{},
		messages: []*memory.Message{
			{Role: "user", Content: "I love Go", CreatedAt: time.Now()},
			{Role: "assistant", Content: "Go is great!", CreatedAt: time.Now()},
		},
	}
	ci := NewContextInjector(store)
	ctx := context.Background()

	messages := []proxy.ChatMessage{
		{Role: "user", Content: "Continue"},
	}

	injected, err := ci.InjectContext(ctx, "user-1", messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(injected) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(injected))
	}
	if injected[0].Role != "system" {
		t.Error("expected system message with recent history")
	}
}

func TestInjectContext_NoData(t *testing.T) {
	store := &mockStore{}
	ci := NewContextInjector(store)
	ctx := context.Background()

	messages := []proxy.ChatMessage{
		{Role: "user", Content: "Hello"},
	}

	injected, err := ci.InjectContext(ctx, "user-1", messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No context to inject — return messages unchanged
	if len(injected) != 1 {
		t.Errorf("expected 1 message (no context), got %d", len(injected))
	}
}

func TestInjectContext_MergesWithExistingSystem(t *testing.T) {
	store := &mockStore{
		memories: []*memory.Memory{
			{Summary: "User is a developer", Theme: "profile", CreatedAt: time.Now()},
		},
	}
	ci := NewContextInjector(store)
	ctx := context.Background()

	messages := []proxy.ChatMessage{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
	}

	injected, err := ci.InjectContext(ctx, "user-1", messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should merge system messages, not duplicate
	if len(injected) != 2 {
		t.Fatalf("expected 2 messages (merged system + user), got %d", len(injected))
	}
	if injected[0].Role != "system" {
		t.Error("expected system message")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long", 10, "this is to..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}
