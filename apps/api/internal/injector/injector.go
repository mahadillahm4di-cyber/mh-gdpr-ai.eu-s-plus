// Package injector handles context injection when users switch between AI providers.
// This is the CORE of the protocol: when you switch from GPT to Claude,
// the injector retrieves your memory and injects it so Claude knows everything GPT knew.
package injector

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/memory"
	"github.com/mahadillahm4di-cyber/mh-gdpr-ai.eu-s-plus/internal/proxy"
)

// ContextInjector tracks provider switches and injects context.
type ContextInjector struct {
	store        memory.Store
	lastProvider map[string]proxy.Provider // user_id → last provider
	mu           sync.RWMutex
}

// NewContextInjector creates a new injector.
func NewContextInjector(store memory.Store) *ContextInjector {
	return &ContextInjector{
		store:        store,
		lastProvider: make(map[string]proxy.Provider),
	}
}

// DetectSwitch checks if the user switched providers and returns true if so.
func (ci *ContextInjector) DetectSwitch(userID string, currentProvider proxy.Provider) bool {
	ci.mu.RLock()
	last, exists := ci.lastProvider[userID]
	ci.mu.RUnlock()

	if !exists {
		ci.mu.Lock()
		ci.lastProvider[userID] = currentProvider
		ci.mu.Unlock()
		return false
	}

	if last != currentProvider {
		ci.mu.Lock()
		ci.lastProvider[userID] = currentProvider
		ci.mu.Unlock()
		return true
	}

	return false
}

// InjectContext adds memory context to the messages when a provider switch is detected.
// It prepends a system message with the user's memory summaries.
func (ci *ContextInjector) InjectContext(ctx context.Context, userID string, messages []proxy.ChatMessage) ([]proxy.ChatMessage, error) {
	// Get user memories
	memories, err := ci.store.GetMemories(ctx, userID)
	if err != nil {
		return messages, fmt.Errorf("inject context: get memories: %w", err)
	}

	if len(memories) == 0 {
		// No memories yet, also try recent messages
		recentMsgs, err := ci.store.GetRecentMessages(ctx, userID, 20)
		if err != nil {
			slog.Warn("inject_context: failed to get recent messages", "error", err)
			return messages, nil
		}
		if len(recentMsgs) == 0 {
			return messages, nil
		}

		// Build context from recent messages
		contextStr := buildContextFromMessages(recentMsgs)
		return prependSystemMessage(messages, contextStr), nil
	}

	// Build context from memories
	contextStr := buildContextFromMemories(memories)
	return prependSystemMessage(messages, contextStr), nil
}

// buildContextFromMemories creates a system message from memory summaries.
func buildContextFromMemories(memories []*memory.Memory) string {
	var sb strings.Builder
	sb.WriteString("You are continuing a conversation with this user. Here is what you know about them from previous conversations:\n\n")

	for i, m := range memories {
		if i >= 10 { // Limit to 10 most recent memories to control token usage
			break
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", m.Theme, m.Summary))
	}

	sb.WriteString("\nUse this context naturally. Don't mention that you received this context.")
	return sb.String()
}

// buildContextFromMessages creates a system message from recent messages.
func buildContextFromMessages(messages []*memory.Message) string {
	var sb strings.Builder
	sb.WriteString("You are continuing a conversation with this user. Here is the recent conversation history:\n\n")

	for _, m := range messages {
		sb.WriteString(fmt.Sprintf("%s: %s\n", m.Role, truncate(m.Content, 200)))
	}

	sb.WriteString("\nContinue naturally from this context. Don't mention that you received this context.")
	return sb.String()
}

// prependSystemMessage adds a system message at the beginning of the messages list.
func prependSystemMessage(messages []proxy.ChatMessage, content string) []proxy.ChatMessage {
	systemMsg := proxy.ChatMessage{
		Role:    "system",
		Content: content,
	}

	// If there's already a system message, merge them
	if len(messages) > 0 && messages[0].Role == "system" {
		messages[0].Content = content + "\n\n" + messages[0].Content
		return messages
	}

	return append([]proxy.ChatMessage{systemMsg}, messages...)
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
