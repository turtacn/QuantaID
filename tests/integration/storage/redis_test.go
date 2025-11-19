package storage

import (
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	redis_storage "github.com/turtacn/QuantaID/internal/storage/redis"
)

type RedisRepoTestSuite struct {
	suite.Suite
	client    *redis.Client
	container testcontainers.Container
	sessionRepo *redis_storage.RedisSessionRepository
}

func (s *RedisRepoTestSuite) SetupSuite() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("could not start container: %v", err)
	}
	s.container = container

	host, err := container.Host(ctx)
	s.Require().NoError(err)
	port, err := container.MappedPort(ctx, "6379")
	s.Require().NoError(err)

	s.client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	// Use a real Redis client for the repository
	realRedisClient := redis_storage.NewRedisClientWrapper(s.client)
	sessionManager := redis_storage.NewSessionManager(
		realRedisClient,
		redis_storage.SessionConfig{DefaultTTL: 2 * time.Second},
		zap.NewNop(),
		&redis_storage.GoogleUUIDGenerator{},
		&redis_storage.RealClock{},
		nil, // No metrics for this test
	)
	s.sessionRepo = redis_storage.NewRedisSessionRepository(realRedisClient, sessionManager)
}

func (s *RedisRepoTestSuite) TearDownSuite() {
	if s.container != nil {
		if err := s.container.Terminate(context.Background()); err != nil {
			log.Fatalf("could not stop container: %v", err)
		}
	}
}

func TestRedisRepoTestSuite(t *testing.T) {
	requireDocker(t)
	suite.Run(t, new(RedisRepoTestSuite))
}

func (s *RedisRepoTestSuite) TestRedisSession() {
	ctx := context.Background()
	userID := "test-user"
	req := httptest.NewRequest("GET", "/", nil)

	// 1. CreateSession
	session, err := s.sessionRepo.CreateSession(ctx, userID, req)
	s.Require().NoError(err)
	s.Require().NotNil(session)

	// Verify in Redis
	key := "session:" + session.ID
	data, err := s.client.Get(ctx, key).Result()
	s.Require().NoError(err)
	s.NotEmpty(data)

	// 2. GetSession
	retrieved, err := s.sessionRepo.GetSession(ctx, session.ID, req)
	s.Require().NoError(err)
	s.Equal(session.ID, retrieved.ID)
	s.Equal(userID, retrieved.UserID)

	// 3. Wait for TTL to expire
	time.Sleep(3 * time.Second)
	_, err = s.client.Get(ctx, key).Result()
	s.Equal(redis.Nil, err)
}
