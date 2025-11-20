package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDatabase(t *testing.T) *gorm.DB {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)
	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, "test", "password", "testdb", port.Port())
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations (simplified for test)
	err = db.AutoMigrate(&audit.AuditEvent{})
	require.NoError(t, err)

	return db
}

func TestE2E_AuditLog_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	db := setupTestDatabase(t)
	repo := postgresql.NewPostgresAuditLogRepository(db)

	// Create a real logger that writes to the test DB
	// We use a small flush interval for the test to get results quickly
	logger := audit.NewAuditLogger(repo, zap.NewNop(), 5, 200*time.Millisecond, 20)
	defer logger.Shutdown()

	// 1. Perform an action that should be audited
	eventToRecord := &audit.AuditEvent{
		EventType: audit.EventLoginFailure,
		Actor:     audit.Actor{ID: "user-123", Type: "user"},
		Action:    "login",
		Result:    audit.ResultFailure,
		Metadata:  map[string]interface{}{"reason": "invalid_password"},
	}
	logger.Record(context.Background(), eventToRecord)

	// 2. Wait for the logger to flush the event to the database
	time.Sleep(300 * time.Millisecond)

	// 3. Query the database to verify the audit log was written correctly
	events, err := repo.Query(context.Background(), audit.QueryFilter{ActorID: "user-123"})
	require.NoError(t, err)

	require.Len(t, events, 1)
	recordedEvent := events[0]

	assert.Equal(t, audit.EventLoginFailure, recordedEvent.EventType)
	assert.Equal(t, "user-123", recordedEvent.Actor.ID)
	assert.Equal(t, audit.ResultFailure, recordedEvent.Result)
	assert.Equal(t, "invalid_password", recordedEvent.Metadata["reason"])
}
