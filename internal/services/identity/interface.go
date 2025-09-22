package identity

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// IApplicationService defines the interface for the identity application service.
type IApplicationService interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (*types.User, *types.Error)
	GetUserByID(ctx context.Context, userID string) (*types.User, *types.Error)
	AddUserToGroup(ctx context.Context, req AddUserToGroupRequest) *types.Error
}

//Personal.AI order the ending
