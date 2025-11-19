package testutils

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/auth/mfa"
	"github.com/turtacn/QuantaID/pkg/types"
)

type MockMFAProvider struct {
	mock.Mock
}

func (m *MockMFAProvider) Enroll(ctx context.Context, user *types.User) (*types.MFAEnrollment, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MFAEnrollment), args.Error(1)
}

func (m *MockMFAProvider) Challenge(ctx context.Context, user *types.User) (*types.MFAChallenge, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MFAChallenge), args.Error(1)
}

func (m *MockMFAProvider) Verify(ctx context.Context, user *types.User, code string) (bool, error) {
	args := m.Called(ctx, user, code)
	return args.Bool(0), args.Error(1)
}

func (m *MockMFAProvider) ListMethods(ctx context.Context, user *types.User) ([]*types.MFAMethod, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.MFAMethod), args.Error(1)
}

func (m *MockMFAProvider) GetStrength() mfa.StrengthLevel {
	args := m.Called()
	return args.Get(0).(mfa.StrengthLevel)
}
