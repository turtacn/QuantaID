package local

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// LocalProvider implements KeyProvider using a local AES key.
// It uses AES-256-GCM for encryption.
type LocalProvider struct {
	key   []byte
	keyID string
}

// New creates a new LocalProvider with the given hex-encoded key.
// The key must be 32 bytes (64 hex characters) for AES-256.
func New(keyHex string) (*LocalProvider, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key hex: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: expected 32 bytes (AES-256), got %d", len(key))
	}

	// For simple local provider, we use a static key ID "v1" or similar.
	// In a real scenario, this might come from config or the key itself.
	return &LocalProvider{
		key:   key,
		keyID: "v1",
	}, nil
}

// Encrypt encrypts the plaintext using AES-GCM.
// It returns the ciphertext (nonce + encrypted data), the key ID, and error.
func (p *LocalProvider) Encrypt(ctx context.Context, plaintext []byte) ([]byte, string, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, p.keyID, nil
}

// Decrypt decrypts the ciphertext using AES-GCM.
func (p *LocalProvider) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}
