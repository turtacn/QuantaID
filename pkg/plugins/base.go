package plugins

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// BasePlugin provides a skeletal implementation of the IPlugin interface
// to minimize the effort required to implement this interface.
// It's intended to be embedded in concrete plugin implementations.
type BasePlugin struct {
	PluginName string
	PluginType types.PluginType
	Logger     utils.Logger
}

// Name returns the configured name of the plugin.
// This is part of the IPlugin interface implementation.
func (b *BasePlugin) Name() string {
	return b.PluginName
}

// Type returns the configured type of the plugin.
// This is part of the IPlugin interface implementation.
func (b *BasePlugin) Type() types.PluginType {
	return b.PluginType
}

// Initialize sets up the plugin with its configuration and a logger.
// This base implementation simply stores the logger.
// Concrete plugins should override this method to add their own initialization logic.
//
// Parameters:
//   - ctx: The context for the initialization process.
//   - config: The configuration for this specific plugin instance.
//   - logger: The logger for the plugin to use.
//
// Returns:
//   An error if initialization fails, otherwise nil.
func (b *BasePlugin) Initialize(ctx context.Context, config types.ConnectorConfig, logger utils.Logger) error {
	b.Logger = logger
	return nil
}

// Start begins the plugin's operation. This base implementation does nothing.
// Concrete plugins should override this to start any background processes or connections.
//
// Parameters:
//   - ctx: The context for the start process.
//
// Returns:
//   An error if starting fails, otherwise nil.
func (b *BasePlugin) Start(ctx context.Context) error {
	return nil
}

// Stop ceases the plugin's operation. This base implementation does nothing.
// Concrete plugins should override this to gracefully shut down processes and connections.
//
// Parameters:
//   - ctx: The context for the stop process.
//
// Returns:
//   An error if stopping fails, otherwise nil.
func (b *BasePlugin) Stop(ctx context.Context) error {
	return nil
}

// HealthCheck reports the operational status of the plugin. This base implementation always returns nil.
// Concrete plugins should override this to provide a meaningful health check.
//
// Parameters:
//   - ctx: The context for the health check.
//
// Returns:
//   An error if the plugin is unhealthy, otherwise nil.
func (b *BasePlugin) HealthCheck(ctx context.Context) error {
	return nil
}
