package ui

import (
	"net/http"
)

// AuthHandler handles UI-based authentication flows.
type AuthHandler struct {
	renderer *Renderer
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(renderer *Renderer) *AuthHandler {
	return &AuthHandler{renderer: renderer}
}

// ShowLoginPage renders the login page.
func (h *AuthHandler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "login.html", nil)
}

// HandleLogin processes the login form submission.
// TODO: Integrate with the actual authentication service.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// --- Placeholder Logic ---
	// In a real implementation, you would call your auth service here.
	// e.g., auth.Authenticate(username, password)
	if username == "admin" && password == "password" {
		// On success, redirect to a dashboard or the original URL.
		// For now, let's assume MFA is required and redirect to the MFA page.
		http.Redirect(w, r, "/auth/mfa", http.StatusFound)
	} else {
		// On failure, re-render the login page with an error.
		data := map[string]string{
			"Error":    "Invalid username or password",
			"Username": username,
		}
		h.renderer.Render(w, r, "login.html", data)
	}
}

// ShowMFAPage renders the MFA challenge page.
func (h *AuthHandler) ShowMFAPage(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "mfa_challenge.html", nil)
}

// HandleMFA processes the MFA submission.
func (h *AuthHandler) HandleMFA(w http.ResponseWriter, r *http.Request) {
	// TODO: Handle TOTP and WebAuthn submissions.
	http.Redirect(w, r, "/dashboard", http.StatusFound) // Placeholder
}

// ShowConsentPage renders the OIDC consent page.
func (h *AuthHandler) ShowConsentPage(w http.ResponseWriter, r *http.Request) {
	// In a real OIDC flow, you would fetch details about the client application
	// and the requested scopes from the request context or a session store.
	dummyData := map[string]interface{}{
		"ClientName": "Example App",
		"Scopes":     []string{"openid", "profile", "email"},
	}
	h.renderer.Render(w, r, "consent.html", dummyData)
}

// HandleConsent processes the user's decision on the OIDC consent page.
func (h *AuthHandler) HandleConsent(w http.ResponseWriter, r *http.Request) {
	// TODO: Process the user's consent (allow/deny) and complete the OIDC flow.
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}
