package auth

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/pkg/plugins/mfa/totp"
	"github.com/turtacn/QuantaID/pkg/types"
)

// MFARepository defines the interface for a persistence layer for MFA configurations.
type MFARepository interface {
	CreateFactor(ctx context.Context, factor *types.MFAFactor) error
	GetFactor(ctx context.Context, factorID uuid.UUID) (*types.MFAFactor, error)
	GetUserFactors(ctx context.Context, userID uuid.UUID) ([]*types.MFAFactor, error)
	UpdateFactor(ctx context.Context, factor *types.MFAFactor) error
	DeleteFactor(ctx context.Context, factorID uuid.UUID) error
	CreateVerificationLog(ctx context.Context, log *types.MFAVerificationLog) error
}

// MFAPolicy handles the business logic for MFA.
type MFAPolicy struct {
	mfaRepo      MFARepository
	redis        *redis.Client
	TotpProvider *totp.TOTPProvider
}

// NewMFAPolicy creates a new MFA policy engine.
func NewMFAPolicy(mfaRepo MFARepository, redis *redis.Client, totpProvider *totp.TOTPProvider) *MFAPolicy {
	return &MFAPolicy{
		mfaRepo:      mfaRepo,
		redis:        redis,
		TotpProvider: totpProvider,
	}
}

// ShouldEnforceMFA determines if a user should be forced to use MFA.
func (mp *MFAPolicy) ShouldEnforceMFA(ctx context.Context, user *types.User) (bool, error) {
	// For now, we'll enforce MFA if the user has any MFA method enabled.
	// In the future, this could be based on user roles or other policies.
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return false, err
	}
	factors, err := mp.mfaRepo.GetUserFactors(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, factor := range factors {
		if factor.Status == "active" {
			return true, nil
		}
	}
	return false, nil
}

// GetAvailableMFAMethods returns a list of available MFA methods for a user.
func (mp *MFAPolicy) GetAvailableMFAMethods(ctx context.Context, userID uuid.UUID) ([]string, error) {
	factors, err := mp.mfaRepo.GetUserFactors(ctx, userID)
	if err != nil {
		return nil, err
	}
	var methods []string
	for _, factor := range factors {
		if factor.Status == "active" {
			methods = append(methods, factor.Type)
		}
	}
	return methods, nil
}

func (mp *MFAPolicy) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	factors, err := mp.mfaRepo.GetUserFactors(ctx, userID)
	if err != nil {
		return false, types.ErrNotFound.WithCause(err)
	}

	for _, factor := range factors {
		if factor.Type == "totp" {
			return mp.TotpProvider.VerifyCode(factor.Secret, code), nil
		}
	}

	return false, types.ErrNotFound
}

// VerifyMFAChallenge verifies an MFA challenge.
func (mp *MFAPolicy) VerifyMFAChallenge(ctx context.Context, challengeID, method, code string) error {
	// 1. From Redis, get challenge information
	challenge, err := mp.redis.Get(ctx, "mfa:challenge:"+challengeID).Result()
	if err != nil {
		return types.ErrInvalidRequest
	}

	var challengeData struct {
		UserID uuid.UUID
		Method string
	}
	json.Unmarshal([]byte(challenge), &challengeData)

	// 2. Verify the code based on the method
	switch method {
	case "totp":
		factors, err := mp.mfaRepo.GetUserFactors(ctx, challengeData.UserID)
		if err != nil {
			return types.ErrUnauthorized
		}

		for _, factor := range factors {
			if factor.Type == "totp" {
				if !mp.TotpProvider.VerifyCode(factor.Secret, code) {
					return types.ErrUnauthorized
				}
			}
		}

	case "sms":
		storedCode, err := mp.redis.Get(ctx, "sms:otp:"+challengeData.UserID.String()).Result()
		if err != nil || storedCode != code {
			return types.ErrUnauthorized
		}
	}

	// 3. Delete the challenge after successful verification
	mp.redis.Del(ctx, "mfa:challenge:"+challengeID)

	return nil
}
