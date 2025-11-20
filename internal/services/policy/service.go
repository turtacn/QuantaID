package policy

import (
	"context"

	"github.com/turtacn/QuantaID/internal/domain/policy"
)

type service struct {
	repo policy.RBACRepository
}

func NewService(repo policy.RBACRepository) PolicyService {
	return &service{repo: repo}
}

func (s *service) CreateRole(ctx context.Context, role *policy.Role) error {
	// Add any validation or business logic here
	return s.repo.CreateRole(ctx, role)
}

func (s *service) ListRoles(ctx context.Context) ([]*policy.Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *service) UpdateRole(ctx context.Context, role *policy.Role) error {
	return s.repo.UpdateRole(ctx, role)
}

func (s *service) DeleteRole(ctx context.Context, roleID uint) error {
	return s.repo.DeleteRole(ctx, roleID)
}

func (s *service) CreatePermission(ctx context.Context, permission *policy.Permission) error {
	return s.repo.CreatePermission(ctx, permission)
}

func (s *service) ListPermissions(ctx context.Context) ([]*policy.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *service) AddPermissionToRole(ctx context.Context, roleID, permissionID uint) error {
	return s.repo.AddPermissionToRole(ctx, roleID, permissionID)
}

func (s *service) AssignRoleToUser(ctx context.Context, userID string, roleID uint) error {
	return s.repo.AssignRoleToUser(ctx, userID, roleID)
}

func (s *service) UnassignRoleFromUser(ctx context.Context, userID string, roleID uint) error {
	return s.repo.UnassignRoleFromUser(ctx, userID, roleID)
}
