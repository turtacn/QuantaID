package types

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// UserMFAConfig stores the MFA configuration for a user.
type UserMFAConfig struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	Method       string         `gorm:"type:varchar(20);not null" json:"method"`
	Config       datatypes.JSON `gorm:"type:jsonb;not null" json:"config"`
	BackupCodes  []string       `gorm:"type:text[];" json:"-"` // Encrypted backup codes
	Enabled      bool           `gorm:"default:true" json:"enabled"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// MFAVerificationLog stores a log of MFA verification attempts.
type MFAVerificationLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Method     string    `gorm:"type:varchar(20);not null" json:"method"`
	Success    bool      `gorm:"not null" json:"success"`
	IPAddress  string    `gorm:"type:inet" json:"ip_address"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (c *UserMFAConfig) TableName() string {
	return "user_mfa_configs"
}

func (l *MFAVerificationLog) TableName() string {
	return "mfa_verification_logs"
}
