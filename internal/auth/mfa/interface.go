package mfa

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// StrengthLevel represents the strength of an MFA provider.
type StrengthLevel string

const (
	// StrengthLevelNormal is for standard MFA providers like TOTP.
	StrengthLevelNormal StrengthLevel = "normal"
	// StrengthLevelStrong is for phishing-resistant providers like WebAuthn.
	StrengthLevelStrong StrengthLevel = "strong"
)

// MFAProvider defines the interface for all MFA providers.
type MFAProvider interface {
	// Challenge generates an MFA challenge for the user.
	Challenge(ctx context.Context, user *types.User) (*types.MFAChallenge, error)
	// Verify verifies the user's response to an MFA challenge.
	Verify(ctx context.Context, user *types.User, code string) (bool, error)
	// ListMethods returns the MFA methods available for the user.
	ListMethods(ctx context.Context, user *types.User) ([]*types.MFAMethod, error)
	// GetStrength returns the strength of the MFA provider.
	GetStrength() StrengthLevel
}
