package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/protocols/saml"
	"github.com/turtacn/QuantaID/internal/services/application"
	auth_service "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	identity_service "github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Server encapsulates the HTTP server for the QuantaID API.
type Server struct {
	httpServer *http.Server
	Router     *mux.Router
	logger     utils.Logger
}

// Config holds the configuration required for the HTTP server.
type Config struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Services is a container for all the application services.
type Services struct {
	AuthService     *auth_service.ApplicationService
	IdentityService *identity_service.ApplicationService
	AuthzService    *authorization.ApplicationService
	AppService      *application.ApplicationService
	SamlService     *saml.Service
	CryptoManager   *utils.CryptoManager
}

// NewServer creates a new HTTP server instance.
func NewServer(config Config, logger utils.Logger, services Services) *Server {
	router := mux.NewRouter()
	server := &Server{
		Router: router,
		logger: logger,
		httpServer: &http.Server{
			Addr:         config.Address,
			Handler:      router,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
		},
	}
	server.registerRoutes(services)
	return server
}

// NewServerWithConfig creates a new server instance based on the provided configuration.
func NewServerWithConfig(httpCfg Config, appCfg *utils.Config, logger utils.Logger, cryptoManager *utils.CryptoManager) (*Server, error) {
	var idRepo identity.UserRepository
	var groupRepo identity.GroupRepository
	var sessionRepo auth.SessionRepository
	var tokenRepo auth.TokenRepository
	var auditRepo auth.AuditLogRepository
	// var policyRepo policy.Repository // Not used yet

	switch appCfg.Storage.Mode {
	case "memory":
		logger.Info(context.Background(), "Using in-memory storage backend")
		memIdRepo := memory.NewIdentityMemoryRepository()
		idRepo = memIdRepo
		groupRepo = memIdRepo

		memAuthRepo := memory.NewAuthMemoryRepository()
		sessionRepo = memAuthRepo
		tokenRepo = memAuthRepo
		auditRepo = memAuthRepo

		// policyRepo = memory.NewPolicyMemoryRepository()
	case "postgres":
		logger.Info(context.Background(), "Using PostgreSQL storage backend")
		db, err := postgresql.NewConnection(appCfg.Postgres)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}
		pgIdRepo := postgresql.NewPostgresIdentityRepository(db)
		idRepo = pgIdRepo
		groupRepo = pgIdRepo
		return nil, fmt.Errorf("postgres repositories for auth and policy are not yet implemented")
	default:
		return nil, fmt.Errorf("invalid storage mode: %s", appCfg.Storage.Mode)
	}

	identityDomainService := identity.NewService(idRepo, groupRepo, cryptoManager, logger)
	identityAppService := identity_service.NewApplicationService(identityDomainService, logger)

	authDomainService := auth.NewService(identityDomainService, sessionRepo, tokenRepo, auditRepo, cryptoManager, logger)

	// Create a no-op tracer for now.
	tracer := trace.NewNoopTracerProvider().Tracer("quantid-test")

	authAppService := auth_service.NewApplicationService(authDomainService, logger, auth_service.Config{
		AccessTokenDuration:  time.Hour, // Placeholder values
		RefreshTokenDuration: time.Hour * 24,
		SessionDuration:      time.Hour * 24,
	}, tracer)

	services := Services{
		IdentityService: identityAppService,
		AuthService:     authAppService,
		CryptoManager:   cryptoManager,
	}

	return NewServer(httpCfg, logger, services), nil
}

// registerRoutes sets up the API routes, their handlers, and associated middleware.
func (s *Server) registerRoutes(services Services) {
	authHandlers := handlers.NewAuthHandlers(services.AuthService, s.logger)
	identityHandlers := handlers.NewIdentityHandlers(services.IdentityService, s.logger)

	loggingMiddleware := middleware.NewLoggingMiddleware(s.logger)
	// Auth middleware would require AuthzService, which is not fully wired yet.
	// authMiddleware := middleware.NewAuthMiddleware(services.AuthzService, services.CryptoManager, s.logger)

	s.Router.Use(loggingMiddleware.Execute)

	apiV1 := s.Router.PathPrefix("/api/v1").Subrouter()
	apiV1.HandleFunc("/auth/login", authHandlers.Login).Methods("POST")

	// For simplicity, we are not protecting routes yet.
	apiV1.HandleFunc("/users", identityHandlers.CreateUser).Methods("POST")
	apiV1.HandleFunc("/users/{id}", identityHandlers.GetUser).Methods("GET")

	s.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
}


// Start begins listening for and serving HTTP requests.
func (s *Server) Start() {
	s.logger.Info(context.Background(), "Starting HTTP server", zap.String("address", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error(context.Background(), "HTTP server failed to start", zap.Error(err))
	}
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) {
	s.logger.Info(ctx, "Shutting down HTTP server")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error(ctx, "HTTP server graceful shutdown failed", zap.Error(err))
	}
}
