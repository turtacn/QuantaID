package auth

import (
	"context"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"time"
)

// Service encapsulates the core business logic for authentication,
// including user login, logout, session management, and token handling.
type Service struct {
	identityService identity.IService
	sessionRepo     SessionRepository
	tokenRepo       TokenRepository
	auditRepo       AuditLogRepository
	crypto          *utils.CryptoManager
	logger          utils.Logger
}

// Config holds configuration for the auth service, specifically token and session lifetimes.
type Config struct {
	// AccessTokenDuration specifies the validity period for access tokens.
	AccessTokenDuration time.Duration
	// RefreshTokenDuration specifies the validity period for refresh tokens.
	RefreshTokenDuration time.Duration
	// SessionDuration specifies the validity period for user sessions.
	SessionDuration time.Duration
}

// NewService creates a new authentication service instance.
// It brings together all the necessary dependencies to handle authentication logic.
//
// Parameters:
//   - identityService: The service for interacting with user identity data.
//   - sessionRepo: The repository for managing user sessions.
//   - tokenRepo: The repository for managing tokens and deny lists.
//   - auditRepo: The repository for recording audit logs.
//   - crypto: The utility for cryptographic operations.
//   - logger: The logger for logging service-level messages.
//
// Returns:
//   A new authentication service instance.
func NewService(
	identityService identity.IService,
	sessionRepo SessionRepository,
	tokenRepo TokenRepository,
	auditRepo AuditLogRepository,
	crypto *utils.CryptoManager,
	logger utils.Logger,
) *Service {
	return &Service{
		identityService: identityService,
		sessionRepo:     sessionRepo,
		tokenRepo:       tokenRepo,
		auditRepo:       auditRepo,
		crypto:          crypto,
		logger:          logger,
	}
}

// LoginWithPassword handles the traditional username and password authentication flow.
// It validates the user's credentials, checks their account status, and if successful,
// creates a new session and issues access and refresh tokens.
//
// Parameters:
//   - ctx: The context for the request.
//   - username: The user's username.
//   - password: The user's plain-text password.
//   - serviceConfig: The configuration containing token and session durations.
//
// Returns:
//   An AuthResponse containing tokens and user info, or an error if login fails.
func (s *Service) LoginWithPassword(ctx context.Context, username, password string, serviceConfig Config) (*types.AuthResponse, error) {
	user, err := s.identityService.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, types.ErrInvalidCredentials.WithCause(err)
	}

	if user.Status != types.UserStatusActive {
		s.logAuthFailure(ctx, user.ID, "login_password", "user_not_active")
		return nil, types.ErrUserDisabled
	}

	if !s.crypto.CheckPasswordHash(password, user.Password) {
		s.logAuthFailure(ctx, user.ID, "login_password", "invalid_password")
		return nil, types.ErrInvalidCredentials
	}

	return s.createSessionAndTokens(ctx, user, serviceConfig)
}

// createSessionAndTokens is a helper function that generates JWTs, creates a user session,
// and constructs the final authentication response.
func (s *Service) createSessionAndTokens(ctx context.Context, user *types.User, serviceConfig Config) (*types.AuthResponse, error) {
	accessToken, err := s.crypto.GenerateJWT(user.ID, serviceConfig.AccessTokenDuration, nil)
	if err != nil {
		s.logger.Error(ctx, "Failed to generate access token", zap.Error(err), zap.String("userID", user.ID))
		return nil, types.ErrInternal.WithCause(err)
	}

	refreshToken := s.crypto.GenerateUUID()
	if err := s.tokenRepo.StoreRefreshToken(ctx, refreshToken, user.ID, serviceConfig.RefreshTokenDuration); err != nil {
		s.logger.Error(ctx, "Failed to store refresh token", zap.Error(err), zap.String("userID", user.ID))
		return nil, types.ErrInternal.WithCause(err)
	}

	session := &types.UserSession{
		ID:        s.crypto.GenerateUUID(),
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(serviceConfig.SessionDuration),
	}
	if err := s.sessionRepo.CreateSession(ctx, session, serviceConfig.SessionDuration); err != nil {
		s.logger.Error(ctx, "Failed to create session", zap.Error(err), zap.String("userID", user.ID))
		return nil, types.ErrInternal.WithCause(err)
	}

	s.logAuthSuccess(ctx, user.ID, "login_password")

	return &types.AuthResponse{
		Success: true,
		Token: &types.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(serviceConfig.AccessTokenDuration.Seconds()),
		},
		User: user,
	}, nil
}

// Logout handles the user logout process.
// It deletes the user's session and adds the provided access token to a deny list
// to prevent its reuse until it expires.
//
// Parameters:
//   - ctx: The context for the request.
//   - sessionID: The ID of the session to be terminated.
//   - accessToken: The access token to be invalidated.
//
// Returns:
//   An error if the process fails, otherwise nil.
func (s *Service) Logout(ctx context.Context, sessionID, accessToken string) error {
	if err := s.sessionRepo.DeleteSession(ctx, sessionID); err != nil {
		s.logger.Warn(ctx, "Failed to delete session on logout", zap.Error(err), zap.String("sessionID", sessionID))
	}

	claims, err := s.crypto.ValidateJWT(accessToken)
	if err != nil {
		return types.ErrInvalidToken.WithCause(err)
	}
	jti, _ := claims["jti"].(string)
	exp, _ := claims["exp"].(float64)
	ttl := time.Until(time.Unix(int64(exp), 0))
	if err := s.tokenRepo.AddToDenyList(ctx, jti, ttl); err != nil {
		s.logger.Error(ctx, "Failed to add token to deny list", zap.Error(err), zap.String("jti", jti))
		return types.ErrInternal.WithCause(err)
	}

	return nil
}

// logAuthSuccess records a successful authentication event to the audit log.
// It runs in a separate goroutine to avoid blocking the main authentication flow.
func (s *Service) logAuthSuccess(ctx context.Context, userID, method string) {
	go func() {
		logEntry := &types.AuditLog{
			ID:        s.crypto.GenerateUUID(),
			ActorID:   userID,
			Action:    string(types.EventUserLoginSuccess),
			Resource:  "user:" + userID,
			Status:    "success",
			Context:   map[string]interface{}{"method": method},
			Timestamp: time.Now().UTC(),
		}
		if err := s.auditRepo.CreateLogEntry(context.Background(), logEntry); err != nil {
			s.logger.Error(context.Background(), "Failed to create audit log for successful auth", zap.Error(err))
		}
	}()
}

// logAuthFailure records a failed authentication attempt to the audit log.
// It runs in a separate goroutine.
func (s *Service) logAuthFailure(ctx context.Context, userID, method, reason string) {
	go func() {
		logEntry := &types.AuditLog{
			ID:        s.crypto.GenerateUUID(),
			ActorID:   userID,
			Action:    string(types.EventUserLoginFailure),
			Resource:  "user:" + userID,
			Status:    "failure",
			Context: map[string]interface{}{
				"method": method,
				"reason": reason,
			},
			Timestamp: time.Now().UTC(),
		}
		if err := s.auditRepo.CreateLogEntry(context.Background(), logEntry); err != nil {
			s.logger.Error(context.Background(), "Failed to create audit log for failed auth", zap.Error(err))
		}
	}()
}
