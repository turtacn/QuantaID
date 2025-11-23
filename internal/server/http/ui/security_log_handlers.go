package ui

import (
	"net/http"

	"github.com/turtacn/QuantaID/internal/server/middleware"
	audit_service "github.com/turtacn/QuantaID/internal/services/audit"
	"go.uber.org/zap"
)

type SecurityLogHandler struct {
	auditService *audit_service.Service
	renderer     *Renderer
	logger       *zap.Logger
}

func NewSecurityLogHandler(auditService *audit_service.Service, renderer *Renderer, logger *zap.Logger) *SecurityLogHandler {
	return &SecurityLogHandler{
		auditService: auditService,
		renderer:     renderer,
		logger:       logger,
	}
}

func (h *SecurityLogHandler) ShowSecurityLog(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey)
	if userID == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	logs, err := h.auditService.GetLogsForUser(r.Context(), userID.(string), 50)
	if err != nil {
		h.logger.Error("Failed to fetch security logs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Logs": logs,
	}

	h.renderer.Render(w, r, "portal/security_log.html", data)
}
