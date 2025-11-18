package utils

import (
	"context"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockLogger is a mock implementation of the Logger interface for testing.
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) With(fields ...zap.Field) Logger {
	args := m.Called(fields)
	return args.Get(0).(Logger)
}
