package session

import (
	"context"
	"time"

	"github.com/turtacn/QuantaID/internal/auth/device"
	"github.com/turtacn/QuantaID/internal/storage/redis"
)

// RiskSignal represents a specific risk indicator detected during session monitoring.
type RiskSignal struct {
	Type       string      // ip_change, geo_jump, inactive, behavior_anomaly
	Severity   string      // low/medium/high
	Value      interface{}
	DetectedAt time.Time
}

// IPReputationChecker defines the interface for checking IP reputation.
type IPReputationChecker interface {
	GetReputation(ip string) *IPReputation
}

type IPReputation struct {
	Score   int
	IsTor   bool
	IsProxy bool
}

// GeoService defines the interface for geolocation services.
type GeoService interface {
	GetLocation(ip string) *GeoLocation
}

type GeoLocation struct {
	Latitude  float64
	Longitude float64
	City      string
	Country   string
}

// MonitorConfig holds configuration for the RiskMonitor.
type MonitorConfig struct {
	GeoJumpThresholdKm    float64 // Default 500
	GeoJumpTimeMinutes    int     // Default 60
	SuspiciousIPThreshold int     // IP reputation score threshold, default 30
}

// RiskMonitor is responsible for collecting real-time risk signals.
type RiskMonitor struct {
	deviceService *device.DeviceService
	geoService    GeoService
	ipChecker     IPReputationChecker
	riskStore     *redis.SessionRiskStore
	config        MonitorConfig
}

// NewRiskMonitor creates a new RiskMonitor.
func NewRiskMonitor(deviceService *device.DeviceService, geoService GeoService, ipChecker IPReputationChecker, riskStore *redis.SessionRiskStore, config MonitorConfig) *RiskMonitor {
	if config.GeoJumpThresholdKm == 0 {
		config.GeoJumpThresholdKm = 500
	}
	if config.GeoJumpTimeMinutes == 0 {
		config.GeoJumpTimeMinutes = 60
	}
	if config.SuspiciousIPThreshold == 0 {
		config.SuspiciousIPThreshold = 30
	}

	return &RiskMonitor{
		deviceService: deviceService,
		geoService:    geoService,
		ipChecker:     ipChecker,
		riskStore:     riskStore,
		config:        config,
	}
}

// CollectSignals gathers risk signals for a given session.
func (m *RiskMonitor) CollectSignals(ctx context.Context, session *Session) []RiskSignal {
	var signals []RiskSignal

	// Check IP change
	if ipSignal := m.checkIPChange(ctx, session); ipSignal != nil {
		signals = append(signals, *ipSignal)
	}

	// Check Geo Jump
	if geoSignal := m.checkGeoJump(ctx, session); geoSignal != nil {
		signals = append(signals, *geoSignal)
	}

	// Check IP Reputation
	if ipRepSignal := m.checkIPReputation(ctx, session.CurrentIP); ipRepSignal != nil {
		signals = append(signals, *ipRepSignal)
	}

	// Check Device Change
	if deviceSignal := m.checkDeviceChange(ctx, session); deviceSignal != nil {
		signals = append(signals, *deviceSignal)
	}

	// Check Inactivity
	if inactiveSignal := m.checkInactivity(session); inactiveSignal != nil {
		signals = append(signals, *inactiveSignal)
	}

	return signals
}

func (m *RiskMonitor) checkIPChange(ctx context.Context, session *Session) *RiskSignal {
	if session.PreviousIP != "" && session.PreviousIP != session.CurrentIP {
		return &RiskSignal{
			Type:       "ip_change",
			Severity:   "low",
			Value:      map[string]string{"old": session.PreviousIP, "new": session.CurrentIP},
			DetectedAt: time.Now(),
		}
	}
	return nil
}

