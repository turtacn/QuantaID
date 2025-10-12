package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/protocols/saml"
	"github.com/turtacn/QuantaID/internal/server/http"
	applicationservice "github.com/turtacn/QuantaID/internal/services/application"
	authservice "github.com/turtacn/QuantaID/internal/services/auth"
	authorizationservice "github.com/turtacn/QuantaID/internal/services/authorization"
	identityservice "github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"math/big"
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
	cfgManager, err := utils.NewConfigManager("./configs", "server", "yaml", logger)
	if err != nil {
		logger.Error(context.Background(), "Failed to load configuration", zap.Error(err))
		os.Exit(1)
	}

	var cfg utils.Config
	if err := cfgManager.Unmarshal(&cfg); err != nil {
		logger.Error(context.Background(), "Failed to unmarshal configuration", zap.Error(err))
		os.Exit(1)
	}

	db, err := postgresql.NewConnection(cfg.Postgres)
	if err != nil {
		logger.Error(context.Background(), "Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}

	if err := postgresql.AutoMigrate(db); err != nil {
		logger.Error(context.Background(), "Failed to auto-migrate database", zap.Error(err))
		os.Exit(1)
	}

	identityRepo := postgresql.NewPostgresIdentityRepository(db)
	policyRepo := postgresql.NewPostgresPolicyRepository(db)
	appRepo := postgresql.NewPostgresApplicationRepository(db)
	auditRepo := postgresql.NewPostgresAuditLogRepository(db)
	sessionRepo := redis.NewInMemorySessionRepository()
	tokenRepo := redis.NewInMemoryTokenRepository()

	// --- Domain Services Setup ---
	// These services contain the core business logic.
	identityDomainSvc := identity.NewService(identityRepo, identityRepo, cryptoManager, logger)
	authDomainSvc := auth.NewService(identityDomainSvc, sessionRepo, tokenRepo, auditRepo, cryptoManager, logger)
	policyDomainSvc := policy.NewService(policyRepo, logger)

	// --- SAML Service Setup ---
	// In a real application, these would be loaded from a secure config/store.
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	certTmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "quantid.dev"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageCertSign,
	}
	certBytes, _ := x509.CreateCertificate(rand.Reader, &certTmpl, &certTmpl, &key.PublicKey, key)
	idpCert, _ := x509.ParseCertificate(certBytes)

	samlSvc, err := saml.NewService(logger, appRepo, identityDomainSvc, cryptoManager, key, idpCert, "http://localhost:8080/saml/metadata")
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize SAML service", zap.Error(err))
		os.Exit(1)
	}

	// --- Application Services Setup ---
	// These services act as a facade over the domain layer for the transport layer.
	identityAppSvc := identityservice.NewApplicationService(identityDomainSvc, logger)
	appAppSvc := applicationservice.NewApplicationService(appRepo, logger, cryptoManager)
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
		AppService:      appAppSvc,  // New application service
		SamlService:     samlSvc,    // New SAML service
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
