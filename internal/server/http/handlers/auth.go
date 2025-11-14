package handlers

import (
	"encoding/json"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/orchestrator"
	auth_service "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"net/http"
	"time"
)

// AuthHandlers provides HTTP handlers for authentication-related endpoints.
// It translates HTTP requests into calls to the authentication application service.
type AuthHandlers struct {
	authService *auth_service.ApplicationService
	engine      *orchestrator.Engine
	logger      utils.Logger
}

// NewAuthHandlers creates a new set of authentication handlers.
//
// Parameters:
//   - authService: The application service that contains the core authentication logic.
//   - logger: The logger for handler-specific messages.
//
// Returns:
//   A new AuthHandlers instance.
func NewAuthHandlers(authService *auth_service.ApplicationService, engine *orchestrator.Engine, logger utils.Logger) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		engine:      engine,
		logger:      logger,
	}
}

// Login is the HTTP handler for the user login endpoint.
// It decodes the login request, calls the authentication service,
// and writes the JSON response or error.
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req auth_service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, types.ErrBadRequest.WithCause(err), http.StatusBadRequest)
		return
	}

	loginCtx := auth.LoginContext{
		Username:  req.Username,
		Password:  req.Password,
		CurrentIP: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Now:       time.Now(),
	}

	initialState := orchestrator.State{"login_ctx": loginCtx}
	finalState, err := h.engine.Execute(r.Context(), "login_workflow", initialState)
	if err != nil {
		appErr, ok := err.(*types.Error)
		if ok {
			WriteJSONError(w, appErr, appErr.HttpStatus)
		} else {
			WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	authResp := finalState["auth_response"].(*auth_service.LoginResponse)
	WriteJSON(w, http.StatusOK, authResp)
}

// Logout is the HTTP handler for the user logout endpoint.
// It decodes the logout request, calls the authentication service to invalidate the session/token,
// and returns a successful status.
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	var req auth_service.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, types.ErrBadRequest.WithCause(err), http.StatusBadRequest)
		return
	}

	if appErr := h.authService.Logout(r.Context(), req); appErr != nil {
		WriteJSONError(w, appErr, appErr.HttpStatus)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
