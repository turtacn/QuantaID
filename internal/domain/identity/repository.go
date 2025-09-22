package identity

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// PaginationQuery defines parameters for paginated queries.
type PaginationQuery struct {
	PageSize int
	Offset   int
}

// UserRepository defines the interface for data access operations related to users.
type UserRepository interface {
	CreateUser(ctx context.Context, user *types.User) error
	GetUserByID(ctx context.Context, id string) (*types.User, error)
	GetUserByUsername(ctx context.Context, username string) (*types.User, error)
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	UpdateUser(ctx context.Context, user *types.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, pq PaginationQuery) ([]*types.User, error)
	FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error)
}

// GroupRepository defines the interface for data access operations related to user groups.
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

// TransactionalRepository defines an interface for managing transactions.
type TransactionalRepository interface {
	BeginTransaction(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

//Personal.AI order the ending
