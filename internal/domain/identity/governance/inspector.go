package governance

import (
	"regexp"

	"github.com/turtacn/QuantaID/pkg/types"
)

// Inspector checks users for data quality issues.
type Inspector struct {
	config DataGovernanceConfig
	// Regex patterns
	emailRegex *regexp.Regexp
	phoneRegex *regexp.Regexp
}

// NewInspector creates a new Data Quality Inspector.
func NewInspector(config DataGovernanceConfig) *Inspector {
	return &Inspector{
		config:     config,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`),
		// Simple phone regex for example
		phoneRegex: regexp.MustCompile(`^\+?[1-9]\d{1,14}$`),
	}
}

// Check evaluates a user for data quality issues.
func (i *Inspector) Check(user *types.User) []QualityIssue {
	var issues []QualityIssue

	// Check 1: Missing Required Fields
	for _, field := range i.config.RequiredFields {
		if i.isFieldMissing(user, field) {
			issues = append(issues, QualityIssue{
				UserID:    user.ID,
				Field:     field,
				IssueType: IssueMissingField,
				Message:   "Required field is missing or empty",
			})
		}
	}

	// Check 2: Format Validation (Email)
	if string(user.Email) != "" && !i.emailRegex.MatchString(string(user.Email)) {
		issues = append(issues, QualityIssue{
			UserID:    user.ID,
			Field:     "email",
			IssueType: IssueInvalidFormat,
			Message:   "Invalid email format",
		})
	}

	// Check 3: Format Validation (Phone)
	if string(user.Phone) != "" && !i.phoneRegex.MatchString(string(user.Phone)) {
		issues = append(issues, QualityIssue{
			UserID:    user.ID,
			Field:     "phone",
			IssueType: IssueInvalidFormat,
			Message:   "Invalid phone format",
		})
	}

	// Check 4: Logical Errors
	if !user.UpdatedAt.IsZero() && user.CreatedAt.After(user.UpdatedAt) {
		issues = append(issues, QualityIssue{
			UserID:    user.ID,
			Field:     "timestamps",
			IssueType: IssueLogicalError,
			Message:   "CreatedAt is after UpdatedAt",
		})
	}

	return issues
}

func (i *Inspector) isFieldMissing(user *types.User, field string) bool {
	switch field {
	case "email":
		return string(user.Email) == ""
	case "username":
		return user.Username == ""
	case "phone":
		return string(user.Phone) == "" // Optional usually, but if configured as required
	}
	// Check attributes
	if val, ok := user.Attributes[field]; ok {
		return val == nil || val == ""
	}
	return true // Missing if not found
}
