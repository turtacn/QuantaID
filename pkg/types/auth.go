package types

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// AuthRequest represents a generic authentication request.
type AuthRequest struct {
	Protocol    ProtocolType           `json:"protocol"`
	Credentials map[string]string      `json:"credentials"`
	Context     map[string]interface{} `json:"context"`
}

// AuthResponse represents a generic authentication response.
type AuthResponse struct {
	Success bool          `json:"success"`
	Token   *Token        `json:"token,omitempty"`
	User    *User         `json:"user,omitempty"`
	Error   *Error        `json:"error,omitempty"`
	Next    *MFAChallenge `json:"next,omitempty"` // For multi-step auth
}

// Token represents an access token and related data.
type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"` // in seconds
	IDToken      string `json:"idToken,omitempty"` // For OIDC
}

// Claims represents the standard claims in a JWT.
type Claims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope,omitempty"`
}

// MFAChallenge represents a challenge for multi-factor authentication.
type MFAChallenge struct {
	MFAProvider AuthMethod               `json:"mfaProvider"`
	ChallengeID string                   `json:"challengeId"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// MFAVerificationRequest is used to submit an MFA response.
type MFAVerificationRequest struct {
	ChallengeID string `json:"challengeId"`
	Code        string `json:"code"`
}

// IdentityProvider represents an external identity source.
type IdentityProvider struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	Name          string                 `json:"name" gorm:"uniqueIndex;not null"`
	Type          ProtocolType           `json:"type" gorm:"not null"`
	Enabled       bool                   `json:"enabled" gorm:"not null;default:true"`
	Configuration map[string]interface{} `json:"configuration" gorm:"type:jsonb"`
	CreatedAt     time.Time              `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time              `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ConnectorConfig holds the configuration for a specific connector instance.
type ConnectorConfig struct {
	InstanceID string                 `json:"instanceId"`
	ProviderID string                 `json:"providerId"`
	Config     map[string]interface{} `json:"config"`
}

//Personal.AI order the ending
