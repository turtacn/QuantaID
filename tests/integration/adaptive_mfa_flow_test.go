package integration

import (
	"context"
	"github.com/turtacn/QuantaID/internal/orchestrator"
	"github.com/turtacn/QuantaID/internal/orchestrator/workflows"
	"github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

import "github.com/turtacn/QuantaID/pkg/utils"

func TestLoginFlow_NoMFA_WhenLowRisk(t *testing.T) {
	engine := orchestrator.NewEngine(utils.NewNoopLogger())
	riskEngine := auth.NewSimpleRiskEngine(auth.SimpleRiskConfig{
		NewDeviceScore:   0.3,
		GeoVelocityScore: 0.3,
		UnusualTimeScore: 0.2,
		MfaThreshold:     0.3,
		BlockThreshold:   0.7,
	})
	workflows.NewAuthWorkflow(engine, nil, riskEngine)

	initialState := orchestrator.State{
		"username":           "testuser",
		"password":           "password",
		"client_ip":          "192.168.1.1",
		"client_country":     "US",
		"last_login_ip":      "192.168.1.1",
		"last_login_country": "US",
	}

	finalState, err := engine.Execute(context.Background(), "standard_auth_flow", initialState)
	assert.NoError(t, err)
	assert.False(t, finalState["mfa_required"].(bool))
}

func TestLoginFlow_RequireMFA_WhenMediumRisk(t *testing.T) {
	engine := orchestrator.NewEngine(utils.NewNoopLogger())
	riskEngine := auth.NewSimpleRiskEngine(auth.SimpleRiskConfig{
		NewDeviceScore:   0.3,
		GeoVelocityScore: 0.3,
		UnusualTimeScore: 0.2,
		MfaThreshold:     0.3,
		BlockThreshold:   0.7,
	})
	workflows.NewAuthWorkflow(engine, nil, riskEngine)

	initialState := orchestrator.State{
		"username":           "testuser",
		"password":           "password",
		"client_ip":          "198.51.100.0",
		"client_country":     "US",
		"last_login_ip":      "192.168.1.1",
		"last_login_country": "US",
	}

	finalState, err := engine.Execute(context.Background(), "standard_auth_flow", initialState)
	assert.NoError(t, err)
	assert.True(t, finalState["mfa_required"].(bool))
}

func TestLoginFlow_Block_WhenHighRisk(t *testing.T) {
	engine := orchestrator.NewEngine(utils.NewNoopLogger())
	riskEngine := auth.NewSimpleRiskEngine(auth.SimpleRiskConfig{
		NewDeviceScore:   0.3,
		GeoVelocityScore: 0.3,
		UnusualTimeScore: 0.2,
		BlockThreshold:   0.7,
		MfaThreshold:     0.3,
	})
	workflows.NewAuthWorkflow(engine, nil, riskEngine)

	initialState := orchestrator.State{
		"username":           "testuser",
		"password":           "password",
		"client_ip":          "203.0.113.0",
		"client_country":     "CN",
		"last_login_ip":      "192.168.1.1",
		"last_login_country": "US",
		"now":                time.Date(2023, 1, 1, 2, 0, 0, 0, time.UTC),
	}

	_, err := engine.Execute(context.Background(), "standard_auth_flow", initialState)
	assert.Error(t, err)
}
