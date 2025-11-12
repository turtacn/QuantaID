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
	CreateUserMFAConfig(ctx context.Context, config *types.UserMFAConfig) error
	GetUserMFAConfig(ctx context.Context, userID uuid.UUID, method string) (*types.UserMFAConfig, error)
	GetUserMFAConfigs(ctx context.Context, userID uuid.UUID) ([]*types.UserMFAConfig, error)
	UpdateUserMFAConfig(ctx context.Context, config *types.UserMFAConfig) error
	DeleteUserMFAConfig(ctx context.Context, userID uuid.UUID, method string) error
	CreateMFAVerificationLog(ctx context.Context, log *types.MFAVerificationLog) error
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
	configs, err := mp.mfaRepo.GetUserMFAConfigs(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, config := range configs {
		if config.Enabled {
			return true, nil
		}
	}
	return false, nil
}

// GetAvailableMFAMethods returns a list of available MFA methods for a user.
func (mp *MFAPolicy) GetAvailableMFAMethods(ctx context.Context, userID uuid.UUID) ([]string, error) {
	configs, err := mp.mfaRepo.GetUserMFAConfigs(ctx, userID)
	if err != nil {
		return nil, err
	}
	var methods []string
	for _, config := range configs {
		if config.Enabled {
			methods = append(methods, config.Method)
		}
	}
	return methods, nil
}

func (mp *MFAPolicy) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	config, err := mp.mfaRepo.GetUserMFAConfig(ctx, userID, "totp")
	if err != nil {
		return false, types.ErrNotFound.WithCause(err)
	}

	var totpConfig struct {
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(config.Config, &totpConfig); err != nil {
		return false, types.ErrInternal.WithCause(err)
	}

	return mp.TotpProvider.VerifyCode(totpConfig.Secret, code), nil
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
		mfaConfig, err := mp.mfaRepo.GetUserMFAConfig(ctx, challengeData.UserID, "totp")
		if err != nil {
			return types.ErrUnauthorized
		}

		var totpConfig struct {
			Secret string `json:"secret"`
		}
		if err := json.Unmarshal(mfaConfig.Config, &totpConfig); err != nil {
			return types.ErrInternal
		}

		if !mp.TotpProvider.VerifyCode(totpConfig.Secret, code) {
			return types.ErrUnauthorized
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
