package engine

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/pkg/utils"
)

func TestOPAProvider(t *testing.T) {
	// 1. Setup Policy File
	policyContent := `package quantaid.authz

default allow = false

allow {
	input.user.roles[_] == "admin"
}

allow {
	input.action == "read"
	input.env.time == "day"
}
`
	tmpFile, err := os.CreateTemp("", "opa_policy_*.rego")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(policyContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// 2. Init Provider
	config := utils.OPAConfig{
		Enabled:    true,
		Mode:       "sdk",
		PolicyFile: tmpFile.Name(),
	}

	provider, err := NewOPAProvider(config)
	assert.NoError(t, err)

	// 3. Test Cases
	tests := []struct {
		name     string
		req      EvaluationRequest
		expected bool
	}{
		{
			name: "Admin Allow",
			req: EvaluationRequest{
				SubjectID: "admin1",
				Context: map[string]interface{}{
					"roles": []string{"admin"},
				},
			},
			expected: true,
		},
		{
			name: "User Deny (No roles, wrong time)",
			req: EvaluationRequest{
				SubjectID: "user1",
				Action:    "read",
				Context: map[string]interface{}{
					"roles": []string{"user"},
					"env": map[string]string{
						"time": "night",
					},
				},
			},
			expected: false,
		},
		{
			name: "User Allow (Correct time)",
			req: EvaluationRequest{
				SubjectID: "user1",
				Action:    "read",
				Context: map[string]interface{}{
					"roles": []string{"user"},
					"env": map[string]string{
						"time": "day",
					},
				},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allowed, err := provider.Evaluate(context.Background(), tc.req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, allowed)
		})
	}
}

func TestOPAProvider_Disabled(t *testing.T) {
	config := utils.OPAConfig{
		Enabled: false,
	}
	provider, err := NewOPAProvider(config)
	assert.NoError(t, err)

	allowed, err := provider.Evaluate(context.Background(), EvaluationRequest{})
	assert.NoError(t, err)
	assert.True(t, allowed, "Should default to true when OPA is disabled")
}
