package protocols

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/plugins"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"gopkg.in/square/go-jose.v2"
)

// OIDCAdapter implements the IProtocolAdapter for OpenID Connect 1.0.
type OIDCAdapter struct {
	plugins.BasePlugin
	oauthAdapter *OAuthAdapter
	logger       utils.Logger
	privateKey   *rsa.PrivateKey
	userRepo     identity.UserRepository
	jwtSecret    []byte
}

// NewOIDCAdapter is the factory function for this plugin.
func NewOIDCAdapter() plugins.IPlugin {
	return &OIDCAdapter{
		BasePlugin: plugins.BasePlugin{
			PluginName: "oidc_adapter",
			PluginType: types.PluginTypeProtocolAdapter,
		},
		oauthAdapter: &OAuthAdapter{},
	}
}

// Initialize sets up the adapter.
func (a *OIDCAdapter) Initialize(ctx context.Context, config types.ConnectorConfig, logger utils.Logger) error {
	a.logger = logger
	a.oauthAdapter.Initialize(ctx, config, logger)

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(config.PrivateKey))
	if err != nil {
		return types.ErrInternal.WithCause(err)
	}
	a.privateKey = privateKey
	a.jwtSecret = []byte(config.JWTSecret)

	a.logger.Info(ctx, "Initializing OIDC Adapter")
	return nil
}

// HandleAuthRequest processes an incoming OIDC authentication request.
func (a *OIDCAdapter) HandleAuthRequest(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error) {
	return a.oauthAdapter.HandleAuthRequest(ctx, request)
}

// HandleTokenRequest processes an incoming OIDC token request.
func (a *OIDCAdapter) HandleTokenRequest(ctx context.Context, request *types.TokenRequest) (*types.TokenResponse, error) {
	return a.oauthAdapter.HandleTokenRequest(ctx, request)
}

func (a *OIDCAdapter) generateIDToken(ctx context.Context, user *types.User, scope, nonce, clientID string) (string, error) {
	claims := jwt.MapClaims{
		"iss":   "https://localhost:8080",
		"sub":   user.ID,
		"aud":   clientID,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
		"iat":   time.Now().Unix(),
		"nonce": nonce,
	}

	if utils.ScopeContains(scope, "email") {
		claims["email"] = user.Email
		claims["email_verified"] = user.Status == types.UserStatusActive
	}

	if utils.ScopeContains(scope, "profile") {
		claims["name"] = user.Attributes["name"]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(a.privateKey)
}

func (a *OIDCAdapter) GetUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})
	if err != nil {
		return nil, types.ErrInvalidToken.WithCause(err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, types.ErrInvalidToken.WithDetails(map[string]string{"error": "invalid claims"})
	}

	userID := claims["sub"].(string)
	user, err := a.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, types.ErrUserNotFound.WithCause(err)
	}

	userInfo := map[string]interface{}{
		"sub": user.ID,
	}
	if utils.ScopeContains(claims["scope"].(string), "profile") {
		userInfo["name"] = claims["name"]
	}
	if utils.ScopeContains(claims["scope"].(string), "email") {
		userInfo["email"] = claims["email"]
		userInfo["email_verified"] = claims["email_verified"]
	}

	return userInfo, nil
}

// GetJWKS returns the JSON Web Key Set.
func (a *OIDCAdapter) GetJWKS() jose.JSONWebKeySet {
	return jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key:       &a.privateKey.PublicKey,
				KeyID:     "1",
				Algorithm: "RS256",
				Use:       "sig",
			},
		},
	}
}

// SetUserRepo sets the user repository for the adapter.
func (a *OIDCAdapter) SetUserRepo(repo identity.UserRepository) {
	a.userRepo = repo
}

// SetPrivateKey sets the private key for the adapter.
func (a *OIDCAdapter) SetPrivateKey(key *rsa.PrivateKey) {
	a.privateKey = key
}
