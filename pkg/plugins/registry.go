package plugins

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"sync"
)

// PluginFactory is a function that creates a new instance of a plugin.
// It is used by the registry to instantiate plugins on demand.
type PluginFactory func() IPlugin

// Registry holds a catalog of available plugins, mapping plugin names to their
// factory functions. It provides a central point for discovering and instantiating plugins.
type Registry struct {
	mu              sync.RWMutex
	pluginFactories map[string]PluginFactory
	logger          utils.Logger
}

// NewRegistry creates a new, empty plugin registry.
//
// Parameters:
//   - logger: The logger for the registry to use.
//
// Returns:
//   A new plugin registry.
func NewRegistry(logger utils.Logger) *Registry {
	return &Registry{
		pluginFactories: make(map[string]PluginFactory),
		logger:          logger,
	}
}

// RegisterPlugin adds a new plugin factory to the registry.
// It is typically called during application startup to build the catalog of available plugins.
//
// Parameters:
//   - name: The name to register the plugin under. This must be unique.
//   - factory: The factory function that creates an instance of the plugin.
//
// Returns:
//   An error if a plugin with the same name is already registered, otherwise nil.
func (r *Registry) RegisterPlugin(name string, factory PluginFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.pluginFactories[name]; exists {
		return fmt.Errorf("plugin with name '%s' is already registered", name)
	}

	r.pluginFactories[name] = factory
	r.logger.Info(context.Background(), "Plugin registered successfully", zap.String("plugin", name))
	return nil
}

// GetPluginFactory retrieves a plugin factory by its name.
// This allows the plugin manager to create new instances of a plugin.
//
// Parameters:
//   - name: The name of the plugin factory to retrieve.
//
// Returns:
//   The plugin factory if found, otherwise an error.
func (r *Registry) GetPluginFactory(name string) (PluginFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.pluginFactories[name]
	if !exists {
		return nil, types.ErrPluginNotFound.WithDetails(map[string]string{"name": name})
	}
	return factory, nil
}

// ListPlugins returns a list of all registered plugin names.
//
// Returns:
//   A slice of strings containing the names of all registered plugins.
func (r *Registry) ListPlugins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.pluginFactories))
	for name := range r.pluginFactories {
		names = append(names, name)
	}
	return names
}

// ListPluginsByType returns a list of registered plugin names of a specific type.
//
// Parameters:
//   - pluginType: The type of plugins to list.
//
// Returns:
//   A slice of strings containing the names of registered plugins that match the specified type.
func (r *Registry) ListPluginsByType(pluginType types.PluginType) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0)
	for name, factory := range r.pluginFactories {
		p := factory()
		if p.Type() == pluginType {
			names = append(names, name)
		}
	}
	return names
}

