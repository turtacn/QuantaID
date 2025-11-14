package application

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

type Service interface {
	GetApplicationByID(ctx context.Context, id string) (*types.Application, *types.Error)
	CreateApplication(ctx context.Context, req CreateApplicationRequest) (*types.Application, *types.Error)
	ListApplications(ctx context.Context) ([]*types.Application, *types.Error)
}
