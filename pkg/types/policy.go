package types

import "time"

// Policy represents a single authorization policy.
type Policy struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Version     int       `json:"version" gorm:"not null;default:1"`
	Effect      Effect    `json:"effect" gorm:"not null"` // Allow or Deny
	Actions     []string  `json:"actions" gorm:"type:text[]"`
	Resources   []string  `json:"resources" gorm:"type:text[]"`
	Subjects    []string  `json:"subjects" gorm:"type:text[]"` // e.g., "user:123", "group:abc"
	Conditions  Condition `json:"conditions,omitempty" gorm:"type:jsonb"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Effect defines the effect of a policy.
type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

// Condition represents a set of conditions for a policy to be evaluated.
type Condition map[string]map[string]interface{}

// Rule is a simplified representation of a policy, often used in RBAC.
type Rule struct {
	ID        string   `json:"id"`
	Subject   string   `json:"subject"` // e.g., a Role or Group ID
	Resource  string   `json:"resource"`
	Actions   []string `json:"actions"`
	Effect    Effect   `json:"effect"`
}

// PolicyEvaluationContext contains all the information needed to evaluate a policy.
type PolicyEvaluationContext struct {
	Subject   map[string]interface{} `json:"subject"`
	Action    string                 `json:"action"`
	Resource  map[string]interface{} `json:"resource"`
	Context   map[string]interface{} `json:"context"`
}

// PolicyDecision represents the outcome of a policy evaluation.
type PolicyDecision struct {
	Allowed         bool     `json:"allowed"`
	Reason          string   `json:"reason,omitempty"`
	MatchingPolicies []string `json:"matchingPolicies,omitempty"`
}

// PolicyType defines the type of policy model.
type PolicyType string

const (
	PolicyTypeRBAC PolicyType = "rbac" // Role-Based Access Control
	PolicyTypeABAC PolicyType = "abac" // Attribute-Based Access Control
)

//Personal.AI order the ending
