package e2e

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/auth"
)

type MockRiskEngine struct {
	mock.Mock
}

func (m *MockRiskEngine) Assess(ctx context.Context, loginCtx auth.LoginContext) (*auth.RiskAssessment, error) {
	args := m.Called(ctx, loginCtx)
	return args.Get(0).(*auth.RiskAssessment), args.Error(1)
}
