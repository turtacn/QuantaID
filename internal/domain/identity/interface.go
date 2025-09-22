package identity

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// IService defines the interface for the identity domain service.
type IService interface {
	CreateUser(ctx context.Context, username, email, password string) (*types.User, error)
	GetUser(ctx context.Context, userID string) (*types.User, error)
	GetUserByUsername(ctx context.Context, username string) (*types.User, error)
	GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error)
	AddUserToGroup(ctx context.Context, userID, groupID string) error
	ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error
}

//Personal.AI order the ending
