package identity

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// PaginationQuery defines parameters for paginated list queries.
type PaginationQuery struct {
	// PageSize specifies the maximum number of items to return.
	PageSize int
	// Offset specifies the number of items to skip before starting to collect the result set.
	Offset int
}

// UserRepository defines the persistence interface for user-related data.
// It outlines the CRUD (Create, Read, Update, Delete) and query operations for user entities.
type UserRepository interface {
	// CreateUser saves a new user to the database.
	CreateUser(ctx context.Context, user *types.User) error
	// GetUserByID retrieves a user by their unique ID.
	GetUserByID(ctx context.Context, id string) (*types.User, error)
	// GetUserByUsername retrieves a user by their unique username.
	GetUserByUsername(ctx context.Context, username string) (*types.User, error)
	// GetUserByEmail retrieves a user by their unique email address.
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	// UpdateUser modifies an existing user's data.
	UpdateUser(ctx context.Context, user *types.User) error
	// DeleteUser removes a user from the database by their ID.
	DeleteUser(ctx context.Context, id string) error
	// ListUsers retrieves a paginated list of all users.
	ListUsers(ctx context.Context, pq PaginationQuery) ([]*types.User, error)
	// FindUsersByAttribute finds users that have a specific attribute with a given value.
	FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error)
}

// GroupRepository defines the persistence interface for user group-related data.
// It outlines the CRUD and membership management operations for group entities.
type GroupRepository interface {
	// CreateGroup saves a new group to the database.
	CreateGroup(ctx context.Context, group *types.UserGroup) error
	// GetGroupByID retrieves a group by its unique ID.
	GetGroupByID(ctx context.Context, id string) (*types.UserGroup, error)
	// GetGroupByName retrieves a group by its unique name.
	GetGroupByName(ctx context.Context, name string) (*types.UserGroup, error)
	// UpdateGroup modifies an existing group's data.
	UpdateGroup(ctx context.Context, group *types.UserGroup) error
	// DeleteGroup removes a group from the database by its ID.
	DeleteGroup(ctx context.Context, id string) error
	// ListGroups retrieves a paginated list of all groups.
	ListGroups(ctx context.Context, pq PaginationQuery) ([]*types.UserGroup, error)
	// AddUserToGroup creates a membership link between a user and a group.
	AddUserToGroup(ctx context.Context, userID, groupID string) error
	// RemoveUserFromGroup removes a membership link between a user and a group.
	RemoveUserFromGroup(ctx context.Context, userID, groupID string) error
	// GetUserGroups retrieves all groups that a specific user is a member of.
	GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error)
}

// TransactionalRepository defines an interface for managing database transactions.
// This allows service-layer operations to be executed atomically.
type TransactionalRepository interface {
	// BeginTransaction starts a new transaction and returns a context that carries it.
	BeginTransaction(ctx context.Context) (context.Context, error)
	// Commit finalizes the transaction carried by the context.
	Commit(ctx context.Context) error
	// Rollback cancels the transaction carried by the context.
	Rollback(ctx context.Context) error
}
