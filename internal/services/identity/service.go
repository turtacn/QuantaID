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
func (s *ApplicationService) CreateUser(ctx context.Context, req CreateUserRequest) (*types.User, *types.Error) {
	user, err := s.identityDomain.CreateUser(ctx, req.Username, req.Email, req.Password)
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
func (s *ApplicationService) GetUserByID(ctx context.Context, userID string) (*types.User, *types.Error) {
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
func (s *ApplicationService) AddUserToGroup(ctx context.Context, req AddUserToGroupRequest) *types.Error {
	err := s.identityDomain.AddUserToGroup(ctx, req.UserID, req.GroupID)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return appErr
		}
		return types.ErrInternal.WithCause(err)
	}
	return nil
}

