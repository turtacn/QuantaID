package mfa

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/pkg/types"
)

// MFAManager manages all MFA providers.
type MFAManager struct {
	providers map[string]MFAProvider
}

// NewMFAManager creates a new MFAManager.
func NewMFAManager() *MFAManager {
	return &MFAManager{
		providers: make(map[string]MFAProvider),
	}
}

// RegisterProvider registers a new MFA provider.
func (m *MFAManager) RegisterProvider(name string, provider MFAProvider) {
	m.providers[name] = provider
}

// GetAvailableMFAMethods returns the MFA methods available for the user.
func (m *MFAManager) GetAvailableMFAMethods(ctx context.Context, user *types.User, requiredStrength StrengthLevel) ([]*types.MFAMethod, error) {
	var availableMethods []*types.MFAMethod

	for _, provider := range m.providers {
		if provider.GetStrength() == requiredStrength {
			methods, err := provider.ListMethods(ctx, user)
			if err != nil {
				return nil, err
			}
			availableMethods = append(availableMethods, methods...)
		}
	}

	return availableMethods, nil
}

// Challenge generates an MFA challenge for the user.
func (m *MFAManager) Challenge(ctx context.Context, user *types.User, providerName string) (*types.MFAChallenge, error) {
	provider, ok := m.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	return provider.Challenge(ctx, user)
}

// Verify verifies the user's response to an MFA challenge.
func (m *MFAManager) Verify(ctx context.Context, user *types.User, providerName, code string) (bool, error) {
	provider, ok := m.providers[providerName]
	if !ok {
		return false, fmt.Errorf("provider not found: %s", providerName)
	}

	return provider.Verify(ctx, user, code)
}
