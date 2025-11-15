package redis

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

type realUUIDGenerator struct{}

func (g *realUUIDGenerator) New() string {
	return uuid.New().String()
}

type realClock struct{}

func (c *realClock) Now() time.Time {
	return time.Now()
}

func setupSessionManagerForBenchmark(b *testing.B) *SessionManager {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		b.Fatalf("could not start redis container: %s", err)
	}
	b.Cleanup(func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			b.Fatalf("could not stop redis container: %s", err)
		}
	})

	host, err := redisContainer.Host(ctx)
	if err != nil {
		b.Fatalf("could not get redis host: %s", err)
	}
	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		b.Fatalf("could not get redis port: %s", err)
	}

	redisConfig := &RedisConfig{
		Host: host,
		Port: port.Int(),
	}

	reg := prometheus.NewRegistry()
	metrics := NewMetrics("benchmark", reg)

	client, err := NewRedisClient(redisConfig, metrics)
	if err != nil {
		b.Fatalf("could not create redis client: %s", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Can't initialize zap logger: %v", err)
	}

	sessionConfig := SessionConfig{
		DefaultTTL: time.Hour,
	}

	return NewSessionManager(client, sessionConfig, logger, &realUUIDGenerator{}, &realClock{}, metrics)
}

func BenchmarkSessionCreation(b *testing.B) {
	sm := setupSessionManagerForBenchmark(b)
	req, _ := http.NewRequest("GET", "/", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := sm.CreateSession(context.Background(), fmt.Sprintf("user-%s", uuid.New().String()), req)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

func BenchmarkSessionRetrieval(b *testing.B) {
	sm := setupSessionManagerForBenchmark(b)
	req, _ := http.NewRequest("GET", "/", nil)
	session, err := sm.CreateSession(context.Background(), "user-benchmark", req)
	if err != nil {
		b.Fatalf("Failed to create session for benchmark: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := sm.GetSession(context.Background(), session.ID, req)
			if err != nil {
				b.Error(err)
			}
		}
	})
}
