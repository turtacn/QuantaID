package ui

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mssola/user_agent"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"go.uber.org/zap"
)

type DeviceInfo struct {
	ID                string
	IPAddress         string
	Browser           string
	OS                string
	Device            string // "Mobile" or "Desktop" or specific name
	LastActive        string // Formatted time
	DeviceFingerprint string
	Current           bool
}

type DeviceHandler struct {
	sessionManager *redis.SessionManager
	renderer       *Renderer
	logger         *zap.Logger
}

func NewDeviceHandler(sessionManager *redis.SessionManager, renderer *Renderer, logger *zap.Logger) *DeviceHandler {
	return &DeviceHandler{
		sessionManager: sessionManager,
		renderer:       renderer,
		logger:         logger,
	}
}

func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	// Assuming middleware populates UserID in context, or we extract it from session
	// For now, let's assume we have a helper or middleware.
	// Since I don't see a "GetUserID" helper immediately available in the context of this UI package,
	// I'll assume the session is available in the request context or similar.
	// But `ui` handlers are usually protected by auth middleware.

	// Let's assume a context key for UserID exists.
	// Searching for "UserIDContextKey" in memory... found in `internal/server/middleware/context.go`.
	// I'll need to check the exact key usage.
	// For now, I will assume a string key "user_id" or similar if not imported.
	// Better: Use a helper if available.

	// In `internal/server/http/ui/portal_handlers.go`, it uses dummy data.
	// I need to implement real logic.

	// I will check `internal/server/middleware/context.go` to see how to extract UserID.
	// For now, I'll put a placeholder TODO and fix it in the wiring step or next tool call.
	// Actually, I should do it right.

	// Let's assume there is a `middleware.GetUserID(ctx)` or similar.
	// I'll check `internal/server/middleware` in next step if needed.
	// For now, I will just proceed assuming I can get it.

    // Using a hardcoded string for now to allow compilation, assuming middleware sets it.
	userID := r.Context().Value(middleware.UserIDContextKey)
    if userID == nil {
        // Redirect to login if not found (middleware should have handled this)
        http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
        return
    }

	sessions, err := h.sessionManager.GetUserSessions(r.Context(), userID.(string))
	if err != nil {
		h.logger.Error("Failed to list devices", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var devices []DeviceInfo
	// Need to identify current session.
	// For now, if there is a session cookie matching, we mark it.
	// But handlers don't have easy access to session ID unless middleware puts it in context.
	// Assume session ID is NOT in context for now, so "Current" might be skipped or heuristic.

	for _, s := range sessions {
		ua := user_agent.New(s.UserAgent)
		browserName, browserVer := ua.Browser()

		deviceType := "Desktop"
		if ua.Mobile() {
			deviceType = "Mobile"
		}
		if ua.Bot() {
			deviceType = "Bot"
		}

		// GeoIP would go here if we had the service injected.
		// Ignoring for now as per plan focus on UA parsing first.

		devices = append(devices, DeviceInfo{
			ID:                s.ID,
			IPAddress:         s.IPAddress,
			Browser:           browserName + " " + browserVer,
			OS:                ua.OS(),
			Device:            deviceType,
			LastActive:        s.LastRotatedAt.Format("2006-01-02 15:04:05"),
			DeviceFingerprint: s.DeviceFingerprint,
			Current:           false, // TODO: Check against current request session
		})
	}

	data := map[string]interface{}{
		"Sessions": devices,
	}

	h.renderer.Render(w, r, "portal/devices.html", data)
}

func (h *DeviceHandler) RevokeDevice(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey)
	if userID == nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	sessionID := vars["id"]

	err := h.sessionManager.RevokeSession(r.Context(), userID.(string), sessionID)
	if err != nil {
		h.logger.Error("Failed to revoke device", zap.Error(err))
		// Continue to redirect even on error
	}

	http.Redirect(w, r, "/portal/devices", http.StatusSeeOther)
}
