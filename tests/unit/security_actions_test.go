package unit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/security/actions"
	"github.com/turtacn/QuantaID/internal/security/automator"
	appredis "github.com/turtacn/QuantaID/internal/storage/redis"
	"go.uber.org/zap"
)

func Test_BlockIP_Action(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub redis connection", err)
	}
	defer s.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	// Wrap the mock client in the application's client wrapper to satisfy the interface.
	appRedisClient := appredis.NewRedisClientWrapper(redisClient)
	action := actions.NewBlockIPAction(appRedisClient)
	input := automator.ActionInput{IP: "192.168.1.1"}

	err = action.Execute(context.Background(), input)
	assert.NoError(t, err)

	key := "security:blacklist:ip:192.168.1.1"
	exists, err := redisClient.Exists(context.Background(), key).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	ttl, err := redisClient.TTL(context.Background(), key).Result()
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(time.Hour), time.Now().Add(ttl), 5*time.Second)
}

func Test_Automator_PolicyMatch(t *testing.T) {
	cfg := &config.SecurityConfig{
		ResponsePolicies: []config.ResponsePolicy{
			{
				Name:          "Block High-Risk IP",
				RiskThreshold: 0.9,
				ActionIDs:     []string{"block_ip"},
				IsBlocking:    true,
			},
			{
				Name:          "Notify Admin",
				RiskThreshold: 0.9,
				ActionIDs:     []string{"notify_admin"},
				IsBlocking:    false,
			},
		},
	}

	logger, _ := zap.NewDevelopment()
	engine := automator.NewEngine(cfg, logger)

	// Mock action
	mockAction := &mockSecurityAction{id: "block_ip"}
	mockNotifyAction := &mockSecurityAction{id: "notify_admin"}
	engine.RegisterAction(mockAction)
	engine.RegisterAction(mockNotifyAction)

	input := automator.ActionInput{RiskScore: 0.95}
	isBlocking, err := engine.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.True(t, isBlocking)
	assert.True(t, mockAction.executed)
	assert.True(t, mockNotifyAction.executed)
}

type mockSecurityAction struct {
	id       string
	executed bool
}

func (m *mockSecurityAction) ID() string {
	return m.id
}

func (m *mockSecurityAction) Execute(ctx context.Context, input automator.ActionInput) error {
	m.executed = true
	return nil
}
