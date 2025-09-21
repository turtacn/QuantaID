package plugins

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// BasePlugin provides a default implementation for the IPlugin interface.
type BasePlugin struct {
	PluginName string
	PluginType types.PluginType
	Logger     utils.Logger
}

func (b *BasePlugin) Name() string {
	return b.PluginName
}

func (b *BasePlugin) Type() types.PluginType {
	return b.PluginType
}

func (b *BasePlugin) Initialize(ctx context.Context, config types.ConnectorConfig, logger utils.Logger) error {
	b.Logger = logger
	return nil
}

func (b *BasePlugin) Start(ctx context.Context) error {
	return nil
}

func (b *BasePlugin) Stop(ctx context.Context) error {
	return nil
}

func (b *BasePlugin) HealthCheck(ctx context.Context) error {
	return nil
}

//Personal.AI order the ending
