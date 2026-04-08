package memory

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"strings"
	"time"
)

const (
	// SummarizeThreshold is the number of messages after which a memory is created.
	SummarizeThreshold = 2
	// MaxFullTextLength is the max length before switching to summary mode.
	MaxFullTextLength = 2000
	// MaxSummaryLength is the maximum length of a generated summary.
	MaxSummaryLength = 500
)

// Summarizer generates memory summaries from conversations.
type Summarizer struct {
	store Store
}

// NewSummarizer creates a new summarizer.
func NewSummarizer(store Store) *Summarizer {
	return &Summarizer{store: store}
}

// CheckAndSummarize checks if a conversation has enough messages
// and generates a memory summary if it does.
func (s *Summarizer) CheckAndSummarize(ctx context.Context, conversationID, userID string) error {
	msgs, err := s.store.GetMessages(ctx, conversationID, userID)
	if err != nil {
		return fmt.Errorf("get messages: %w", err)
	}

	if len(msgs) < SummarizeThreshold {
		return nil
	}

	// Only summarize if we haven't already (check if memory exists for this conversation)
	existing, err := s.store.GetMemories(ctx, userID)
	if err != nil {
		return fmt.Errorf("get memories: %w", err)
	}

	for _, m := range existing {
		if strings.Contains(m.SourceConversationIDs, conversationID) {
			return nil // Already summarized
		}
	}

	// Save full conversation text, or summarize if too long
	fullText := buildFullConversation(msgs)
	var summary string
	if len(fullText) <= MaxFullTextLength {
		summary = fullText
	} else {
		summary = generateLocalSummary(msgs)
	}
	theme := detectTheme(msgs)

	mem := &Memory{
		ID:                    NewID(),
		UserID:                userID,
		Summary:               summary,
		SourceConversationIDs: fmt.Sprintf(`["%s"]`, conversationID),
		Theme:                 theme,
		Importance:            calculateImportance(msgs),
		PositionX:             (rand.Float64() - 0.5) * 8,
		PositionY:             (rand.Float64() - 0.5) * 6,
		PositionZ:             (rand.Float64() - 0.5) * 6,
		CreatedAt:             time.Now(),
	}

	if err := s.store.SaveMemory(ctx, mem); err != nil {
		return fmt.Errorf("save memory: %w", err)
	}

	slog.Info("memory_created",
		"memory_id", mem.ID,
		"user_id", userID,
		"conversation_id", conversationID,
		"theme", theme,
		"importance", mem.Importance,
	)

	return nil
}

// buildFullConversation returns the full conversation text, word for word.
func buildFullConversation(msgs []*Message) string {
	var sb strings.Builder
	for _, m := range msgs {
		switch m.Role {
		case "user":
			sb.WriteString("User: ")
		case "assistant":
			sb.WriteString("AI: ")
		case "system":
			continue
		}
		sb.WriteString(m.Content)
		sb.WriteString("\n\n")
	}
	return strings.TrimSpace(sb.String())
}

// generateLocalSummary creates a summary from conversation messages without calling an AI.
// In the future, this can be enhanced to use the proxy to ask an AI for a better summary.
func generateLocalSummary(msgs []*Message) string {
	var sb strings.Builder
	sb.WriteString("Conversation topics: ")

	// Extract key topics from user messages
	userMsgs := make([]string, 0)
	for _, m := range msgs {
		if m.Role == "user" {
			content := m.Content
			if len(content) > 100 {
				content = content[:100]
			}
			userMsgs = append(userMsgs, content)
		}
	}

	if len(userMsgs) == 0 {
		return "Empty conversation"
	}

	// Take first and last user messages as representative
	if len(userMsgs) == 1 {
		sb.WriteString(userMsgs[0])
	} else {
		sb.WriteString(userMsgs[0])
		sb.WriteString(" ... ")
		sb.WriteString(userMsgs[len(userMsgs)-1])
	}

	result := sb.String()
	if len(result) > MaxSummaryLength {
		result = result[:MaxSummaryLength]
	}
	return result
}

// detectTheme identifies the primary theme of a conversation based on the provider used.
func detectTheme(msgs []*Message) string {
	if len(msgs) == 0 {
		return "general"
	}

	// Use the provider of the majority of messages
	providerCount := make(map[string]int)
	for _, m := range msgs {
		if m.Provider != "" {
			providerCount[m.Provider]++
		}
	}

	bestProvider := "general"
	bestCount := 0
	for provider, count := range providerCount {
		if count > bestCount {
			bestProvider = provider
			bestCount = count
		}
	}

	return bestProvider
}

// calculateImportance assigns an importance score based on conversation characteristics.
func calculateImportance(msgs []*Message) float64 {
	if len(msgs) == 0 {
		return 0.5
	}

	// Longer conversations are more important
	lengthFactor := math.Min(float64(len(msgs))/20.0, 1.0)

	// More content = more important
	totalContent := 0
	for _, m := range msgs {
		totalContent += len(m.Content)
	}
	contentFactor := math.Min(float64(totalContent)/5000.0, 1.0)

	importance := 0.3 + 0.4*lengthFactor + 0.3*contentFactor
	return math.Round(importance*100) / 100
}
