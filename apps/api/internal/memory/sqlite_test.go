package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	store, err := NewSQLiteStore(dbPath, "")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	// Create a test user
	_, err = store.db.Exec(
		`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		"user-1", "test@example.com", "hashedpw",
	)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return store
}

func TestNewSQLiteStore_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "subdir", "test.db")
	store, err := NewSQLiteStore(dbPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer store.Close()

	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}

func TestSQLiteStore_SaveAndGetConversation(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	conv := &Conversation{
		ID:        NewID(),
		UserID:    "user-1",
		Title:     "Test Conversation",
		Provider:  "openai",
		Model:     "gpt-4o",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.SaveConversation(ctx, conv); err != nil {
		t.Fatalf("save error: %v", err)
	}

	got, err := store.GetConversation(ctx, conv.ID, "user-1")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if got == nil {
		t.Fatal("expected conversation, got nil")
	}
	if got.Title != "Test Conversation" {
		t.Errorf("expected title 'Test Conversation', got %q", got.Title)
	}
	if got.Provider != "openai" {
		t.Errorf("expected provider 'openai', got %q", got.Provider)
	}
}

func TestSQLiteStore_GetConversation_WrongUser(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	conv := &Conversation{
		ID:        NewID(),
		UserID:    "user-1",
		Title:     "Private",
		Provider:  "openai",
		Model:     "gpt-4o",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.SaveConversation(ctx, conv)

	got, err := store.GetConversation(ctx, conv.ID, "user-other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil for wrong user, got conversation")
	}
}

func TestSQLiteStore_ListConversations(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		store.SaveConversation(ctx, &Conversation{
			ID:        NewID(),
			UserID:    "user-1",
			Title:     "Conv",
			Provider:  "openai",
			Model:     "gpt-4o",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	convs, err := store.ListConversations(ctx, "user-1", 3, 0)
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(convs) != 3 {
		t.Errorf("expected 3 conversations, got %d", len(convs))
	}

	convs, err = store.ListConversations(ctx, "user-1", 10, 3)
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(convs) != 2 {
		t.Errorf("expected 2 conversations with offset, got %d", len(convs))
	}
}

func TestSQLiteStore_DeleteConversation(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	conv := &Conversation{
		ID:        NewID(),
		UserID:    "user-1",
		Title:     "To Delete",
		Provider:  "openai",
		Model:     "gpt-4o",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.SaveConversation(ctx, conv)

	if err := store.DeleteConversation(ctx, conv.ID, "user-1"); err != nil {
		t.Fatalf("delete error: %v", err)
	}

	got, _ := store.GetConversation(ctx, conv.ID, "user-1")
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestSQLiteStore_SaveAndGetMessages(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	conv := &Conversation{
		ID: NewID(), UserID: "user-1", Title: "Test", Provider: "openai", Model: "gpt-4o",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	store.SaveConversation(ctx, conv)

	msg := &Message{
		ID:             NewID(),
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "Hello, how are you?",
		Provider:       "openai",
		TokenCount:     5,
		CreatedAt:      time.Now(),
	}

	if err := store.SaveMessage(ctx, msg); err != nil {
		t.Fatalf("save message error: %v", err)
	}

	msgs, err := store.GetMessages(ctx, conv.ID, "user-1")
	if err != nil {
		t.Fatalf("get messages error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "Hello, how are you?" {
		t.Errorf("expected original content, got %q", msgs[0].Content)
	}
	if msgs[0].Role != "user" {
		t.Errorf("expected role 'user', got %q", msgs[0].Role)
	}
}

func TestSQLiteStore_MessageEncryption(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "encrypted.db")
	key := generateTestKey()

	store, err := NewSQLiteStore(dbPath, key)
	if err != nil {
		t.Fatalf("create store error: %v", err)
	}
	defer store.Close()

	store.db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		"user-1", "test@example.com", "hash")

	ctx := context.Background()
	conv := &Conversation{
		ID: NewID(), UserID: "user-1", Provider: "openai", Model: "gpt-4o",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	store.SaveConversation(ctx, conv)

	msg := &Message{
		ID:             NewID(),
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "This should be encrypted at rest",
		Provider:       "openai",
		CreatedAt:      time.Now(),
	}
	store.SaveMessage(ctx, msg)

	// Read raw from DB — should be encrypted
	var rawContent string
	store.db.QueryRow(`SELECT content FROM messages WHERE id = ?`, msg.ID).Scan(&rawContent)
	if rawContent == "This should be encrypted at rest" {
		t.Error("content should be encrypted in DB, but found plaintext")
	}

	// Read through store — should be decrypted
	msgs, err := store.GetMessages(ctx, conv.ID, "user-1")
	if err != nil {
		t.Fatalf("get messages error: %v", err)
	}
	if msgs[0].Content != "This should be encrypted at rest" {
		t.Errorf("expected decrypted content, got %q", msgs[0].Content)
	}
}

func TestSQLiteStore_SaveAndGetMemory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	mem := &Memory{
		ID:                    NewID(),
		UserID:                "user-1",
		Summary:               "User discussed AI architecture",
		SourceConversationIDs: `["conv-1","conv-2"]`,
		Theme:                 "architecture",
		Importance:            0.8,
		PositionX:             1.0,
		PositionY:             2.0,
		PositionZ:             3.0,
		CreatedAt:             time.Now(),
	}

	if err := store.SaveMemory(ctx, mem); err != nil {
		t.Fatalf("save memory error: %v", err)
	}

	got, err := store.GetMemory(ctx, mem.ID, "user-1")
	if err != nil {
		t.Fatalf("get memory error: %v", err)
	}
	if got == nil {
		t.Fatal("expected memory, got nil")
	}
	if got.Summary != "User discussed AI architecture" {
		t.Errorf("expected summary, got %q", got.Summary)
	}
	if got.Theme != "architecture" {
		t.Errorf("expected theme 'architecture', got %q", got.Theme)
	}
	if got.Importance != 0.8 {
		t.Errorf("expected importance 0.8, got %f", got.Importance)
	}
}

func TestSQLiteStore_GetMemories(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		store.SaveMemory(ctx, &Memory{
			ID:                    NewID(),
			UserID:                "user-1",
			Summary:               "Memory",
			SourceConversationIDs: "[]",
			Theme:                 "test",
			Importance:            0.5,
			CreatedAt:             time.Now(),
		})
	}

	mems, err := store.GetMemories(ctx, "user-1")
	if err != nil {
		t.Fatalf("get memories error: %v", err)
	}
	if len(mems) != 3 {
		t.Errorf("expected 3 memories, got %d", len(mems))
	}
}

func TestSQLiteStore_SearchMemories(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	store.SaveMemory(ctx, &Memory{
		ID: NewID(), UserID: "user-1", Summary: "Discussed Go programming",
		SourceConversationIDs: "[]", Theme: "coding", Importance: 0.7, CreatedAt: time.Now(),
	})
	store.SaveMemory(ctx, &Memory{
		ID: NewID(), UserID: "user-1", Summary: "Talked about cooking",
		SourceConversationIDs: "[]", Theme: "food", Importance: 0.3, CreatedAt: time.Now(),
	})

	// Search by theme
	results, err := store.SearchMemories(ctx, "user-1", "coding")
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'coding', got %d", len(results))
	}

	// Search by summary content (dev mode — no encryption)
	results, err = store.SearchMemories(ctx, "user-1", "cooking")
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'cooking', got %d", len(results))
	}
}

func TestSQLiteStore_DeleteMemory(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	mem := &Memory{
		ID: NewID(), UserID: "user-1", Summary: "To delete",
		SourceConversationIDs: "[]", Importance: 0.5, CreatedAt: time.Now(),
	}
	store.SaveMemory(ctx, mem)

	if err := store.DeleteMemory(ctx, mem.ID, "user-1"); err != nil {
		t.Fatalf("delete error: %v", err)
	}

	got, _ := store.GetMemory(ctx, mem.ID, "user-1")
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestSQLiteStore_UpdateConversationTitle(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	conv := &Conversation{
		ID: NewID(), UserID: "user-1", Title: "Original", Provider: "openai", Model: "gpt-4o",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	store.SaveConversation(ctx, conv)

	if err := store.UpdateConversationTitle(ctx, conv.ID, "user-1", "Updated Title"); err != nil {
		t.Fatalf("update error: %v", err)
	}

	got, _ := store.GetConversation(ctx, conv.ID, "user-1")
	if got.Title != "Updated Title" {
		t.Errorf("expected 'Updated Title', got %q", got.Title)
	}
}

func TestSQLiteStore_GetRecentMessages(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	conv := &Conversation{
		ID: NewID(), UserID: "user-1", Provider: "openai", Model: "gpt-4o",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	store.SaveConversation(ctx, conv)

	for i := 0; i < 5; i++ {
		store.SaveMessage(ctx, &Message{
			ID:             NewID(),
			ConversationID: conv.ID,
			Role:           "user",
			Content:        "message",
			Provider:       "openai",
			CreatedAt:      time.Now(),
		})
	}

	msgs, err := store.GetRecentMessages(ctx, "user-1", 3)
	if err != nil {
		t.Fatalf("get recent error: %v", err)
	}
	if len(msgs) != 3 {
		t.Errorf("expected 3 recent messages, got %d", len(msgs))
	}
}

func TestSQLiteStore_SaveConversationRecord(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.SaveConversationRecord(ctx, NewID(), "user-1", "Auto Title", "anthropic", "claude-sonnet")
	if err != nil {
		t.Fatalf("save conversation record error: %v", err)
	}

	convs, _ := store.ListConversations(ctx, "user-1", 10, 0)
	if len(convs) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(convs))
	}
	if convs[0].Title != "Auto Title" {
		t.Errorf("expected title 'Auto Title', got %q", convs[0].Title)
	}
}

func TestSQLiteStore_SaveMessageRecord(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	convID := NewID()
	store.SaveConversationRecord(ctx, convID, "user-1", "Test", "openai", "gpt-4o")

	err := store.SaveMessageRecord(ctx, NewID(), convID, "user", "Hello world", "openai", 5)
	if err != nil {
		t.Fatalf("save message record error: %v", err)
	}

	msgs, _ := store.GetMessages(ctx, convID, "user-1")
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", msgs[0].Content)
	}
}
