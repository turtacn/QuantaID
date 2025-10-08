package plugins

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// IPlugin is the fundamental interface for all plugins in the QuantaID system.
// It defines the basic lifecycle and identification methods that every plugin must implement.
type IPlugin interface {
	// Name returns the unique name of the plugin instance.
	Name() string
	// Type returns the category of the plugin.
	Type() types.PluginType
	// Initialize configures the plugin with its specific settings and a logger.
	Initialize(ctx context.Context, config types.ConnectorConfig, logger utils.Logger) error
	// Start activates the plugin, allowing it to begin its operations.
	Start(ctx context.Context) error
	// Stop gracefully shuts down the plugin.
	Stop(ctx context.Context) error
	// HealthCheck verifies the operational status of the plugin.
	HealthCheck(ctx context.Context) error
}

// IIdentityConnector defines the interface for plugins that connect to external identity sources,
// such as LDAP, Active Directory, or a SQL database.
type IIdentityConnector interface {
	IPlugin
	// Authenticate validates user credentials against the external identity source.
	Authenticate(ctx context.Context, credentials map[string]string) (*types.AuthResponse, error)
	// GetUser retrieves a user's profile from the identity source.
	GetUser(ctx context.Context, identifier string) (*types.User, error)
	// GetGroup retrieves a group's profile from the identity source.
	GetGroup(ctx context.Context, identifier string) (*types.UserGroup, error)
	// SyncUsers fetches all users from the identity source for synchronization.
	SyncUsers(ctx context.Context) ([]*types.User, error)
	// SyncGroups fetches all groups from the identity source for synchronization.
	SyncGroups(ctx context.Context) ([]*types.UserGroup, error)
}

// IMFAProvider defines the interface for plugins that provide multi-factor authentication methods,
// like TOTP, SMS, or FIDO2/WebAuthn.
type IMFAProvider interface {
	IPlugin
	// SendChallenge initiates an MFA challenge for a user, e.g., sending a code via SMS.
	SendChallenge(ctx context.Context, user *types.User) (*types.MFAChallenge, error)
	// VerifyChallenge checks if the user's response to an MFA challenge is correct.
	VerifyChallenge(ctx context.Context, challengeID string, code string) (bool, error)
}

// IProtocolAdapter defines the interface for plugins that handle specific authentication protocols
// like OAuth 2.0, SAML 2.0, or OIDC.
type IProtocolAdapter interface {
	IPlugin
	// HandleAuthRequest processes an incoming authentication request for a specific protocol.
	HandleAuthRequest(ctx context.Context, request *types.AuthRequest) (*types.AuthResponse, error)
}

// IEventHandler defines the interface for plugins that react to system events.
// This allows for custom workflows, notifications, or logging based on events.
type IEventHandler interface {
	IPlugin
	// HandleEvent is called when a subscribed event occurs.
	HandleEvent(ctx context.Context, eventType types.EventType, payload interface{}) error
	// SubscribedEvents returns a list of event types this plugin is interested in.
	SubscribedEvents() []types.EventType
}
