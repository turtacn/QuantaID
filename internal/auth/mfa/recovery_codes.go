package mfa

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// RecoveryCodeProvider handles the generation and verification of MFA recovery codes.
type RecoveryCodeProvider struct {
	repo   MFAFactorRepository
	crypto utils.CryptoManagerInterface
}

// NewRecoveryCodeProvider creates a new RecoveryCodeProvider.
func NewRecoveryCodeProvider(repo MFAFactorRepository, crypto utils.CryptoManagerInterface) *RecoveryCodeProvider {
	return &RecoveryCodeProvider{
		repo:   repo,
		crypto: crypto,
	}
}

// GenerateAndStore creates a new set of recovery codes for a user, hashes them, and stores them.
// It returns the plaintext codes for the user to save.
func (p *RecoveryCodeProvider) GenerateAndStore(ctx context.Context, factor *types.MFAFactor) ([]string, error) {
	plaintextCodes, err := p.crypto.GenerateRecoveryCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate recovery codes: %w", err)
	}

	hashedCodes := make([]string, len(plaintextCodes))
	for i, code := range plaintextCodes {
		hashedCode, err := p.crypto.HashRecoveryCode(code)
		if err != nil {
			return nil, fmt.Errorf("failed to hash recovery code: %w", err)
		}
		hashedCodes[i] = hashedCode
	}

	codesJSON, err := json.Marshal(hashedCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recovery codes: %w", err)
	}

	factor.BackupCodes = codesJSON
	if err := p.repo.UpdateMFAFactor(ctx, factor); err != nil {
		return nil, fmt.Errorf("failed to update MFA factor with recovery codes: %w", err)
	}

	return plaintextCodes, nil
}

// Verify checks if a given recovery code is valid for a user.
// If the code is valid, it is removed from the list of available codes.
func (p *RecoveryCodeProvider) Verify(ctx context.Context, user *types.User, code string) (bool, error) {
	factors, err := p.repo.GetMFAFactorsByUserID(ctx, user.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get MFA factors: %w", err)
	}

	for _, factor := range factors {
		if len(factor.BackupCodes) == 0 {
			continue
		}

		var hashedCodes []string
		if err := json.Unmarshal(factor.BackupCodes, &hashedCodes); err != nil {
			// Log the error, but don't reveal the issue to the user
			fmt.Printf("failed to unmarshal backup codes for user %s: %v\n", user.ID, err)
			continue
		}

		for i, hashedCode := range hashedCodes {
			if p.crypto.CheckPasswordHash(code, hashedCode) {
				// Remove the used code from the list
				hashedCodes = append(hashedCodes[:i], hashedCodes[i+1:]...)
				codesJSON, err := json.Marshal(hashedCodes)
				if err != nil {
					return false, fmt.Errorf("failed to marshal remaining recovery codes: %w", err)
				}
				factor.BackupCodes = codesJSON
				if err := p.repo.UpdateMFAFactor(ctx, factor); err != nil {
					return false, fmt.Errorf("failed to update MFA factor after using recovery code: %w", err)
				}
				return true, nil
			}
		}
	}

	return false, nil
}
