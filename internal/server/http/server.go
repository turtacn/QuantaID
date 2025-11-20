package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
	"github.com/turtacn/QuantaID/internal/api/admin"
	i_audit "github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/auth/adaptive"
	"github.com/turtacn/QuantaID/internal/auth/mfa"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/metrics"
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
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Server encapsulates the HTTP server for the QuantaID API.
type Server struct {
	httpServer *http.Server
	Router     *mux.Router
	logger     utils.Logger
	Services   Services
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
	AuditLogger           *i_audit.AuditLogger
}

// NewServer creates a new HTTP server instance.
func NewServer(config Config, logger utils.Logger, services Services, appCfg *utils.Config) *Server {
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
		Services: services,
	}
	server.registerRoutes(services, appCfg)
	return server
}

// NewServerWithConfig creates a new server instance based on the provided configuration.
func NewServerWithConfig(httpCfg Config, appCfg *utils.Config, logger utils.Logger, cryptoManager *utils.CryptoManager) (*Server, error) {
	var idRepo identity.UserRepository
	var groupRepo identity.GroupRepository
	var sessionRepo auth.SessionRepository
	var tokenRepo auth.TokenRepository
	var auditRepo auth.AuditLogRepository
	var policyRepo policy.PolicyRepository
	var db *gorm.DB
	var err error
	var redisClient redis.RedisClientInterface

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
		logger.Info(context.Background(), "Using PostgreSQL storage backend",
			zap.String("host", appCfg.Postgres.Host),
			zap.String("dbname", appCfg.Postgres.DbName),
		)
		// Postgres Connection
		db, err = postgresql.NewConnection(appCfg.Postgres)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}

		// Postgres Repositories
		pgIdRepo := postgresql.NewPostgresIdentityRepository(db)
		idRepo = pgIdRepo
		groupRepo = pgIdRepo
		_, auditRepo = postgresql.NewPostgresAuthRepository(db)
		policyRepo = postgresql.NewPostgresPolicyRepository(db)

		// Redis Connection
		redisMetrics := metrics.NewRedisMetrics("quantid")
		redisConfig := &redis.RedisConfig{
			Host:     appCfg.Redis.Host,
			Port:     appCfg.Redis.Port,
			Password: appCfg.Redis.Password,
			DB:       appCfg.Redis.DB,
		}
		redisClient, err = redis.NewRedisClient(redisConfig, redisMetrics)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to redis: %w", err)
		}

		// Redis Repositories
		sessionRepo = redis.NewRedisSessionRepository(redisClient)
		tokenRepo = redis.NewRedisTokenRepository(redisClient)
	default:
		return nil, fmt.Errorf("invalid storage mode: %s", appCfg.Storage.Mode)
	}

	// Audit Service Setup
	auditPipeline := i_audit.NewPipeline(logger.(*utils.ZapLogger).Logger)
	auditService := audit_service.NewService(auditPipeline)

	evaluator := authorization.NewDefaultEvaluator(policyRepo)
	authzService := authorization.NewService(evaluator, auditService)

	identityDomainService := identity.NewService(idRepo, groupRepo, cryptoManager, logger)
	identityAppService := identity_service.NewApplicationService(identityDomainService, logger)

	riskEngine := adaptive.NewRiskEngine(appCfg.Security.Risk, redisClient, logger.(*utils.ZapLogger).Logger)
	mfaManager := &mfa.MFAManager{}
	policyEngine := authorization.NewPolicyEngine(evaluator)

	authDomainService := auth.NewService(identityDomainService, sessionRepo, tokenRepo, auditRepo, nil, cryptoManager, logger, riskEngine, policyEngine, mfaManager)
	tracer := trace.NewNoopTracerProvider().Tracer("quantid-test")

	authAppService := auth_service.NewApplicationService(authDomainService, auditService, logger, auth_service.Config{
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24,
		SessionDuration:      time.Hour * 24,
	}, tracer)

	appService := application.NewApplicationService(postgresql.NewPostgresApplicationRepository(db), logger, cryptoManager)
	devCenterSvc := platform.NewDevCenterService(appService, nil, authzService, nil)

	// Audit Logger
	auditRepoForLogger := postgresql.NewPostgresAuditLogRepository(db)
	auditLogger := i_audit.NewAuditLogger(auditRepoForLogger, logger.(*utils.ZapLogger).Logger, 100, 5*time.Second, 1000)

	services := Services{
		IdentityService:       identityAppService,
		AuthService:           authAppService,
		AuthzService:          authzService,
		CryptoManager:         cryptoManager,
		IdentityDomainService: identityDomainService,
		DevCenterService:      devCenterSvc,
		AuditLogger:           auditLogger,
	}

	return NewServer(httpCfg, logger, services, appCfg), nil
}

// registerRoutes sets up the API routes, their handlers, and associated middleware.
func (s *Server) registerRoutes(services Services, appCfg *utils.Config) {
	authHandlers := handlers.NewAuthHandlers(services.AuthService, s.logger)
	identityHandlers := handlers.NewIdentityHandlers(services.IdentityService, s.logger)
	adminUserHandlers := admin.NewAdminUserHandler(services.IdentityDomainService, services.AuditLogger)

	loggingMiddleware := middleware.NewLoggingMiddleware(s.logger)
	auditorMiddleware := middleware.AuditorMiddleware(services.AuditLogger)
	authMiddleware := middleware.NewAuthMiddleware(services.CryptoManager, s.logger, services.IdentityDomainService)
	authzUserReadMiddleware := middleware.NewAuthorizationMiddleware(services.AuthzService, policy.Action("users.read"), "user")
	authzUserUpdateMiddleware := middleware.NewAuthorizationMiddleware(services.AuthzService, policy.Action("user:update"), "user")

	s.Router.Use(loggingMiddleware.Execute)
	if appCfg.Metrics.Enabled {
		metricsMiddleware := middleware.NewMetricsMiddleware(prometheus.DefaultRegisterer)
		s.Router.Use(metricsMiddleware.Execute)
	}
	s.Router.Use(auditorMiddleware)

	s.Router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	apiV1 := s.Router.PathPrefix("/api/v1").Subrouter()
	apiV1.HandleFunc("/auth/login", authHandlers.Login).Methods("POST")
	apiV1.HandleFunc("/users", identityHandlers.CreateUser).Methods("POST")

	// Protected route for getting a user
	getUserHandler := http.HandlerFunc(identityHandlers.GetUser)
	protectedGetUserRoute := authMiddleware.Execute(authzUserReadMiddleware.Execute(getUserHandler))
	apiV1.Handle("/users/{id}", protectedGetUserRoute).Methods("GET")

	// Admin routes for user management
	adminRouter := apiV1.PathPrefix("/admin").Subrouter()
	adminRouter.Use(authMiddleware.Execute, authzUserUpdateMiddleware.Execute)
	adminRouter.HandleFunc("/users", adminUserHandlers.ListUsers).Methods("GET")
	adminRouter.HandleFunc("/users/{userID}/ban", adminUserHandlers.BanUser).Methods("POST")
	adminRouter.HandleFunc("/users/{userID}/unban", adminUserHandlers.UnbanUser).Methods("POST")

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
