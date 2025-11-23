package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
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

// SaveLoginGeo stores the user's login location using Redis GEOADD and a Sorted Set for timeline.
// It stores the location with the timestamp as the member name in the Geo Set,
// and maintains a separate Sorted Set to index these timestamps for easy retrieval of the latest one.
func (m *GeoManager) SaveLoginGeo(ctx context.Context, userID string, lat, lon float64, timestamp time.Time) error {
	geoKey := fmt.Sprintf("user:geo:%s", userID)
	timelineKey := fmt.Sprintf("user:geo:timeline:%s", userID)
	tsStr := fmt.Sprintf("%d", timestamp.Unix())

	// 1. Add to Geo Set: GEOADD key lon lat member
	// Member is the timestamp string
	_, err := m.client.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Name:      tsStr,
		Longitude: lon,
		Latitude:  lat,
	})
	if err != nil {
		return fmt.Errorf("failed to add geo location: %w", err)
	}

	// 2. Add to Timeline Sorted Set: ZADD key score member
	// Score is the timestamp, member is the timestamp string
	err = m.client.ZAdd(ctx, timelineKey, redis.Z{
		Score:  float64(timestamp.Unix()),
		Member: tsStr,
	})
	if err != nil {
		return fmt.Errorf("failed to update geo timeline: %w", err)
	}

	// Set TTL to 30 days for both keys
	ttl := 30 * 24 * time.Hour
	m.client.Expire(ctx, geoKey, ttl)
	m.client.Expire(ctx, timelineKey, ttl)

	return nil
}

// GetLastLoginLocation retrieves the last login location for the user.
// Returns an error if not found.
func (m *GeoManager) GetLastLoginLocation(ctx context.Context, userID string) (*GeoLocation, error) {
	timelineKey := fmt.Sprintf("user:geo:timeline:%s", userID)
	geoKey := fmt.Sprintf("user:geo:%s", userID)

	// 1. Get the latest timestamp from the timeline ZSET
	// ZREVRANGE key 0 0 -> gets the member with the highest score
	result, err := m.client.ZRange(ctx, timelineKey, -1, -1)
	if err != nil {
		return nil, fmt.Errorf("failed to get last login timestamp: %w", err)
	}

	if len(result) == 0 {
		return nil, types.ErrNotFound
	}

	lastTsStr := result[0]

	// 2. Get the location from the Geo Set using the timestamp as the member name
	geoPos, err := m.client.GeoPos(ctx, geoKey, lastTsStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get geo position: %w", err)
	}

	if len(geoPos) == 0 || geoPos[0] == nil {
		// Inconsistency: found in timeline but not in geo set
		return nil, types.ErrNotFound
	}

	// Parse timestamp
	ts, err := strconv.ParseInt(lastTsStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format in storage: %w", err)
	}

	return &GeoLocation{
		Lat:       geoPos[0].Latitude,
		Lon:       geoPos[0].Longitude,
		Timestamp: time.Unix(ts, 0),
	}, nil
}
