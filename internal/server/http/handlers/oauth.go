package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	auth_service "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
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

// NewOAuthHandlers creates a new OAuthHandler.
func NewOAuthHandlers(authService *auth_service.ApplicationService, identityService identity.IService, logger utils.Logger) *OAuthHandler {
	logger.Info(context.Background(), "Creating new OAuth handlers",
		zap.Any("authService", authService != nil),
		zap.Any("identityService", identityService != nil),
		zap.Any("userRepo", authService.GetUserRepo() != nil),
		zap.Any("appRepo", authService.GetAppRepo() != nil),
		zap.Any("redisClient", authService.GetRedisClient() != nil),
		zap.Any("cryptoManager", authService.GetCryptoManager() != nil),
	)
	adapter := protocols.NewOAuthAdapter().(*protocols.OAuthAdapter)
	adapter.SetUserRepo(authService.GetUserRepo())
	adapter.SetAppRepo(authService.GetAppRepo())
	adapter.SetRedis(authService.GetRedisClient())
	adapter.SetOIDCAdapter(protocols.NewOIDCAdapter(authService.GetCryptoManager().(*utils.CryptoManager)))

	return &OAuthHandler{
		oauthAdapter: adapter,
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

func (h *OAuthHandler) Discovery(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *OAuthHandler) JWKS(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
