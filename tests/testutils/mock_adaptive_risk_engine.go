package testutils

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/auth/adaptive"
)

type MockAdaptiveRiskEngine struct {
	mock.Mock
}

func (m *MockAdaptiveRiskEngine) Evaluate(ctx context.Context, event *adaptive.AuthEvent) (*adaptive.RiskScore, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adaptive.RiskScore), args.Error(1)
}
