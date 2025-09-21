package identity

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// service implements the IService interface.
type service struct {
	userRepo  UserRepository
	groupRepo GroupRepository
	crypto    *utils.CryptoManager
	logger    utils.Logger
}

// NewService creates a new identity service.
func NewService(userRepo UserRepository, groupRepo GroupRepository, crypto *utils.CryptoManager, logger utils.Logger) IService {
	return &service{
		userRepo:  userRepo,
		groupRepo: groupRepo,
		crypto:    crypto,
		logger:    logger,
	}
}

func (s *service) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, types.ErrValidation.WithDetails(map[string]string{"field": "username/email/password", "error": "cannot be empty"})
	}

	if existing, _ := s.userRepo.GetUserByUsername(ctx, username); existing != nil {
		return nil, types.ErrConflict.WithDetails(map[string]string{"field": "username", "value": username})
	}
	if existing, _ := s.userRepo.GetUserByEmail(ctx, email); existing != nil {
		return nil, types.ErrConflict.WithDetails(map[string]string{"field": "email", "value": email})
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

func (s *service) GetUser(ctx context.Context, userID string) (*types.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"userID": userID})
	}
	return user, nil
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, types.ErrNotFound.WithCause(err).WithDetails(map[string]string{"username": username})
	}
	return user, nil
}

func (s *service) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	groups, err := s.groupRepo.GetUserGroups(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get user groups from repository", zap.Error(err), zap.String("userID", userID))
		return nil, types.ErrInternal.WithCause(err)
	}
	return groups, nil
}

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

//Personal.AI order the ending
