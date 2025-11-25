package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/domain/identity/governance"
	"github.com/turtacn/QuantaID/pkg/types"
)

func Test_Quality_MissingEmail(t *testing.T) {
	config := governance.DataGovernanceConfig{
		RequiredFields: []string{"email"},
	}
	inspector := governance.NewInspector(config)

	user := &types.User{
		ID:       "u1",
		Username: "user1",
		// Email is empty
	}

	issues := inspector.Check(user)
	assert.NotEmpty(t, issues)

	found := false
	for _, issue := range issues {
		if issue.Field == "email" && issue.IssueType == governance.IssueMissingField {
			found = true
			break
		}
	}
	assert.True(t, found, "Should detect missing email")
}

func Test_Quality_InvalidFormat(t *testing.T) {
	config := governance.DataGovernanceConfig{
		RequiredFields: []string{},
	}
	inspector := governance.NewInspector(config)

	user := &types.User{
		ID:       "u2",
		Username: "user2",
		Email:    "invalid-email", // Missing @ and domain
	}

	issues := inspector.Check(user)
	assert.NotEmpty(t, issues)

	found := false
	for _, issue := range issues {
		if issue.Field == "email" && issue.IssueType == governance.IssueInvalidFormat {
			found = true
			break
		}
	}
	assert.True(t, found, "Should detect invalid email format")
}

func Test_Quality_LogicalError(t *testing.T) {
	inspector := governance.NewInspector(governance.DataGovernanceConfig{})

	user := &types.User{
		ID:        "u3",
		Username:  "user3",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now().Add(-1 * time.Hour), // Updated before Created
	}

	issues := inspector.Check(user)
	assert.NotEmpty(t, issues)

	found := false
	for _, issue := range issues {
		if issue.IssueType == governance.IssueLogicalError {
			found = true
			break
		}
	}
	assert.True(t, found, "Should detect logical error (Created > Updated)")
}
