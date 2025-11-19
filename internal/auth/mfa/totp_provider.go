package mfa

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// MFAFactorRepository defines the interface for storing and retrieving MFA factors.
type MFAFactorRepository interface {
	CreateMFAFactor(ctx context.Context, factor *types.MFAFactor) error
	GetMFAFactorsByUserID(ctx context.Context, userID string) ([]*types.MFAFactor, error)
	UpdateMFAFactor(ctx context.Context, factor *types.MFAFactor) error
}

// TOTPProvider implements the MFAProvider interface for Time-based One-Time Passwords.
type TOTPProvider struct {
	repo   MFAFactorRepository
	crypto utils.CryptoManagerInterface
}

// NewTOTPProvider creates a new TOTPProvider.
func NewTOTPProvider(repo MFAFactorRepository, crypto utils.CryptoManagerInterface) *TOTPProvider {
	return &TOTPProvider{
		repo:   repo,
		crypto: crypto,
	}
}

// Enroll starts the enrollment process for a new TOTP factor.
func (p *TOTPProvider) Enroll(ctx context.Context, user *types.User) (*types.MFAEnrollment, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "QuantaID",
		AccountName: user.Username,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	encryptedSecret, err := p.crypto.Encrypt(key.Secret())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	factor := &types.MFAFactor{
		UserID: userID,
		Type:   "totp",
		Status: "pending",
		Secret: encryptedSecret,
	}

	if err := p.repo.CreateMFAFactor(ctx, factor); err != nil {
		return nil, fmt.Errorf("failed to save MFA factor: %w", err)
	}

	return &types.MFAEnrollment{
		Secret: key.Secret(),
		URL:    key.URL(),
	}, nil
}

// Challenge generates a new TOTP challenge.
func (p *TOTPProvider) Challenge(ctx context.Context, user *types.User) (*types.MFAChallenge, error) {
	// For TOTP, the challenge is implicit. The user provides the code from their app.
	return &types.MFAChallenge{
		ChallengeID: "totp-challenge",
		MFAProvider: "totp",
	}, nil
}

// Verify validates a TOTP code.
func (p *TOTPProvider) Verify(ctx context.Context, user *types.User, code string) (bool, error) {
	factors, err := p.repo.GetMFAFactorsByUserID(ctx, user.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get MFA factors: %w", err)
	}

	for _, factor := range factors {
		if factor.Type == "totp" {
			secret, err := p.crypto.Decrypt(factor.Secret)
			if err != nil {
				// Log the error, but don't reveal that the secret was invalid
				fmt.Printf("failed to decrypt secret for user %s: %v\n", user.ID, err)
				continue
			}

			if totp.Validate(code, secret) {
				return true, nil
			}
		}
	}

	return false, nil
}

// ListMethods returns the available TOTP methods for the user.
func (p *TOTPProvider) ListMethods(ctx context.Context, user *types.User) ([]*types.MFAMethod, error) {
	factors, err := p.repo.GetMFAFactorsByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MFA factors: %w", err)
	}

	var methods []*types.MFAMethod
	for _, factor := range factors {
		if factor.Type == "totp" && factor.Status == "enrolled" {
			methods = append(methods, &types.MFAMethod{
				ID:   factor.ID.String(),
				Type: "totp",
			})
		}
	}

	return methods, nil
}

// GetStrength returns the strength of the TOTP provider.
func (p *TOTPProvider) GetStrength() StrengthLevel {
	return StrengthLevelNormal
}
