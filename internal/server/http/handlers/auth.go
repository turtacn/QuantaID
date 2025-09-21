package handlers

import (
	"encoding/json"
	"github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"net/http"
)

// AuthHandlers provides HTTP handlers for authentication-related endpoints.
type AuthHandlers struct {
	authService *auth.ApplicationService
	logger      utils.Logger
}

// NewAuthHandlers creates a new set of auth handlers.
func NewAuthHandlers(authService *auth.ApplicationService, logger utils.Logger) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		logger:      logger,
	}
}

// Login is the HTTP handler for the user login endpoint.
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, types.ErrBadRequest.WithCause(err), http.StatusBadRequest)
		return
	}

	loginResp, appErr := h.authService.Login(r.Context(), req)
	if appErr != nil {
		WriteJSONError(w, appErr, appErr.HttpStatus)
		return
	}

	WriteJSON(w, http.StatusOK, loginResp)
}

// Logout is the HTTP handler for the user logout endpoint.
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	var req auth.LogoutRequest
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

//Personal.AI order the ending
