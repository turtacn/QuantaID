package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	domain_auth "github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/audit"
	services_audit "github.com/turtacn/QuantaID/internal/services/audit"
	services_auth "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/tests/testutils"
	"go.uber.org/zap"
)


func TestSimpleRiskEngine_LowRisk(t *testing.T) {
	cfg := services_auth.SimpleRiskConfig{
		NewDeviceScore:   0.3,
		GeoVelocityScore: 0.3,
		UnusualTimeScore: 0.2,
		MfaThreshold:     0.3,
		BlockThreshold:   0.7,
	}
	logger, _ := zap.NewDevelopment()
	auditPipeline := audit.NewPipeline(logger, &testutils.MockSink{})
	auditService := services_audit.NewService(auditPipeline)
	engine := services_auth.NewSimpleRiskEngine(cfg, auditService)

	loginCtx := domain_auth.LoginContext{
		CurrentIP:        "192.168.1.1",
		CurrentCountry:   "US",
		Now:              time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
		LastLoginIP:      "192.168.1.1",
		LastLoginCountry: "US",
	}

	assessment, err := engine.Assess(context.Background(), loginCtx)
	assert.NoError(t, err)
	assert.Equal(t, domain_auth.RiskScore(0.0), assessment.Score)
	assert.Empty(t, assessment.Factors)
	assert.Equal(t, domain_auth.RiskDecisionAllow, assessment.Decision)
}

func TestSimpleRiskEngine_MediumRisk(t *testing.T) {
	cfg := services_auth.SimpleRiskConfig{
		NewDeviceScore:   0.3,
		GeoVelocityScore: 0.3,
		UnusualTimeScore: 0.2,
		MfaThreshold:     0.3,
		BlockThreshold:   0.7,
	}
	logger, _ := zap.NewDevelopment()
	auditPipeline := audit.NewPipeline(logger, &testutils.MockSink{})
	auditService := services_audit.NewService(auditPipeline)
	engine := services_auth.NewSimpleRiskEngine(cfg, auditService)

	loginCtx := domain_auth.LoginContext{
		CurrentIP:        "198.51.100.0",
		CurrentCountry:   "US",
		Now:              time.Date(2023, 1, 1, 23, 0, 0, 0, time.UTC),
		LastLoginIP:      "192.168.1.1",
		LastLoginCountry: "US",
	}

	assessment, err := engine.Assess(context.Background(), loginCtx)
	assert.NoError(t, err)
	assert.Equal(t, domain_auth.RiskScore(0.5), assessment.Score)
	assert.Contains(t, assessment.Factors, domain_auth.RiskFactorNewDevice)
	assert.Contains(t, assessment.Factors, domain_auth.RiskFactorUnusualTime)
	assert.Equal(t, domain_auth.RiskDecisionRequireMFA, assessment.Decision)
}

func TestSimpleRiskEngine_HighRisk(t *testing.T) {
	cfg := services_auth.SimpleRiskConfig{
		NewDeviceScore:   0.3,
		GeoVelocityScore: 0.3,
		UnusualTimeScore: 0.2,
		MfaThreshold:     0.3,
		BlockThreshold:   0.7,
	}
	logger, _ := zap.NewDevelopment()
	auditPipeline := audit.NewPipeline(logger, &testutils.MockSink{})
	auditService := services_audit.NewService(auditPipeline)
	engine := services_auth.NewSimpleRiskEngine(cfg, auditService)

	loginCtx := domain_auth.LoginContext{
		CurrentIP:        "203.0.113.0",
		CurrentCountry:   "CN",
		Now:              time.Date(2023, 1, 1, 2, 0, 0, 0, time.UTC),
		LastLoginIP:      "192.168.1.1",
		LastLoginCountry: "US",
	}

	assessment, err := engine.Assess(context.Background(), loginCtx)
	assert.NoError(t, err)
	assert.Equal(t, domain_auth.RiskScore(0.8), assessment.Score)
	assert.Contains(t, assessment.Factors, domain_auth.RiskFactorNewDevice)
	assert.Contains(t, assessment.Factors, domain_auth.RiskFactorGeoVelocity)
	assert.Contains(t, assessment.Factors, domain_auth.RiskFactorUnusualTime)
	assert.Equal(t, domain_auth.RiskDecisionDeny, assessment.Decision)
}
