package e2e

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/auth"
)

type MockServiceRiskEngine struct {
	mock.Mock
}

func (m *MockServiceRiskEngine) Assess(ctx context.Context, loginCtx auth.LoginContext) (*auth.RiskAssessment, error) {
	args := m.Called(ctx, loginCtx)
	return args.Get(0).(*auth.RiskAssessment), args.Error(1)
}
