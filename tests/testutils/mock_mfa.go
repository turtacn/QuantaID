package testutils

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/auth/mfa"
)

type MockMFAManager struct {
	mock.Mock
}

func (m *MockMFAManager) EnrollFactor(ctx context.Context, userID string, mfaType mfa.MFAType, params mfa.EnrollParams) (*mfa.EnrollResult, error) {
	args := m.Called(ctx, userID, mfaType, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mfa.EnrollResult), args.Error(1)
}

func (m *MockMFAManager) VerifyFactor(ctx context.Context, userID string, mfaType mfa.MFAType, credential string) (bool, error) {
	args := m.Called(ctx, userID, mfaType, credential)
	return args.Bool(0), args.Error(1)
}

func (m *MockMFAManager) ActivateFactor(ctx context.Context, userID string, mfaType mfa.MFAType, credential string) error {
	args := m.Called(ctx, userID, mfaType, credential)
	return args.Error(0)
}

func (m *MockMFAManager) GetRequiredFactors(ctx context.Context, userID string) ([]mfa.MFAType, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]mfa.MFAType), args.Error(1)
}
