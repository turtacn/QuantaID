package testutils

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/auth"
)

type MockAdaptiveRiskEngine struct {
	mock.Mock
}

func (m *MockAdaptiveRiskEngine) Evaluate(ctx context.Context, event auth.AuthContext) (auth.RiskScore, auth.RiskLevel, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return 0, "", args.Error(2)
	}
	return args.Get(0).(auth.RiskScore), args.Get(1).(auth.RiskLevel), args.Error(2)
}
