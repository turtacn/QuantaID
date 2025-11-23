//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/policy/engine"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/utils"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Mock RBAC Provider for Hybrid Flow Test
type MockRBACProvider struct {
	Allowed bool
}

func (m *MockRBACProvider) IsAllowed(ctx context.Context, subjectID, action, resource string) (bool, error) {
	return m.Allowed, nil
}

func TestHybridFlow(t *testing.T) {
	// 1. Setup OPA Policy
	policyContent := `package quantaid.authz
default allow = false
default deny = false

# Allow if action is view
allow {
	input.action == "view"
}

# Deny if IP is blocked
deny {
	input.env.ip == "1.1.1.1"
}
`
	tmpFile, err := os.CreateTemp("", "hybrid_policy_*.rego")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString(policyContent)
	require.NoError(t, err)
	tmpFile.Close()

	// 2. Init Evaluator
	opaConfig := utils.OPAConfig{
		Enabled:    true,
		Mode:       "sdk",
		PolicyFile: tmpFile.Name(),
	}
	opaProvider, err := engine.NewOPAProvider(opaConfig)
	require.NoError(t, err)

	// Scenario A: RBAC Allows, OPA Allows (implied by not denying) -> ALLOW
	t.Run("RBAC Allow, OPA Neutral", func(t *testing.T) {
		// Create evaluator with RBAC allowing
		rbac := &MockRBACProvider{Allowed: true}
		evaluator := engine.NewHybridEvaluator(rbac, nil, opaProvider)

		req := engine.EvaluationRequest{
			SubjectID: "alice",
			Action:    "edit", // OPA doesn't explicitly allow 'edit', but doesn't deny it either
			Resource:  "doc:1",
			Context:   map[string]interface{}{},
		}

		allowed, err := evaluator.Evaluate(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, allowed, "Should be allowed because RBAC allows and OPA doesn't deny")
	})

	// Scenario B: RBAC Denies, OPA Allows -> ALLOW (Override)
	t.Run("RBAC Deny, OPA Allow", func(t *testing.T) {
		rbac := &MockRBACProvider{Allowed: false}
		evaluator := engine.NewHybridEvaluator(rbac, nil, opaProvider)

		req := engine.EvaluationRequest{
			SubjectID: "bob",
			Action:    "view", // OPA explicitly allows 'view'
			Resource:  "report:1",
			Context:   map[string]interface{}{},
		}

		allowed, err := evaluator.Evaluate(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, allowed, "Should be allowed because OPA overrides RBAC deny")
	})

	// Scenario C: RBAC Allows, OPA Denies -> DENY
	t.Run("RBAC Allow, OPA Deny", func(t *testing.T) {
		rbac := &MockRBACProvider{Allowed: true}
		evaluator := engine.NewHybridEvaluator(rbac, nil, opaProvider)

		req := engine.EvaluationRequest{
			SubjectID: "charlie",
			Action:    "edit",
			Resource:  "doc:1",
			Context: map[string]interface{}{
				"env": map[string]interface{}{
					"ip": "1.1.1.1", // Trigger OPA deny
				},
			},
		}

		allowed, err := evaluator.Evaluate(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, allowed, "Should be denied because OPA explicitly denies")
	})
}
