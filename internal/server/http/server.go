package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/auth/adaptive"
	"github.com/turtacn/QuantaID/internal/auth/mfa"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/internal/orchestrator"
	"github.com/turtacn/QuantaID/internal/protocols/saml"
	"github.com/turtacn/QuantaID/internal/services/application"
	audit_service "github.com/turtacn/QuantaID/internal/services/audit"
	auth_service "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	identity_service "github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/internal/services/platform"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
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
	AuthService           *auth_service.ApplicationService
	IdentityService       *identity_service.ApplicationService
	AuthzService          *authorization.Service
	AppService            *application.ApplicationService
	SamlService           *saml.Service
	CryptoManager         *utils.CryptoManager
	IdentityDomainService identity.IService
	DevCenterService      *platform.DevCenterService
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

type PolicyConfig struct {
	Rules []authorization.Rule `yaml:"rules"`
}

// NewServerWithConfig creates a new server instance based on the provided configuration.
func NewServerWithConfig(httpCfg Config, appCfg *utils.Config, logger utils.Logger, cryptoManager *utils.CryptoManager) (*Server, error) {
	var idRepo identity.UserRepository
	var groupRepo identity.GroupRepository
	var sessionRepo auth.SessionRepository
	var tokenRepo auth.TokenRepository

	switch appCfg.Storage.Mode {
	case "memory":
		logger.Info(context.Background(), "Using in-memory storage backend")
		memIdRepo := memory.NewIdentityMemoryRepository()
		idRepo = memIdRepo
		groupRepo = memIdRepo
		memAuthRepo := memory.NewAuthMemoryRepository()
		sessionRepo = memAuthRepo
		tokenRepo = memAuthRepo
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

	// Authorization Service Setup
	policyData, err := os.ReadFile("configs/policy/basic.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}
	var policyCfg PolicyConfig
	if err := yaml.Unmarshal(policyData, &policyCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy data: %w", err)
	}
	// Audit Service Setup
	auditCfg, err := config.LoadAuditConfig("configs/audit/pipeline.jules.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load audit config: %w", err)
	}
	zapLogger, _ := zap.NewProduction()
	auditPipeline, err := config.NewPipelineFromConfig(auditCfg, zapLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit pipeline: %w", err)
	}
	auditService := audit_service.NewService(auditPipeline)

	evaluator := authorization.NewDefaultEvaluator(policyCfg.Rules)
	authzService := authorization.NewService(evaluator, auditService)

	identityDomainService := identity.NewService(idRepo, groupRepo, cryptoManager, logger)
	identityAppService := identity_service.NewApplicationService(identityDomainService, logger)

	riskEngine := &adaptive.RiskEngine{}
	mfaManager := &mfa.MFAManager{}

	authDomainService := auth.NewService(identityDomainService, sessionRepo, tokenRepo, nil, cryptoManager, logger, riskEngine, mfaManager)
	tracer := trace.NewNoopTracerProvider().Tracer("quantid-test")

	authAppService := auth_service.NewApplicationService(authDomainService, auditService, logger, auth_service.Config{
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24,
		SessionDuration:      time.Hour * 24,
	}, tracer)

	appService := application.NewApplicationService(nil, logger, cryptoManager)
	devCenterSvc := platform.NewDevCenterService(appService, nil, authzService, nil)

	services := Services{
		IdentityService:       identityAppService,
		AuthService:           authAppService,
		AuthzService:          authzService,
		CryptoManager:         cryptoManager,
		IdentityDomainService: identityDomainService,
		DevCenterService:      devCenterSvc,
	}

	return NewServer(httpCfg, logger, services), nil
}

// registerRoutes sets up the API routes, their handlers, and associated middleware.
func (s *Server) registerRoutes(services Services) {
	engine := orchestrator.NewEngine(s.logger)
	authHandlers := handlers.NewAuthHandlers(services.AuthService, engine, s.logger)
	identityHandlers := handlers.NewIdentityHandlers(services.IdentityService, s.logger)

	loggingMiddleware := middleware.NewLoggingMiddleware(s.logger)
	metricsMiddleware := metrics.NewHTTPMetricsMiddleware(prometheus.DefaultRegisterer)
	authMiddleware := middleware.NewAuthMiddleware(services.CryptoManager, s.logger, services.IdentityDomainService)
	authzUserReadMiddleware := middleware.NewAuthorizationMiddleware(services.AuthzService, policy.Action("users.read"), "user")

	s.Router.Use(loggingMiddleware.Execute)
	s.Router.Use(metricsMiddleware.Execute)

	s.Router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	apiV1 := s.Router.PathPrefix("/api/v1").Subrouter()
	apiV1.HandleFunc("/auth/login", authHandlers.Login).Methods("POST")
	apiV1.HandleFunc("/users", identityHandlers.CreateUser).Methods("POST")

	// Protected route for getting a user
	getUserHandler := http.HandlerFunc(identityHandlers.GetUser)
	protectedGetUserRoute := authMiddleware.Execute(authzUserReadMiddleware.Execute(getUserHandler))
	apiV1.Handle("/users/{id}", protectedGetUserRoute).Methods("GET")

	devcenterHandlers := handlers.NewDevCenterHandler(services.DevCenterService)
	devcenterAdminMiddleware := middleware.NewAuthorizationMiddleware(services.AuthzService, policy.Action("devcenter.admin"), "devcenter")
	devcenterRouter := apiV1.PathPrefix("/devcenter").Subrouter()
	devcenterRouter.Use(authMiddleware.Execute, devcenterAdminMiddleware.Execute)
	devcenterHandlers.RegisterRoutes(devcenterRouter)

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
