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

// Server encapsulates the HTTP server for the QuantaID API.
// It manages the server's lifecycle, routing, and dependencies.
type Server struct {
	httpServer *http.Server
	// Router is the main router for the HTTP server. It is public to allow for extension,
	// for example, in tests or other specialized configurations.
	Router *mux.Router
	logger utils.Logger
}

// Config holds the configuration required for the HTTP server.
type Config struct {
	// Address is the TCP address for the server to listen on (e.g., ":8080").
	Address string
	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration
}

// Services is a container for all the application services that the server's handlers will depend on.
// This struct is used to inject dependencies into the server and its handlers.
type Services struct {
	AuthService     *auth.ApplicationService
	IdentityService *identity.ApplicationService
	AuthzService    *authorization.ApplicationService
	CryptoManager   *utils.CryptoManager
}

// NewServer creates and configures a new HTTP server instance.
// It initializes the router, sets up the server with the given configuration,
// and registers all the API routes and their corresponding handlers.
//
// Parameters:
//   - config: The configuration for the server (address, timeouts).
//   - logger: The logger for server-level messages.
//   - services: A container for all the application services required by the handlers.
//
// Returns:
//   A new, configured but not yet started, Server instance.
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

// registerRoutes sets up the API routes, their handlers, and associated middleware.
// It defines the structure of the API, including versioning and protected routes.
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

// Start begins listening for and serving HTTP requests.
// This is a blocking call. It will only return on a server error (other than ErrServerClosed).
func (s *Server) Start() {
	s.logger.Info(context.Background(), "Starting HTTP server", zap.String("address", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error(context.Background(), "HTTP server failed to start", zap.Error(err))
	}
}

// Stop gracefully shuts down the HTTP server without interrupting any active connections.
// It waits for a given context to be done before forcing a shutdown.
//
// Parameters:
//   - ctx: A context to control the shutdown duration.
func (s *Server) Stop(ctx context.Context) {
	s.logger.Info(ctx, "Shutting down HTTP server")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error(ctx, "HTTP server graceful shutdown failed", zap.Error(err))
	}
}
