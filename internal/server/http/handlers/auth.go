package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// AuthServiceInterface defines the interface for the authentication service.
type AuthServiceInterface interface {
	LoginWithPassword(ctx context.Context, req auth.AuthnRequest, serviceConfig auth.Config) (*types.AuthResult, error)
	VerifyMFAChallenge(ctx context.Context, req *types.VerifyMFARequest, serviceConfig auth.Config) (*types.AuthResult, error)
}

// AuthHandlers provides HTTP handlers for authentication-related endpoints.
type AuthHandlers struct {
	authService AuthServiceInterface
	logger      utils.Logger
}

// NewAuthHandlers creates a new set of authentication handlers.
func NewAuthHandlers(authService AuthServiceInterface, logger utils.Logger) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		logger:      logger,
	}
}

// Login is the HTTP handler for the user login endpoint.
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.AuthnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, types.ErrBadRequest.WithCause(err), http.StatusBadRequest)
		return
	}

	authResult, err := h.authService.LoginWithPassword(r.Context(), req, auth.Config{})
	if err != nil {
		metrics.AuthLoginTotal.WithLabelValues("fail").Inc()
		if appErr, ok := err.(*types.Error); ok {
			WriteJSONError(w, appErr, appErr.HttpStatus)
		} else {
			WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	if authResult.IsMfaRequired {
		WriteJSON(w, http.StatusOK, authResult.MFAChallenge)
		return
	}

	metrics.AuthLoginTotal.WithLabelValues("success").Inc()
	WriteJSON(w, http.StatusOK, authResult.Token)
}

// VerifyMFA is the HTTP handler for the MFA verification endpoint.
func (h *AuthHandlers) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	var req types.VerifyMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, types.ErrBadRequest.WithCause(err), http.StatusBadRequest)
		return
	}

	authResult, err := h.authService.VerifyMFAChallenge(r.Context(), &req, auth.Config{})
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			WriteJSONError(w, appErr, appErr.HttpStatus)
		} else {
			WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	WriteJSON(w, http.StatusOK, authResult.Token)
}

// Logout is the HTTP handler for the user logout endpoint.
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	// To be implemented
	w.WriteHeader(http.StatusNoContent)
}
