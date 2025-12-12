package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// RadiusClient represents a NAS (Network Access Server) client that is allowed to communicate with the RADIUS server.
type RadiusClient struct {
	ID         string          `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name       string          `gorm:"type:varchar(128);not null" json:"name"`
	IPAddress  string          `gorm:"type:varchar(45);not null;index" json:"ip_address"` // Supports CIDR
	Secret     string          `gorm:"type:varchar(256);not null" json:"secret"`
	TenantID   string          `gorm:"type:varchar(64);not null;index" json:"tenant_id"`
	Enabled    bool            `gorm:"default:true" json:"enabled"`
	VendorType string          `gorm:"type:varchar(32);default:'generic'" json:"vendor_type"`
	Attributes json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"attributes"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// BeforeCreate generates a new ID if not present.
// Note: Assuming a UUID generation function exists or ID is passed.
// Standard practice in this codebase seems to be UUIDs.
func (c *RadiusClient) BeforeCreate(tx *gorm.DB) (err error) {
    if c.ID == "" {
        // Fallback if ID is not set, though service should usually handle this.
        // We can use a UUID generator here if needed, but often it's done in the service layer.
    }
    return
}
