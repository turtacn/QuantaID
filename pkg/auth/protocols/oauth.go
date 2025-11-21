package protocols

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/plugins"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// OAuthAdapter implements the IProtocolAdapter for OAuth 2.1.
type OAuthAdapter struct {
	plugins.BasePlugin
	logger      utils.Logger
	redis       redis.RedisClientInterface
	appRepo     types.ApplicationRepository
	userRepo    identity.UserRepository
	jwtSecret   []byte
	oidcAdapter *OIDCAdapter
}

// NewOAuthAdapter is the factory function for this plugin.
func NewOAuthAdapter() plugins.IPlugin {
	return &OAuthAdapter{
		BasePlugin: plugins.BasePlugin{
			PluginName: "oauth2_adapter",
			PluginType: types.PluginTypeProtocolAdapter,
		},
	}
}

// Initialize sets up the adapter.
func (a *OAuthAdapter) Initialize(ctx context.Context, config types.ConnectorConfig, logger utils.Logger) error {
	a.logger = logger
	redisMetrics := redis.NewMetrics("oauth", prometheus.DefaultRegisterer)
	redisClient, err := redis.NewRedisClient(&redis.RedisConfig{
		Host: config.RedisURL,
	}, redisMetrics)
	if err != nil {
		return err
	}
	a.redis = redisClient
	a.jwtSecret = []byte(config.JWTSecret)
	a.logger.Info(ctx, "Initializing OAuth 2.1 Adapter")
	return nil
}

// HandleAuthRequest processes an incoming OAuth 2.1 authentication request.
func (a *OAuthAdapter) HandleAuthRequest(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error) {
	if a.appRepo == nil {
		panic("OAuthAdapter.appRepo is nil in HandleAuthRequest")
	}
	oauthRequest := request.Credentials
	a.logger.Info(ctx, "handling auth request", zap.Any("request", oauthRequest))
	if oauthRequest["response_type"] != "code" {
		return nil, types.ErrInvalidRequest.WithDetails(map[string]string{"error": "unsupported response_type"})
	}

	app, err := a.appRepo.GetApplicationByClientID(ctx, oauthRequest["client_id"])
	if err != nil {
		return nil, types.ErrInvalidClient.WithCause(err)
	}

	redirectURIs, ok := app.ProtocolConfig["redirect_uris"].([]string)
	if !ok {
		return nil, types.ErrInvalidClient.WithDetails(map[string]string{"error": "invalid client configuration"})
	}

	validRedirectURI := false
	for _, uri := range redirectURIs {
		if uri == oauthRequest["redirect_uri"] {
			validRedirectURI = true
			break
		}
	}
	if !validRedirectURI {
		return nil, types.ErrInvalidRequest.WithDetails(map[string]string{"error": "invalid redirect_uri"})
	}

	if app.ClientType == types.ClientTypePublic {
		if oauthRequest["code_challenge"] == "" || oauthRequest["code_challenge_method"] == "" {
			return nil, types.ErrInvalidRequest.WithDetails(map[string]string{"error": "public clients must use pkce"})
		}
	}

	code := generateRandomString(32)
	authCodeData := map[string]interface{}{
		"user_id":          oauthRequest["user_id"],
		"client_id":        oauthRequest["client_id"],
		"redirect_uri":     oauthRequest["redirect_uri"],
		"code_challenge":   oauthRequest["code_challenge"],
		"challenge_method": oauthRequest["code_challenge_method"],
		"scope":            oauthRequest["scope"],
		"nonce":            oauthRequest["nonce"],
	}
	authCodeJSON, _ := json.Marshal(authCodeData)

	if err := a.redis.Set(ctx, "authcode:"+code, authCodeJSON, 10*time.Minute); err != nil {
		return nil, types.ErrInternal.WithCause(err)
	}

	return &types.AuthResponse{
		Code:        code,
		RedirectURI: oauthRequest["redirect_uri"],
		State:       oauthRequest["state"],
	}, nil
}

// HandleTokenRequest processes an incoming OAuth 2.1 token request.
func (a *OAuthAdapter) HandleTokenRequest(ctx context.Context, request *types.TokenRequest) (*types.TokenResponse, error) {
	switch request.GrantType {
	case "authorization_code":
		return a.handleAuthorizationCode(ctx, request)
	case "refresh_token":
		return a.handleRefreshToken(ctx, request)
	case "client_credentials":
		return a.handleClientCredentials(ctx, request)
	default:
		return nil, types.ErrUnsupportedGrantType
	}
}

