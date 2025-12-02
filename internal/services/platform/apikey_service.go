package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/internal/domain/apikey"
	"github.com/turtacn/QuantaID/pkg/types"
	"golang.org/x/crypto/bcrypt"
)

// APIKeyService manages API key generation, validation, and lifecycle.
type APIKeyService struct {
	repo apikey.Repository
}

// NewAPIKeyService creates a new APIKeyService.
func NewAPIKeyService(repo apikey.Repository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// GenerateKey creates a new API key for the given application.
// It returns the plaintext key (to be shown once) and the created APIKey model.
// Format: prefix_{keyID_16_hex}{secret_32_hex}
func (s *APIKeyService) GenerateKey(ctx context.Context, appID string, prefix string, scopes []string, ttl time.Duration) (string, *apikey.APIKey, error) {
	if prefix == "" {
		prefix = "qid_live_"
	}

	// Generate random component
	// 1. KeyID: 8 bytes (16 hex chars)
	keyIDBytes := make([]byte, 8)
	if _, err := rand.Read(keyIDBytes); err != nil {
		return "", nil, types.ErrInternal.WithCause(err)
	}
	keyID := hex.EncodeToString(keyIDBytes)

	// 2. Secret: 16 bytes (32 hex chars)
	secretBytes := make([]byte, 16)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", nil, types.ErrInternal.WithCause(err)
	}
	secret := hex.EncodeToString(secretBytes)

	// Combine: prefix + keyID + secret
	plaintextKey := fmt.Sprintf("%s%s%s", prefix, keyID, secret)

	// Hash the plaintext using bcrypt
	// (Alternatively, we could hash just the secret, but hashing full key is simpler for validation consistency if KeyID is also in plaintext)
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(plaintextKey), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, types.ErrInternal.WithCause(err)
	}

	now := time.Now()
	var expiresAt *time.Time
	if ttl > 0 {
		t := now.Add(ttl)
		expiresAt = &t
	}

	id := generateID() // Helper to generate internal DB ID (UUID)

	key := &apikey.APIKey{
		ID:        id,
		AppID:     appID,
		KeyID:     keyID, // Store KeyID for O(1) lookup
		KeyHash:   string(hashBytes),
		Prefix:    prefix,
		Scopes:    scopes,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
		Revoked:   false,
	}

	if err := s.repo.Create(ctx, key); err != nil {
		return "", nil, types.ErrInternal.WithCause(err)
	}

	return plaintextKey, key, nil
}

// ValidateKey validates the provided plaintext key and returns the associated APIKey if valid.
func (s *APIKeyService) ValidateKey(ctx context.Context, plaintextKey string) (*apikey.APIKey, error) {
	// Key Structure: [Prefix][KeyID(16)][Secret(32)]
	// Total random part length = 16 + 32 = 48 chars.
	const randomPartLen = 48
	const keyIDLen = 16

	if len(plaintextKey) <= randomPartLen {
		return nil, types.ErrInvalidToken
	}

	// Extract KeyID
	// Position: It starts after Prefix. Prefix length is (Total - 48).
	prefixLen := len(plaintextKey) - randomPartLen
	if prefixLen < 0 {
		return nil, types.ErrInvalidToken
	}

	// KeyID is the first 16 chars of the random part
	keyID := plaintextKey[prefixLen : prefixLen+keyIDLen]

	// 2. Lookup by KeyID (O(1))
	key, err := s.repo.GetByKeyID(ctx, keyID)
	if err != nil {
		// If not found, return invalid token (avoid leaking existence via error types if possible, but repo might return NotFound)
		// Usually we want constant time response, but bcrypt dominates.
		// If key not found, we should probably do a fake comparison?
		// For now, fail fast.
		return nil, types.ErrInvalidToken
	}

	// 3. Verify checks
	if key.Revoked {
		return nil, types.ErrInvalidToken
	}
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, types.ErrInvalidToken
	}

	// 4. Verify hash
	if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(plaintextKey)); err != nil {
		return nil, types.ErrInvalidToken
	}

	return key, nil
}

// GetRateLimitPolicy returns the rate limit policy for an application.
func (s *APIKeyService) GetRateLimitPolicy(ctx context.Context, appID string) (*apikey.RateLimitPolicy, error) {
	return s.repo.GetRateLimitPolicy(ctx, appID)
}

// RevokeKey revokes an API key.
func (s *APIKeyService) RevokeKey(ctx context.Context, id string) error {
	key, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	key.Revoked = true
	key.UpdatedAt = time.Now()

	return s.repo.Update(ctx, key)
}

func generateID() string {
	// Simple random ID generation
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
