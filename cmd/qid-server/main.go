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

	cryptoManager := utils.NewCryptoManager("my-super-secret-jwt-key")

	identityRepo := postgresql.NewInMemoryIdentityRepository()
	authDbRepo := postgresql.NewInMemoryAuthRepository()
	policyRepo := postgresql.NewInMemoryPolicyRepository()
	sessionRepo := redis.NewInMemorySessionRepository()
	tokenRepo := redis.NewInMemoryTokenRepository()

	identityDomainSvc := identity.NewService(identityRepo, identityRepo, cryptoManager, logger)
	authDomainSvc := auth.NewService(identityDomainSvc, sessionRepo, tokenRepo, authDbRepo, cryptoManager, logger)
	policyDomainSvc := policy.NewService(policyRepo, logger)

	identityAppSvc := identityservice.NewApplicationService(identityDomainSvc, logger)
	authAppSvc := authservice.NewApplicationService(authDomainSvc, logger, authservice.Config{
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		SessionDuration:      1 * time.Hour,
	})
	authzAppSvc := authorizationservice.NewApplicationService(policyDomainSvc, identityDomainSvc, logger)

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

	go httpServer.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(context.Background(), "Shutdown signal received, starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpServer.Stop(ctx)
	logger.Info(context.Background(), "Server gracefully stopped.")
}

//Personal.AI order the ending
