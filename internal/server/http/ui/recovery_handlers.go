package ui

import (
	"net/http"

	"github.com/turtacn/QuantaID/internal/domain/auth"
	"go.uber.org/zap"
)

type RecoveryHandler struct {
	service  *auth.RecoveryService
	renderer *Renderer
	logger   *zap.Logger
}

func NewRecoveryHandler(service *auth.RecoveryService, renderer *Renderer, logger *zap.Logger) *RecoveryHandler {
	return &RecoveryHandler{
		service:  service,
		renderer: renderer,
		logger:   logger,
	}
}

func (h *RecoveryHandler) ShowForgotPassword(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "auth/forgot_password.html", nil)
}

func (h *RecoveryHandler) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")

	// Call service regardless of result to prevent enumeration
	// (Service handles this, but good to be explicit in handler flow)
	err := h.service.InitiateRecovery(r.Context(), email)
	if err != nil {
		h.logger.Error("Failed to initiate recovery", zap.Error(err))
		// Even on error, show the reset page (or a success message saying "If that email exists...")
		// to prevent enumeration.
	}

	// Render the reset page pre-filled with email
	data := map[string]interface{}{
		"Email": email,
	}
	h.renderer.Render(w, r, "auth/reset_password.html", data)
}

func (h *RecoveryHandler) ShowResetPassword(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "auth/reset_password.html", nil)
}

func (h *RecoveryHandler) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	code := r.FormValue("code")
	password := r.FormValue("password")

	err := h.service.VerifyAndReset(r.Context(), email, code, password)
	if err != nil {
		h.logger.Warn("Password reset failed", zap.Error(err))
		data := map[string]interface{}{
			"Error": "Invalid code or request failed.",
			"Email": email,
		}
		h.renderer.Render(w, r, "auth/reset_password.html", data)
		return
	}

	// Redirect to login with success message
	http.Redirect(w, r, "/auth/login?reset=success", http.StatusSeeOther)
}
