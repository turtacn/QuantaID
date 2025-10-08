package types

import "time"

// Policy represents a single authorization policy, forming the core of the authorization engine.
// It defines who can do what on which resources, under what conditions.
type Policy struct {
	// ID is the unique identifier for the policy.
	ID string `json:"id" gorm:"primaryKey"`
	// Version allows for tracking changes to the policy over time.
	Version int `json:"version" gorm:"not null;default:1"`
	// Effect determines whether the policy grants (allows) or revokes (denies) permission.
	Effect Effect `json:"effect" gorm:"not null"`
	// Actions is a list of operations that the policy applies to (e.g., "read", "write").
	Actions []string `json:"actions" gorm:"type:text[]"`
	// Resources is a list of resources that the policy affects (e.g., "urn:myapp:documents:123").
	Resources []string `json:"resources" gorm:"type:text[]"`
	// Subjects is a list of users, groups, or roles to whom the policy applies.
	Subjects []string `json:"subjects" gorm:"type:text[]"`
	// Conditions specifies a set of attribute-based conditions that must be met for the policy to apply.
	Conditions Condition `json:"conditions,omitempty" gorm:"type:jsonb"`
	// Description is a human-readable explanation of the policy's purpose.
	Description string `json:"description,omitempty"`
	// CreatedAt is the timestamp when the policy was created.
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	// UpdatedAt is the timestamp of the last update.
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// Effect defines whether a policy allows or denies an action.
type Effect string

// Supported policy effects.
const (
	// EffectAllow grants permission if the policy matches.
	EffectAllow Effect = "allow"
	// EffectDeny revokes permission if the policy matches. Deny policies typically override allow policies.
	EffectDeny Effect = "deny"
)

// Condition represents a set of complex, attribute-based conditions for a policy to be evaluated.
// The outer map key is the condition operator (e.g., "StringEquals", "NumericGreaterThan"),
// and the inner map specifies the attributes and values to compare.
type Condition map[string]map[string]interface{}

// Rule is a simplified representation of a policy, often used in Role-Based Access Control (RBAC) systems.
// It links a subject (like a role) to a set of permissions on a resource.
type Rule struct {
	// ID is the unique identifier for the rule.
	ID string `json:"id"`
	// Subject identifies who the rule applies to (e.g., a role or group ID).
	Subject string `json:"subject"`
	// Resource is the entity the rule applies to.
	Resource string `json:"resource"`
	// Actions is a list of operations permitted or denied by the rule.
	Actions []string `json:"actions"`
	// Effect specifies whether the rule allows or denies the actions.
	Effect Effect `json:"effect"`
}

// PolicyEvaluationContext contains all the information required by the authorization engine
// to evaluate whether a request should be allowed or denied.
type PolicyEvaluationContext struct {
	// Subject contains attributes of the user or principal making the request.
	Subject map[string]interface{} `json:"subject"`
	// Action is the operation being attempted by the subject.
	Action string `json:"action"`
	// Resource contains attributes of the resource being accessed.
	Resource map[string]interface{} `json:"resource"`
	// Context contains environmental data, such as time of day or IP address.
	Context map[string]interface{} `json:"context"`
}

// PolicyDecision represents the outcome of a policy evaluation.
type PolicyDecision struct {
	// Allowed is true if the request is permitted, and false otherwise.
	Allowed bool `json:"allowed"`
	// Reason provides an explanation for the decision, especially useful for denials.
	Reason string `json:"reason,omitempty"`
	// MatchingPolicies lists the IDs of the policies that were used to make the decision.
	MatchingPolicies []string `json:"matchingPolicies,omitempty"`
}

// PolicyType defines the type of policy model being used.
type PolicyType string

// Supported policy model types.
const (
	// PolicyTypeRBAC represents Role-Based Access Control.
	PolicyTypeRBAC PolicyType = "rbac"
	// PolicyTypeABAC represents Attribute-Based Access Control.
	PolicyTypeABAC PolicyType = "abac"
)
