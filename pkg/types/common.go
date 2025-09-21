package types

import "time"

// AuditLog represents a single audit trail event.
type AuditLog struct {
	ID        string                 `json:"id"`
	ActorID   string                 `json:"actorId"`   // The user who performed the action
	Action    string                 `json:"action"`    // e.g., "user.login", "policy.update"
	Resource  string                 `json:"resource"`  // The resource that was affected, e.g., "user:123"
	Status    string                 `json:"status"`    // "success" or "failure"
	Context   map[string]interface{} `json:"context"`   // Additional contextual data, e.g., IP address
	Timestamp time.Time              `json:"timestamp"`
}

// PaginationQuery defines parameters for paginated list queries.
type PaginationQuery struct {
	PageSize int
	Offset   int
}

//Personal.AI order the ending
