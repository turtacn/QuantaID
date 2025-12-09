package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// SessionEventType defines the type of session lifecycle event.
type SessionEventType string

const (
	SessionEventCreated         SessionEventType = "created"
	SessionEventEvaluated       SessionEventType = "evaluated"
	SessionEventDowngraded      SessionEventType = "downgraded"
	SessionEventUpgraded        SessionEventType = "upgraded"
	SessionEventStepUpRequired  SessionEventType = "step_up_required"
	SessionEventReauthenticated SessionEventType = "reauthenticated"
	SessionEventTerminated      SessionEventType = "terminated"
	SessionEventExpired         SessionEventType = "expired"
	SessionEventRiskChanged     SessionEventType = "risk_changed"
)

// SessionState represents the state of a session at a specific point in time.
type SessionState struct {
	Status          string    `json:"status"`
	RiskLevel       string    `json:"risk_level"`
	Permissions     []string  `json:"permissions,omitempty"`
	AuthLevel       int       `json:"auth_level"`
	LastEvaluatedAt time.Time `json:"last_evaluated_at"`
}

// Value implements the driver.Valuer interface for SessionState.
func (s SessionState) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for SessionState.
func (s *SessionState) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &s)
}

// SessionEvent records a significant event in a session's lifecycle.
type SessionEvent struct {
	ID            string           `gorm:"primaryKey;type:varchar(64)"`
	SessionID     string           `gorm:"index;type:varchar(64);not null"`
	TenantID      string           `gorm:"index;type:varchar(64);not null"`
	UserID        string           `gorm:"index;type:varchar(64);not null"`
	EventType     SessionEventType `gorm:"type:varchar(32);not null"`
	PreviousState SessionState     `gorm:"type:jsonb"`
	NewState      SessionState     `gorm:"type:jsonb"`
	TriggerReason string           `gorm:"type:varchar(256)"`
	RiskScore     int              `gorm:"type:int"`
	IPAddress     string           `gorm:"type:varchar(45)"`
	DeviceID      string           `gorm:"type:varchar(64)"`
	Metadata      map[string]any   `gorm:"type:jsonb;serializer:json"`
	CreatedAt     time.Time        `gorm:"index;autoCreateTime"`
}
