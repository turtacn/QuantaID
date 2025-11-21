package types

import (
	"context"
	"time"
)

// Application represents a client application that integrates with QuantaID
// for authentication and authorization. This could be a web app, a mobile app,
// or a third-party service.
type Application struct {
	// ID is the unique identifier for the application.
	ID string `json:"id" gorm:"primaryKey"`
	// Name is a human-readable name for the application.
	Name string `json:"name" gorm:"uniqueIndex;not null"`
	// Description provides more details about the application's purpose.
	Description string `json:"description,omitempty"`
	// Status indicates the current state of the application.
	Status ApplicationStatus `json:"status" gorm:"not null;default:'active'"`
	// ClientType specifies the application's type, determining its security profile.
	ClientType ClientType `json:"clientType" gorm:"not null;default:'confidential'"`
	// Protocol specifies the primary authentication protocol used by the application (e.g., saml, oidc).
	Protocol ProtocolType `json:"protocol" gorm:"not null"`
	// ProtocolConfig stores protocol-specific settings in a flexible JSONB format.
	// For SAML, this would include ACS URL, Entity ID, etc.
	// For OIDC, this would include redirect URIs, grant types, etc.
	ProtocolConfig JSONB `json:"protocolConfig" gorm:"type:jsonb"`
	// CreatedAt is the timestamp when the application was registered.
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	// UpdatedAt is the timestamp of the last update.
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ApplicationRepository defines the interface for application persistence.
type ApplicationRepository interface {
	CreateApplication(ctx context.Context, app *Application) error
	GetApplicationByID(ctx context.Context, id string) (*Application, error)
	GetApplicationByClientID(ctx context.Context, clientID string) (*Application, error)
	GetApplicationByName(ctx context.Context, name string) (*Application, error)
	UpdateApplication(ctx context.Context, app *Application) error
	DeleteApplication(ctx context.Context, id string) error
	ListApplications(ctx context.Context, pq PaginationQuery) ([]*Application, error)
}

// ApplicationStatus defines the possible states of an application.
type ApplicationStatus string

// Possible application statuses.
const (
	ApplicationStatusActive   ApplicationStatus = "active"
	ApplicationStatusInactive ApplicationStatus = "inactive"
)

// ClientType defines the application type (e.g., confidential, public).
type ClientType string

// Possible client types.
const (
	ClientTypeConfidential ClientType = "confidential"
	ClientTypePublic       ClientType = "public"
)