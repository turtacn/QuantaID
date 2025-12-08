package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/auth/device"
	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// MockGeoService implements device.GeoService
type MockGeoService struct {
	Locations map[string][2]float64
}

func (m *MockGeoService) GetLocation(ip string) (float64, float64, error) {
	if loc, ok := m.Locations[ip]; ok {
		return loc[0], loc[1], nil
	}
	return 0, 0, nil
}

func TestTrustScorer_NewDevice_LowScore(t *testing.T) {
	config := utils.DeviceTrustConfig{
		BaseScore: 20,
	}
	scorer := device.NewTrustScorer(config)

	d := &models.Device{
		CreatedAt: time.Now(),
	}

	score := scorer.CalculateScore(d)
	assert.LessOrEqual(t, score, 30)
	assert.GreaterOrEqual(t, score, 20)
}

func TestTrustScorer_OldDevice_HighScore(t *testing.T) {
	config := utils.DeviceTrustConfig{
		BaseScore:   20,
		AgeBonus:    1,
		MaxAgeBonus: 30,
	}
	scorer := device.NewTrustScorer(config)

	d := &models.Device{
		CreatedAt: time.Now().AddDate(0, 0, -45), // 45 days old
	}

	score := scorer.CalculateScore(d)
	// Base 20 + MaxAge 30 = 50
	assert.Equal(t, 50, score)
}

func TestTrustScorer_BoundDevice_Bonus(t *testing.T) {
	config := utils.DeviceTrustConfig{
		BaseScore:  20,
		BoundBonus: 20,
	}
	scorer := device.NewTrustScorer(config)

	now := time.Now()
	d := &models.Device{
		CreatedAt: now,
		UserID:    "user1",
		BoundAt:   &now,
	}

	score := scorer.CalculateScore(d)
	// Base 20 + Bound 20 = 40
	assert.Equal(t, 40, score)
}

func TestAnomalyDetector_GeoJump_Detected(t *testing.T) {
	geo := &MockGeoService{
		Locations: map[string][2]float64{
			"1.1.1.1": {39.9042, 116.4074}, // Beijing
			"2.2.2.2": {40.7128, -74.0060}, // New York
		},
	}
	detector := device.NewAnomalyDetector(geo, 900, 0.7)

	prevIP := "1.1.1.1"
	currIP := "2.2.2.2"
	now := time.Now()

	// 1 hour difference, impossible to travel Beijing -> NY
	prevTime := now.Add(-1 * time.Hour)
	currTime := now

	result := detector.DetectGeoJump(prevIP, prevTime, currIP, currTime)
	assert.True(t, result.Detected)
	assert.Equal(t, device.AnomalyGeoJump, result.Type)
}

func TestAnomalyDetector_FingerprintChange_Partial(t *testing.T) {
	detector := device.NewAnomalyDetector(nil, 900, 0.7)

	oldFP := map[string]interface{}{
		"screen_resolution": "1920x1080",
		"timezone":          "Asia/Shanghai",
		"plugins":           "pluginA,pluginB",
	}
	newFP := map[string]interface{}{
		"screen_resolution": "1920x1080",
		"timezone":          "Asia/Shanghai",
		"plugins":           "pluginA,pluginC", // Changed
	}

	result := detector.DetectFingerprintChange(oldFP, newFP)
	assert.False(t, result.Detected)
}

func TestAnomalyDetector_FingerprintChange_Major(t *testing.T) {
	detector := device.NewAnomalyDetector(nil, 900, 0.7)

	oldFP := map[string]interface{}{
		"screen_resolution": "1920x1080",
		"timezone":          "Asia/Shanghai",
		"platform":          "Windows",
	}
	newFP := map[string]interface{}{
		"screen_resolution": "2560x1440", // Changed
		"timezone":          "America/New_York", // Changed
		"platform":          "MacOS", // Changed
	}

	result := detector.DetectFingerprintChange(oldFP, newFP)
	assert.True(t, result.Detected)
	assert.Equal(t, "critical", result.Severity)
}
