package sync

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

type LDAPConnector interface {
	SyncUsers(ctx context.Context) ([]*types.User, error)
	SearchUsers(ctx context.Context, filter string) ([]*types.User, error)
}

type AuditService interface {
	RecordAdminAction(ctx context.Context, userID, ip, resource, action, traceID string, details map[string]any)
}
