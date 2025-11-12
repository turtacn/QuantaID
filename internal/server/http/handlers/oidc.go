package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/turtacn/QuantaID/pkg/auth/protocols"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// OIDCHandler handles OIDC requests.
type OIDCHandler struct {
	oidcAdapter *protocols.OIDCAdapter
	logger      utils.Logger
	baseURL     string
}

// NewOIDCHandler creates a new OIDCHandler.
func NewOIDCHandler(oidcAdapter *protocols.OIDCAdapter, logger utils.Logger, baseURL string) *OIDCHandler {
	return &OIDCHandler{
		oidcAdapter: oidcAdapter,
		logger:      logger,
		baseURL:     baseURL,
	}
}

// UserInfo handles the userinfo endpoint.
func (h *OIDCHandler) UserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}
	accessToken := parts[1]

	userInfo, err := h.oidcAdapter.GetUserInfo(r.Context(), accessToken)
	if err != nil {
		h.logger.Error(r.Context(), "Error getting user info", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// Discovery handles the .well-known/openid-configuration endpoint.
func (h *OIDCHandler) Discovery(w http.ResponseWriter, r *http.Request) {
	discovery := map[string]interface{}{
		"issuer":                 h.baseURL,
		"authorization_endpoint": h.baseURL + "/oauth/authorize",
		"token_endpoint":         h.baseURL + "/oauth/token",
		"userinfo_endpoint":      h.baseURL + "/oauth/userinfo",
		"jwks_uri":               h.baseURL + "/.well-known/jwks.json",
		"response_types_supported": []string{"code"},
		"subject_types_supported":  []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":       []string{"openid", "profile", "email"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post"},
		"claims_supported":       []string{"sub", "iss", "aud", "exp", "iat", "name", "email", "email_verified"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discovery)
}

// JWKS handles the .well-known/jwks.json endpoint.
func (h *OIDCHandler) JWKS(w http.ResponseWriter, r *http.Request) {
	jwks := h.oidcAdapter.GetJWKS()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jwks)
}
