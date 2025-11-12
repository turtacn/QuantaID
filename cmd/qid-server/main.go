package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/turtacn/QuantaID/internal/domain/auth"
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
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize router
	router := http.NewServeMux()

	// Register handlers
	RegisterOAuthHandlers(router, logger)

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
