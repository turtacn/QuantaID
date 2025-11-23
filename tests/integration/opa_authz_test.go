//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/policy/engine"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// Mock providers for integration test
type MockRBAC struct{}

func (m *MockRBAC) IsAllowed(ctx context.Context, subjectID, action, resource string) (bool, error) {
	return true, nil // RBAC always allows in this test, so we can test OPA
}

type MockABAC struct{}

func (m *MockABAC) Evaluate(ctx context.Context, requestContext map[string]interface{}) (bool, error) {
	return true, nil // ABAC always allows
}

func TestOPAIntegration_SDK(t *testing.T) {
	// 1. Setup Policy
	policyContent := `package quantaid.authz

default allow = false

allow {
	input.action == "read"
	input.env.safe_ip == true
}
`
	tmpFile, err := os.CreateTemp("", "opa_integ_*.rego")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(policyContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// 2. Setup Config
	cfg := utils.OPAConfig{
		Enabled:    true,
		Mode:       "sdk",
		PolicyFile: tmpFile.Name(),
	}

	// 3. Setup Components
	opaProvider, err := engine.NewOPAProvider(cfg)
	assert.NoError(t, err)

	rbac := &MockRBAC{}
	abac := &MockABAC{}
	evaluator := engine.NewHybridEvaluator(rbac, abac, opaProvider)

	// 4. Test Scenarios
	ctx := context.Background()

	// Scenario 1: Allowed by OPA
	reqAllow := engine.EvaluationRequest{
		SubjectID: "user1",
		Action:    "read",
		Resource:  "data",
		Context: map[string]interface{}{
			"env": map[string]interface{}{
				"safe_ip": true,
			},
		},
	}
	allowed, err := evaluator.Evaluate(ctx, reqAllow)
	assert.NoError(t, err)
	assert.True(t, allowed, "Should be allowed by OPA")

	// Scenario 2: Denied by OPA
	reqDeny := engine.EvaluationRequest{
		SubjectID: "user1",
		Action:    "read",
		Resource:  "data",
		Context: map[string]interface{}{
			"env": map[string]interface{}{
				"safe_ip": false,
			},
		},
	}
	allowed, err = evaluator.Evaluate(ctx, reqDeny)
	assert.NoError(t, err)
	assert.False(t, allowed, "Should be denied by OPA")

	// Scenario 3: Hot Update Policy (Simulated by reloading OPA provider,
	// typically the service would watch file changes, but here we just re-init or check if we can update file)
	// Note: The current OPAProvider implementation loads policy once at startup.
	// To support hot reload, we would need to implement a watcher or method to reload.
	// For this test, we can verify that a new provider picks up the changes.

	newPolicyContent := `package quantaid.authz
default allow = true
# Changed default to true for test
`
	err = os.WriteFile(tmpFile.Name(), []byte(newPolicyContent), 0644)
	assert.NoError(t, err)

	// Re-init provider to pick up changes (in a real app, we'd have a watcher)
	opaProviderReloaded, err := engine.NewOPAProvider(cfg)
	assert.NoError(t, err)
	evaluatorReloaded := engine.NewHybridEvaluator(rbac, abac, opaProviderReloaded)

	allowed, err = evaluatorReloaded.Evaluate(ctx, reqDeny) // Was denied, now allowed by default
	assert.NoError(t, err)
	assert.True(t, allowed, "Should be allowed after policy update")
}
