package types

import "time"

// AuditLog represents a single audit trail event, capturing a record of an action performed within the system.
type AuditLog struct {
	// ID is the unique identifier for the audit log entry.
	ID string `json:"id"`
	// ActorID identifies the user or system principal that performed the action.
	ActorID string `json:"actorId"`
	// Action is a string describing the action that was performed (e.g., "user.login", "policy.update").
	Action string `json:"action"`
	// Resource identifies the entity that was affected by the action (e.g., "user:123").
	Resource string `json:"resource"`
	// Status indicates whether the action was successful or failed.
	Status string `json:"status"`
	// Context contains additional contextual data about the event, such as IP address or user agent.
	Context map[string]interface{} `json:"context"`
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
