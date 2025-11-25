package lifecycle

// Operator defines the comparison operator for rules.
type Operator string

const (
	OpEq       Operator = "eq"
	OpNeq      Operator = "neq"
	OpGt       Operator = "gt" // Greater than (for numbers/dates)
	OpLt       Operator = "lt" // Less than (for numbers/dates)
	OpGte      Operator = "gte"
	OpLte      Operator = "lte"
	OpContains Operator = "contains"
)

// ActionType defines the action to take when a rule is matched.
type ActionType string

const (
	ActionDisable ActionType = "disable"
	ActionDelete  ActionType = "delete"
	ActionNotify  ActionType = "notify"
	ActionArchive ActionType = "archive"
)

// LifecycleRule represents a rule for identity lifecycle management.
type LifecycleRule struct {
	Name        string      `json:"name" yaml:"name"`
	Description string      `json:"description,omitempty" yaml:"description"`
	Conditions  []Condition `json:"conditions" yaml:"conditions"`
	Actions     []Action    `json:"actions" yaml:"actions"`
}

// Condition represents a single check against a user attribute.
type Condition struct {
	Attribute string      `json:"attribute" yaml:"attribute"` // e.g. "lastLoginAt", "status", "attributes.department"
	Operator  Operator    `json:"operator" yaml:"operator"`
	Value     interface{} `json:"value" yaml:"value"`
}

// Action represents an operation to perform.
type Action struct {
	Type   ActionType             `json:"type" yaml:"type"`
	Params map[string]interface{} `json:"params,omitempty" yaml:"params"`
}

// ExecutionResult contains the result of evaluating rules.
type ExecutionResult struct {
	Actions []Action
	Rule    string
}
