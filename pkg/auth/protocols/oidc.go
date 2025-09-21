package protocols

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/plugins"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"strings"
)

// OIDCAdapter implements the IProtocolAdapter for OpenID Connect 1.0.
type OIDCAdapter struct {
	plugins.BasePlugin
	oauthAdapter *OAuthAdapter
	logger       utils.Logger
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
	a.logger.Info(ctx, "Initializing OIDC Adapter")
	return nil
}

// HandleAuthRequest processes an incoming OIDC authentication request.
func (a *OIDCAdapter) HandleAuthRequest(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error) {
	authResponse, err := a.oauthAdapter.HandleAuthRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	if authResponse.Success {
		scopes, ok := request.Credentials["scope"]
		if !ok || !strings.Contains(scopes, "openid") {
			return authResponse, nil
		}

		idToken, idErr := a.generateIDToken(ctx, authResponse.User, scopes)
		if idErr != nil {
			return nil, types.ErrInternal.WithCause(idErr)
		}
		authResponse.Token.IDToken = idToken
	}

	return authResponse, nil
}

func (a *OIDCAdapter) generateIDToken(ctx context.Context, user *types.User, scopes string) (string, error) {
	a.logger.Info(ctx, "Generating ID Token for user", zap.String("userID", user.ID))
	return "stub_id_token", types.ErrInternal.WithDetails(map[string]string{"status": "not implemented"})
}

func (a *OIDCAdapter) GetUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	a.logger.Info(ctx, "Handling UserInfo request")
	return nil, types.ErrInternal.WithDetails(map[string]string{"status": "not implemented"})
}

//Personal.AI order the ending
