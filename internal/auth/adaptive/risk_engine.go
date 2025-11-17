package adaptive

import (
	"context"
	"net"
	"time"

	"github.com/turtacn/QuantaID/pkg/utils"
)

type RiskEngine struct {
	geoIP         GeoIPReader
	logger        utils.Logger
	// deviceFP      *DeviceFingerprint // TODO:
	// behaviorModel *BehaviorModel // TODO:
	config        RiskConfig
}

type RiskConfig struct {
	Weights RiskWeights `yaml:"weights"`
}

type RiskWeights struct {
	IPReputation    float64 `yaml:"ip_reputation"`
	GeoAnomaly      float64 `yaml:"geo_anomaly"`
	DeviceChange    float64 `yaml:"device_change"`
	TimeAnomaly     float64 `yaml:"time_anomaly"`
	VelocityAnomaly float64 `yaml:"velocity_anomaly"`
}

type RiskScore struct {
	TotalScore     float64
	Level          RiskLevel
	Factors        map[string]float64
	Recommendation string
}

type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

type AuthEvent struct {
	UserID            string
	IPAddress         string
	DeviceFingerprint string
	Timestamp         time.Time
}

func NewRiskEngine(geoIP GeoIPReader, logger utils.Logger) *RiskEngine {
	return &RiskEngine{
		geoIP:  geoIP,
		logger: logger,
	}
}

func (re *RiskEngine) Evaluate(ctx context.Context, event *AuthEvent) (*RiskScore, error) {
	factors := make(map[string]float64)

	ipScore := re.evaluateIPReputation(event.IPAddress)
	factors["ip_reputation"] = ipScore

	geoScore := re.evaluateGeoAnomaly(ctx, event.UserID, event.IPAddress)
	factors["geo_anomaly"] = geoScore

	// deviceScore := re.evaluateDeviceChange(ctx, event.UserID, event.DeviceFingerprint)
	// factors["device_change"] = deviceScore

	// timeScore := re.evaluateTimeAnomaly(ctx, event.UserID, event.Timestamp)
	// factors["time_anomaly"] = timeScore

	// velocityScore := re.evaluateVelocity(ctx, event.UserID, event.Timestamp)
	// factors["velocity_anomaly"] = velocityScore

	totalScore := ipScore*re.config.Weights.IPReputation +
		geoScore*re.config.Weights.GeoAnomaly
		// deviceScore*re.config.Weights.DeviceChange +
		// timeScore*re.config.Weights.TimeAnomaly +
		// velocityScore*re.config.Weights.VelocityAnomaly

	var level RiskLevel
	var recommendation string
	switch {
	case totalScore > 60:
		level = RiskLevelHigh
		recommendation = "require_mfa_and_notify"
	case totalScore > 30:
		level = RiskLevelMedium
		recommendation = "require_mfa"
	default:
		level = RiskLevelLow
		recommendation = "allow"
	}

	return &RiskScore{
		TotalScore:     totalScore,
		Level:          level,
		Factors:        factors,
		Recommendation: recommendation,
	}, nil
}

func (re *RiskEngine) evaluateIPReputation(ip string) float64 {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0.5 // Unknown IP format, medium risk
	}

	if re.geoIP == nil {
		return 0.0
	}

	record, err := re.geoIP.City(parsedIP)
	if err != nil {
		return 0.2 // Couldn't determine location, slight risk
	}

	score := 0.0
	if record.Traits.IsAnonymousProxy {
		score += 0.4
	}

	highRiskCountries := map[string]bool{
		"CN": true, "RU": true, "KP": true,
	}

	if highRiskCountries[record.Country.IsoCode] {
		score += 0.8
	}

	return score
}

func (re *RiskEngine) evaluateGeoAnomaly(ctx context.Context, userID, ip string) float64 {
	// TODO:
	return 0
}
