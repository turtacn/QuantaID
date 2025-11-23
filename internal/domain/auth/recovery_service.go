package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/turtacn/QuantaID/internal/auth/mfa"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

var (
	ErrInvalidCode = errors.New("invalid or expired OTP code")
)

// RecoveryService handles password recovery operations.
type RecoveryService struct {
	userRepo       identity.UserRepository
	otpProvider    *mfa.OTPProvider
	cryptoManager  *utils.CryptoManager
	sessionManager *redis.SessionManager
	logger         *zap.Logger
}

// NewRecoveryService creates a new RecoveryService.
func NewRecoveryService(
	userRepo identity.UserRepository,
	otpProvider *mfa.OTPProvider,
	cryptoManager *utils.CryptoManager,
	sessionManager *redis.SessionManager,
	logger *zap.Logger,
) *RecoveryService {
	return &RecoveryService{
		userRepo:       userRepo,
		otpProvider:    otpProvider,
		cryptoManager:  cryptoManager,
		sessionManager: sessionManager,
		logger:         logger,
	}
}

// InitiateRecovery starts the password recovery process by sending an OTP to the user's email.
func (s *RecoveryService) InitiateRecovery(ctx context.Context, email string) error {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, types.ErrUserNotFound) {
			// Maintain silence to prevent email enumeration
			s.logger.Info("Password recovery requested for non-existent email", zap.String("email", email))
			return nil
		}
		s.logger.Error("Error looking up user for recovery", zap.Error(err))
		return err
	}

	// Generate and send OTP via email
	// Note: We use "email" as the method. P1's OTPProvider should handle this.
	_, err = s.otpProvider.Challenge(ctx, user.ID, email, "email")
	if err != nil {
		s.logger.Error("Failed to generate/send recovery OTP", zap.Error(err))
		return err
	}

	s.logger.Info("Password recovery initiated", zap.String("userID", user.ID))
	return nil
}

// VerifyAndReset verifies the OTP and resets the user's password.
func (s *RecoveryService) VerifyAndReset(ctx context.Context, email, code, newPassword string) error {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, types.ErrUserNotFound) {
			return types.ErrInvalidCredentials // Generic error
		}
		return err
	}

	// Verify OTP
	valid, err := s.otpProvider.Verify(ctx, user.ID, code)
	if err != nil {
		s.logger.Error("Error verifying OTP", zap.Error(err))
		return err
	}
	if !valid {
		return ErrInvalidCode
	}

	// Hash new password
	hashedPassword, err := s.cryptoManager.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.Password = hashedPassword
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	// Revoke all existing sessions to enforce re-login
	if err := s.sessionManager.RevokeAllUserSessions(ctx, user.ID); err != nil {
		// Log but don't fail the request, password is already changed
		s.logger.Warn("Failed to revoke sessions after password reset", zap.Error(err))
	}

	s.logger.Info("Password reset successful", zap.String("userID", user.ID))
	return nil
}
