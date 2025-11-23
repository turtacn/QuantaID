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
	webhook_service "github.com/turtacn/QuantaID/internal/services/webhook"
	"github.com/turtacn/QuantaID/internal/domain/webhook"
	"github.com/turtacn/QuantaID/internal/worker"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Server encapsulates the HTTP server for the QuantaID API.
type Server struct {
	httpServer  *http.Server
	Router      *mux.Router
	logger      utils.Logger
	Services    Services
	db          *gorm.DB
	redisClient redis.RedisClientInterface
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
	WebhookService        *webhook_service.Service
	WebhookWorker         *worker.WebhookSender
}

// NewServer creates a new HTTP server instance.
func NewServer(config Config, logger utils.Logger, services Services, appCfg *utils.Config, db *gorm.DB, redisClient redis.RedisClientInterface) *Server {
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
		Services:    services,
		db:          db,
		redisClient: redisClient,
	}

	// Start Webhook Worker
	if services.WebhookWorker != nil {
		services.WebhookWorker.Start(context.Background())
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
	var appRepo types.ApplicationRepository
	var webhookRepo webhook.Repository
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
		appRepo = postgresql.NewPostgresApplicationRepository(db)
		webhookRepo = postgresql.NewWebhookRepository(db)

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

	// Webhook Worker Setup
	webhookWorker := worker.NewWebhookSender(webhookRepo, logger, 100)
	webhookDispatcher := webhook_service.NewDispatcher(webhookRepo, webhookWorker, logger)
	webhookService := webhook_service.NewService(webhookRepo)

	// Audit Service Setup
	auditPipeline := i_audit.NewPipeline(logger.(*utils.ZapLogger).Logger)
	auditService := audit_service.NewService(auditPipeline, webhookDispatcher)

	evaluator := authorization.NewDefaultEvaluator(policyRepo)
	authzService := authorization.NewService(evaluator, auditService)

	identityDomainService := identity.NewService(idRepo, groupRepo, cryptoManager, logger)
	identityAppService := identity_service.NewApplicationService(identityDomainService, auditService, logger)

	riskEngine := adaptive.NewRiskEngine(appCfg.Security.Risk, redisClient, logger.(*utils.ZapLogger).Logger)
	mfaManager := &mfa.MFAManager{}
	policyEngine := authorization.NewPolicyEngine(evaluator)

	authDomainService := auth.NewService(identityDomainService, sessionRepo, tokenRepo, auditRepo, nil, cryptoManager, logger, riskEngine, policyEngine, mfaManager, appRepo, redisClient)
	tracer := trace.NewNoopTracerProvider().Tracer("quantid-test")

	authAppService := auth_service.NewApplicationService(authDomainService, auditService, logger, auth_service.Config{
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24,
		SessionDuration:      time.Hour * 24,
	}, tracer)

	appService := application.NewApplicationService(appRepo, logger, cryptoManager)
	devCenterSvc := platform.NewDevCenterService(appService, nil, authzService, nil)

	// Audit Logger
	auditRepoForLogger := postgresql.NewPostgresAuditLogRepository(db)
	auditLogger := i_audit.NewAuditLogger(auditRepoForLogger, logger.(*utils.ZapLogger).Logger, 100, 5*time.Second, 1000)

	services := Services{
		IdentityService:       identityAppService,
		AuthService:           authAppService,
		AuthzService:          authzService,
		AppService:            appService,
		CryptoManager:         cryptoManager,
		IdentityDomainService: identityDomainService,
		DevCenterService:      devCenterSvc,
		AuditLogger:           auditLogger,
		WebhookService:        webhookService,
		WebhookWorker:         webhookWorker,
	}

	return NewServer(httpCfg, logger, services, appCfg, db, redisClient), nil
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

	oauthHandlers := handlers.NewOAuthHandlers(services.AuthService, services.IdentityService, s.logger)
	s.Router.HandleFunc("/oauth/authorize", oauthHandlers.Authorize).Methods("GET")
	s.Router.HandleFunc("/oauth/token", oauthHandlers.Token).Methods("POST")
	s.Router.HandleFunc("/.well-known/openid-configuration", oauthHandlers.Discovery).Methods("GET")
	s.Router.HandleFunc("/.well-known/jwks.json", oauthHandlers.JWKS).Methods("GET")

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

	// SCIM v2 routes
	scimHandler := handlers.NewSCIMHandler(services.IdentityDomainService, s.logger)
	// SCIM routes require authentication (usually Bearer token).
	// For P2-T5: "Add Bearer Token auth middleware".
	// We reuse authMiddleware for now, or we should create a specific one if it needs long-lived tokens (API Keys).
	// Assuming standard authMiddleware works with Bearer tokens.
	scimRouter := s.Router.PathPrefix("/scim/v2").Subrouter()
	scimRouter.Use(authMiddleware.Execute)
	scimHandler.RegisterRoutes(scimRouter)

	// Webhook Management
	webhookHandler := admin.NewWebhookHandler(services.WebhookService, s.logger)
	webhookRouter := adminRouter.PathPrefix("/webhooks").Subrouter()
	webhookRouter.HandleFunc("", webhookHandler.CreateSubscription).Methods("POST")
	webhookRouter.HandleFunc("", webhookHandler.ListSubscriptions).Methods("GET")
	webhookRouter.HandleFunc("/{id}", webhookHandler.DeleteSubscription).Methods("DELETE")
	webhookRouter.HandleFunc("/{id}/rotate-secret", webhookHandler.RotateSecret).Methods("POST")

	// Health and readiness probes
	s.Router.HandleFunc("/healthz", s.healthzHandler).Methods("GET")
	s.Router.HandleFunc("/readyz", s.readyzHandler).Methods("GET")
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

// healthzHandler is a liveness probe.
func (s *Server) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// readyzHandler is a readiness probe.
func (s *Server) readyzHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check DB connection if it exists
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			http.Error(w, "failed to get database connection pool", http.StatusInternalServerError)
			return
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			s.logger.Error(ctx, "database ping failed for readiness probe", zap.Error(err))
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
	}

	// Check Redis connection if it exists
	if s.redisClient != nil {
		if err := s.redisClient.HealthCheck(ctx); err != nil {
			s.logger.Error(ctx, "redis health check failed for readiness probe", zap.Error(err))
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
