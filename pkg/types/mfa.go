package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// MFAFactor stores the configuration for a user's MFA factor.
type MFAFactor struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	Type         string         `gorm:"type:varchar(20);not null" json:"type"`
	Status       string         `gorm:"type:varchar(20);not null" json:"status"`
	CredentialID string         `gorm:"type:varchar(255)" json:"credential_id"`
	PublicKey    []byte         `gorm:"type:bytea" json:"public_key"`
	Secret       string         `gorm:"type:text" json:"secret"`
	PhoneNumber  string         `gorm:"type:varchar(20)" json:"phone_number"`
	BackupCodes  datatypes.JSON `gorm:"type:jsonb" json:"-"`
	Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	LastUsedAt   *time.Time     `gorm:"type:timestamp" json:"last_used_at"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// MFAVerificationLog stores a log of MFA verification attempts.
type MFAVerificationLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FactorID    uuid.UUID `gorm:"type:uuid;not null" json:"factor_id"`
	Success     bool      `gorm:"not null" json:"success"`
	ErrorReason string    `gorm:"type:text" json:"error_reason"`
	IPAddress   string    `gorm:"type:inet" json:"ip_address"`
	UserAgent   string    `gorm:"type:text" json:"user_agent"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (f *MFAFactor) TableName() string {
	return "mfa_factors"
}

func (l *MFAVerificationLog) TableName() string {
	return "mfa_verification_logs"
}

// VerifyMFARequest represents a request to verify an MFA challenge.
type VerifyMFARequest struct {
	ChallengeID string `json:"challenge_id"`
	Code        string `json:"code"`
	UserID      string `json:"user_id"`
}

// MFAMethod represents an MFA method that is available to the user.
type MFAMethod struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// MFAEnrollment holds the necessary information for a user to complete MFA enrollment.
// This is a temporary object and should not be stored.
type MFAEnrollment struct {
	// Secret is the plaintext secret key for the user to enter into their authenticator app.
	Secret string `json:"secret"`
	// URL is the provisioning URL for the QR code.
	URL string `json:"url"`
	// RecoveryCodes are the one-time codes the user can use if they lose their device.
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}
