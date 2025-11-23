package kms

import "context"

// KeyProvider defines the interface for a Key Management Service provider.
// It handles encryption and decryption of data using underlying keys.
type KeyProvider interface {
	// Encrypt encrypts the given plaintext and returns the ciphertext,
	// the key ID used for encryption, and any error encountered.
	Encrypt(ctx context.Context, plaintext []byte) (ciphertext []byte, keyID string, err error)

	// Decrypt decrypts the given ciphertext and returns the plaintext.
	Decrypt(ctx context.Context, ciphertext []byte) (plaintext []byte, err error)
}
