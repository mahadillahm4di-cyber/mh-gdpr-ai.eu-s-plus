package memory

import (
	"context"
	"testing"
	"time"
)

func TestGenerateLocalSummary_SingleMessage(t *testing.T) {
	msgs := []*Message{
		{Role: "user", Content: "How do I write Go tests?"},
	}
	summary := generateLocalSummary(msgs)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	if len(summary) > MaxSummaryLength {
		t.Errorf("summary too long: %d > %d", len(summary), MaxSummaryLength)
	}
}

func TestGenerateLocalSummary_MultipleMessages(t *testing.T) {
	msgs := []*Message{
		{Role: "user", Content: "Tell me about Go"},
		{Role: "assistant", Content: "Go is a programming language"},
		{Role: "user", Content: "What about concurrency?"},
		{Role: "assistant", Content: "Go has goroutines and channels"},
	}
	summary := generateLocalSummary(msgs)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestGenerateLocalSummary_NoUserMessages(t *testing.T) {
	msgs := []*Message{
		{Role: "assistant", Content: "I can help you"},
	}
	summary := generateLocalSummary(msgs)
	if summary != "Empty conversation" {
		t.Errorf("expected 'Empty conversation', got %q", summary)
	}
}

func TestGenerateLocalSummary_Empty(t *testing.T) {
	summary := generateLocalSummary([]*Message{})
	if summary != "Empty conversation" {
		t.Errorf("expected 'Empty conversation', got %q", summary)
	}
}

func TestDetectTheme(t *testing.T) {
	msgs := []*Message{
		{Provider: "openai"},
		{Provider: "openai"},
		{Provider: "anthropic"},
	}
	theme := detectTheme(msgs)
	if theme != "openai" {
		t.Errorf("expected 'openai', got %q", theme)
	}
}

func TestDetectTheme_Empty(t *testing.T) {
	theme := detectTheme([]*Message{})
	if theme != "general" {
		t.Errorf("expected 'general', got %q", theme)
	}
}

func TestCalculateImportance(t *testing.T) {
	// Short conversation
	short := []*Message{
		{Content: "hello"},
	}
	imp := calculateImportance(short)
	if imp < 0.0 || imp > 1.0 {
		t.Errorf("importance out of range: %f", imp)
	}

	// Long conversation with lots of content
	long := make([]*Message, 20)
	for i := range long {
		long[i] = &Message{Content: "This is a long message with a lot of content to test importance calculation"}
	}
	impLong := calculateImportance(long)
	if impLong <= imp {
		t.Errorf("longer conversation should have higher importance: %f <= %f", impLong, imp)
	}
}

func TestCalculateImportance_Empty(t *testing.T) {
	imp := calculateImportance([]*Message{})
	if imp != 0.5 {
		t.Errorf("expected 0.5 for empty, got %f", imp)
	}
}

func TestSummarizer_CheckAndSummarize_BelowThreshold(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	conv := &Conversation{
		ID: NewID(), UserID: "user-1", Provider: "openai", Model: "gpt-4o",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	store.SaveConversation(ctx, conv)

	// Add fewer messages than threshold
	for i := 0; i < SummarizeThreshold-1; i++ {
		store.SaveMessage(ctx, &Message{
			ID: NewID(), ConversationID: conv.ID, Role: "user",
			Content: "test", Provider: "openai", CreatedAt: time.Now(),
		})
	}

	summarizer := NewSummarizer(store)
	if err := summarizer.CheckAndSummarize(ctx, conv.ID, "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not have created a memory
	mems, _ := store.GetMemories(ctx, "user-1")
	if len(mems) != 0 {
		t.Errorf("expected no memories below threshold, got %d", len(mems))
	}
}

func TestSummarizer_CheckAndSummarize_AboveThreshold(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	conv := &Conversation{
		ID: NewID(), UserID: "user-1", Provider: "openai", Model: "gpt-4o",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	store.SaveConversation(ctx, conv)

	// Add enough messages
	for i := 0; i < SummarizeThreshold+5; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		store.SaveMessage(ctx, &Message{
			ID: NewID(), ConversationID: conv.ID, Role: role,
			Content: "test message content", Provider: "openai", CreatedAt: time.Now(),
		})
	}

	summarizer := NewSummarizer(store)
	if err := summarizer.CheckAndSummarize(ctx, conv.ID, "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have created a memory
	mems, _ := store.GetMemories(ctx, "user-1")
	if len(mems) != 1 {
		t.Fatalf("expected 1 memory, got %d", len(mems))
	}
	if mems[0].Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestSummarizer_CheckAndSummarize_NoDuplicate(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	conv := &Conversation{
		ID: NewID(), UserID: "user-1", Provider: "openai", Model: "gpt-4o",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	store.SaveConversation(ctx, conv)

	for i := 0; i < SummarizeThreshold+5; i++ {
		store.SaveMessage(ctx, &Message{
			ID: NewID(), ConversationID: conv.ID, Role: "user",
			Content: "test", Provider: "openai", CreatedAt: time.Now(),
		})
	}

	summarizer := NewSummarizer(store)
	summarizer.CheckAndSummarize(ctx, conv.ID, "user-1")
	summarizer.CheckAndSummarize(ctx, conv.ID, "user-1") // Second call

	mems, _ := store.GetMemories(ctx, "user-1")
	if len(mems) != 1 {
		t.Errorf("expected 1 memory (no duplicate), got %d", len(mems))
	}
}
