package postgresql

import (
	"context"
	"errors"

	"github.com/turtacn/QuantaID/internal/domain/policy"
	"gorm.io/gorm"
)

type rbacRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) policy.RBACRepository {
	return &rbacRepository{db: db}
}

// Role management
func (r *rbacRepository) CreateRole(ctx context.Context, role *policy.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *rbacRepository) GetRoleByCode(ctx context.Context, code string) (*policy.Role, error) {
	var role policy.Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Consider returning a domain-specific error
		}
		return nil, err
	}
	return &role, nil
}

func (r *rbacRepository) UpdateRole(ctx context.Context, role *policy.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *rbacRepository) DeleteRole(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).Delete(&policy.Role{}, roleID).Error
}

func (r *rbacRepository) ListRoles(ctx context.Context) ([]*policy.Role, error) {
	var roles []*policy.Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

// Permission management
func (r *rbacRepository) CreatePermission(ctx context.Context, permission *policy.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *rbacRepository) GetPermission(ctx context.Context, resource, action string) (*policy.Permission, error) {
	var perm policy.Permission
	err := r.db.WithContext(ctx).Where("resource = ? AND action = ?", resource, action).First(&perm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Consider returning a domain-specific error
		}
		return nil, err
	}
	return &perm, nil
}

func (r *rbacRepository) ListPermissions(ctx context.Context) ([]*policy.Permission, error) {
	var permissions []*policy.Permission
	err := r.db.WithContext(ctx).Find(&permissions).Error
	return permissions, err
}

// Assignment management
func (r *rbacRepository) AddPermissionToRole(ctx context.Context, roleID, permissionID uint) error {
	role := policy.Role{ID: roleID}
	permission := policy.Permission{ID: permissionID}
	return r.db.WithContext(ctx).Model(&role).Association("Permissions").Append(&permission)
}

func (r *rbacRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint) error {
	role := policy.Role{ID: roleID}
	permission := policy.Permission{ID: permissionID}
	return r.db.WithContext(ctx).Model(&role).Association("Permissions").Delete(&permission)
}

func (r *rbacRepository) AssignRoleToUser(ctx context.Context, userID string, roleID uint) error {
	userRole := policy.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.WithContext(ctx).Create(&userRole).Error
}

func (r *rbacRepository) UnassignRoleFromUser(ctx context.Context, userID string, roleID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&policy.UserRole{}).Error
}

// Query methods
func (r *rbacRepository) GetRolesForUser(ctx context.Context, userID string) ([]*policy.Role, error) {
	var roles []*policy.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Preload("Permissions").
		Find(&roles).Error
	return roles, err
}

func (r *rbacRepository) GetPermissionsForUser(ctx context.Context, userID string) ([]*policy.Permission, error) {
	var permissions []*policy.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct().
		Find(&permissions).Error
	return permissions, err
}

func (r *rbacRepository) GetPermissionsForRole(ctx context.Context, roleID uint) ([]*policy.Permission, error) {
	var role policy.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, roleID).Error
	if err != nil {
		return nil, err
	}
	return role.Permissions, nil
}
