package auth

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/pkg/plugins"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// Engine is the central component for orchestrating authentication and authorization flows.
// It uses a plugin-based architecture to support various authentication protocols and identity providers.
type Engine struct {
	pluginManager *plugins.Manager
	logger        utils.Logger
}

// NewEngine creates a new authentication engine.
//
// Parameters:
//   - pluginManager: The plugin manager to use for retrieving authentication plugins.
//   - logger: The logger to use for logging messages.
//
// Returns:
//   A new authentication engine.
func NewEngine(pluginManager *plugins.Manager, logger utils.Logger) *Engine {
	return &Engine{
		pluginManager: pluginManager,
		logger:        logger,
	}
}

// Authenticate processes a generic authentication request.
// It dynamically selects the appropriate protocol adapter plugin based on the request's protocol type
// and delegates the authentication handling to the plugin.
//
// Parameters:
//   - ctx: The context for the authentication request.
//   - request: The authentication request to process.
//
// Returns:
//   The authentication response, or an error if the authentication process fails.
func (e *Engine) Authenticate(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error) {
	e.logger.Info(ctx, "Starting authentication process", zap.String("protocol", string(request.Protocol)))

	pluginName := e.getPluginNameForProtocol(request.Protocol)

	plugin, err := e.pluginManager.GetPlugin(pluginName)
	if err != nil {
		e.logger.Error(ctx, "Failed to get plugin for protocol", zap.Error(err), zap.String("protocol", string(request.Protocol)))
		return nil, types.ErrPluginNotFound.WithCause(err)
	}

	adapter, ok := plugin.(plugins.IProtocolAdapter)
	if !ok {
		err := fmt.Errorf("plugin '%s' does not implement IProtocolAdapter", pluginName)
		e.logger.Error(ctx, "Plugin type mismatch", zap.Error(err))
		return nil, types.ErrPluginLoadFailed.WithCause(err)
	}

	authResponse, err := adapter.HandleAuthRequest(ctx, request)
	if err != nil {
		e.logger.Error(ctx, "Protocol adapter failed to handle auth request", zap.Error(err), zap.String("plugin", pluginName))
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}

	e.logger.Info(ctx, "Authentication process completed", zap.String("protocol", string(request.Protocol)), zap.Bool("success", authResponse.Success))
	return authResponse, nil
}

func (e *Engine) getPluginNameForProtocol(protocol types.ProtocolType) string {
	switch protocol {
	case types.ProtocolOAuth:
		return "oauth2_adapter"
	case types.ProtocolSAML:
		return "saml_adapter"
	case types.ProtocolOIDC:
		return "oidc_adapter"
	default:
		return "default_password_connector"
	}
}

