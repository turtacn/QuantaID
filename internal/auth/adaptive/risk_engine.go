package adaptive

import (
	"context"
	"fmt"

	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"go.uber.org/zap"
)

// RiskEngine evaluates the risk of an authentication attempt.
type RiskEngine struct {
	config      config.RiskConfig
	redisClient redis.RedisClientInterface
	logger      *zap.Logger
}

// NewRiskEngine creates a new risk engine with the given configuration and dependencies.
func NewRiskEngine(cfg config.RiskConfig, redisClient redis.RedisClientInterface, logger *zap.Logger) *RiskEngine {
	return &RiskEngine{
		config:      cfg,
		redisClient: redisClient,
		logger:      logger,
	}
}

// Evaluate assesses the risk of the authentication attempt based on the AuthContext.
func (e *RiskEngine) Evaluate(ctx context.Context, ac auth.AuthContext) (auth.RiskScore, auth.RiskLevel, error) {
	factors, err := e.buildRiskFactors(ctx, ac)
	if err != nil {
		return 0, "", fmt.Errorf("failed to build risk factors: %w", err)
	}

	score := factors.ToScore(e.config)
	level := score.Level(e.config.Thresholds)

	e.logger.Info("Risk evaluation complete",
		zap.Float64("score", float64(score)),
		zap.String("level", string(level)),
	)

	return score, level, nil
}

// buildRiskFactors assembles the risk factors from the AuthContext.
func (e *RiskEngine) buildRiskFactors(ctx context.Context, ac auth.AuthContext) (auth.RiskFactors, error) {
	isKnown, err := e.isKnownDevice(ctx, ac.UserID, ac.DeviceFingerprint)
	if err != nil {
		return auth.RiskFactors{}, err
	}

	return auth.RiskFactors{
		IPReputation:   e.evaluateIPReputation(ac.IPAddress),
		IsKnownDevice:  isKnown,
		GeoReputation:  0.2, // Placeholder
		GeoVelocity:    0.1, // Placeholder
		UserAgent:      ac.UserAgent,
		IPAddress:      ac.IPAddress,
		AcceptLanguage: ac.AcceptLanguage,
		TimeWindow:     ac.Timestamp,
	}, nil
}

// isKnownDevice checks if the device fingerprint is known for the user.
func (e *RiskEngine) isKnownDevice(ctx context.Context, userID, fingerprint string) (bool, error) {
	key := fmt.Sprintf("user:%s:devices", userID)
	return e.redisClient.SIsMember(ctx, key, fingerprint).Result()
}

// getFailedLogins returns the number of failed logins for the user in the last hour.
func (e *RiskEngine) getFailedLogins(ctx context.Context, userID string) (int, error) {
	key := fmt.Sprintf("user:%s:failed_logins", userID)
	val, err := e.redisClient.Get(ctx, key)
	if err != nil {
		return 0, nil
	}
	var count int
	_, err = fmt.Sscanf(val, "%d", &count)
	return count, err
}

// evaluateIPReputation provides a placeholder score for IP reputation.
func (e *RiskEngine) evaluateIPReputation(ip string) float64 {
	// In a real implementation, this would involve lookups against threat intelligence feeds.
	if ip == "8.8.8.8" { // Known good IP
		return 0.1
	}
	if ip == "1.2.3.4" { // Known bad IP
		return 0.9
	}
	return 0.4 // Neutral
}
