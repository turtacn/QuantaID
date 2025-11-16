package postgresql

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
)

type PostgresIdentityRepository struct {
	db *gorm.DB
}

func NewPostgresIdentityRepository(db *gorm.DB) *PostgresIdentityRepository {
	return &PostgresIdentityRepository{db: db}
}

func (r *PostgresIdentityRepository) CreateUser(ctx context.Context, user *types.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *PostgresIdentityRepository) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	var user types.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	return &user, err
}

func (r *PostgresIdentityRepository) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	var user types.User
	err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error
	return &user, err
}

func (r *PostgresIdentityRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	return &user, err
}

func (r *PostgresIdentityRepository) UpdateUser(ctx context.Context, user *types.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *PostgresIdentityRepository) DeleteUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.User{}, "id = ?", id).Error
}

func (r *PostgresIdentityRepository) ListUsers(ctx context.Context, pq identity.PaginationQuery) ([]*types.User, error) {
	var users []*types.User
	err := r.db.WithContext(ctx).Offset(pq.Offset).Limit(pq.PageSize).Find(&users).Error
	return users, err
}

func (r *PostgresIdentityRepository) FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error) {
	var users []*types.User
	err := r.db.WithContext(ctx).Where("attributes ->> ? = ?", attribute, value).Find(&users).Error
	return users, err
}

func (r *PostgresIdentityRepository) UpsertBatch(ctx context.Context, users []*types.User) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoUpdates: clause.AssignmentColumns([]string{"username", "attributes", "status"}),
	}).Create(&users).Error
}

func (r *PostgresIdentityRepository) CreateGroup(ctx context.Context, group *types.UserGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *PostgresIdentityRepository) GetGroupByID(ctx context.Context, id string) (*types.UserGroup, error) {
	var group types.UserGroup
	err := r.db.WithContext(ctx).First(&group, "id = ?", id).Error
	return &group, err
}

func (r *PostgresIdentityRepository) GetGroupByName(ctx context.Context, name string) (*types.UserGroup, error) {
	var group types.UserGroup
	err := r.db.WithContext(ctx).First(&group, "name = ?", name).Error
	return &group, err
}

func (r *PostgresIdentityRepository) UpdateGroup(ctx context.Context, group *types.UserGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *PostgresIdentityRepository) DeleteGroup(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.UserGroup{}, "id = ?", id).Error
}

func (r *PostgresIdentityRepository) ListGroups(ctx context.Context, pq identity.PaginationQuery) ([]*types.UserGroup, error) {
	var groups []*types.UserGroup
	err := r.db.WithContext(ctx).Offset(pq.Offset).Limit(pq.PageSize).Find(&groups).Error
	return groups, err
}

func (r *PostgresIdentityRepository) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	return r.db.WithContext(ctx).Model(&types.User{ID: userID}).Association("Groups").Append(&types.UserGroup{ID: groupID})
}

func (r *PostgresIdentityRepository) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	return r.db.WithContext(ctx).Model(&types.User{ID: userID}).Association("Groups").Delete(&types.UserGroup{ID: groupID})
}

func (r *PostgresIdentityRepository) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	var user types.User
	err := r.db.WithContext(ctx).Preload("Groups").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	var groups []*types.UserGroup
	for i := range user.Groups {
		groups = append(groups, &user.Groups[i])
	}
	return groups, nil
}
