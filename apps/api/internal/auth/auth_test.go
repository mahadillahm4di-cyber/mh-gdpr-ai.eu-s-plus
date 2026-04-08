package auth

import (
	"testing"
	"time"
)

func TestHashPassword_MinLength(t *testing.T) {
	_, err := HashPassword("short")
	if err == nil {
		t.Error("expected error for password shorter than 8 chars")
	}
}

func TestHashPassword_Valid(t *testing.T) {
	hash, err := HashPassword("validpassword123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if hash == "validpassword123" {
		t.Error("hash should not equal plaintext")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "securepassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Error("expected CheckPassword to return true for correct password")
	}

	if CheckPassword("wrongpassword", hash) {
		t.Error("expected CheckPassword to return false for wrong password")
	}
}

func TestGenerateAccessToken(t *testing.T) {
	secret := "test-secret-key-32-chars-minimum!"
	userID := "user-123"

	token, err := GenerateAccessToken(userID, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Validate the token
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, claims.UserID)
	}
	if claims.Issuer != "mh-gdpr-ai" {
		t.Errorf("expected issuer mh-gdpr-ai, got %s", claims.Issuer)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	secret := "test-secret-key-32-chars-minimum!"
	userID := "user-456"

	token, err := GenerateRefreshToken(userID, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, claims.UserID)
	}

	// Refresh token should expire in ~7 days
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 6*24*time.Hour || remaining > 8*24*time.Hour {
		t.Errorf("expected ~7 day expiry, got %v", remaining)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, err := GenerateAccessToken("user-123", "correct-secret-key-32-chars!!!!!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = ValidateToken(token, "wrong-secret-key-32-chars!!!!!!")
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	_, err := ValidateToken("not-a-valid-jwt", "some-secret")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestAccessTokenExpiry(t *testing.T) {
	secret := "test-secret-key-32-chars-minimum!"
	token, err := GenerateAccessToken("user-123", secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	// Access token should expire in ~15 minutes
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 14*time.Minute || remaining > 16*time.Minute {
		t.Errorf("expected ~15 minute expiry, got %v", remaining)
	}
}
