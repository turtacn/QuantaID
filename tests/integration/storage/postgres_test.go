package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/types"
)

type PostgresRepoTestSuite struct {
	suite.Suite
	db        *gorm.DB
	container testcontainers.Container
	authRepo  *postgresql.PostgresAuthRepository
}

func (s *PostgresRepoTestSuite) SetupSuite() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
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
	port, err := container.MappedPort(ctx, "5432")
	s.Require().NoError(err)

	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpassword dbname=testdb sslmode=disable", host, port.Port())
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	s.Require().NoError(err)
	s.db = db

	// Run migrations
	err = db.AutoMigrate(&types.IdentityProvider{}, &types.AuditLog{})
	s.Require().NoError(err)

	s.authRepo = postgresql.NewPostgresAuthRepository(s.db)
}

func (s *PostgresRepoTestSuite) TearDownSuite() {
	if s.container != nil {
		if err := s.container.Terminate(context.Background()); err != nil {
			log.Fatalf("could not stop container: %v", err)
		}
	}
}

func TestPostgresRepoTestSuite(t *testing.T) {
	requireDocker(t)
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
