package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	redis_storage "github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/types"
)

type PostgresRepoTestSuite struct {
	suite.Suite
	db             *gorm.DB
	pgContainer    testcontainers.Container
	redisContainer testcontainers.Container
	authRepo       auth.IdentityProviderRepository
	sessionRepo    *redis_storage.RedisSessionRepository
	redisClient    redis_storage.RedisClientInterface
}

func (s *PostgresRepoTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test in short mode")
	}
	ctx := context.Background()

	// Start PostgreSQL container
	pgReq := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pgReq,
		Started:          true,
	})
	s.Require().NoError(err)
	s.pgContainer = pgContainer

	pgHost, err := pgContainer.Host(ctx)
	s.Require().NoError(err)
	pgPort, err := pgContainer.MappedPort(ctx, "5432")
	s.Require().NoError(err)
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpassword dbname=testdb sslmode=disable", pgHost, pgPort.Port())
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	s.Require().NoError(err)
	s.db = db
	err = db.AutoMigrate(&types.IdentityProvider{}, &types.AuditLog{})
	s.Require().NoError(err)
	idpRepo, _ := postgresql.NewPostgresAuthRepository(s.db)
	s.authRepo = idpRepo

	// Start Redis container
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	s.Require().NoError(err)
	s.redisContainer = redisContainer

	redisHost, err := redisContainer.Host(ctx)
	s.Require().NoError(err)
	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	s.Require().NoError(err)

	redisMetrics := redis_storage.NewMetrics("test", prometheus.NewRegistry())
	redisClient, err := redis_storage.NewRedisClient(&redis_storage.RedisConfig{
		Host: redisHost,
		Port: redisPort.Int(),
	}, redisMetrics)
	s.Require().NoError(err)
	s.redisClient = redisClient

	sessionManager := redis_storage.NewSessionManager(
		redisClient,
		redis_storage.SessionConfig{DefaultTTL: 2 * time.Second}, // Short TTL for test
		zap.NewNop(),
		&redis_storage.GoogleUUIDGenerator{},
		&redis_storage.RealClock{},
		redisMetrics,
	)
	s.sessionRepo = redis_storage.NewRedisSessionRepository(redisClient, sessionManager)
}

func (s *PostgresRepoTestSuite) TearDownSuite() {
	if s.pgContainer != nil {
		if err := s.pgContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("could not stop postgres container: %v", err)
		}
	}
	if s.redisContainer != nil {
		if err := s.redisContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("could not stop redis container: %v", err)
		}
	}
}

func TestPostgresRepoTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresRepoTestSuite))
}

func (s *PostgresRepoTestSuite) TestPostgresAuthRepo_CreateProvider() {
	ctx := context.Background()
	provider := &types.IdentityProvider{
		ID:      uuid.New().String(),
		Name:    "test-provider",
		Type:    types.ProtocolOIDC,
		Enabled: true,
		Config:  json.RawMessage(`{"client_id": "test-client"}`),
	}

	err := s.authRepo.CreateProvider(ctx, provider)
	s.Require().NoError(err)

	// Verify
	retrieved, err := s.authRepo.GetProviderByID(ctx, provider.ID)
	s.Require().NoError(err)
	s.Require().NotNil(retrieved)
	s.Equal("test-provider", retrieved.Name)
	s.JSONEq(`{"client_id": "test-client"}`, string(retrieved.Config))
}

func (s *PostgresRepoTestSuite) TestPostgresAuthRepo_ListProviders() {
	ctx := context.Background()
	// Clear table
	s.db.Exec("DELETE FROM identity_providers")

	// Create two providers
	p1 := &types.IdentityProvider{ID: uuid.New().String(), Name: "p1", Type: types.ProtocolOIDC, Config: json.RawMessage(`{}`)}
	p2 := &types.IdentityProvider{ID: uuid.New().String(), Name: "p2", Type: types.ProtocolSAML, Config: json.RawMessage(`{}`)}
	s.Require().NoError(s.authRepo.CreateProvider(ctx, p1))
	s.Require().NoError(s.authRepo.CreateProvider(ctx, p2))

	providers, err := s.authRepo.ListProviders(ctx)
	s.Require().NoError(err)
	s.Len(providers, 2)
}

func (s *PostgresRepoTestSuite) TestPostgresAuthRepo_GetProvider_NotFound() {
	ctx := context.Background()
	_, err := s.authRepo.GetProviderByID(ctx, uuid.New().String())
	s.ErrorIs(err, types.ErrNotFound)
}

func (s *PostgresRepoTestSuite) TestRedisSession() {
	ctx := context.Background()
	userID := uuid.New().String()
	req := httptest.NewRequest("GET", "/", nil)

	// 1. CreateSession
	session, err := s.sessionRepo.CreateSession(ctx, userID, req)
	s.Require().NoError(err)
	s.Require().NotNil(session)

	// 2. GetSession & Verify TTL
	key := "session:" + session.ID
	ttl, err := s.redisClient.Client().TTL(ctx, key).Result()
	s.Require().NoError(err)
	s.InDelta(2*time.Second, ttl, float64(time.Millisecond*200))

	retrieved, err := s.sessionRepo.GetSession(ctx, session.ID, req)
	s.Require().NoError(err)
	s.Require().NotNil(retrieved)
	s.Equal(userID, retrieved.UserID)

	// 3. Wait for TTL to expire
	time.Sleep(3 * time.Second)
	_, err = s.sessionRepo.GetSession(ctx, session.ID, req)
	s.ErrorIs(err, types.ErrSessionExpired)
}
