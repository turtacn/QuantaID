package strategies

import (
	"context"
	"math"
	"time"

	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/internal/storage/redis"
)

// Strategy defines the interface for risk calculation strategies.
type Strategy interface {
	CalculateRisk(ctx context.Context, userID string, currentLat, currentLon float64) (float64, error)
}

const MaxTravelSpeedKmH = 800.0

// ImpossibleTravelStrategy detects impossible travel between logins.
type ImpossibleTravelStrategy struct {
	geoRepo *redis.GeoManager
}

// NewImpossibleTravelStrategy creates a new ImpossibleTravelStrategy.
func NewImpossibleTravelStrategy(geoRepo *redis.GeoManager) *ImpossibleTravelStrategy {
	return &ImpossibleTravelStrategy{
		geoRepo: geoRepo,
	}
}

// CalculateRisk calculates the risk based on travel speed.
func (s *ImpossibleTravelStrategy) CalculateRisk(ctx context.Context, userID string, currentLat, currentLon float64) (float64, error) {
	lastGeo, err := s.geoRepo.GetLastLoginLocation(ctx, userID)
	if err != nil {
		if err == types.ErrNotFound {
			return 0.0, nil // First login or no history, no risk
		}
		return 0.0, err
	}

	distance := haversine(lastGeo.Lat, lastGeo.Lon, currentLat, currentLon)
	duration := time.Since(lastGeo.Timestamp).Hours()

	if duration <= 0 {
		// If duration is 0 (same second), check distance.
		if distance > 10 { // > 10km in 0 time is definitely suspicious
			return 1.0, nil // Teleportation
		}
		return 0.0, nil
	}

	speed := distance / duration
	if speed > MaxTravelSpeedKmH {
		return 1.0, nil
	}

	return 0.0, nil
}

// haversine calculates the great-circle distance between two points in km.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	lat1Rad := lat1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
