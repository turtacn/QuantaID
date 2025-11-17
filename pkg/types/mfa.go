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
