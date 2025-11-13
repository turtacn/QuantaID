package audit

import "time"

// AuditEvent defines the universal event model for audit logging.
type AuditEvent struct {
	ID        string         `json:"id"`
	Timestamp time.Time      `json:"ts"`
	Category  string         `json:"category"` // e.g., auth, policy, admin, mfa, risk
	Action    string         `json:"action"`   // e.g., login_success, policy_evaluated
	UserID    string         `json:"user_id,omitempty"`
	IP        string         `json:"ip,omitempty"`
	Resource  string         `json:"resource,omitempty"`
	Result    string         `json:"result"` // e.g., success, fail, deny
	TraceID   string         `json:"trace_id,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}
