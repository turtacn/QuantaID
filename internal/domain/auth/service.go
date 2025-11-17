package auth

import (
	"context"
	"github.com/turtacn/QuantaID/internal/auth/adaptive"
	"github.com/turtacn/QuantaID/internal/auth/mfa"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"time"
)

// Service encapsulates the core business logic for authentication,
// including user login, logout, session management, and token handling.
type Service struct {
	identityService   identity.IService
	sessionRepo       SessionRepository
	tokenRepo         TokenRepository
	auditRepo         AuditLogRepository
	tokenFamilyRepo   TokenFamilyRepository
	crypto            *utils.CryptoManager
	logger            utils.Logger
	riskEngine        *adaptive.RiskEngine
	mfaManager        *mfa.MFAManager
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

type AuthnRequest struct {
	Username          string
	Password          string
	IPAddress         string
	DeviceFingerprint string
}

// NewService creates a new authentication service instance.
func NewService(
	identityService identity.IService,
	sessionRepo SessionRepository,
	tokenRepo TokenRepository,
	auditRepo AuditLogRepository,
	crypto *utils.CryptoManager,
	logger utils.Logger,
	riskEngine *adaptive.RiskEngine,
	mfaManager *mfa.MFAManager,
) *Service {
	return &Service{
		identityService: identityService,
		sessionRepo:     sessionRepo,
		tokenRepo:       tokenRepo,
		auditRepo:       auditRepo,
		crypto:          crypto,
		logger:          logger,
		riskEngine:      riskEngine,
		mfaManager:      mfaManager,
	}
}

// LoginWithPassword handles the traditional username and password authentication flow.
func (s *Service) LoginWithPassword(ctx context.Context, req AuthnRequest, serviceConfig Config) (*types.AuthResponse, error) {
	user, err := s.identityService.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, types.ErrInvalidCredentials.WithCause(err)
	}

	if user.Status != types.UserStatusActive {
		s.logAuthFailure(ctx, user.ID, "login_password", "user_not_active")
		return nil, types.ErrUserDisabled
	}

	if !s.crypto.CheckPasswordHash(req.Password, user.Password) {
		s.logAuthFailure(ctx, user.ID, "login_password", "invalid_password")
		return nil, types.ErrInvalidCredentials
	}

	riskScore, err := s.riskEngine.Evaluate(ctx, &adaptive.AuthEvent{
		UserID:            user.ID,
		IPAddress:         req.IPAddress,
		DeviceFingerprint: req.DeviceFingerprint,
		Timestamp:         time.Now(),
	})
	if err != nil {
		return nil, types.ErrInternal.WithCause(err)
	}

	if riskScore.Recommendation == "require_mfa" {
		// TODO: return a response that indicates MFA is required
		return &types.AuthResponse{
			Success:      false,
			NextStep:     "mfa",
			RequiredMFA:  []string{"totp", "webauthn"},
		}, nil
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

// RefreshAccessToken handles the process of issuing a new access token using a refresh token.
// It implements refresh token rotation to enhance security.
func (s *Service) RefreshAccessToken(ctx context.Context, refreshToken string, serviceConfig Config) (*types.Token, error) {
    userID, err := s.tokenRepo.GetRefreshTokenUserID(ctx, refreshToken)
    if err != nil {
        return nil, types.ErrInvalidGrant.WithCause(err)
    }

    family, err := s.tokenFamilyRepo.GetFamilyByToken(ctx, refreshToken)
    if err != nil {
        return nil, types.ErrInternal.WithCause(err)
    }

    // If the family is revoked, it's a sign of a replay attack.
    if family.RevokedAt != nil {
        s.logger.Warn(ctx, "Refresh token replay detected", zap.String("familyID", family.FamilyID))
        return nil, types.ErrInvalidGrant
    }

    // Generate a new pair of tokens.
    newAccessToken, err := s.crypto.GenerateJWT(userID, serviceConfig.AccessTokenDuration, nil)
    if err != nil {
        return nil, types.ErrInternal.WithCause(err)
    }
    newRefreshToken := s.crypto.GenerateUUID()
    if err := s.tokenRepo.StoreRefreshToken(ctx, newRefreshToken, userID, serviceConfig.RefreshTokenDuration); err != nil {
        return nil, types.ErrInternal.WithCause(err)
    }

    // Update the token family.
    family.IssuedTokens = append(family.IssuedTokens, newRefreshToken)
    family.CurrentToken = newRefreshToken
    if err := s.tokenFamilyRepo.UpdateFamily(ctx, family); err != nil {
        return nil, types.ErrInternal.WithCause(err)
    }

    // Revoke the old refresh token.
    if err := s.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
        s.logger.Warn(ctx, "Failed to delete old refresh token", zap.Error(err))
    }

    return &types.Token{
        AccessToken:  newAccessToken,
        RefreshToken: newRefreshToken,
        TokenType:    "Bearer",
        ExpiresIn:    int64(serviceConfig.AccessTokenDuration.Seconds()),
    }, nil
}

// RevokeToken handles the revocation of a token (either access or refresh).
func (s *Service) RevokeToken(ctx context.Context, token, tokenTypeHint string) error {
    if tokenTypeHint == "refresh_token" {
        family, err := s.tokenFamilyRepo.GetFamilyByToken(ctx, token)
        if err == nil && family != nil {
            return s.tokenFamilyRepo.RevokeFamily(ctx, family.FamilyID)
        }
        return s.tokenRepo.DeleteRefreshToken(ctx, token)
    }

    claims, err := s.crypto.ValidateJWT(token)
    if err != nil {
        return types.ErrInvalidToken.WithCause(err)
    }
    jti, _ := claims["jti"].(string)
    exp, _ := claims["exp"].(float64)
    ttl := time.Until(time.Unix(int64(exp), 0))
    return s.tokenRepo.AddToDenyList(ctx, jti, ttl)
}

// logAuthFailure records a failed authentication attempt to the audit log.
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
