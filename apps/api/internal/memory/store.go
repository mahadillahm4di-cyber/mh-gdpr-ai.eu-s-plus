package memory

import "context"

// Store defines the interface for memory persistence.
// All implementations must encrypt content at rest.
type Store interface {
	// ── Conversations ──
	SaveConversation(ctx context.Context, conv *Conversation) error
	GetConversation(ctx context.Context, id, userID string) (*Conversation, error)
	ListConversations(ctx context.Context, userID string, limit, offset int) ([]*Conversation, error)
	UpdateConversationTitle(ctx context.Context, id, userID, title string) error
	DeleteConversation(ctx context.Context, id, userID string) error

	// ── Messages ──
	SaveMessage(ctx context.Context, msg *Message) error
	GetMessages(ctx context.Context, conversationID, userID string) ([]*Message, error)
	GetRecentMessages(ctx context.Context, userID string, limit int) ([]*Message, error)

	// ── Memories ──
	SaveMemory(ctx context.Context, mem *Memory) error
	GetMemory(ctx context.Context, id, userID string) (*Memory, error)
	GetMemories(ctx context.Context, userID string) ([]*Memory, error)
	SearchMemories(ctx context.Context, userID, query string) ([]*Memory, error)
	DeleteMemory(ctx context.Context, id, userID string) error

	// ── Lifecycle ──
	Close() error
}
