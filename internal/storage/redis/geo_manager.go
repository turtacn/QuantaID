package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/turtacn/QuantaID/pkg/types"
)

// GeoLocation represents a user's location at a specific time.
type GeoLocation struct {
	Lat       float64
	Lon       float64
	Timestamp time.Time
}

// GeoManager handles the storage and retrieval of user geographic login history.
type GeoManager struct {
	client RedisClientInterface
}

// NewGeoManager creates a new GeoManager instance.
func NewGeoManager(client RedisClientInterface) *GeoManager {
	return &GeoManager{
		client: client,
	}
}

// SaveLoginGeo stores the user's login location with a TTL of 7 days.
func (m *GeoManager) SaveLoginGeo(ctx context.Context, userID string, lat, lon float64, timestamp time.Time) error {
	key := fmt.Sprintf("user:geo:%s", userID)
	data := map[string]interface{}{
		"lat":       lat,
		"lon":       lon,
		"timestamp": timestamp.Unix(),
	}

	// Use HMSet to store the struct fields
	if err := m.client.HMSet(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("failed to save geo location: %w", err)
	}

	// Set TTL to 7 days
	if err := m.client.Expire(ctx, key, 7*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set TTL for geo location: %w", err)
	}

	return nil
}

// GetLastLoginGeo retrieves the last login location for the user.
// Returns an error if not found.
func (m *GeoManager) GetLastLoginGeo(ctx context.Context, userID string) (*GeoLocation, error) {
	key := fmt.Sprintf("user:geo:%s", userID)

	result, err := m.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get geo location: %w", err)
	}

	if len(result) == 0 {
		return nil, types.ErrNotFound
	}

	lat, err := strconv.ParseFloat(result["lat"], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid lat: %w", err)
	}

	lon, err := strconv.ParseFloat(result["lon"], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid lon: %w", err)
	}

	ts, err := strconv.ParseInt(result["timestamp"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	return &GeoLocation{
		Lat:       lat,
		Lon:       lon,
		Timestamp: time.Unix(ts, 0),
	}, nil
}
