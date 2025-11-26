package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/security/actions"
	"github.com/turtacn/QuantaID/internal/security/automator"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	appredis "github.com/turtacn/QuantaID/internal/storage/redis"
	"go.uber.org/zap"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupMiniRedis() (*miniredis.Miniredis, appredis.RedisClientInterface) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	// Wrap the mock client in the application's client wrapper to satisfy the interface.
	return s, appredis.NewRedisClientWrapper(client)
}

func Test_HighRisk_Login_Flow(t *testing.T) {
	// 1. Setup
	s, redisClient := setupMiniRedis()
	defer s.Close()

	cfg := &config.SecurityConfig{
		ResponsePolicies: []config.ResponsePolicy{
			{
				Name:          "Block High-Risk IP",
				RiskThreshold: 0.9,
				ActionIDs:     []string{"block_ip"},
				IsBlocking:    true,
			},
		},
	}

	logger, _ := zap.NewDevelopment()
	engine := automator.NewEngine(cfg, logger)
	blockIPAction := actions.NewBlockIPAction(redisClient)
	engine.RegisterAction(blockIPAction)

	// 2. Simulate High-Risk Login
	input := automator.ActionInput{IP: "1.2.3.4", RiskScore: 0.95}
	isBlocking, err := engine.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, isBlocking)

	// 3. Verify IP is Blocked
	blacklistMiddleware := middleware.IPBlacklistMiddleware(redisClient)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	testHandler := blacklistMiddleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4"
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	// 4. Verify a non-blocked IP is allowed
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "5.6.7.8"
	rr = httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
