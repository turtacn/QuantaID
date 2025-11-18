package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/turtacn/QuantaID/pkg/auth/protocols"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// OAuthHandler handles OAuth 2.1 requests.
type OAuthHandler struct {
	oauthAdapter *protocols.OAuthAdapter
	logger       utils.Logger
}

// NewOAuthHandler creates a new OAuthHandler.
func NewOAuthHandler(oauthAdapter *protocols.OAuthAdapter, logger utils.Logger) *OAuthHandler {
	return &OAuthHandler{
		oauthAdapter: oauthAdapter,
		logger:       logger,
	}
}

// Authorize handles the authorization endpoint.
func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement proper session management to get the user ID
	req := &types.AuthRequest{
		Protocol: "oauth",
		Credentials: map[string]string{
			"response_type":         r.URL.Query().Get("response_type"),
			"client_id":             r.URL.Query().Get("client_id"),
			"redirect_uri":          r.URL.Query().Get("redirect_uri"),
			"scope":                 r.URL.Query().Get("scope"),
			"state":                 r.URL.Query().Get("state"),
			"code_challenge":        r.URL.Query().Get("code_challenge"),
			"code_challenge_method": r.URL.Query().Get("code_challenge_method"),
			"user_id":               "test_user",
		},
	}

	resp, err := h.oauthAdapter.HandleAuthRequest(r.Context(), req)
	if err != nil {
		h.logger.Error(r.Context(), "Error handling auth request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	redirectURL := resp.RedirectURI + "?code=" + resp.Code + "&state=" + resp.State
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// Token handles the token endpoint.
func (h *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	req := types.TokenRequest{
		GrantType:    r.Form.Get("grant_type"),
		Code:         r.Form.Get("code"),
		RedirectURI:  r.Form.Get("redirect_uri"),
		ClientID:     r.Form.Get("client_id"),
		ClientSecret: r.Form.Get("client_secret"),
		CodeVerifier: r.Form.Get("code_verifier"),
		RefreshToken: r.Form.Get("refresh_token"),
	}

	resp, err := h.oauthAdapter.HandleTokenRequest(r.Context(), &req)
	if err != nil {
		h.logger.Error(r.Context(), "Error handling token request", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
