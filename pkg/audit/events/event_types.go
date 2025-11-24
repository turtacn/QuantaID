package events

import "time"

// EventType defines the type of audit event.
type EventType string

// --- Event Type Constants ---
const (
	// Authentication Events
	EventLoginSuccess    EventType = "auth.login.success"
	EventLoginFailure    EventType = "auth.login.failure"
	EventLogout          EventType = "auth.logout"
	EventMFAVerified     EventType = "auth.mfa.verified"
	EventPasswordChanged EventType = "auth.password.changed"
	AuthRiskEvaluated    EventType = "auth.risk.evaluated"
	MFAChallenged        EventType = "auth.mfa.challenged"
	MFACompleted         EventType = "auth.mfa.completed"

	// Authorization Events
	EventPermissionGranted EventType = "authz.permission.granted"
	EventPermissionRevoked EventType = "authz.permission.revoked"
	EventRoleAssigned      EventType = "authz.role.assigned"
	EventRoleRemoved       EventType = "authz.role.removed"

	// Data Access Events
	EventDataRead     EventType = "data.read"
	EventDataExport   EventType = "data.export"
	EventDataModified EventType = "data.modified"
	EventDataDeleted  EventType = "data.deleted"

	// System Events
	EventConfigChanged EventType = "system.config.changed"
	EventKeyRotated    EventType = "system.key.rotated"
	EventBackupCreated EventType = "system.backup.created"
)

// Actor represents the entity that performed the action.
type Actor struct {
	ID   string `json:"id"`
	Type string `json:"type"` // 'user', 'service', 'system'
	Name string `json:"name,omitempty"`
}

// Target represents the entity on which the action was performed.
type Target struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// Result indicates the outcome of the action.
type Result string

const (
	ResultSuccess Result = "success"
	ResultFailure Result = "failure"
)

// AuditEvent defines the structured log for an auditable action.
type AuditEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType EventType              `json:"event_type"`
	Actor     Actor                  `json:"actor"`
	Target    Target                 `json:"target"`
	Action    string                 `json:"action"`
	Result    Result                 `json:"result"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Category  string                 `json:"category,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	Details   map[string]any         `json:"details,omitempty"`
	Resource  string                 `json:"resource,omitempty"`
}

// --- Standardized Metadata Structs ---

// LoginMetadata contains details for authentication events.
type LoginMetadata struct {
	Protocol   string   `json:"protocol"` // "oauth2" | "saml" | "ldap"
	ClientID   string   `json:"client_id,omitempty"`
	Scopes     []string `json:"scopes,omitempty"`
	FailReason string   `json:"fail_reason,omitempty"`
}

// PermissionChangeMetadata contains details for authorization changes.
type PermissionChangeMetadata struct {
	Resource      string   `json:"resource"`
	Permissions   []string `json:"permissions"`
	ChangedBy     string   `json:"changed_by"`
	Justification string   `json:"justification,omitempty"`
}
