package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// AuditLog represents a single audit trail event, capturing a record of an action performed within the system.
type AuditLog struct {
	// ID is the unique identifier for the audit log entry.
	ID string `json:"id" gorm:"primaryKey"`
	// ActorID identifies the user or system principal that performed the action.
	ActorID string `json:"actorId" gorm:"index"`
	// Action is a string describing the action that was performed (e.g., "user.login", "policy.update").
	Action string `json:"action" gorm:"index"`
	// Resource identifies the entity that was affected by the action (e.g., "user:123").
	Resource string `json:"resource" gorm:"index"`
	// Status indicates whether the action was successful or failed.
	Status string `json:"status"`
	// Context contains additional contextual data about the event, such as IP address or user agent.
	Context JSONB `json:"context" gorm:"type:jsonb"`
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
}

// PaginationQuery defines parameters for paginated list queries,
// allowing clients to retrieve large datasets in smaller chunks.
type PaginationQuery struct {
	// PageSize specifies the maximum number of items to return in a single page.
	PageSize int
	// Offset is the number of items to skip before starting to collect the result set.
	Offset int
}

// JSONB represents a JSONB database type, which can be used for flexible data storage.
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface, allowing the JSONB type to be written to the database.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface, allowing the JSONB type to be read from the database.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("scan source is not []byte, but %T", value)
	}
	if len(bytes) == 0 {
		*j = nil
		return nil
	}
	return json.Unmarshal(bytes, j)
}
