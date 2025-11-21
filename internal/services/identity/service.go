package identity

import (
	"context"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// ApplicationService provides application-level use cases for identity management.
type ApplicationService struct {
	identityDomain identity.IService
	logger         utils.Logger
}

// NewApplicationService creates a new identity application service.
func NewApplicationService(identityDomain identity.IService, logger utils.Logger) *ApplicationService {
	return &ApplicationService{
		identityDomain: identityDomain,
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

