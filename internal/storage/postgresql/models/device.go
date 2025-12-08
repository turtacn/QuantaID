package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONMap represents a JSON object stored in the database
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for JSONMap
func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONMap
func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &j)
}

// Device represents a user's device
type Device struct {
	ID             string     `gorm:"primaryKey"`
	TenantID       string     `gorm:"index"`
	UserID         string     `gorm:"index"`          // Can be empty if not bound
	Fingerprint    string     `gorm:"uniqueIndex"`    // Device fingerprint hash
	FingerprintRaw JSONMap    `gorm:"type:jsonb"`     // Raw fingerprint data
	DeviceName     string     // User visible name
	DeviceType     string     // mobile/desktop/tablet
	OS             string
	Browser        string
	TrustScore     int        `gorm:"default:0"`      // 0-100
	LastIP         string
	LastLocation   string     // City level, optional
	LastActiveAt   time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	BoundAt        *time.Time // Binding time, used for trust calculation
}
