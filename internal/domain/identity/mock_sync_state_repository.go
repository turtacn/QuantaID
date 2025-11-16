package identity

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
	"time"
)

type MockSyncStateRepository struct {
	mock.Mock
}

func (m *MockSyncStateRepository) GetLastSyncState(ctx context.Context, sourceID string) (*types.SyncState, error) {
	args := m.Called(ctx, sourceID)
	return args.Get(0).(*types.SyncState), args.Error(1)
}

func (m *MockSyncStateRepository) UpdateProgress(ctx context.Context, sourceID string, processed int) error {
	args := m.Called(ctx, sourceID, processed)
	return args.Error(0)
}

func (m *MockSyncStateRepository) MarkCompleted(ctx context.Context, sourceID string, completedAt time.Time) error {
	args := m.Called(ctx, sourceID, completedAt)
	return args.Error(0)
}
