package main

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/server/http"
	authservice "github.com/turtacn/QuantaID/internal/services/auth"
	authorizationservice "github.com/turtacn/QuantaID/internal/services/authorization"
	identityservice "github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/utils"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// main is the entry point for the QuantaID server.
// It is responsible for the complete setup and teardown of the application.
// This includes:
// - Initializing configuration and utilities (logger, crypto).
// - Setting up the persistence layer (in-memory repositories in this case).
// - Wiring together the domain and application services with their dependencies.
// - Creating and starting the HTTP server.
// - Handling graceful shutdown on interrupt signals (SIGINT, SIGTERM).
func main() {
	loggerConfig := &utils.LoggerConfig{
		Level:   "debug",
		Format:  "console",
		Console: utils.ConsoleConfig{Enabled: true},
	}
	logger, err := utils.NewZapLogger(loggerConfig)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Initialize cryptographic utilities with a secret key.
	// In a real application, this should come from a secure configuration source.
	cryptoManager := utils.NewCryptoManager("my-super-secret-jwt-key")

	// --- Persistence Layer Setup ---
	// For this example, we are using in-memory repositories.
	// In a production environment, these would be replaced with actual database-backed implementations.
	identityRepo := postgresql.NewInMemoryIdentityRepository()
	authDbRepo := postgresql.NewInMemoryAuthRepository()
	policyRepo := postgresql.NewInMemoryPolicyRepository()
	sessionRepo := redis.NewInMemorySessionRepository()
	tokenRepo := redis.NewInMemoryTokenRepository()

	// --- Domain Services Setup ---
	// These services contain the core business logic.
	identityDomainSvc := identity.NewService(identityRepo, identityRepo, cryptoManager, logger)
	authDomainSvc := auth.NewService(identityDomainSvc, sessionRepo, tokenRepo, authDbRepo, cryptoManager, logger)
	policyDomainSvc := policy.NewService(policyRepo, logger)

	// --- Application Services Setup ---
	// These services act as a facade over the domain layer for the transport layer.
	identityAppSvc := identityservice.NewApplicationService(identityDomainSvc, logger)
	authAppSvc := authservice.NewApplicationService(authDomainSvc, logger, authservice.Config{
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		SessionDuration:      1 * time.Hour,
	})
	authzAppSvc := authorizationservice.NewApplicationService(policyDomainSvc, identityDomainSvc, logger)

	// --- HTTP Server Setup ---
	serverConfig := http.Config{
		Address:      ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	services := http.Services{
		AuthService:     authAppSvc,
		IdentityService: identityAppSvc,
		AuthzService:    authzAppSvc,
		CryptoManager:   cryptoManager,
	}
	httpServer := http.NewServer(serverConfig, logger, services)

	// --- Server Lifecycle Management ---
	// Start the server in a separate goroutine so it doesn't block.
	go httpServer.Start()

	// Wait for an interrupt signal (e.g., Ctrl+C) to gracefully shut down.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(context.Background(), "Shutdown signal received, starting graceful shutdown...")

	// Create a context with a timeout to allow for a graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpServer.Stop(ctx)
	logger.Info(context.Background(), "Server gracefully stopped.")
}
