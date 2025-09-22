package auth

import (
	"context"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"time"
)

// Service encapsulates the business logic for authentication.
type Service struct {
	identityService identity.IService
	sessionRepo     SessionRepository
	tokenRepo       TokenRepository
	auditRepo       AuditLogRepository
	crypto          *utils.CryptoManager
	logger          utils.Logger
}

// Config holds configuration for the auth service.
type Config struct {
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	SessionDuration      time.Duration
}

// NewService creates a new authentication service.
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

// LoginWithPassword handles the password-based login flow.
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

func (s *Service) logAuthSuccess(ctx context.Context, userID, method string) {
	go func() {
		_ = s.auditRepo.CreateLogEntry(context.Background(), &types.AuditLog{})
	}()
}

func (s *Service) logAuthFailure(ctx context.Context, userID, method, reason string) {
	go func() {
		_ = s.auditRepo.CreateLogEntry(context.Background(), &types.AuditLog{})
	}()
}

//Personal.AI order the ending
