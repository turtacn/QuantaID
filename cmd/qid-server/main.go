package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/plugins/mfa/totp"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := utils.NewZapLogger(&utils.LoggerConfig{
		Level: "info",
		Format: "json",
	})
	if err != nil {
		// If logger fails to initialize, we can't use it. Fall back to standard log.
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Initialize router
	router := http.NewServeMux()

	// Initialize repositories
	db, err := postgresql.NewConnection(utils.PostgresConfig{
		DSN: "postgres://user:password@localhost:5432/quantid?sslmode=disable",
	})
	if err != nil {
		logger.Error(context.Background(), "Failed to connect to database", zap.Error(err))
		return
	}
	userRepo := postgresql.NewPostgresIdentityRepository(db)
	auditRepo := postgresql.NewPostgresAuditLogRepository(db)

	// Register handlers
	RegisterOAuthHandlers(router, logger)
	handlers.RegisterAdminHandlers(router, userRepo, auditRepo)

	// Initialize CryptoManager
	cryptoManager := utils.NewCryptoManager("your-jwt-secret")

	// In a real application, you would initialize the MFA policy with its dependencies.
	totpProvider := &totp.TOTPProvider{}
	mfaPolicy := auth.NewMFAPolicy(nil, nil, totpProvider)
	RegisterMFAHandlers(router, mfaPolicy, logger, cryptoManager)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Info(context.Background(), fmt.Sprintf("Starting server on port %s", port))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		logger.Error(context.Background(), "Failed to start server", zap.Error(err))
	}
}
