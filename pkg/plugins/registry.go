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
type PluginFactory func() IPlugin

// Registry holds a catalog of available plugins.
type Registry struct {
	mu              sync.RWMutex
	pluginFactories map[string]PluginFactory
	logger          utils.Logger
}

// NewRegistry creates a new plugin registry.
func NewRegistry(logger utils.Logger) *Registry {
	return &Registry{
		pluginFactories: make(map[string]PluginFactory),
		logger:          logger,
	}
}

// RegisterPlugin adds a new plugin factory to the registry.
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

//Personal.AI order the ending
