package postgresql

import (
	"context"
	"errors"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		return nil, err
	}
	return &user, err
}

func (r *PostgresIdentityRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrUserNotFound
		}
		return nil, err
	}
	return &user, err
}

func (r *PostgresIdentityRepository) UpdateUser(ctx context.Context, user *types.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *PostgresIdentityRepository) DeleteUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.User{}, "id = ?", id).Error
}

func (r *PostgresIdentityRepository) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) {
	var users []*types.User
	var total int64

	query := r.db.WithContext(ctx).Model(&types.User{})

	if filter.Query != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}

	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortOrder == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	}

	offset := (filter.Page - 1) * filter.PageSize
	err = query.Offset(offset).Limit(filter.PageSize).Find(&users).Error

	return users, int(total), err
}

func (r *PostgresIdentityRepository) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error {
	return r.db.WithContext(ctx).Model(&types.User{}).Where("id = ?", userID).Update("status", newStatus).Error
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

func (r *PostgresIdentityRepository) CreateBatch(ctx context.Context, users []*types.User) error {
	return r.db.WithContext(ctx).Create(&users).Error
}

func (r *PostgresIdentityRepository) UpdateBatch(ctx context.Context, users []*types.User) error {
	return r.db.WithContext(ctx).Save(&users).Error
}

func (r *PostgresIdentityRepository) DeleteBatch(ctx context.Context, userIDs []string) error {
	return r.db.WithContext(ctx).Delete(&types.User{}, "id IN ?", userIDs).Error
}

func (r *PostgresIdentityRepository) FindUsersBySource(ctx context.Context, sourceID string) ([]*types.User, error) {
	// This is a placeholder implementation. In a real scenario, you'd have a way
	// to associate users with a source. For now, we return all users.
	var users []*types.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}
