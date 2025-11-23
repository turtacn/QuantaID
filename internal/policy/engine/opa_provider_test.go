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
default deny = false

allow {
	input.user.roles[_] == "admin"
}

allow {
	input.action == "read"
	input.env.time == "day"
}

deny {
	input.env.ip == "1.1.1.1"
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
		name        string
		req         EvaluationRequest
		expectedAllow bool
		expectedDeny  bool
	}{
		{
			name: "Admin Allow",
			req: EvaluationRequest{
				SubjectID: "admin1",
				Context: map[string]interface{}{
					"roles": []string{"admin"},
				},
			},
			expectedAllow: true,
			expectedDeny:  false,
		},
		{
			name: "User Deny (No roles, wrong time)",
			req: EvaluationRequest{
				SubjectID: "user1",
				Action:    "read",
				Context: map[string]interface{}{
					"roles": []string{"user"},
					"env": map[string]interface{}{
						"time": "night",
					},
				},
			},
			expectedAllow: false,
			expectedDeny:  false,
		},
		{
			name: "User Allow (Correct time)",
			req: EvaluationRequest{
				SubjectID: "user1",
				Action:    "read",
				Context: map[string]interface{}{
					"roles": []string{"user"},
					"env": map[string]interface{}{
						"time": "day",
					},
				},
			},
			expectedAllow: true,
			expectedDeny:  false,
		},
		{
			name: "Explicit Deny (Blocked IP)",
			req: EvaluationRequest{
				SubjectID: "user1",
				Context: map[string]interface{}{
					"roles": []string{"admin"}, // Even admin is blocked
					"env": map[string]interface{}{
						"ip": "1.1.1.1",
					},
				},
			},
			expectedAllow: true, // Matches admin allow rule
			expectedDeny:  true, // Matches deny rule
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allowed, denied, err := provider.Evaluate(context.Background(), tc.req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAllow, allowed, "Allow mismatch")
			assert.Equal(t, tc.expectedDeny, denied, "Deny mismatch")
		})
	}
}

func TestOPAProvider_Disabled(t *testing.T) {
	config := utils.OPAConfig{
		Enabled: false,
	}
	provider, err := NewOPAProvider(config)
	assert.NoError(t, err)

	allowed, denied, err := provider.Evaluate(context.Background(), EvaluationRequest{})
	assert.NoError(t, err)
	assert.False(t, allowed, "Should default to false when OPA is disabled")
	assert.False(t, denied, "Should default to false when OPA is disabled")
}