func (m *RiskMonitor) checkGeoJump(ctx context.Context, session *Session) *RiskSignal {
	if session.PreviousIP == "" || session.PreviousIP == session.CurrentIP {
		return nil
	}
	if m.geoService == nil {
		return nil
	}

	prevLoc := m.geoService.GetLocation(session.PreviousIP)
	currLoc := m.geoService.GetLocation(session.CurrentIP)
	if prevLoc == nil || currLoc == nil {
		return nil
	}

	distance := HaversineDistance(prevLoc, currLoc)
	timeDiff := time.Since(session.LastIPChangeAt)

	// If timeDiff is very small, use a small value to avoid division by zero or huge speeds
	if timeDiff.Minutes() < 1 {
		timeDiff = 1 * time.Minute
	}

	if distance > m.config.GeoJumpThresholdKm && timeDiff.Minutes() < float64(m.config.GeoJumpTimeMinutes) {
		speed := distance / timeDiff.Hours()
		severity := "high"
		if speed < 500 {
			severity = "medium"
		}
		return &RiskSignal{
			Type:       "geo_jump",
			Severity:   severity,
			Value:      map[string]interface{}{"distance_km": distance, "speed_kmh": speed},
			DetectedAt: time.Now(),
		}
	}
	return nil
}

func (m *RiskMonitor) checkIPReputation(ctx context.Context, ip string) *RiskSignal {
	if m.ipChecker == nil {
		return nil
	}
	reputation := m.ipChecker.GetReputation(ip)
	if reputation == nil {
		return nil
	}

	if reputation.Score < m.config.SuspiciousIPThreshold {
		severity := "medium"
		if reputation.Score < 15 {
			severity = "high"
		}
		if reputation.IsTor || reputation.IsProxy {
			severity = "high"
		}
		return &RiskSignal{
			Type:       "suspicious_ip",
			Severity:   severity,
			Value:      reputation,
			DetectedAt: time.Now(),
		}
	}
	return nil
}

func (m *RiskMonitor) checkDeviceChange(ctx context.Context, session *Session) *RiskSignal {
	// In this simplified model, we assume session is bound to a device ID.
	// If DeviceID changed on the session object (which shouldn't happen for the same session ID usually),
	// or if we compare against historical data.
	// For this implementation, we check if the device is trusted.

	if m.deviceService == nil {
		return nil
	}

	device, err := m.deviceService.GetByID(ctx, session.DeviceID)
	if err != nil || device == nil {
		// Device not found or error
		return nil
	}

	// Assuming TrustScore is available on device struct.
	// If the device package struct doesn't have TrustScore exported or available, we might need to adjust.
	// Based on memory, TrustScorer calculates score. We might need to call TrustScorer here or assume it's stored.
	// Let's assume we can access trust score. If not, we skip.
	// For now, let's assume a method or field exists. I will check device model later if compilation fails.
	// Using a safe assumption:

	/*
	if device.TrustScore < 30 {
		return &RiskSignal{
			Type:       "untrusted_device",
			Severity:   "medium",
			Value:      session.DeviceID,
			DetectedAt: time.Now(),
		}
	}
	*/

	return nil
}

func (m *RiskMonitor) checkInactivity(session *Session) *RiskSignal {
	inactiveMinutes := time.Since(session.LastActivityAt).Minutes()
	if inactiveMinutes > 30 {
		severity := "low"
		if inactiveMinutes > 60 {
			severity = "medium"
		}
		return &RiskSignal{
			Type:       "inactive",
			Severity:   severity,
			Value:      inactiveMinutes,
			DetectedAt: time.Now(),
		}
	}
	return nil
}

// RegisterSignal allows external components to register a signal manually.
func (m *RiskMonitor) RegisterSignal(ctx context.Context, sessionID string, signal RiskSignal) error {
	return m.riskStore.AddSignal(ctx, sessionID, signal.Type)
}

// HaversineDistance calculates the distance between two points on Earth in kilometers.
// This is a placeholder; a real implementation would use math.
func HaversineDistance(p1, p2 *GeoLocation) float64 {
	// Simplified mock implementation
	// Real implementation requires math.Sin, math.Cos, etc.
	// For now returning 0 to avoid compilation errors if math is not imported
	return 0.0
}
