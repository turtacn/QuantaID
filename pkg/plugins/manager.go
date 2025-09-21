package plugins

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"sync"
)

// Manager handles the lifecycle of all plugins.
type Manager struct {
	mu             sync.RWMutex
	registry       *Registry
	logger         utils.Logger
	config         *utils.ConfigManager
	activePlugins  map[string]IPlugin
}

// NewManager creates a new plugin manager.
func NewManager(registry *Registry, logger utils.Logger, config *utils.ConfigManager) *Manager {
	return &Manager{
		registry:      registry,
		logger:        logger,
		config:        config,
		activePlugins: make(map[string]IPlugin),
	}
}

// LoadAndStartPlugins loads all plugins specified in the configuration and starts them.
func (m *Manager) LoadAndStartPlugins(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var pluginConfigs []PluginConfig
	// In a real system, you'd unmarshal a "plugins.providers" section from your config file.
	// m.config.UnmarshalKey("plugins.providers", &pluginConfigs)

	for _, pconf := range pluginConfigs {
		if !pconf.Enabled {
			m.logger.Info(ctx, "Skipping disabled plugin", zap.String("plugin", pconf.Name))
			continue
		}

		m.logger.Info(ctx, "Loading plugin", zap.String("plugin", pconf.Name))
		factory, err := m.registry.GetPluginFactory(pconf.Name)
		if err != nil {
			m.logger.Error(ctx, "Failed to get plugin factory", zap.String("plugin", pconf.Name), zap.Error(err))
			return err
		}

		plugin := factory()

		connectorConfig := types.ConnectorConfig{
			InstanceID: plugin.Name(),
			ProviderID: plugin.Name(),
			Config:     pconf.Settings,
		}

		if err := plugin.Initialize(ctx, connectorConfig, m.logger); err != nil {
			m.logger.Error(ctx, "Failed to initialize plugin", zap.String("plugin", plugin.Name()), zap.Error(err))
			return err
		}

		if err := plugin.Start(ctx); err != nil {
			m.logger.Error(ctx, "Failed to start plugin", zap.String("plugin", plugin.Name()), zap.Error(err))
			return err
		}

		m.activePlugins[plugin.Name()] = plugin
		m.logger.Info(ctx, "Plugin started successfully", zap.String("plugin", plugin.Name()))
	}

	return nil
}

// StopAllPlugins gracefully stops all running plugins.
func (m *Manager) StopAllPlugins(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, plugin := range m.activePlugins {
		m.logger.Info(ctx, "Stopping plugin", zap.String("plugin", name))
		if err := plugin.Stop(ctx); err != nil {
			m.logger.Error(ctx, "Failed to stop plugin", zap.String("plugin", name), zap.Error(err))
		}
	}
}

// GetPlugin retrieves a running plugin by its name.
func (m *Manager) GetPlugin(name string) (IPlugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.activePlugins[name]
	if !exists {
		return nil, types.ErrPluginNotFound.WithDetails(map[string]string{"name": name})
	}

	return plugin, nil
}

// PluginConfig represents the configuration for a single plugin instance.
type PluginConfig struct {
	Name     string                 `yaml:"name"`
	Enabled  bool                   `yaml:"enabled"`
	Settings map[string]interface{} `yaml:"settings"`
}

//Personal.AI order the ending
