package auth

import (
	"context"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"time"
)

// ApplicationService provides application-level use cases for authentication.
type ApplicationService struct {
	authDomain *auth.Service
	logger     utils.Logger
	config     Config
}

// Config holds the application-level configuration for the auth service.
type Config struct {
	AccessTokenDuration  time.Duration `yaml:"accessTokenDuration"`
	RefreshTokenDuration time.Duration `yaml:"refreshTokenDuration"`
	SessionDuration      time.Duration `yaml:"sessionDuration"`
}

// NewApplicationService creates a new auth application service.
func NewApplicationService(authDomain *auth.Service, logger utils.Logger, config Config) *ApplicationService {
	return &ApplicationService{
		authDomain: authDomain,
		logger:     logger,
		config:     config,
	}
}

// LoginRequest defines the DTO for a login request.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse defines the DTO for a successful login response.
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
	User         *UserDTO `json:"user"`
}

// UserDTO is a safe representation of a user for API responses.
type UserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Login handles the primary authentication use case.
func (s *ApplicationService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, *types.Error) {
	domainConfig := auth.Config{
		AccessTokenDuration:  s.config.AccessTokenDuration,
		RefreshTokenDuration: s.config.RefreshTokenDuration,
		SessionDuration:      s.config.SessionDuration,
	}

	authResp, err := s.authDomain.LoginWithPassword(ctx, req.Username, req.Password, domainConfig)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}

	return &LoginResponse{
		AccessToken:  authResp.Token.AccessToken,
		RefreshToken: authResp.Token.RefreshToken,
		TokenType:    authResp.Token.TokenType,
		ExpiresIn:    authResp.Token.ExpiresIn,
		User: &UserDTO{
			ID:       authResp.User.ID,
			Username: authResp.User.Username,
			Email:    authResp.User.Email,
		},
	}, nil
}

// LogoutRequest defines the DTO for a logout request.
type LogoutRequest struct {
	SessionID   string `json:"sessionId"`
	AccessToken string `json:"accessToken"`
}

// Logout handles the session and token invalidation use case.
func (s *ApplicationService) Logout(ctx context.Context, req LogoutRequest) *types.Error {
	err := s.authDomain.Logout(ctx, req.SessionID, req.AccessToken)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return appErr
		}
		return types.ErrInternal.WithCause(err)
	}
	return nil
}

//Personal.AI order the ending
