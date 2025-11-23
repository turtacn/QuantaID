package types

import (
	"context"
	"encoding/base64"
	"testing"
)

// MockProvider is a mock KMS provider for testing.
type MockProvider struct{}

func (m *MockProvider) Encrypt(ctx context.Context, plaintext []byte) ([]byte, string, error) {
	// Simple mock encryption: reverse the bytes
	ciphertext := make([]byte, len(plaintext))
	for i, b := range plaintext {
		ciphertext[len(plaintext)-1-i] = b
	}
	return ciphertext, "mock-key", nil
}

func (m *MockProvider) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	// Simple mock decryption: reverse the bytes again
	plaintext := make([]byte, len(ciphertext))
	for i, b := range ciphertext {
		plaintext[len(ciphertext)-1-i] = b
	}
	return plaintext, nil
}

func TestEncryptedString(t *testing.T) {
	// Initialize global KMS with mock
	SetGlobalKMS(&MockProvider{})

	original := EncryptedString("Hello World")

	// Test Value() (Encryption)
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() failed: %v", err)
	}

	strVal, ok := val.(string)
	if !ok {
		t.Fatalf("Value() returned non-string: %T", val)
	}

	// Verify encryption (reversed "Hello World" -> "dlroW olleH" -> Base64)
	expectedBytes := []byte("dlroW olleH")
	expectedBase64 := base64.StdEncoding.EncodeToString(expectedBytes)

	if strVal != expectedBase64 {
		t.Errorf("Encryption result mismatch. Expected %s, got %s", expectedBase64, strVal)
	}

	// Test Scan() (Decryption)
	var scanned EncryptedString
	err = scanned.Scan(strVal)
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	if scanned != original {
		t.Errorf("Decryption result mismatch. Expected '%s', got '%s'", original, scanned)
	}

	// Test Scan() with []byte
	var scannedBytes EncryptedString
	err = scannedBytes.Scan([]byte(strVal))
	if err != nil {
		t.Fatalf("Scan() with []byte failed: %v", err)
	}
	if scannedBytes != original {
		t.Errorf("Decryption result mismatch with []byte. Expected '%s', got '%s'", original, scannedBytes)
	}
}

func TestEncryptedString_Uninitialized(t *testing.T) {
	// Temporarily unset global KMS
	oldKMS := globalKMS
	SetGlobalKMS(nil)
	defer SetGlobalKMS(oldKMS)

	s := EncryptedString("test")
	_, err := s.Value()
	if err == nil || err.Error() != "global KMS not initialized" {
		t.Errorf("Expected 'global KMS not initialized' error, got %v", err)
	}

	err = s.Scan("test")
	if err == nil || err.Error() != "global KMS not initialized" {
		t.Errorf("Expected 'global KMS not initialized' error, got %v", err)
	}
}
