package identity

import (
	"context"
	"errors"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// service implements the IService interface.
// It provides the core business logic for managing user identities and groups.
type service struct {
	userRepo  UserRepository
	groupRepo GroupRepository
	crypto    *utils.CryptoManager
	logger    utils.Logger
}

// NewService creates a new identity service instance.
// It combines the user and group repositories with crypto and logging utilities
// to provide a complete service for identity management.
//
// Parameters:
//   - userRepo: The repository for user data access.
//   - groupRepo: The repository for group data access.
//   - crypto: The utility for cryptographic operations like password hashing.
//   - logger: The logger for logging service-level messages.
//
// Returns:
//   A new instance of the identity service that implements the IService interface.
func NewService(userRepo UserRepository, groupRepo GroupRepository, crypto *utils.CryptoManager, logger utils.Logger) IService {
	return &service{
		userRepo:  userRepo,
		groupRepo: groupRepo,
		crypto:    crypto,
		logger:    logger,
	}
}

// CreateUser handles the business logic for creating a new user.
// It validates input, checks for existing users with the same username or email,
// hashes the password, and persists the new user to the repository.
//
// Parameters:
//   - ctx: The context for the request.
//   - username: The desired username for the new user.
//   - email: The desired email for the new user.
//   - password: The plain-text password for the new user.
//
// Returns:
//   The newly created user object, or an error if the creation fails.
func (s *service) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, types.ErrValidation.WithDetails(map[string]string{"field": "username/email/password", "error": "cannot be empty"})
	}

	_, err := s.userRepo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, types.ErrConflict.WithDetails(map[string]string{"field": "username", "value": username})
	} else if !errors.Is(err, types.ErrUserNotFound) {
		s.logger.Error(ctx, "failed to check for existing user by username", zap.Error(err))
		return nil, types.ErrInternal.WithCause(err)
	}

	_, err = s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, types.ErrConflict.WithDetails(map[string]string{"field": "email", "value": email})
	} else if !errors.Is(err, types.ErrUserNotFound) {
		s.logger.Error(ctx, "failed to check for existing user by email", zap.Error(err))
		return nil, types.ErrInternal.WithCause(err)
	}

	hashedPassword, err := s.crypto.HashPassword(password)
	if err != nil {
		s.logger.Error(ctx, "Failed to hash password", zap.Error(err))
		return nil, types.ErrInternal.WithCause(err)
	}

	user := &types.User{
		ID:       s.crypto.GenerateUUID(),
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Status:   types.UserStatusActive,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to create user in repository", zap.Error(err))
		return nil, types.ErrInternal.WithCause(err)
	}

	s.logger.Info(ctx, "User created successfully", zap.String("userID", user.ID), zap.String("username", user.Username))
	return user, nil
}

// GetUser retrieves a user by their unique ID.
//
// Parameters:
//   - ctx: The context for the request.
//   - userID: The ID of the user to retrieve.
//
// Returns:
//   The user object if found, or an error.
func (s *service) GetUser(ctx context.Context, userID string) (*types.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"userID": userID})
	}
	return user, nil
}

// ListUsers retrieves a paginated list of users, with optional filtering.
func (s *service) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) {
	s.logger.Info(ctx, "Listing users with filter", zap.Any("filter", filter))
	users, total, err := s.userRepo.ListUsers(ctx, filter)
	if err != nil {
		s.logger.Error(ctx, "Failed to list users", zap.Error(err))
	}
	return users, total, err
}

// GetUserByID retrieves a user by their unique ID.
//
// Parameters:
//   - ctx: The context for the request.
//   - userID: The ID of the user to retrieve.
//
// Returns:
//   The user object if found, or an error.
func (s *service) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"userID": userID})
	}
	return user, nil
}

// GetUserByUsername retrieves a user by their unique username.
//
// Parameters:
//   - ctx: The context for the request.
//   - username: The username of the user to retrieve.
//
// Returns:
//   The user object if found, or an error.
func (s *service) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"username": username})
	}
	return user, nil
}

// GetUserGroups retrieves all groups a user is a member of.
//
// Parameters:
//   - ctx: The context for the request.
//   - userID: The ID of the user whose groups are to be retrieved.
//
// Returns:
//   A slice of user groups, or an error.
func (s *service) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	groups, err := s.groupRepo.GetUserGroups(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get user groups from repository", zap.Error(err), zap.String("userID", userID))
		return nil, types.ErrInternal.WithCause(err)
	}
	return groups, nil
}

// AddUserToGroup creates a membership link between a user and a group.
// It first validates that both the user and the group exist before creating the link.
//
// Parameters:
//   - ctx: The context for the request.
//   - userID: The ID of the user to add to the group.
//   - groupID: The ID of the group to which the user will be added.
//
// Returns:
//   An error if the user or group is not found, or if the operation fails.
func (s *service) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"user_id": userID})
	}
	_, err = s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		return types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"group_id": groupID})
	}

	err = s.groupRepo.AddUserToGroup(ctx, userID, groupID)
	if err != nil {
		s.logger.Error(ctx, "Failed to add user to group", zap.Error(err), zap.String("userID", userID), zap.String("groupID", groupID))
		return types.ErrInternal.WithCause(err)
	}

	s.logger.Info(ctx, "User added to group", zap.String("userID", userID), zap.String("groupID", groupID))
	return nil
}

