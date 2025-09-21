package http

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	"github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Server is the HTTP server for the QuantaID API.
type Server struct {
	httpServer *http.Server
	Router     *mux.Router // Changed to be public
	logger     utils.Logger
}

// Config holds the configuration for the HTTP server.
type Config struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Services holds the application services that the server will depend on.
type Services struct {
	AuthService     *auth.ApplicationService
	IdentityService *identity.ApplicationService
	AuthzService    *authorization.ApplicationService
	CryptoManager   *utils.CryptoManager
}

// NewServer creates a new HTTP server.
func NewServer(config Config, logger utils.Logger, services Services) *Server {
	router := mux.NewRouter()

	server := &Server{
		Router: router, // Changed to be public
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

// registerRoutes sets up the API routes and their handlers.
func (s *Server) registerRoutes(services Services) {
	authHandlers := handlers.NewAuthHandlers(services.AuthService, s.logger)
	identityHandlers := handlers.NewIdentityHandlers(services.IdentityService, s.logger)

	loggingMiddleware := middleware.NewLoggingMiddleware(s.logger)
	authMiddleware := middleware.NewAuthMiddleware(services.AuthzService, services.CryptoManager, s.logger)

	s.Router.Use(loggingMiddleware.Execute)

	apiV1 := s.Router.PathPrefix("/api/v1").Subrouter()

	apiV1.HandleFunc("/auth/login", authHandlers.Login).Methods("POST")

	protected := apiV1.PathPrefix("/").Subrouter()
	protected.Use(authMiddleware.Execute)
	protected.HandleFunc("/users", identityHandlers.CreateUser).Methods("POST")
	protected.HandleFunc("/users/{id}", identityHandlers.GetUser).Methods("GET")

	s.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
}

// Start runs the HTTP server.
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

//Personal.AI order the ending
