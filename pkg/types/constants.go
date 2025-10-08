package types

// Build-time variables, injected by the linker.
var (
	Version   = "dev"     // Version number of the application.
	BuildDate = "unknown" // BuildDate is the date when the binary was built.
	Commit    = "unknown" // Commit is the git commit hash of the build.
)

// ProtocolType defines the types of authentication protocols supported.
type ProtocolType string

// Supported authentication protocols.
const (
	ProtocolSAML  ProtocolType = "saml"
	ProtocolOAuth ProtocolType = "oauth2"
	ProtocolOIDC  ProtocolType = "oidc"
	ProtocolLDAP  ProtocolType = "ldap"
)

// AuthMethod defines the types of authentication methods or factors.
type AuthMethod string

// Supported authentication methods.
const (
	AuthMethodPassword AuthMethod = "password"
	AuthMethodTOTP     AuthMethod = "totp"
	AuthMethodSMS      AuthMethod = "sms"
	AuthMethodEmailOTP AuthMethod = "email_otp"
	AuthMethodHardware AuthMethod = "hardware_token"
	AuthMethodBiometric AuthMethod = "biometric"
	AuthMethodJWT      AuthMethod = "jwt"
)

// PermissionLevel defines the roles or permission levels within the system.
type PermissionLevel string

// System permission levels.
const (
	PermissionLevelAdmin  PermissionLevel = "admin"
	PermissionLevelUser   PermissionLevel = "user"
	PermissionLevelGuest  PermissionLevel = "guest"
	PermissionLevelSystem PermissionLevel = "system"
)

// ConfigKey represents a key in the application's configuration.
type ConfigKey string

// Common configuration keys.
const (
	ConfigKeyServerAddress  ConfigKey = "server.address"
	ConfigKeyDatabaseURL    ConfigKey = "database.url"
	ConfigKeyRedisURL       ConfigKey = "redis.url"
	ConfigKeyLogLevel       ConfigKey = "log.level"
	ConfigKeyJWTSecret      ConfigKey = "jwt.secret"
	ConfigKeyPluginDir      ConfigKey = "plugins.directory"
)

// SystemLimit defines various system-wide limits.
type SystemLimit int

// System-wide limits.
const (
	LimitMaxUsernameLength SystemLimit = 255
	LimitMinPasswordLength SystemLimit = 8
	LimitMaxSessionPerPage SystemLimit = 100
)

// PluginType defines the categories of plugins.
type PluginType string

// Supported plugin types.
const (
	PluginTypeIdentityConnector PluginType = "identity_connector"
	PluginTypeMFAProvider       PluginType = "mfa_provider"
	PluginTypeProtocolAdapter   PluginType = "protocol_adapter"
	PluginTypeEventHandler      PluginType = "event_handler"
)

// EventType defines the types of system events that can be audited or handled.
type EventType string

// System event types.
const (
	EventUserLoginSuccess EventType = "user.login.success"
	EventUserLoginFailure EventType = "user.login.failure"
	EventUserCreated      EventType = "user.created"
	EventUserDeleted      EventType = "user.deleted"
	EventPolicyUpdated    EventType = "policy.updated"
)

// CryptoAlgorithm defines supported cryptographic algorithms, primarily for hashing.
type CryptoAlgorithm string

// Supported cryptographic hashing algorithms.
const (
	CryptoAlgoBcrypt CryptoAlgorithm = "bcrypt"
	CryptoAlgoScrypt CryptoAlgorithm = "scrypt"
	CryptoAlgoArgon2 CryptoAlgorithm = "argon2"
	CryptoAlgoSHA256 CryptoAlgorithm = "sha256"
)
