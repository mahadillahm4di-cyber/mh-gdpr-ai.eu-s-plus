package memory

import (
	"encoding/hex"
	"testing"
)

func TestNewID(t *testing.T) {
	id1 := NewID()
	id2 := NewID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	if len(id1) != 36 {
		t.Errorf("expected 36-char UUID, got %d chars", len(id1))
	}
}

// generateTestKey returns a valid 32-byte hex-encoded key for testing.
func generateTestKey() string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return hex.EncodeToString(key)
}

func TestEncrypt_Decrypt_Roundtrip(t *testing.T) {
	key := generateTestKey()
	plaintext := "Hello, this is a secret message!"

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	if encrypted == plaintext {
		t.Error("encrypted text should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncrypt_EmptyKey_PassthroughDev(t *testing.T) {
	// In dev mode (empty key), encrypt/decrypt should pass through
	plaintext := "test message"

	encrypted, err := Encrypt(plaintext, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if encrypted != plaintext {
		t.Errorf("expected passthrough, got %q", encrypted)
	}

	decrypted, err := Decrypt(encrypted, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("expected passthrough, got %q", decrypted)
	}
}

func TestEncrypt_InvalidKey(t *testing.T) {
	_, err := Encrypt("test", "not-hex")
	if err == nil {
		t.Error("expected error for non-hex key")
	}
}

func TestEncrypt_WrongKeyLength(t *testing.T) {
	// 16 bytes (32 hex chars) instead of 32 bytes (64 hex chars)
	shortKey := hex.EncodeToString(make([]byte, 16))
	_, err := Encrypt("test", shortKey)
	if err == nil {
		t.Error("expected error for wrong key length")
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	key := generateTestKey()
	_, err := Decrypt("not-hex-encoded", key)
	if err == nil {
		t.Error("expected error for invalid ciphertext")
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	key := generateTestKey()
	_, err := Decrypt("aabb", key)
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := generateTestKey()
	key2 := hex.EncodeToString(make([]byte, 32))
	// Make key2 different
	key2Bytes, _ := hex.DecodeString(key2)
	key2Bytes[0] = 0xFF
	key2 = hex.EncodeToString(key2Bytes)

	encrypted, err := Encrypt("secret", key1)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	_, err = Decrypt(encrypted, key2)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
}

func TestEncrypt_UniqueNonce(t *testing.T) {
	key := generateTestKey()
	plaintext := "same message"

	enc1, _ := Encrypt(plaintext, key)
	enc2, _ := Encrypt(plaintext, key)

	if enc1 == enc2 {
		t.Error("encrypting the same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestEncrypt_EmptyMessage(t *testing.T) {
	key := generateTestKey()

	encrypted, err := Encrypt("", key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if decrypted != "" {
		t.Errorf("expected empty string, got %q", decrypted)
	}
}

func TestEncrypt_LargeMessage(t *testing.T) {
	key := generateTestKey()
	// 10KB message
	plaintext := make([]byte, 10240)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	encrypted, err := Encrypt(string(plaintext), key)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if decrypted != string(plaintext) {
		t.Error("roundtrip failed for large message")
	}
}
