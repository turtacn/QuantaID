package auth

import (
	"context"
	"github.com/turtacn/QuantaID/internal/services/audit"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.opentelemetry.io/otel/trace"
	"time"
)

// ApplicationService provides application-level use cases for authentication. It acts as
// a facade over the authentication domain service, handling data transfer objects (DTOs)
// and coordinating with the domain layer.
type ApplicationService struct {
	authDomain   *auth.Service
	auditService *audit.Service
	logger       utils.Logger
	config       Config
	tracer       trace.Tracer
}

// Config holds the application-level configuration for the auth service,
// primarily related to token and session lifetimes.
type Config struct {
	AccessTokenDuration  time.Duration `yaml:"accessTokenDuration"`
	RefreshTokenDuration time.Duration `yaml:"refreshTokenDuration"`
	SessionDuration      time.Duration `yaml:"sessionDuration"`
}

// NewApplicationService creates a new authentication application service.
func NewApplicationService(authDomain *auth.Service, auditService *audit.Service, logger utils.Logger, config Config, tracer trace.Tracer) *ApplicationService {
	return &ApplicationService{
		authDomain:   authDomain,
		auditService: auditService,
		logger:       logger,
		config:       config,
		tracer:       tracer,
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
func (s *ApplicationService) LoginWithPassword(ctx context.Context, req auth.AuthnRequest, serviceConfig auth.Config) (*types.AuthResult, error) {
	ctx, span := s.tracer.Start(ctx, "ApplicationService.LoginWithPassword")
	defer span.End()

	// TODO: Extract IP and TraceID from context
	ip := "not_implemented"
	traceID := "not_implemented"

	authResp, err := s.authDomain.LoginWithPassword(ctx, req, serviceConfig)
	if err != nil {
		span.RecordError(err)
		s.auditService.RecordLoginFailed(ctx, req.Username, ip, traceID, err.Error(), nil)
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}

	metrics.OauthTokensIssuedTotal.Inc()
	return authResp, nil
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

func (s *ApplicationService) VerifyMFAChallenge(ctx context.Context, req *types.VerifyMFARequest, serviceConfig auth.Config) (*types.AuthResult, error) {
	return s.authDomain.VerifyMFAChallenge(ctx, req, serviceConfig)
}
