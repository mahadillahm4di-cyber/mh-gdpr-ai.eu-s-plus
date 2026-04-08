package memory

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using SQLite for local-first storage.
// SECURITY: All message content and memory summaries are encrypted with AES-256-GCM.
type SQLiteStore struct {
	db            *sql.DB
	encryptionKey string
}

// NewSQLiteStore creates a new SQLite store at the given path.
func NewSQLiteStore(dbPath, encryptionKey string) (*SQLiteStore, error) {
	// Ensure data directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	// SECURITY: WAL mode for performance + busy timeout to avoid locks
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(1) // SQLite is single-writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	store := &SQLiteStore{db: db, encryptionKey: encryptionKey}

	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	slog.Info("sqlite_store_ready", "path", dbPath)
	return store, nil
}

// migrate creates tables if they don't exist.
func (s *SQLiteStore) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id),
			title TEXT DEFAULT '',
			provider TEXT NOT NULL,
			model TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_user ON conversations(user_id)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
			role TEXT NOT NULL CHECK(role IN ('system', 'user', 'assistant')),
			content TEXT NOT NULL,
			provider TEXT NOT NULL,
			token_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id)`,
		`CREATE TABLE IF NOT EXISTS memories (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id),
			summary TEXT NOT NULL,
			source_conversation_ids TEXT NOT NULL DEFAULT '[]',
			theme TEXT DEFAULT '',
			importance REAL DEFAULT 0.5,
			position_x REAL DEFAULT 0.0,
			position_y REAL DEFAULT 0.0,
			position_z REAL DEFAULT 0.0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_user ON memories(user_id)`,
		`CREATE TABLE IF NOT EXISTS user_settings (
			user_id TEXT PRIMARY KEY REFERENCES users(id),
			openai_api_key TEXT DEFAULT '',
			anthropic_api_key TEXT DEFAULT '',
			groq_api_key TEXT DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("migration failed: %w\nQuery: %s", err, q)
		}
	}

	// Add groq_api_key column if missing (migration for existing databases)
	s.db.Exec(`ALTER TABLE user_settings ADD COLUMN groq_api_key TEXT DEFAULT ''`)

	// Create default local user for anonymous/dev mode access
	_, _ = s.db.Exec(
		`INSERT OR IGNORE INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		"local-user", "local@localhost", "no-password-local-only",
	)

	return nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

// DB returns the underlying *sql.DB for shared access (e.g., auth store).
func (s *SQLiteStore) DB() *sql.DB { return s.db }

// ── Conversations ──

func (s *SQLiteStore) SaveConversation(ctx context.Context, conv *Conversation) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO conversations (id, user_id, title, provider, model, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		conv.ID, conv.UserID, conv.Title, conv.Provider, conv.Model, conv.CreatedAt, conv.UpdatedAt,
	)
	return err
}

func (s *SQLiteStore) GetConversation(ctx context.Context, id, userID string) (*Conversation, error) {
	conv := &Conversation{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, provider, model, created_at, updated_at
		 FROM conversations WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.Provider, &conv.Model, &conv.CreatedAt, &conv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return conv, err
}

func (s *SQLiteStore) ListConversations(ctx context.Context, userID string, limit, offset int) ([]*Conversation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, title, provider, model, created_at, updated_at
		 FROM conversations WHERE user_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []*Conversation
	for rows.Next() {
		c := &Conversation{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Provider, &c.Model, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		convs = append(convs, c)
	}
	return convs, rows.Err()
}

func (s *SQLiteStore) UpdateConversationTitle(ctx context.Context, id, userID, title string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE conversations SET title = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		title, time.Now(), id, userID,
	)
	return err
}

func (s *SQLiteStore) DeleteConversation(ctx context.Context, id, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM conversations WHERE id = ? AND user_id = ?`, id, userID,
	)
	return err
}

// ── Messages ──

