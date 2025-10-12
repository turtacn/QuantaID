package postgresql

import (
	"context"

	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// PostgresIdentityRepository provides a GORM-based implementation of the identity-related repositories.
type PostgresIdentityRepository struct {
	db *gorm.DB
}

// NewPostgresIdentityRepository creates a new PostgreSQL identity repository.
func NewPostgresIdentityRepository(db *gorm.DB) *PostgresIdentityRepository {
	return &PostgresIdentityRepository{db: db}
}

// --- UserRepository Implementation ---

// CreateUser adds a new user to the database.
func (r *PostgresIdentityRepository) CreateUser(ctx context.Context, user *types.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetUserByID retrieves a user by their ID from the database.
func (r *PostgresIdentityRepository) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	var user types.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	return &user, err
}

// GetUserByUsername searches for a user by their username in the database.
func (r *PostgresIdentityRepository) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	var user types.User
	err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error
	return &user, err
}

// GetUserByEmail searches for a user by their email in the database.
func (r *PostgresIdentityRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	return &user, err
}

// UpdateUser updates an existing user in the database.
func (r *PostgresIdentityRepository) UpdateUser(ctx context.Context, user *types.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// DeleteUser removes a user from the database.
func (r *PostgresIdentityRepository) DeleteUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.User{}, "id = ?", id).Error
}

// ListUsers returns a paginated list of all users from the database.
func (r *PostgresIdentityRepository) ListUsers(ctx context.Context, pq identity.PaginationQuery) ([]*types.User, error) {
	var users []*types.User
	err := r.db.WithContext(ctx).Offset(pq.Offset).Limit(pq.PageSize).Find(&users).Error
	return users, err
}

// FindUsersByAttribute is not yet implemented for the PostgreSQL repository.
func (r *PostgresIdentityRepository) FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error) {
	// This is a complex query to implement with GORM and JSONB, so it's deferred.
	return nil, types.ErrNotImplemented
}

// --- GroupRepository Implementation ---

// CreateGroup adds a new group to the database.
func (r *PostgresIdentityRepository) CreateGroup(ctx context.Context, group *types.UserGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroupByID retrieves a group by its ID from the database.
func (r *PostgresIdentityRepository) GetGroupByID(ctx context.Context, id string) (*types.UserGroup, error) {
	var group types.UserGroup
	err := r.db.WithContext(ctx).First(&group, "id = ?", id).Error
	return &group, err
}

// GetGroupByName searches for a group by its name in the database.
func (r *PostgresIdentityRepository) GetGroupByName(ctx context.Context, name string) (*types.UserGroup, error) {
	var group types.UserGroup
	err := r.db.WithContext(ctx).First(&group, "name = ?", name).Error
	return &group, err
}

// UpdateGroup updates an existing group in the database.
func (r *PostgresIdentityRepository) UpdateGroup(ctx context.Context, group *types.UserGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteGroup removes a group from the database.
func (r *PostgresIdentityRepository) DeleteGroup(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.UserGroup{}, "id = ?", id).Error
}

// ListGroups returns a paginated list of all groups from the database.
func (r *PostgresIdentityRepository) ListGroups(ctx context.Context, pq identity.PaginationQuery) ([]*types.UserGroup, error) {
	var groups []*types.UserGroup
	err := r.db.WithContext(ctx).Offset(pq.Offset).Limit(pq.PageSize).Find(&groups).Error
	return groups, err
}

// AddUserToGroup creates a membership link between a user and a group in the database.
func (r *PostgresIdentityRepository) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	user := types.User{ID: userID}
	group := types.UserGroup{ID: groupID}
	return r.db.WithContext(ctx).Model(&group).Association("Users").Append(&user)
}

// RemoveUserFromGroup removes a membership link between a user and a group in the database.
func (r *PostgresIdentityRepository) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	user := types.User{ID: userID}
	group := types.UserGroup{ID: groupID}
	return r.db.WithContext(ctx).Model(&group).Association("Users").Delete(&user)
}

// GetUserGroups retrieves all groups a user is a member of from the database.
func (r *PostgresIdentityRepository) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	var user types.User
	err := r.db.WithContext(ctx).Preload("Groups").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	groups := make([]*types.UserGroup, len(user.Groups))
	for i := range user.Groups {
		groups[i] = &user.Groups[i]
	}
	return groups, nil
}