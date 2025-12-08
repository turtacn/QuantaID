package device

import (
	"math"
	"time"
)

// AnomalyType defines the type of anomaly detected
type AnomalyType string

const (
	AnomalyGeoJump           AnomalyType = "geo_jump"
	AnomalyFingerprintChange AnomalyType = "fingerprint_change"
	AnomalyUnusualTime       AnomalyType = "unusual_time"
	AnomalyNewLocation       AnomalyType = "new_location"
)

// AnomalyResult represents the result of an anomaly detection check
type AnomalyResult struct {
	Detected   bool
	Type       AnomalyType
	Severity   string // low, medium, high, critical
	Details    string
	Confidence float64 // 0.0 - 1.0
}

// GeoService interface for looking up location data
type GeoService interface {
	GetLocation(ip string) (lat, lon float64, err error)
}

// AnomalyDetector detects anomalous behavior
type AnomalyDetector struct {
	geoService           GeoService
	maxSpeedKmH          float64
	fingerprintThreshold float64
}

// NewAnomalyDetector creates a new AnomalyDetector
func NewAnomalyDetector(geoService GeoService, maxSpeedKmH float64, fingerprintThreshold float64) *AnomalyDetector {
	if maxSpeedKmH == 0 {
		maxSpeedKmH = 900 // Default to ~plane speed
	}
	if fingerprintThreshold == 0 {
		fingerprintThreshold = 0.7
	}
	return &AnomalyDetector{
		geoService:           geoService,
		maxSpeedKmH:          maxSpeedKmH,
		fingerprintThreshold: fingerprintThreshold,
	}
}

// DetectGeoJump detects impossible travel between two locations
func (d *AnomalyDetector) DetectGeoJump(prevIP string, prevTime time.Time, currIP string, currTime time.Time) *AnomalyResult {
	if d.geoService == nil || prevIP == "" || currIP == "" {
		return &AnomalyResult{Detected: false}
	}

	lat1, lon1, err1 := d.geoService.GetLocation(prevIP)
	lat2, lon2, err2 := d.geoService.GetLocation(currIP)

	if err1 != nil || err2 != nil {
		// Log error?
		return &AnomalyResult{Detected: false}
	}

	distance := haversineDistance(lat1, lon1, lat2, lon2)
	timeDiff := currTime.Sub(prevTime).Hours()

	if timeDiff <= 0 {
		// Concurrent or time anomaly? Treat as instant travel if locations differ significantly
		if distance > 100 { // 100km
			return &AnomalyResult{
				Detected:   true,
				Type:       AnomalyGeoJump,
				Severity:   "high",
				Details:    "Distance > 100km with zero or negative time difference",
				Confidence: 1.0,
			}
		}
		return &AnomalyResult{Detected: false}
	}

	speed := distance / timeDiff

	if speed > d.maxSpeedKmH {
		return &AnomalyResult{
			Detected:   true,
			Type:       AnomalyGeoJump,
			Severity:   "high",
			Details:    "Speed exceeds threshold",
			Confidence: 1.0,
		}
	}

	return &AnomalyResult{Detected: false}
}

// DetectFingerprintChange detects significant changes in device fingerprint
func (d *AnomalyDetector) DetectFingerprintChange(oldFP, newFP map[string]interface{}) *AnomalyResult {
	// Core fields that shouldn't change often
	coreFields := []string{"screen_resolution", "timezone", "platform"}
	coreChanges := 0
	for _, field := range coreFields {
		v1, ok1 := oldFP[field]
		v2, ok2 := newFP[field]
		// Use simple string comparison for now.
		// In a real scenario, we might want deeper comparison.
		// For simplicity, converting to string via JSON (similar to device_fingerprint.go helper) or just string cast if possible.
		// Assuming simple types for these fields.
		if ok1 && ok2 && v1 != v2 {
			coreChanges++
		}
	}

	if coreChanges >= 2 {
		return &AnomalyResult{
			Detected:   true,
			Type:       AnomalyFingerprintChange,
			Severity:   "critical",
			Details:    "Multiple core fingerprint fields changed",
			Confidence: 0.9,
		}
	}

	// Minor fields
	minorFields := []string{"plugins", "fonts", "canvas_hash"}
	minorChanges := 0
	for _, field := range minorFields {
		v1, ok1 := oldFP[field]
		v2, ok2 := newFP[field]
		if ok1 && ok2 && v1 != v2 {
			minorChanges++
		}
	}

	if minorChanges >= 3 {
		return &AnomalyResult{
			Detected:   true,
			Type:       AnomalyFingerprintChange,
			Severity:   "medium",
			Details:    "Multiple minor fingerprint fields changed",
			Confidence: 0.7,
		}
	}

	return &AnomalyResult{Detected: false}
}

// DetectUnusualTime checks if access time is unusual for the user
func (d *AnomalyDetector) DetectUnusualTime(accessTime time.Time, historicalTimes []time.Time) *AnomalyResult {
	if len(historicalTimes) < 10 {
		return &AnomalyResult{Detected: false} // Not enough data
	}

	// Simplified logic: Check hour distribution
	hourCounts := make(map[int]int)
	total := 0
	for _, t := range historicalTimes {
		hourCounts[t.Hour()]++
		total++
	}

	currentHour := accessTime.Hour()
	count := hourCounts[currentHour]
	probability := float64(count) / float64(total)

	if probability < 0.05 { // Less than 5% chance
		return &AnomalyResult{
			Detected:   true,
			Type:       AnomalyUnusualTime,
			Severity:   "low",
			Details:    "Access time is unusual for this user",
			Confidence: 0.6,
		}
	}

	return &AnomalyResult{Detected: false}
}

// haversineDistance calculates distance in km
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
