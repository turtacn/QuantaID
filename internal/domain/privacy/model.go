package privacy

import (
	"context"
	"time"
)

// ConsentAction represents the type of consent action.
type ConsentAction string

const (
	// ConsentActionGranted indicates the user granted consent.
	ConsentActionGranted ConsentAction = "GRANTED"
	// ConsentActionRevoked indicates the user revoked consent.
	ConsentActionRevoked ConsentAction = "REVOKED"
)

// ConsentRecord stores a user's consent action for a specific policy version.
type ConsentRecord struct {
	ID            string        `json:"id" gorm:"primaryKey"`
	UserID        string        `json:"userId" gorm:"index"`
	PolicyID      string        `json:"policyId"`
	PolicyVersion string        `json:"policyVersion"`
	Action        ConsentAction `json:"action"`
	UserAgent     string        `json:"userAgent"`
	IPAddress     string        `json:"ipAddress"`
	CreatedAt     time.Time     `json:"createdAt" gorm:"autoCreateTime"`
}

// DSRRequestStatus defines the possible states of a DSR request.
type DSRRequestStatus string

const (
	DSRRequestStatusPending   DSRRequestStatus = "PENDING"
	DSRRequestStatusProcessing DSRRequestStatus = "PROCESSING"
	DSRRequestStatusCompleted DSRRequestStatus = "COMPLETED"
	DSRRequestStatusFailed    DSRRequestStatus = "FAILED"
)

// DSRRequestType defines the type of DSR request.
type DSRRequestType string

const (
	DSRRequestTypeExport  DSRRequestType = "EXPORT"
	DSRRequestTypeErasure DSRRequestType = "ERASURE"
)

// DSRRequest represents a Data Subject Right request.
type DSRRequest struct {
	ID        string           `json:"id" gorm:"primaryKey"`
	UserID    string           `json:"userId" gorm:"index"`
	Type      DSRRequestType   `json:"type"`
	Status    DSRRequestStatus `json:"status"`
	CreatedAt time.Time        `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time        `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Repository defines the interface for privacy-related database operations.
type Repository interface {
	CreateConsentRecord(ctx context.Context, record *ConsentRecord) error
	GetLastConsentRecord(ctx context.Context, userID, policyID string) (*ConsentRecord, error)
	GetConsentHistory(ctx context.Context, userID string) ([]*ConsentRecord, error)
	CreateDSRRequest(ctx context.Context, request *DSRRequest) error
	GetDSRRequest(ctx context.Context, requestID string) (*DSRRequest, error)
	UpdateDSRRequestStatus(ctx context.Context, requestID string, status DSRRequestStatus) error
}
