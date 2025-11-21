//go:build integration
package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/utils"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start Postgres Container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:13-alpine"),
		postgres.WithDatabase("quantaid_test"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		fmt.Printf("Failed to start postgres container: %s\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			fmt.Printf("Failed to terminate postgres container: %s\n", err)
		}
	}()

	pgHost, err := pgContainer.Host(ctx)
	if err != nil {
		fmt.Printf("Failed to get postgres host: %s\n", err)
		os.Exit(1)
	}
	pgPort, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Printf("Failed to get postgres port: %s\n", err)
		os.Exit(1)
	}
	os.Setenv("QID_POSTGRES_HOST", pgHost)
	os.Setenv("QID_POSTGRES_PORT", pgPort.Port())
	os.Setenv("QID_POSTGRES_USER", "user")
	os.Setenv("QID_POSTGRES_PASSWORD", "password")
	os.Setenv("QID_POSTGRES_DBNAME", "quantaid_test")

	// Run migrations
	db, err := postgresql.NewConnection(utils.PostgresConfig{
		Host:            pgHost,
		Port:            pgPort.Int(),
		User:            "user",
		Password:        "password",
		DbName:          "quantaid_test",
		SSLMode:         "disable",
		ConnMaxLifetime: "1h",
	})
	if err != nil {
		fmt.Printf("Failed to connect to test postgres for migration: %s\n", err)
		os.Exit(1)
	}
	if err := postgresql.AutoMigrate(db); err != nil {
		fmt.Printf("Failed to run auto-migration: %s\n", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	sqlDB.Close()

	// 2. Start Redis Container
	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("redis:5-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		fmt.Printf("Failed to start redis container: %s\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			fmt.Printf("Failed to terminate redis container: %s\n", err)
		}
	}()

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		fmt.Printf("Failed to get redis host: %s\n", err)
		os.Exit(1)
	}
	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		fmt.Printf("Failed to get redis port: %s\n", err)
		os.Exit(1)
	}
	os.Setenv("QID_REDIS_HOST", redisHost)
	os.Setenv("QID_REDIS_PORT", redisPort.Port())

	// Set other env vars for the application to use
	os.Setenv("QID_STORAGE_MODE", "postgres")
	os.Setenv("QID_LOG_LEVEL", "error") // Keep logs quiet during tests

	// 4. Run Tests
	exitCode := m.Run()

	// 5. Exit
	os.Exit(exitCode)
}

func waitForServer(url string, timeout time.Duration) error {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	startTime := time.Now()
	for {
		if time.Since(startTime) > timeout {
			return fmt.Errorf("server did not start within the timeout period")
		}
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}
