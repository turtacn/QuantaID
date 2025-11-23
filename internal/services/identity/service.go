package identity

import (
	"context"

	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/services/audit"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// ApplicationService provides application-level use cases for identity management.
type ApplicationService struct {
	identityDomain identity.IService
	auditService   *audit.Service
	logger         utils.Logger
}

// NewApplicationService creates a new identity application service.
func NewApplicationService(identityDomain identity.IService, auditService *audit.Service, logger utils.Logger) *ApplicationService {
	return &ApplicationService{
		identityDomain: identityDomain,
		auditService:   auditService,
		logger:         logger,
	}
}

// CreateUserRequest defines the request structure for creating a new user.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUser handles the user creation use case.
func (s *ApplicationService) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) {
	user, err := s.identityDomain.CreateUser(ctx, username, email, password)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}

	if s.auditService != nil {
		s.auditService.RecordUserCreated(ctx, user, "unknown", "") // IP and TraceID might be needed from Context
	}

	user.Password = "" // Never expose password hash
	return user, nil
}

// GetUserByID handles the use case of retrieving a single user.
func (s *ApplicationService) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	user, err := s.identityDomain.GetUser(ctx, userID)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}

	user.Password = ""
	return user, nil
}

// AddUserToGroupRequest defines the request structure for adding a user to a group.
type AddUserToGroupRequest struct {
	UserID  string `json:"userId"`
	GroupID string `json:"groupId"`
}

// AddUserToGroup handles the use case of adding a user to a group.
func (s *ApplicationService) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	err := s.identityDomain.AddUserToGroup(ctx, userID, groupID)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return appErr
		}
		return types.ErrInternal.WithCause(err)
	}
	return nil
}

// Implement missing interface methods by delegating to identityDomain

func (s *ApplicationService) UpdateUser(ctx context.Context, user *types.User) error {
	return s.identityDomain.UpdateUser(ctx, user)
}

func (s *ApplicationService) DeleteUser(ctx context.Context, userID string) error {
	return s.identityDomain.DeleteUser(ctx, userID)
}

func (s *ApplicationService) GetUserByExternalID(ctx context.Context, externalID string) (*types.User, error) {
	return s.identityDomain.GetUserByExternalID(ctx, externalID)
}

func (s *ApplicationService) CreateGroup(ctx context.Context, group *types.UserGroup) error {
	return s.identityDomain.CreateGroup(ctx, group)
}

func (s *ApplicationService) GetGroup(ctx context.Context, groupID string) (*types.UserGroup, error) {
	return s.identityDomain.GetGroup(ctx, groupID)
}

func (s *ApplicationService) UpdateGroup(ctx context.Context, group *types.UserGroup) error {
	return s.identityDomain.UpdateGroup(ctx, group)
}

func (s *ApplicationService) DeleteGroup(ctx context.Context, groupID string) error {
	return s.identityDomain.DeleteGroup(ctx, groupID)
}

func (s *ApplicationService) ListGroups(ctx context.Context, offset, limit int) ([]*types.UserGroup, error) {
	return s.identityDomain.ListGroups(ctx, offset, limit)
}

func (s *ApplicationService) GetUser(ctx context.Context, userID string) (*types.User, error) {
	return s.identityDomain.GetUser(ctx, userID)
}

func (s *ApplicationService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	return s.identityDomain.GetUserByUsername(ctx, username)
}

func (s *ApplicationService) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	return s.identityDomain.GetUserGroups(ctx, userID)
}

func (s *ApplicationService) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) {
	return s.identityDomain.ListUsers(ctx, filter)
}

func (s *ApplicationService) GetUserRepo() identity.UserRepository {
	return s.identityDomain.GetUserRepo()
}

// ChangeUserStatusRequest defines the request structure for changing a user's status.
type ChangeUserStatusRequest struct {
	UserID    string           `json:"userId"`
	NewStatus types.UserStatus `json:"newStatus"`
}

// ChangeUserStatus handles the use case of changing a user's status.
func (s *ApplicationService) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error {
	err := s.identityDomain.ChangeUserStatus(ctx, userID, newStatus)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return appErr
		}
		return types.ErrInternal.WithCause(err)
	}
	return nil
}

