package governance

// IssueType defines the category of the data quality issue.
type IssueType string

const (
	IssueMissingField   IssueType = "missing_field"
	IssueInvalidFormat  IssueType = "invalid_format"
	IssueLogicalError   IssueType = "logical_error"
	IssueDuplicateValue IssueType = "duplicate_value"
)

// QualityIssue represents a detected data quality problem.
type QualityIssue struct {
	UserID    string    `json:"userId"`
	Field     string    `json:"field"`
	IssueType IssueType `json:"issueType"`
	Message   string    `json:"message"`
}

// DataGovernanceConfig holds configuration for the inspector.
type DataGovernanceConfig struct {
	RequiredFields []string `yaml:"required_fields"`
	// Could add regex patterns here per field
}
