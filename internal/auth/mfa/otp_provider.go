package mfa

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/notification"
	"github.com/turtacn/QuantaID/pkg/utils"
)

type OTPConfig struct {
	TTL    time.Duration
	Length int
}

type OTPProvider struct {
	redisClient    redis.RedisClientInterface
	notifierManager notification.Manager
	cryptoManager  *utils.CryptoManager
	config         OTPConfig
}

func NewOTPProvider(redisClient redis.RedisClientInterface, notifierManager notification.Manager, cryptoManager *utils.CryptoManager, config OTPConfig) *OTPProvider {
	return &OTPProvider{
		redisClient:     redisClient,
		notifierManager: notifierManager,
		cryptoManager:   cryptoManager,
		config:          config,
	}
}

func (p *OTPProvider) Challenge(ctx context.Context, userID string, target string, method string) (string, error) {
	notifier, err := p.notifierManager.GetNotifier(method)
	if err != nil {
		return "", fmt.Errorf("failed to get notifier for method %s: %w", method, err)
	}

	code, err := p.generateRandomNumberString(p.config.Length)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	key := fmt.Sprintf("mfa:otp:%s", userID)
	if err := p.redisClient.SetEx(ctx, key, code, p.config.TTL).Err(); err != nil {
		return "", fmt.Errorf("failed to store OTP in Redis: %w", err)
	}

	msg := notification.Message{
		Recipient: target,
		Subject:   "Your OTP Code",
		Body:      fmt.Sprintf("Your One-Time Password is: <b>%s</b>. It expires in %d minutes.", code, int(p.config.TTL.Minutes())),
		Type:      notification.MessageTypeOTP,
	}

	if err := notifier.Send(ctx, msg); err != nil {
		return "", fmt.Errorf("failed to send OTP notification: %w", err)
	}

	// In OTP scenarios, we usually don't return the code or a challenge ID to the frontend
	// beyond a generic session identifier which should already exist.
	// Returning empty string as challenge ID for now.
	return "", nil
}

func (p *OTPProvider) Verify(ctx context.Context, userID string, code string) (bool, error) {
	key := fmt.Sprintf("mfa:otp:%s", userID)

	storedCode, err := p.redisClient.Get(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve OTP from Redis (might be expired): %w", err)
	}

	if storedCode == code {
		// Verify successful, delete the key to prevent replay
		_ = p.redisClient.Del(ctx, key)
		return true, nil
	}

	return false, nil
}

// generateRandomNumberString generates a numeric string of given length using crypto/rand
func (p *OTPProvider) generateRandomNumberString(length int) (string, error) {
	const digits = "0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = digits[b%byte(len(digits))]
	}
	return string(bytes), nil
}
