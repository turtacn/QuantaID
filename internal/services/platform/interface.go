package platform

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

type Service interface {
	ListApps(ctx context.Context) ([]*types.DevCenterAppDTO, error)
	CreateApp(ctx context.Context, req types.CreateAppRequest) (*types.DevCenterAppDTO, error)
	ListConnectors(ctx context.Context) ([]*types.DevCenterConnectorDTO, error)
	EnableConnector(ctx context.Context, id string) error
	Diagnostics(ctx context.Context) (*types.DiagnosticsDTO, error)
}