// SaveMessage encrypts content and stores the message.
// SECURITY: Content is encrypted before writing to disk.
func (s *SQLiteStore) SaveMessage(ctx context.Context, msg *Message) error {
	encrypted, err := Encrypt(msg.Content, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt message: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO messages (id, conversation_id, role, content, provider, token_count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.ConversationID, msg.Role, encrypted, msg.Provider, msg.TokenCount, msg.CreatedAt,
	)
	return err
}

// GetMessages retrieves and decrypts all messages for a conversation.
func (s *SQLiteStore) GetMessages(ctx context.Context, conversationID, userID string) ([]*Message, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT m.id, m.conversation_id, m.role, m.content, m.provider, m.token_count, m.created_at
		 FROM messages m
		 JOIN conversations c ON m.conversation_id = c.id
		 WHERE m.conversation_id = ? AND c.user_id = ?
		 ORDER BY m.created_at ASC`,
		conversationID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*Message
	for rows.Next() {
		m := &Message{}
		var encrypted string
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &encrypted, &m.Provider, &m.TokenCount, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Content, err = Decrypt(encrypted, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt message: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (s *SQLiteStore) GetRecentMessages(ctx context.Context, userID string, limit int) ([]*Message, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT m.id, m.conversation_id, m.role, m.content, m.provider, m.token_count, m.created_at
		 FROM messages m
		 JOIN conversations c ON m.conversation_id = c.id
		 WHERE c.user_id = ?
		 ORDER BY m.created_at DESC LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*Message
	for rows.Next() {
		m := &Message{}
		var encrypted string
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &encrypted, &m.Provider, &m.TokenCount, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Content, err = Decrypt(encrypted, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt message: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// ── Memories ──

func (s *SQLiteStore) SaveMemory(ctx context.Context, mem *Memory) error {
	encrypted, err := Encrypt(mem.Summary, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt memory: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO memories (id, user_id, summary, source_conversation_ids, theme, importance, position_x, position_y, position_z, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		mem.ID, mem.UserID, encrypted, mem.SourceConversationIDs, mem.Theme,
		mem.Importance, mem.PositionX, mem.PositionY, mem.PositionZ, mem.CreatedAt,
	)
	return err
}

func (s *SQLiteStore) GetMemory(ctx context.Context, id, userID string) (*Memory, error) {
	m := &Memory{}
	var encrypted string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, summary, source_conversation_ids, theme, importance, position_x, position_y, position_z, created_at
		 FROM memories WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&m.ID, &m.UserID, &encrypted, &m.SourceConversationIDs, &m.Theme,
		&m.Importance, &m.PositionX, &m.PositionY, &m.PositionZ, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m.Summary, err = Decrypt(encrypted, s.encryptionKey)
	return m, err
}

func (s *SQLiteStore) GetMemories(ctx context.Context, userID string) ([]*Memory, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, summary, source_conversation_ids, theme, importance, position_x, position_y, position_z, created_at
		 FROM memories WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mems []*Memory
	for rows.Next() {
		m := &Memory{}
		var encrypted string
		if err := rows.Scan(&m.ID, &m.UserID, &encrypted, &m.SourceConversationIDs, &m.Theme,
			&m.Importance, &m.PositionX, &m.PositionY, &m.PositionZ, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Summary, err = Decrypt(encrypted, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt memory: %w", err)
		}
		mems = append(mems, m)
	}
	return mems, rows.Err()
}

func (s *SQLiteStore) SearchMemories(ctx context.Context, userID, query string) ([]*Memory, error) {
	// SECURITY: Use parameterized queries — never concatenate user input
	searchTerm := "%" + strings.ToLower(query) + "%"
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, summary, source_conversation_ids, theme, importance, position_x, position_y, position_z, created_at
		 FROM memories WHERE user_id = ? AND (LOWER(theme) LIKE ? OR LOWER(summary) LIKE ?)
		 ORDER BY importance DESC`, userID, searchTerm, searchTerm,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mems []*Memory
	for rows.Next() {
		m := &Memory{}
		var encrypted string
		if err := rows.Scan(&m.ID, &m.UserID, &encrypted, &m.SourceConversationIDs, &m.Theme,
			&m.Importance, &m.PositionX, &m.PositionY, &m.PositionZ, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Summary, err = Decrypt(encrypted, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt memory: %w", err)
		}
		mems = append(mems, m)
	}
	return mems, rows.Err()
}

func (s *SQLiteStore) DeleteMemory(ctx context.Context, id, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM memories WHERE id = ? AND user_id = ?`, id, userID,
	)
	return err
}

// ── Auto-save helpers (used by the proxy handler) ──

// SaveConversationRecord creates a conversation record for auto-save.
func (s *SQLiteStore) SaveConversationRecord(ctx context.Context, id, userID, title, provider, model string) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO conversations (id, user_id, title, provider, model, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, userID, title, provider, model, now, now,
	)
	return err
}

// SaveMessageRecord encrypts and saves a message for auto-save.
func (s *SQLiteStore) SaveMessageRecord(ctx context.Context, id, conversationID, role, content, provider string, tokenCount int) error {
	encrypted, err := Encrypt(content, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt message: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO messages (id, conversation_id, role, content, provider, token_count, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, conversationID, role, encrypted, provider, tokenCount, time.Now(),
	)
	return err
}

// UserSettings holds API keys for a user.
type UserSettings struct {
	OpenAIKey    string `json:"openai_api_key"`
	AnthropicKey string `json:"anthropic_api_key"`
	GroqKey      string `json:"groq_api_key"`
}

// SaveUserSettings stores encrypted API keys for a user.
func (s *SQLiteStore) SaveUserSettings(ctx context.Context, userID string, settings *UserSettings) error {
	encOpenAI, err := Encrypt(settings.OpenAIKey, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt openai key: %w", err)
	}
	encAnthropic, err := Encrypt(settings.AnthropicKey, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt anthropic key: %w", err)
	}
	encGroq, err := Encrypt(settings.GroqKey, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt groq key: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO user_settings (user_id, openai_api_key, anthropic_api_key, groq_api_key, updated_at)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(user_id) DO UPDATE SET
		 openai_api_key = excluded.openai_api_key,
		 anthropic_api_key = excluded.anthropic_api_key,
		 groq_api_key = excluded.groq_api_key,
		 updated_at = CURRENT_TIMESTAMP`,
		userID, encOpenAI, encAnthropic, encGroq,
	)
	return err
}

// GetUserSettings retrieves decrypted API keys for a user.
func (s *SQLiteStore) GetUserSettings(ctx context.Context, userID string) (*UserSettings, error) {
	var encOpenAI, encAnthropic, encGroq string
	err := s.db.QueryRowContext(ctx,
		`SELECT openai_api_key, anthropic_api_key, groq_api_key FROM user_settings WHERE user_id = ?`, userID,
	).Scan(&encOpenAI, &encAnthropic, &encGroq)
	if err == sql.ErrNoRows {
		return &UserSettings{}, nil
	}
	if err != nil {
		return nil, err
	}

	settings := &UserSettings{}
	settings.OpenAIKey, _ = Decrypt(encOpenAI, s.encryptionKey)
	settings.AnthropicKey, _ = Decrypt(encAnthropic, s.encryptionKey)
	settings.GroqKey, _ = Decrypt(encGroq, s.encryptionKey)
	return settings, nil
}

// GetUserAPIKey returns the decrypted API key for a specific provider and user.
// Returns "" if no key is configured. Used by the proxy to prioritize user keys over defaults.
func (s *SQLiteStore) GetUserAPIKey(ctx context.Context, userID string, provider string) string {
	settings, err := s.GetUserSettings(ctx, userID)
	if err != nil || settings == nil {
		return ""
	}
	switch provider {
	case "groq":
		return settings.GroqKey
	case "openai":
		return settings.OpenAIKey
	case "anthropic":
		return settings.AnthropicKey
	}
	return ""
}
