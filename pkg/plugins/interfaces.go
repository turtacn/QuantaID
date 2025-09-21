package plugins

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// IPlugin is the base interface for all plugins.
type IPlugin interface {
	Name() string
	Type() types.PluginType
	Initialize(ctx context.Context, config types.ConnectorConfig, logger utils.Logger) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	HealthCheck(ctx context.Context) error
}

// IIdentityConnector defines the interface for plugins that connect to external identity sources.
type IIdentityConnector interface {
	IPlugin
	Authenticate(ctx context.Context, credentials map[string]string) (*types.AuthResponse, error)
	GetUser(ctx context.Context, identifier string) (*types.User, error)
	GetGroup(ctx context.Context, identifier string) (*types.UserGroup, error)
	SyncUsers(ctx context.Context) ([]*types.User, error)
	SyncGroups(ctx context.Context) ([]*types.UserGroup, error)
}

// IMFAProvider defines the interface for plugins that provide multi-factor authentication methods.
type IMFAProvider interface {
	IPlugin
	SendChallenge(ctx context.Context, user *types.User) (*types.MFAChallenge, error)
	VerifyChallenge(ctx context.Context, challengeID string, code string) (bool, error)
}

// IProtocolAdapter defines the interface for plugins that handle specific authentication protocols.
type IProtocolAdapter interface {
	IPlugin
	HandleAuthRequest(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error)
}

// IEventHandler defines the interface for plugins that react to system events.
type IEventHandler interface {
	IPlugin
	HandleEvent(ctx context.Context, eventType types.EventType, payload interface{}) error
	SubscribedEvents() []types.EventType
}

//Personal.AI order the ending
