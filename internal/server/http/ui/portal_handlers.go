package ui

import (
	"net/http"
)

// PortalHandler handles the user self-service portal.
type PortalHandler struct {
	renderer *Renderer
}

// NewPortalHandler creates a new PortalHandler.
func NewPortalHandler(renderer *Renderer) *PortalHandler {
	return &PortalHandler{renderer: renderer}
}

// ShowProfilePage renders the main user profile page.
func (h *PortalHandler) ShowProfilePage(w http.ResponseWriter, r *http.Request) {
	// In a real application, you would fetch the current user's data
	// and pass it to the template.
	// For now, we pass a dummy user object.
	dummyData := map[string]interface{}{
		"Username": "johndoe",
		"Email":    "johndoe@example.com",
		"MFAFactors": []map[string]string{
			{"Type": "TOTP", "Identifier": "Enabled"},
		},
	}

	h.renderer.Render(w, r, "profile.html", dummyData)
}

// HandleAddTOTP begins the process of adding a new TOTP device.
func (h *PortalHandler) HandleAddTOTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Generate a new TOTP secret and QR code, then render a page to display it.
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}

// HandleAddWebAuthn begins the process of adding a new WebAuthn device.
// This is an API endpoint called by the frontend JS.
func (h *PortalHandler) HandleAddWebAuthn(w http.ResponseWriter, r *http.Request) {
	// This will be handled by the API handlers, not the UI handlers.
	// The button on the profile page will call the WebAuthn JS functions directly.
	http.Error(w, "This should be handled by an API endpoint.", http.StatusNotImplemented)
}
