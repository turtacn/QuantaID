package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
)

const (
	// PKCEMethodS256 is the S256 PKCE challenge method.
	PKCEMethodS256 = "S256"
	// PKCEMethodPlain is the plain PKCE challenge method.
	PKCEMethodPlain = "plain"
)

var (
	// ErrPKCERequired is returned when a PKCE challenge is required but not provided.
	ErrPKCERequired = errors.New("pkce challenge required for public clients")
	// ErrUnsupportedChallengeMethod is returned when an unsupported challenge method is used.
	ErrUnsupportedChallengeMethod = errors.New("unsupported pkce challenge method")
	// ErrInvalidVerifierLength is returned when the code_verifier has an invalid length.
	ErrInvalidVerifierLength = errors.New("invalid code_verifier length")
	// ErrInvalidCodeVerifier is returned when the code_verifier does not match the challenge.
	ErrInvalidCodeVerifier = errors.New("invalid code_verifier")
	// ErrPKCEChallengeNotFound is returned when no PKCE challenge is found for a given auth code.
	ErrPKCEChallengeNotFound = errors.New("pkce challenge not found for auth code")
)

// PKCEConfig holds the configuration for PKCE validation.
type PKCEConfig struct {
	// EnforceForPublicClients makes PKCE mandatory for all public clients.
	EnforceForPublicClients bool
	// AllowedMethods is a list of a supported challenge methods (e.g., "S256", "plain").
	AllowedMethods []string
	// MinVerifierLength is the minimum allowed length for a code verifier.
	MinVerifierLength int
	// MaxVerifierLength is the maximum allowed length for a code verifier.
	MaxVerifierLength int
}

// PKCEChallenge represents the stored PKCE challenge data.
type PKCEChallenge struct {
	Challenge       string
	ChallengeMethod string
}

// PKCERepository defines the interface for storing and retrieving PKCE challenges.
type PKCERepository interface {
	GetAuthCodeChallenge(ctx context.Context, authCode string) (*PKCEChallenge, error)
}

// PKCEValidator provides methods for validating PKCE challenges and verifiers.
type PKCEValidator struct {
	config PKCEConfig
	repo   PKCERepository
}

// NewPKCEValidator creates a new PKCEValidator.
func NewPKCEValidator(config PKCEConfig, repo PKCERepository) *PKCEValidator {
	return &PKCEValidator{config: config, repo: repo}
}

// VerifyCodeVerifier validates the provided code verifier against the stored challenge for a given authorization code.
func (v *PKCEValidator) VerifyCodeVerifier(ctx context.Context, authCode, verifier string) error {
	storedChallenge, err := v.repo.GetAuthCodeChallenge(ctx, authCode)
	if err != nil {
		return err // Propagate repository errors
	}

	if storedChallenge == nil {
		// If no challenge was stored, we cannot verify. This implies the authorization
		// request did not include a PKCE challenge. The decision to allow or deny this
		// should be made during the authorization step based on client type.
		return ErrPKCEChallengeNotFound
	}

	// Validate verifier length according to RFC 7636 (43-128 characters).
	if len(verifier) < v.config.MinVerifierLength || len(verifier) > v.config.MaxVerifierLength {
		return ErrInvalidVerifierLength
	}

	var computedChallenge string
	switch storedChallenge.ChallengeMethod {
	case PKCEMethodS256:
		hash := sha256.Sum256([]byte(verifier))
		computedChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
	case PKCEMethodPlain:
		computedChallenge = verifier
	default:
		return ErrUnsupportedChallengeMethod
	}

	// Use constant-time comparison to prevent timing attacks.
	if subtle.ConstantTimeCompare([]byte(computedChallenge), []byte(storedChallenge.Challenge)) != 1 {
		return ErrInvalidCodeVerifier
	}

	return nil
}
