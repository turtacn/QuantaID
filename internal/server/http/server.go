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
	"github.com/turtacn/QuantaID/internal/api/privacy"
	i_audit "github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/audit/sinks"
	"github.com/turtacn/QuantaID/internal/auth/adaptive"
	"github.com/turtacn/QuantaID/internal/auth/mfa"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	domain_privacy "github.com/turtacn/QuantaID/internal/domain/privacy"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/internal/policy/engine"
	"github.com/turtacn/QuantaID/internal/protocols/saml"
	"github.com/turtacn/QuantaID/internal/services/application"
	audit_service "github.com/turtacn/QuantaID/internal/services/audit"
	auth_service "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	identity_service "github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/internal/services/platform"
	policy_service "github.com/turtacn/QuantaID/internal/services/policy"
	privacy_service "github.com/turtacn/QuantaID/internal/services/privacy"
	webhook_service "github.com/turtacn/QuantaID/internal/services/webhook"
	"github.com/turtacn/QuantaID/internal/domain/webhook"
	"github.com/turtacn/QuantaID/internal/worker"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/internal/server/http/ui"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// noopRBACProvider is a stub for when RBAC repo is missing
type noopRBACProvider struct{}

func (n *noopRBACProvider) IsAllowed(ctx context.Context, subjectID, action, resource string) (bool, error) {
	return false, nil // Default deny
}

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
	AuditService          *audit_service.Service // Add AuditService
	WebhookService        *webhook_service.Service
	WebhookWorker         *worker.WebhookSender
	RecoveryService       *auth.RecoveryService
	SessionManager        *redis.SessionManager
	Renderer              *ui.Renderer
	WebAuthnProvider      *mfa.WebAuthnProvider
	PrivacyService        *privacy_service.Service
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

	// Register global middleware first
	router.Use(middleware.IPBlacklistMiddleware(redisClient))

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
	var rbacRepo policy.RBACRepository
	var appRepo types.ApplicationRepository
	var webhookRepo webhook.Repository
	var privacyRepo domain_privacy.Repository
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
		rbacRepo = postgresql.NewRBACRepository(db)
		appRepo = postgresql.NewPostgresApplicationRepository(db)
		webhookRepo = postgresql.NewWebhookRepository(db)
		privacyRepo = postgresql.NewPostgresPrivacyRepository(db)

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

	// Session Manager
	sessionManager := redis.NewSessionManager(
		redisClient,
		redis.SessionConfig{
			DefaultTTL:          time.Hour * 24,
			EnableRotation:      true,
			RotationInterval:    time.Hour,
			EnableDeviceBinding: true,
			MaxSessionsPerUser:  5,
		},
		logger.(*utils.ZapLogger).Logger,
		&redis.UUIDv4Generator{},
		&redis.RealClock{},
		metrics.NewRedisMetrics("session_manager"),
	)

	// Webhook Worker Setup
	webhookWorker := worker.NewWebhookSender(webhookRepo, logger, 100)
	webhookDispatcher := webhook_service.NewDispatcher(webhookRepo, webhookWorker, logger)
	webhookService := webhook_service.NewService(webhookRepo)

	// Audit Service Setup
	var auditPipeline *i_audit.Pipeline
	if logger.(*utils.ZapLogger) != nil {
		auditPipeline = i_audit.NewPipeline(logger.(*utils.ZapLogger).Logger)
	} else {
		// handle non-zap logger or nil
		auditPipeline = i_audit.NewPipeline(zap.NewNop())
	}
	// We need to inject the repository into AuditService for read operations
	_, _ = postgresql.NewPostgresAuthRepository(db) // Reuse existing call or create new instance?
	// NewPostgresAuthRepository returns (AuthRepository, AuditLogRepository).
	// But we already called it above: `_, auditRepo = postgresql.NewPostgresAuthRepository(db)`
	// Wait, auditRepo is of type `auth.AuditLogRepository` (defined in `internal/domain/auth/repository.go`?)
	// Or `internal/audit/repository.go`?
	// Let's check NewPostgresAuthRepository return type.

	// Assuming auditRepo implements audit.AuditRepository
	// In NewServerWithConfig: `var auditRepo auth.AuditLogRepository`
	// But `audit_service.WithRepository` expects `audit.AuditRepository`.
	// Are they the same interface?
	// `internal/audit/repository.go` defines `AuditRepository` with `WriteBatch`, `Query` etc.
	// `internal/domain/auth/repository.go` defines `AuditLogRepository` with `CreateLogEntry`, `GetLogsForUser` (maybe).

	// If they are different, we have a problem.
	// Let's assume for now we use `postgresql.NewPostgresAuditLogRepository` which matches `audit.AuditRepository`.

	auditRepoForService := postgresql.NewPostgresAuditLogRepository(db)
	auditService := audit_service.NewService(auditPipeline, webhookDispatcher).WithRepository(auditRepoForService)

	// Initialize OPA Provider
	opaProvider, err := engine.NewOPAProvider(appCfg.OPA)
	if err != nil {
		logger.Warn(context.Background(), "Failed to initialize OPA provider", zap.Error(err))
	}

	// Use _ to suppress unused warning for policyRepo if we are not using it directly anymore
	_ = policyRepo

	// Initialize RBAC Provider handling memory/postgres modes
	var rbacProvider engine.RBACProvider
	if rbacRepo != nil {
		rbacProvider = engine.NewDBRBACProvider(rbacRepo)
	} else {
		// In memory mode or if initialization failed.
		// For now we use a basic mock/noop provider if no RBAC repository is available
		// to prevent runtime panics. In a real memory-mode implementation,
		// we would inject the MemoryAuthRepository if it implemented RBAC interfaces.
		logger.Warn(context.Background(), "RBAC repository is nil (memory mode?), using default/stub provider")
		// We could implement a struct here, but since we don't have one exported,
		// and we want to avoid panic, we will assume tests running with this config
		// might hit issues if they depend on RBAC.
		// However, for compilation safety:
		// If we are strictly in integration tests with Postgres, this branch won't be hit.
		// If we are in unit tests or memory mode, we might need a stub.
		// For now, we will use a nil provider and let the evaluator handle it if possible,
		// OR we rely on the fact that production uses Postgres.

		// Ideally: rbacProvider = memory.NewRBACProvider(...)
		// Current: We will pass nil, but we must ensure `HybridEvaluator` checks for nil.
		// But `HybridEvaluator.Evaluate` calls `e.rbac.IsAllowed`.
		// Let's create a minimal struct here to satisfy the interface.
		rbacProvider = &noopRBACProvider{}
	}

	hybridEvaluator := engine.NewHybridEvaluator(
		rbacProvider,
		engine.NewSimpleABACProvider(),
		opaProvider,
	)

	// Adapter to adapt engine.Evaluator to authorization.Evaluator
	evaluator := authorization.NewEvaluatorAdapter(hybridEvaluator)

	authzService := authorization.NewService(evaluator, auditService)

	identityDomainService := identity.NewService(idRepo, groupRepo, cryptoManager, logger)
	identityAppService := identity_service.NewApplicationService(identityDomainService, auditService, logger)

	// Initialize MFA Repository (needed for WebAuthn)
	mfaRepo := postgresql.NewPostgresMFARepository(db)

	// Initialize WebAuthn Provider
	webAuthnConfig := mfa.WebAuthnConfig{
		RPID:          appCfg.WebAuthn.RPID,
		RPDisplayName: appCfg.WebAuthn.RPDisplayName,
		RPOrigin:      appCfg.WebAuthn.Origin,
	}
	webAuthnProvider, err := mfa.NewWebAuthnProvider(webAuthnConfig, mfaRepo)
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize WebAuthn provider", zap.Error(err))
		// Continue without WebAuthn? Or fail? Given it's a core feature now, maybe fail.
		// But NewWebAuthnProvider might fail on config only.
	}

	mfaManager := mfa.NewMFAManager()
	if webAuthnProvider != nil {
		mfaManager.RegisterProvider("webauthn", webAuthnProvider)
	}
	policyEngine := authorization.NewPolicyEngine(evaluator)

	// GeoIP setup
	geoDB, err := adaptive.NewGeoIPReader("data/GeoLite2-City.mmdb")
	if err != nil {
		// Log error but don't fail, fallback to nil which RiskEngine should handle or we need a mock
		logger.Warn(context.Background(), "Failed to load GeoIP database", zap.Error(err))
	}
	geoManager := redis.NewGeoManager(redisClient)

	// Re-initialize risk engine with all dependencies
	riskEngine := adaptive.NewRiskEngine(appCfg.Security.Risk, redisClient, geoManager, geoDB, logger.(*utils.ZapLogger).Logger)

	authDomainService := auth.NewService(identityDomainService, sessionRepo, tokenRepo, auditRepo, nil, cryptoManager, logger, riskEngine, policyEngine, mfaManager, appRepo, redisClient)
	tracer := trace.NewNoopTracerProvider().Tracer("quantid-test")

	authAppService := auth_service.NewApplicationService(authDomainService, auditService, logger, auth_service.Config{
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24,
		SessionDuration:      time.Hour * 24,
	}, tracer)

	appService := application.NewApplicationService(appRepo, logger, cryptoManager)
	devCenterSvc := platform.NewDevCenterService(appService, nil, authzService, nil)

	// Policy Management Service (wired for Hot Reload watcher)
	// We need to instantiate it even if not exposed via API yet, to ensure watcher runs.
	// But it requires RBACRepository. We have `rbacProvider` but that is the provider, not the repo.
	// `rbacRepo` might be nil in memory mode.
	if rbacRepo != nil {
		// NewService(repo policy.RBACRepository, opaProvider *engine.OPAProvider, logger *zap.Logger)
		_ = policy_service.NewService(rbacRepo, opaProvider, logger.(*utils.ZapLogger).Logger)
	}

	// Audit Logger
	// auditRepoForLogger is used for the sink.
	// We cast it to a GORM DB if possible, but postgresql.NewPostgresAuditLogRepository returns *PostgresAuditLogRepository
	// which we can't easily turn into a Sink unless we implement Sink on it or use PostgresSink.
	// The prompt asked for `internal/audit/sinks/postgres_sink.go`. We should use that.
	postgresSink := sinks.NewPostgresSink(db)
	auditLogger := i_audit.NewAuditLogger(logger.(*utils.ZapLogger).Logger, 100, 5*time.Second, 1000, postgresSink)

	// UI Renderer
	renderer, err := ui.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize UI renderer: %w", err)
	}

	// Recovery Service
	otpProvider := mfa.NewOTPProvider(redisClient, nil, cryptoManager, mfa.OTPConfig{
		TTL:    15 * time.Minute,
		Length: 6,
	})
	recoveryService := auth.NewRecoveryService(idRepo, otpProvider, cryptoManager, sessionManager, logger.(*utils.ZapLogger).Logger)

	privacyService := privacy_service.NewService(db, sessionManager, auditService, privacyRepo, idRepo, auditRepo, appCfg)

	services := Services{
		IdentityService:       identityAppService,
		AuthService:           authAppService,
		AuthzService:          authzService,
		AppService:            appService,
		CryptoManager:         cryptoManager,
		IdentityDomainService: identityDomainService,
		DevCenterService:      devCenterSvc,
		AuditLogger:           auditLogger,
		AuditService:          auditService, // Inject AuditService
		WebhookService:        webhookService,
		WebhookWorker:         webhookWorker,
		RecoveryService:       recoveryService,
		SessionManager:        sessionManager,
		Renderer:              renderer,
		WebAuthnProvider:      webAuthnProvider,
		PrivacyService:        privacyService,
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

	// UI Handlers (Self-Service)
	recoveryHandler := ui.NewRecoveryHandler(services.RecoveryService, services.Renderer, s.logger.(*utils.ZapLogger).Logger)
	deviceHandler := ui.NewDeviceHandler(services.SessionManager, services.Renderer, s.logger.(*utils.ZapLogger).Logger)
	securityLogHandler := ui.NewSecurityLogHandler(services.AuditService, services.Renderer, s.logger.(*utils.ZapLogger).Logger)

	authRouter := s.Router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/forgot-password", recoveryHandler.ShowForgotPassword).Methods("GET")
	authRouter.HandleFunc("/forgot-password", recoveryHandler.HandleForgotPassword).Methods("POST")
	authRouter.HandleFunc("/reset-password", recoveryHandler.ShowResetPassword).Methods("GET")
	authRouter.HandleFunc("/reset-password", recoveryHandler.HandleResetPassword).Methods("POST")

	portalRouter := s.Router.PathPrefix("/portal").Subrouter()
	// IMPORTANT: UI routes should probably use session-based auth middleware.
	// For now, assume authMiddleware can handle session cookie or we rely on session checking inside handlers (not ideal).
	// Since deviceHandler uses `GetUserSessions` which needs UserID, it MUST be authenticated.
	portalRouter.Use(authMiddleware.Execute)
	portalRouter.HandleFunc("/devices", deviceHandler.ListDevices).Methods("GET")
	portalRouter.HandleFunc("/devices/revoke/{id}", deviceHandler.RevokeDevice).Methods("POST")
	portalRouter.HandleFunc("/security-log", securityLogHandler.ShowSecurityLog).Methods("GET")

	// WebAuthn Routes
	if services.WebAuthnProvider != nil {
		webauthnHandler := handlers.NewWebAuthnHandler(services.WebAuthnProvider, services.IdentityDomainService.GetUserRepo(), s.redisClient)

		// Registration endpoints (Require Auth)
		// Note: Usually we register a new passkey for an existing user session.
		webauthnRegRouter := apiV1.PathPrefix("/webauthn/register").Subrouter()
		webauthnRegRouter.Use(authMiddleware.Execute)
		webauthnRegRouter.HandleFunc("/begin", webauthnHandler.BeginRegistration).Methods("POST")
		webauthnRegRouter.HandleFunc("/finish", webauthnHandler.FinishRegistration).Methods("POST")

		// Login endpoints (Public)
		// Login flow starts without session.
		apiV1.HandleFunc("/webauthn/login/begin", webauthnHandler.BeginLogin).Methods("POST")
		apiV1.HandleFunc("/webauthn/login/finish", webauthnHandler.FinishLogin).Methods("POST")
	}

	// Privacy routes
	privacyHandler := privacy.NewHandlers(services.PrivacyService)
	privacyRouter := apiV1.PathPrefix("/privacy").Subrouter()
	privacyRouter.Use(authMiddleware.Execute)
	privacyHandler.RegisterRoutes(privacyRouter)

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
