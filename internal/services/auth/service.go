package auth

import (
	"context"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"time"
)

// ApplicationService provides application-level use cases for authentication. It acts as
// a facade over the authentication domain service, handling data transfer objects (DTOs)
// and coordinating with the domain layer.
import (
	"go.opentelemetry.io/otel/trace"
)

type ApplicationService struct {
	authDomain *auth.Service
	logger     utils.Logger
	config     Config
	tracer     trace.Tracer
}

// Config holds the application-level configuration for the auth service,
// primarily related to token and session lifetimes.
type Config struct {
	AccessTokenDuration  time.Duration `yaml:"accessTokenDuration"`
	RefreshTokenDuration time.Duration `yaml:"refreshTokenDuration"`
	SessionDuration      time.Duration `yaml:"sessionDuration"`
}

// NewApplicationService creates a new authentication application service.
//
// Parameters:
//   - authDomain: The domain service containing the core authentication logic.
//   - logger: The logger for service-level messages.
//   - config: The configuration for token and session durations.
//
// Returns:
//   A new instance of ApplicationService.
func NewApplicationService(authDomain *auth.Service, logger utils.Logger, config Config, tracer trace.Tracer) *ApplicationService {
	return &ApplicationService{
		authDomain: authDomain,
		logger:     logger,
		config:     config,
		tracer:     tracer,
	}
}

// LoginRequest defines the Data Transfer Object (DTO) for a login request.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse defines the DTO for a successful login response.
// It contains the necessary tokens and user information for the client.
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
	User         *UserDTO `json:"user"`
}

// UserDTO is a safe representation of a user, suitable for exposing in API responses.
// It omits sensitive information like password hashes.
type UserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Login handles the primary authentication use case. It takes a login request DTO,
// passes the credentials to the domain service, and maps the domain response
// back to a response DTO.
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The login request DTO containing user credentials.
//
// Returns:
//   A LoginResponse DTO on success, or an application error on failure.
func (s *ApplicationService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, *types.Error) {
	ctx, span := s.tracer.Start(ctx, "ApplicationService.Login")
	defer span.End()

	domainConfig := auth.Config{
		AccessTokenDuration:  s.config.AccessTokenDuration,
		RefreshTokenDuration: s.config.RefreshTokenDuration,
		SessionDuration:      s.config.SessionDuration,
	}

	authResp, err := s.authDomain.LoginWithPassword(ctx, req.Username, req.Password, domainConfig)
	if err != nil {
		span.RecordError(err)
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}

	metrics.OauthTokensIssuedTotal.Inc()
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
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The logout request DTO containing the session and token to invalidate.
//
// Returns:
//   An application error if the logout process fails.
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
