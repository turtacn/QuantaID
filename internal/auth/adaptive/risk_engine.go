package adaptive

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/turtacn/QuantaID/internal/auth/adaptive/strategies"
	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"go.uber.org/zap"
)

// RiskEngine evaluates the risk of an authentication attempt.
type RiskEngine struct {
	config           config.RiskConfig
	redisClient      redis.RedisClientInterface
	geoManager       *redis.GeoManager
	geoIPReader      GeoIPReader
	logger           *zap.Logger
	travelStrategy   *strategies.ImpossibleTravelStrategy
	ipReputation     *strategies.IPReputationStrategy
}

// NewRiskEngine creates a new risk engine with the given configuration and dependencies.
func NewRiskEngine(
	cfg config.RiskConfig,
	redisClient redis.RedisClientInterface,
	geoManager *redis.GeoManager,
	geoIPReader GeoIPReader,
	logger *zap.Logger,
) *RiskEngine {
	return &RiskEngine{
		config:           cfg,
		redisClient:      redisClient,
		geoManager:       geoManager,
		geoIPReader:      geoIPReader,
		logger:           logger,
		travelStrategy:   strategies.NewImpossibleTravelStrategy(geoManager),
		ipReputation:     strategies.NewIPReputationStrategy(),
	}
}

// Evaluate assesses the risk of the authentication attempt based on the AuthContext.
func (e *RiskEngine) Evaluate(ctx context.Context, ac auth.AuthContext) (auth.RiskScore, auth.RiskLevel, error) {
	// 1. Resolve Geo Location from IP
	lat, lon := e.resolveGeo(ac.IPAddress)

	// 2. Parallel execution of strategies
	var wg sync.WaitGroup
	var travelRisk, ipRisk float64
	var errTravel, errIP error

	wg.Add(2)

	go func() {
		defer wg.Done()
		if ac.UserID != "" {
			travelRisk, errTravel = e.travelStrategy.CalculateRisk(ctx, ac.UserID, lat, lon)
			if errTravel != nil {
				e.logger.Error("Failed to calculate travel risk", zap.Error(errTravel))
			}
		}
	}()

	go func() {
		defer wg.Done()
		ipRisk, errIP = e.ipReputation.CalculateRisk(ctx, ac.IPAddress)
		if errIP != nil {
			e.logger.Error("Failed to calculate IP risk", zap.Error(errIP))
		}
	}()

	wg.Wait()

	// 3. Check Device Risk (Synchronous for now, or could also be parallel)
	isKnown, err := e.isKnownDevice(ctx, ac.UserID, ac.DeviceFingerprint)
	if err != nil {
		e.logger.Warn("Failed to check device fingerprint", zap.Error(err))
	}

	// deviceRisk calculation was not used previously, but now we use isKnown in factors
	// deviceRisk := 0.0
	// if !isKnown {
	// 	deviceRisk = 1.0
	// }

	// 4. Update AuthContext with derived data if needed (though passed by value here)
	// Ideally we return the factors.

	factors := auth.RiskFactors{
		IPReputation:   ipRisk,
		IsKnownDevice:  isKnown,
		GeoReputation:  0.2, // Placeholder default
		GeoVelocity:    travelRisk,
		UserAgent:      ac.UserAgent,
		IPAddress:      ac.IPAddress,
		AcceptLanguage: ac.AcceptLanguage,
		TimeWindow:     ac.Timestamp,
	}

	score := factors.ToScore(e.config)
	level := score.Level(e.config.Thresholds)

	// 5. Async update Geo Location
	if ac.UserID != "" && e.geoManager != nil {
		go func() {
			bgCtx := context.Background()
			if err := e.geoManager.SaveLoginGeo(bgCtx, ac.UserID, lat, lon, ac.Timestamp); err != nil {
				e.logger.Error("Failed to save login geo", zap.Error(err))
			}
		}()
	}

	e.logger.Info("Risk evaluation complete",
		zap.Float64("score", float64(score)),
		zap.String("level", string(level)),
		zap.Float64("travel_risk", travelRisk),
		zap.Float64("ip_risk", ipRisk),
		zap.Bool("is_known_device", isKnown),
	)

	return score, level, nil
}

// resolveGeo resolves IP to Latitude and Longitude. Returns 0,0 if failed.
func (e *RiskEngine) resolveGeo(ipStr string) (float64, float64) {
	if e.geoIPReader == nil {
		return 0, 0
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, 0
	}
	city, err := e.geoIPReader.City(ip)
	if err != nil || city == nil {
		return 0, 0
	}
	return city.Location.Latitude, city.Location.Longitude
}

// isKnownDevice checks if the device fingerprint is known for the user.
func (e *RiskEngine) isKnownDevice(ctx context.Context, userID, fingerprint string) (bool, error) {
	if userID == "" || fingerprint == "" {
		return false, nil
	}
	key := fmt.Sprintf("user:%s:devices", userID)
	return e.redisClient.SIsMember(ctx, key, fingerprint).Result()
}
