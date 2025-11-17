package mfa

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/types"
	"time"
)

type MFAManager struct {
	providers   map[MFAType]MFAProvider
	repo        *postgresql.PostgresMFARepository
	// rateLimiter RateLimiter // TODO: Implement rate limiting
	config      MFAConfig
}

type MFAType string

const (
	MFATypeTOTP     MFAType = "totp"
	MFATypeSMS      MFAType = "sms"
	MFATypeEmail    MFAType = "email"
	MFATypeWebAuthn MFAType = "webauthn"
)

type EnrollParams struct {
	Email string
}

type EnrollResult struct {
	CredentialID  string
	Secret        string
	QRCodeImage   string
	BackupCodes   []string
	SetupURL      string
	Challenge     string
	// RegistrationOptions for WebAuthn
}

type MFAProvider interface {
	Enroll(ctx context.Context, userID string, params EnrollParams) (*EnrollResult, error)
	Verify(ctx context.Context, userID string, credential string) (bool, error)
	Revoke(ctx context.Context, userID string, credentialID string) error
}

type MFAConfig struct {
	EnabledProviders []MFAType `yaml:"enabled_providers"`
	RequireSetup     bool      `yaml:"require_setup"`
	GracePeriod      int       `yaml:"grace_period"`
	MaxAttempts      int       `yaml:"max_attempts"`
}

func (mm *MFAManager) EnrollFactor(ctx context.Context, userID string, mfaType MFAType, params EnrollParams) (*EnrollResult, error) {
	// TODO: check if user already has a factor of this type
	provider, ok := mm.providers[mfaType]
	if !ok {
		return nil, fmt.Errorf("MFA provider not found: %s", mfaType)
	}

	result, err := provider.Enroll(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	factor := &types.MFAFactor{
		UserID:       types.MustParseUUID(userID),
		Type:         string(mfaType),
		Status:       "pending",
		CredentialID: result.CredentialID,
		Secret:       result.Secret,
		CreatedAt:    time.Now(),
	}
	err = mm.repo.CreateFactor(ctx, factor)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (mm *MFAManager) VerifyFactor(ctx context.Context, userID string, mfaType MFAType, credential string) (bool, error) {
	// TODO: Rate limiting
	factors, err := mm.repo.GetUserFactorsByType(ctx, types.MustParseUUID(userID), string(mfaType))
	if err != nil {
		return false, err
	}
	if len(factors) == 0 {
		return false, fmt.Errorf("no factor enrolled")
	}

	for _, factor := range factors {
		provider := mm.providers[MFAType(factor.Type)]
		valid, err := provider.Verify(ctx, userID, credential)
		if err != nil {
			// mm.recordVerificationAttempt(ctx, factor.ID, false, err)
			continue
		}
		if valid {
			// mm.recordVerificationAttempt(ctx, factor.ID, true, nil)
			// mm.updateLastUsed(ctx, factor.ID)
			return true, nil
		}
	}

	return false, fmt.Errorf("invalid credential")
}

func (mm *MFAManager) ActivateFactor(ctx context.Context, userID string, mfaType MFAType, credential string) error {
	// TODO:
	return nil
}

func (mm *MFAManager) GetRequiredFactors(ctx context.Context, userID string) ([]MFAType, error) {
	// TODO:
	return nil, nil
}
