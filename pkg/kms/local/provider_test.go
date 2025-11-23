package local

import (
	"context"
	"testing"
)

func TestLocalProvider(t *testing.T) {
	// Generate a valid 32-byte key
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 64 hex chars

	provider, err := New(key)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx := context.Background()
	originalText := []byte("Sensitive Data")

	// Test Encryption
	ciphertext, keyID, err := provider.Encrypt(ctx, originalText)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if keyID != "v1" {
		t.Errorf("Expected keyID 'v1', got '%s'", keyID)
	}

	if len(ciphertext) == 0 {
		t.Error("Ciphertext is empty")
	}

	// Test Decryption
	decryptedText, err := provider.Decrypt(ctx, ciphertext)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decryptedText) != string(originalText) {
		t.Errorf("Decrypted text mismatch. Expected '%s', got '%s'", string(originalText), string(decryptedText))
	}

	// Test Randomness (IV)
	ciphertext2, _, _ := provider.Encrypt(ctx, originalText)
	if string(ciphertext) == string(ciphertext2) {
		t.Error("Ciphertext should differ for same plaintext due to random IV")
	}
}

func TestInvalidKey(t *testing.T) {
	_, err := New("shortkey")
	if err == nil {
		t.Error("Expected error for invalid key length, got nil")
	}

	_, err = New("invalid-hex-string")
	if err == nil {
		t.Error("Expected error for invalid hex string, got nil")
	}
}
