// Package memory provides encrypted, local-first storage for AI conversations and memories.
// SECURITY: All content is encrypted at rest with AES-256-GCM. No plaintext stored.
package memory

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

// Conversation represents a chat session with a specific provider.
type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a single message in a conversation.
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"` // Decrypted in memory, encrypted in DB
	Provider       string    `json:"provider"`
	TokenCount     int       `json:"token_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// Memory represents a summarized piece of knowledge extracted from conversations.
// These are used for context injection when switching providers.
type Memory struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	Summary               string    `json:"summary"` // Decrypted in memory, encrypted in DB
	SourceConversationIDs string    `json:"source_conversation_ids"` // JSON array
	Theme                 string    `json:"theme"`
	Importance            float64   `json:"importance"` // 0.0 to 1.0
	PositionX             float64   `json:"position_x"` // 3D position for dashboard
	PositionY             float64   `json:"position_y"`
	PositionZ             float64   `json:"position_z"`
	CreatedAt             time.Time `json:"created_at"`
}

// NewID generates a new UUID v4.
func NewID() string {
	return uuid.New().String()
}

// ── Encryption ──

// Encrypt encrypts plaintext using AES-256-GCM.
// SECURITY: Uses authenticated encryption with random nonce.
func Encrypt(plaintext, keyHex string) (string, error) {
	if keyHex == "" {
		// In development mode without encryption key, return plaintext
		return plaintext, nil
	}

	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid encryption key: %w", err)
	}

	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes (64 hex chars)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cipher error: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm error: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce error: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts AES-256-GCM encrypted text.
func Decrypt(ciphertextHex, keyHex string) (string, error) {
	if keyHex == "" {
		return ciphertextHex, nil
	}

	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid encryption key: %w", err)
	}

	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cipher error: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm error: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt error: %w", err)
	}

	return string(plaintext), nil
}