func (a *OAuthAdapter) handleAuthorizationCode(ctx context.Context, request *types.TokenRequest) (*types.TokenResponse, error) {
	val, err := a.redis.Get(ctx, "authcode:"+request.Code)
	if err != nil {
		return nil, types.ErrInvalidGrant.WithCause(err)
	}

	var authCodeData map[string]interface{}
	json.Unmarshal([]byte(val), &authCodeData)

	if !auth.VerifyPKCE(request.CodeVerifier, authCodeData["code_challenge"].(string), authCodeData["challenge_method"].(string)) {
		return nil, types.ErrInvalidGrant.WithDetails(map[string]string{"error": "pkce verification failed"})
	}

	if err := a.redis.Del(ctx, "authcode:"+request.Code); err != nil {
		a.logger.Error(ctx, "failed to delete auth code", zap.Error(err))
	}

	user, err := a.userRepo.GetUserByID(ctx, authCodeData["user_id"].(string))
	if err != nil {
		return nil, types.ErrInternal.WithCause(err)
	}

	return a.generateTokens(ctx, user, authCodeData["scope"].(string), authCodeData["nonce"].(string), authCodeData["client_id"].(string))
}

func (a *OAuthAdapter) handleRefreshToken(ctx context.Context, request *types.TokenRequest) (*types.TokenResponse, error) {
	val, err := a.redis.Get(ctx, "refresh_token:"+request.RefreshToken)
	if err != nil {
		return nil, types.ErrInvalidGrant.WithCause(err)
	}

	var tokenData map[string]interface{}
	json.Unmarshal([]byte(val), &tokenData)

	user, err := a.userRepo.GetUserByID(ctx, tokenData["user_id"].(string))
	if err != nil {
		return nil, types.ErrInternal.WithCause(err)
	}

	if err := a.redis.Del(ctx, "refresh_token:"+request.RefreshToken); err != nil {
		a.logger.Error(ctx, "failed to delete refresh token", zap.Error(err))
	}

	return a.generateTokens(ctx, user, tokenData["scope"].(string), "", tokenData["client_id"].(string))
}

func (a *OAuthAdapter) handleClientCredentials(ctx context.Context, request *types.TokenRequest) (*types.TokenResponse, error) {
	app, err := a.appRepo.GetApplicationByClientID(ctx, request.ClientID)
	if err != nil {
		return nil, types.ErrInvalidClient.WithCause(err)
	}

	clientSecret, ok := app.ProtocolConfig["client_secret"].(string)
	if !ok {
		return nil, types.ErrInvalidClient.WithDetails(map[string]string{"error": "invalid client configuration"})
	}
	if clientSecret != request.ClientSecret {
		return nil, types.ErrInvalidClient.WithDetails(map[string]string{"error": "invalid client_secret"})
	}

	return a.generateTokens(ctx, nil, "", "", request.ClientID)
}

func (a *OAuthAdapter) generateTokens(ctx context.Context, user *types.User, scope, nonce, clientID string) (*types.TokenResponse, error) {
	accessToken, err := a.generateAccessToken(ctx, user, scope)
	if err != nil {
		return nil, err
	}

	refreshToken := generateRandomString(64)
	if user != nil {
		refreshTokenData := map[string]interface{}{
			"user_id":   user.ID,
			"scope":     scope,
			"client_id": clientID,
		}
		marshalledData, _ := json.Marshal(refreshTokenData)
		if err := a.redis.Set(ctx, "refresh_token:"+refreshToken, marshalledData, 7*24*time.Hour); err != nil {
			return nil, types.ErrInternal.WithCause(err)
		}
	}

	idToken := ""
	if user != nil && utils.ScopeContains(scope, "openid") {
		idToken, err = a.oidcAdapter.generateIDToken(ctx, user, scope, nonce, clientID)
		if err != nil {
			return nil, err
		}
	}

	return &types.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		Scope:        scope,
	}, nil
}

func (a *OAuthAdapter) generateAccessToken(ctx context.Context, user *types.User, scope string) (string, error) {
	claims := jwt.MapClaims{
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
		"iat":   time.Now().Unix(),
		"scope": scope,
	}
	if user != nil {
		claims["sub"] = user.ID
	}

	if utils.ScopeContains(scope, "email") {
		claims["email"] = user.Email
		claims["email_verified"] = user.Status == types.UserStatusActive
	}

	if utils.ScopeContains(scope, "profile") {
		claims["name"] = user.Attributes["name"]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

// SetUserRepo sets the user repository for the adapter.
func (a *OAuthAdapter) SetUserRepo(repo identity.UserRepository) {
	a.userRepo = repo
}

// SetAppRepo sets the application repository for the adapter.
func (a *OAuthAdapter) SetAppRepo(repo types.ApplicationRepository) {
	a.appRepo = repo
}

// SetRedis sets the redis client for the adapter.
func (a *OAuthAdapter) SetRedis(redis redis.RedisClientInterface) {
	a.redis = redis
}

// SetOIDCAdapter sets the oidc adapter for the adapter.
func (a *OAuthAdapter) SetOIDCAdapter(adapter *OIDCAdapter) {
	a.oidcAdapter = adapter
}
