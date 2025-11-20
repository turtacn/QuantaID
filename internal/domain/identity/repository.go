package identity

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

type PaginationQuery struct {
	Offset   int
	PageSize int
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *types.User) error
	GetUserByID(ctx context.Context, id string) (*types.User, error)
	GetUserByUsername(ctx context.Context, username string) (*types.User, error)
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	UpdateUser(ctx context.Context, user *types.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error)
	ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error
	FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error)
	UpsertBatch(ctx context.Context, users []*types.User) error
	CreateBatch(ctx context.Context, users []*types.User) error
	UpdateBatch(ctx context.Context, users []*types.User) error
	DeleteBatch(ctx context.Context, userIDs []string) error
	FindUsersBySource(ctx context.Context, sourceID string) ([]*types.User, error)
}

type GroupRepository interface {
	CreateGroup(ctx context.Context, group *types.UserGroup) error
	GetGroupByID(ctx context.Context, id string) (*types.UserGroup, error)
	GetGroupByName(ctx context.Context, name string) (*types.UserGroup, error)
	UpdateGroup(ctx context.Context, group *types.UserGroup) error
	DeleteGroup(ctx context.Context, id string) error
	ListGroups(ctx context.Context, pq PaginationQuery) ([]*types.UserGroup, error)
	AddUserToGroup(ctx context.Context, userID, groupID string) error
	RemoveUserFromGroup(ctx context.Context, userID, groupID string) error
	GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error)
}

type IdentityRepository interface {
	UserRepository
	GroupRepository
}
