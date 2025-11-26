package config

// SecurityConfig holds the security-related configurations for the application.
type SecurityConfig struct {
	ResponsePolicies []ResponsePolicy `yaml:"response_policies"`
}

// ResponsePolicy defines a rule for triggering security actions based on a risk score.
type ResponsePolicy struct {
	// Name is a human-readable identifier for the policy.
	Name string `yaml:"name"`
	// RiskThreshold is the minimum risk score (inclusive) that triggers this policy.
	RiskThreshold float64 `yaml:"risk_threshold"`
	// ActionIDs is a list of action identifiers to be executed when the policy is triggered.
	ActionIDs []string `yaml:"actions"`
	// IsBlocking indicates if the actions in this policy should block the current workflow.
	IsBlocking bool `yaml:"is_blocking"`
}
