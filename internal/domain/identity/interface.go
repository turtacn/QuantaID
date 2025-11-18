package identity

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// IService defines the public contract for the identity domain service.
// It exposes methods for managing users and their group memberships, abstracting
// the underlying implementation details. This interface is the primary entry point
// for other services to interact with user identity data.
type IService interface {
	// CreateUser creates a new user with the given details.
	CreateUser(ctx context.Context, username, email, password string) (*types.User, error)
	// GetUser retrieves a user by their unique ID.
	GetUser(ctx context.Context, userID string) (*types.User, error)
	// GetUserByID retrieves a user by their unique ID.
	GetUserByID(ctx context.Context, userID string) (*types.User, error)
	// GetUserByUsername retrieves a user by their unique username.
	GetUserByUsername(ctx context.Context, username string) (*types.User, error)
	// GetUserGroups retrieves all groups a user is a member of.
	GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error)
	// AddUserToGroup adds a user to a specified group.
	AddUserToGroup(ctx context.Context, userID, groupID string) error
	// ChangeUserStatus updates the status of a user's account (e.g., active, locked).
	ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error
}
