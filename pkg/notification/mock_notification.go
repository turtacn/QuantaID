package notification

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Send(ctx context.Context, msg Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockNotifier) Type() string {
	args := m.Called()
	return args.String(0)
}

type MockNotificationManager struct {
	mock.Mock
}

func (m *MockNotificationManager) GetNotifier(method string) (Notifier, error) {
	args := m.Called(method)
	return args.Get(0).(Notifier), args.Error(1)
}