// ChangeUserStatus updates the status of a user's account.
//
// Parameters:
//   - ctx: The context for the request.
//   - userID: The ID of the user whose status is to be changed.
//   - newStatus: The new status for the user account.
//
// Returns:
//   An error if the user is not found or if the update fails.
func (s *service) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"userID": userID})
	}

	user.Status = newStatus
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to update user status", zap.Error(err), zap.String("userID", userID))
		return types.ErrInternal.WithCause(err)
	}

	s.logger.Info(ctx, "User status changed", zap.String("userID", userID), zap.String("newStatus", string(newStatus)))
	return nil
}

// UpdateUser updates an existing user's details.
//
// Parameters:
//   - ctx: The context for the request.
//   - user: The user object with updated details.
//
// Returns:
//   An error if the update fails.
func (s *service) UpdateUser(ctx context.Context, user *types.User) error {
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to update user", zap.Error(err), zap.String("userID", user.ID))
		return types.ErrInternal.WithCause(err)
	}
	s.logger.Info(ctx, "User updated", zap.String("userID", user.ID))
	return nil
}

// DeleteUser deletes a user by their unique ID.
//
// Parameters:
//   - ctx: The context for the request.
//   - userID: The ID of the user to delete.
//
// Returns:
//   An error if the delete fails.
func (s *service) DeleteUser(ctx context.Context, userID string) error {
	if err := s.userRepo.DeleteUser(ctx, userID); err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"userID": userID})
		}
		s.logger.Error(ctx, "Failed to delete user", zap.Error(err), zap.String("userID", userID))
		return types.ErrInternal.WithCause(err)
	}
	s.logger.Info(ctx, "User deleted", zap.String("userID", userID))
	return nil
}

// CreateGroup creates a new group.
func (s *service) CreateGroup(ctx context.Context, group *types.UserGroup) error {
	if err := s.groupRepo.CreateGroup(ctx, group); err != nil {
		s.logger.Error(ctx, "Failed to create group", zap.Error(err))
		return types.ErrInternal.WithCause(err)
	}
	s.logger.Info(ctx, "Group created", zap.String("groupID", group.ID))
	return nil
}

// GetGroup retrieves a group by ID.
func (s *service) GetGroup(ctx context.Context, groupID string) (*types.UserGroup, error) {
	group, err := s.groupRepo.GetGroupByID(ctx, groupID)
	if err != nil {
		// Assuming repo returns error if not found, mapping to ErrNotFound handled by caller or wrapper?
		// Repos usually return generic error or sql.ErrNoRows.
		// Ideally repo should return domain error or we map it here.
		// Given I can't see repo implementation fully, I'll assume standard behavior or wrap.
		// But reusing existing pattern:
		return nil, types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"groupID": groupID})
	}
	return group, nil
}

// UpdateGroup updates an existing group.
func (s *service) UpdateGroup(ctx context.Context, group *types.UserGroup) error {
	if err := s.groupRepo.UpdateGroup(ctx, group); err != nil {
		s.logger.Error(ctx, "Failed to update group", zap.Error(err), zap.String("groupID", group.ID))
		return types.ErrInternal.WithCause(err)
	}
	s.logger.Info(ctx, "Group updated", zap.String("groupID", group.ID))
	return nil
}

// DeleteGroup deletes a group by ID.
func (s *service) DeleteGroup(ctx context.Context, groupID string) error {
	if err := s.groupRepo.DeleteGroup(ctx, groupID); err != nil {
		s.logger.Error(ctx, "Failed to delete group", zap.Error(err), zap.String("groupID", groupID))
		return types.ErrInternal.WithCause(err)
	}
	s.logger.Info(ctx, "Group deleted", zap.String("groupID", groupID))
	return nil
}

// ListGroups lists groups with pagination.
func (s *service) ListGroups(ctx context.Context, offset, limit int) ([]*types.UserGroup, error) {
	groups, err := s.groupRepo.ListGroups(ctx, PaginationQuery{Offset: offset, PageSize: limit})
	if err != nil {
		s.logger.Error(ctx, "Failed to list groups", zap.Error(err))
		return nil, types.ErrInternal.WithCause(err)
	}
	return groups, nil
}

func (s *service) GetUserRepo() UserRepository {
	return s.userRepo
}

// GetUserByExternalID retrieves a user by their external ID.
// This assumes ExternalID is stored in the attributes JSONB column.
//
// Parameters:
//   - ctx: The context for the request.
//   - externalID: The external ID of the user to retrieve.
//
// Returns:
//   The user object if found, or an error.
func (s *service) GetUserByExternalID(ctx context.Context, externalID string) (*types.User, error) {
	// FindUsersByAttribute is available in UserRepository
	users, err := s.userRepo.FindUsersByAttribute(ctx, "externalId", externalID)
	if err != nil {
		return nil, types.ErrInternal.WithCause(err)
	}
	if len(users) == 0 {
		return nil, types.ErrNotFound.WithDetails(map[string]string{"externalId": externalID})
	}
	if len(users) > 1 {
		s.logger.Warn(ctx, "Multiple users found with same externalId", zap.String("externalId", externalID))
	}
	return users[0], nil
}
