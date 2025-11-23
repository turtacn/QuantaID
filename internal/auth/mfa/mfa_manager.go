package mfa

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/notification"
)

// MFAManager manages all MFA providers.
type MFAManager struct {
	providers       map[string]MFAProvider
	otpProvider     *OTPProvider
	notifierManager notification.Manager
}

// NewMFAManager creates a new MFAManager.
func NewMFAManager() *MFAManager {
	return &MFAManager{
		providers: make(map[string]MFAProvider),
	}
}

// SetOTPProvider sets the OTP provider for the manager
func (m *MFAManager) SetOTPProvider(otpProvider *OTPProvider) {
	m.otpProvider = otpProvider
}

// SetNotifierManager sets the Notifier manager
func (m *MFAManager) SetNotifierManager(manager notification.Manager) {
	m.notifierManager = manager
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
    // Note: OTP methods (Email/SMS) might need to be appended here if they are not treated as standard providers in the map
    // For now assuming they are registered or handled separately.

	return availableMethods, nil
}

// Challenge generates an MFA challenge for the user.
func (m *MFAManager) Challenge(ctx context.Context, user *types.User, providerName string) (*types.MFAChallenge, error) {
	// Handle OTP providers (Email, SMS) specifically if they are not in the standard provider map
	// Or assuming "email" and "sms" are passed as providerName
	if providerName == "email" || providerName == "sms" {
		if m.otpProvider == nil {
			return nil, fmt.Errorf("otp provider not configured")
		}

		target := user.Email
		method := "email"
		if providerName == "sms" {
			target = user.Phone
			method = "sms"
		}

		if target == "" {
			return nil, fmt.Errorf("user contact info missing for %s", providerName)
		}

		_, err := m.otpProvider.Challenge(ctx, user.ID, target, method)
		if err != nil {
			return nil, err
		}

		// Return a generic challenge response
		authMethod := types.AuthMethodEmailOTP
		if providerName == "sms" {
			authMethod = types.AuthMethodSMS
		}

		return &types.MFAChallenge{
			MFAProvider: authMethod,
			// ChallengeID is optional/empty here as OTPProvider doesn't return one yet
		}, nil
	}

	provider, ok := m.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	return provider.Challenge(ctx, user)
}

// Verify verifies the user's response to an MFA challenge.
func (m *MFAManager) Verify(ctx context.Context, user *types.User, providerName, code string) (bool, error) {
	if providerName == "email" || providerName == "sms" {
		if m.otpProvider == nil {
			return false, fmt.Errorf("otp provider not configured")
		}
		return m.otpProvider.Verify(ctx, user.ID, code)
	}

	provider, ok := m.providers[providerName]
	if !ok {
		return false, fmt.Errorf("provider not found: %s", providerName)
	}

	return provider.Verify(ctx, user, code)
}
