package policy

import "time"

// Subject represents the entity (e.g., user, service account) performing an action.
type Subject struct {
	UserID     string            `json:"user_id" yaml:"user_id"`
	Groups     []string          `json:"groups" yaml:"groups"`
	Attributes map[string]string `json:"attributes" yaml:"attributes"`
}

// Resource represents the object upon which an action is being performed.
type Resource struct {
	Type       string            `json:"type" yaml:"type"`
	ID         string            `json:"id" yaml:"id"`
	Attributes map[string]string `json:"attributes" yaml:"attributes"`
}

// Action represents the operation being performed.
type Action string

// Environment captures the contextual information of a request.
type Environment struct {
	IP          string    `json:"ip" yaml:"ip"`
	Time        time.Time `json:"time" yaml:"time"`
	DeviceTrust string    `json:"device_trust" yaml:"device_trust"`
}

// EvaluationContext bundles all information needed for a policy decision.
type EvaluationContext struct {
	Subject     Subject     `json:"subject" yaml:"subject"`
	Resource    Resource    `json:"resource" yaml:"resource"`
	Action      Action      `json:"action" yaml:"action"`
	Environment Environment `json:"environment" yaml:"environment"`
}

// Decision represents the outcome of a policy evaluation.
type Decision string

const (
	// DecisionAllow signifies that the request is permitted.
	DecisionAllow Decision = "allow"
	// DecisionDeny signifies that the request is denied.
	DecisionDeny Decision = "deny"
)