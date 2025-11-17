package adaptive

import (
	"context"
	"testing"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRiskEngine_Evaluate(t *testing.T) {
	mockGeoIP := new(MockGeoIPReader)
	mockGeoIP.On("City", mock.Anything).Return(&geoip2.City{}, nil)

	engine := &RiskEngine{
		geoIP: mockGeoIP,
		config: RiskConfig{
			Weights: RiskWeights{
				IPReputation: 0.5,
				GeoAnomaly:   0.5,
			},
		},
	}

	event := &AuthEvent{
		UserID:    "testuser",
		IPAddress: "127.0.0.1",
		Timestamp: time.Now(),
	}

	score, err := engine.Evaluate(context.Background(), event)
	assert.NoError(t, err)
	assert.NotNil(t, score)
	assert.Equal(t, 0.0, score.TotalScore)
}
