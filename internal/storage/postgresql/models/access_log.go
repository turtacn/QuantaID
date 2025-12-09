package models

import (
	"time"
)

// AccessLog represents an access event in the system.
type AccessLog struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TenantID    string    `json:"tenantId" gorm:"index"`
	UserID      string    `json:"userId" gorm:"index"`
	DeviceID    string    `json:"deviceId" gorm:"index"`
	EventType   string    `json:"eventType"` // login, logout, mfa_verify, etc.
	Action      string    `json:"action"` // legacy alias for EventType often used
	IPAddress   string    `json:"ipAddress"`
	UserAgent   string    `json:"userAgent"`
	Location    string    `json:"location"`
	Success     bool      `json:"success"`
	MFAVerified bool      `json:"mfaVerified"`
	Metadata    JSONMap   `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time `json:"createdAt" gorm:"index"`
	Timestamp   time.Time `json:"timestamp" gorm:"-"` // Alias for CreatedAt in some contexts
}

// AccessLogFilter defines criteria for querying access logs.
type AccessLogFilter struct {
	UserID    string
	DeviceID  string
	EventType string
	StartTime *time.Time
	EndTime   *time.Time
	Success   *bool
	Limit     int
	Offset    int
}
