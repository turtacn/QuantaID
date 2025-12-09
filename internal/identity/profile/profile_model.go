package profile

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// UserProfile represents the core identity profile model
type UserProfile struct {
	ID        string `gorm:"primaryKey;type:varchar(64)"`
	UserID    string `gorm:"uniqueIndex;type:varchar(64);not null"`
	TenantID  string `gorm:"index;type:varchar(64);not null"`

	// Behavior metrics
	Behavior BehaviorMetrics `gorm:"type:jsonb"`

	// Risk indicators
	Risk      RiskIndicators `gorm:"type:jsonb"`
	RiskScore int            // 0-100 summary score
	RiskLevel string         // low/medium/high/critical

	// Tags
	AutoTags   StringSlice `gorm:"type:jsonb"` // System auto tags
	ManualTags StringSlice `gorm:"type:jsonb"` // Admin manual tags

	// Data Quality
	QualityScore   int            // 0-100
	QualityDetails QualityDetails `gorm:"type:jsonb"`

	// Timestamps
	LastActivityAt   *time.Time
	LastRiskUpdateAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// BehaviorMetrics tracks user behavior statistics
type BehaviorMetrics struct {
	TotalLogins         int64      `json:"total_logins"`
	FailedLogins        int64      `json:"failed_logins"`
	UniqueDevices       int        `json:"unique_devices"`
	UniqueLocations     int        `json:"unique_locations"`
	UniqueIPs           int        `json:"unique_ips"`
	AvgSessionDuration  float64    `json:"avg_session_duration"` // Minutes
	PeakActivityHours   []int      `json:"peak_activity_hours"`  // Active hours 0-23
	LastLocations       []string   `json:"last_locations"`       // Recent 5 locations
	LoginFrequency      float64    `json:"login_frequency"`      // Logins per week
	MFAUsageRate        float64    `json:"mfa_usage_rate"`       // MFA usage rate 0-1
	PasswordChangeCount int        `json:"password_change_count"`
	LastPasswordChange  *time.Time `json:"last_password_change,omitempty"`
}

// RiskIndicators tracks risk-related events
type RiskIndicators struct {
	AnomalyCount         int        `json:"anomaly_count"`
	LastAnomalyAt        *time.Time `json:"last_anomaly_at,omitempty"`
	GeoJumpCount         int        `json:"geo_jump_count"`
	FailedMFACount       int        `json:"failed_mfa_count"`
	SuspiciousIPCount    int        `json:"suspicious_ip_count"`
	UnusualTimeAccess    int        `json:"unusual_time_access"`
	NewDeviceCount30d    int        `json:"new_device_count_30d"`
	CompromiseIndicators []string   `json:"compromise_indicators"`
}

// QualityDetails details the data quality of the user profile
type QualityDetails struct {
	HasEmail         bool       `json:"has_email"`
	EmailVerified    bool       `json:"email_verified"`
	HasPhone         bool       `json:"has_phone"`
	PhoneVerified    bool       `json:"phone_verified"`
	HasMFA           bool       `json:"has_mfa"`
	HasRecoveryEmail bool       `json:"has_recovery_email"`
	ProfileComplete  float64    `json:"profile_complete"` // 0-1
	LastVerification *time.Time `json:"last_verification,omitempty"`
}

// StringSlice is a helper type for JSONB array storage
type StringSlice []string

// Value implements driver.Valuer interface for BehaviorMetrics
func (b BehaviorMetrics) Value() (driver.Value, error) {
	return json.Marshal(b)
}

// Scan implements sql.Scanner interface for BehaviorMetrics
func (b *BehaviorMetrics) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, b)
}

// Value implements driver.Valuer interface for RiskIndicators
func (r RiskIndicators) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements sql.Scanner interface for RiskIndicators
func (r *RiskIndicators) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, r)
}

// Value implements driver.Valuer interface for QualityDetails
func (q QualityDetails) Value() (driver.Value, error) {
	return json.Marshal(q)
}

// Scan implements sql.Scanner interface for QualityDetails
func (q *QualityDetails) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, q)
}

// Value implements driver.Valuer interface for StringSlice
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner interface for StringSlice
func (s *StringSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// Predefined tag constants
const (
	TagFrequentTraveler  = "frequent_traveler"
	TagHighValueUser     = "high_value"
	TagDormantUser       = "dormant"
	TagNewUser           = "new_user"
	TagSecurityConscious = "security_conscious"
	TagHighRisk          = "high_risk"
	TagVIP               = "vip"
)
