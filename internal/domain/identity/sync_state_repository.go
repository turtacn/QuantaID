package identity

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"time"
)

type SyncStateRepository interface {
	GetLastSyncState(ctx context.Context, sourceID string) (*types.SyncState, error)
	UpdateProgress(ctx context.Context, sourceID string, processed int) error
	MarkCompleted(ctx context.Context, sourceID string, completedAt time.Time) error
}
