package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/server/http"
	"github.com/turtacn/QuantaID/pkg/kms/local"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := utils.NewZapLogger(&utils.LoggerConfig{Level: "info", Format: "json"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	configManager, err := utils.NewConfigManager("./configs", "server", "yaml", logger)
	if err != nil {
		logger.Error(context.Background(), "Failed to load configuration", zap.Error(err))
		os.Exit(1)
	}
	var appCfg utils.Config
	if err := configManager.Unmarshal(&appCfg); err != nil {
		logger.Error(context.Background(), "Failed to unmarshal configuration", zap.Error(err))
		os.Exit(1)
	}

	// Initialize CryptoManager
	cryptoManager := utils.NewCryptoManager("your-jwt-secret") // Use a real secret from config in production

	// Initialize Data Encryption (KMS)
	if appCfg.DataEncryption.Key != "" {
		kmsProvider, err := local.New(appCfg.DataEncryption.Key)
		if err != nil {
			logger.Error(context.Background(), "Failed to initialize KMS", zap.Error(err))
			os.Exit(1)
		}
		types.SetGlobalKMS(kmsProvider)
	} else {
		// Warn if no key is provided, but allow running if FLE is not used (or for initial setup)
		// However, types.EncryptedString will fail if used.
		logger.Warn(context.Background(), "Data Encryption Key not configured. Encrypted fields will cause errors.")
	}

	// Create server with config
	httpCfg := http.Config{
		Address:      ":8080", // Get from config
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	server, err := http.NewServerWithConfig(httpCfg, &appCfg, logger, cryptoManager)
	if err != nil {
		logger.Error(context.Background(), "Failed to create server", zap.Error(err))
		os.Exit(1)
	}

	// Start audit log retention manager
	retentionManager := audit.NewRetentionManager(server.Services.AuditLogger.GetRepo(), logger.(*utils.ZapLogger).Logger)
	retentionManager.Start(context.Background(), appCfg.Audit.RetentionDays)

	// Start server in a goroutine
	go server.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	retentionManager.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Stop(ctx)
}
