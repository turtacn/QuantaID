package types

var (
	Version   = "dev"
	BuildDate = "unknown"
	Commit    = "unknown"
)

type ProtocolType string

const (
	ProtocolSAML  ProtocolType = "saml"
	ProtocolOAuth ProtocolType = "oauth2"
	ProtocolOIDC  ProtocolType = "oidc"
	ProtocolLDAP  ProtocolType = "ldap"
)

type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "password"
	AuthMethodTOTP     AuthMethod = "totp"
	AuthMethodSMS      AuthMethod = "sms"
	AuthMethodEmailOTP AuthMethod = "email_otp"
	AuthMethodHardware AuthMethod = "hardware_token"
	AuthMethodBiometric AuthMethod = "biometric"
	AuthMethodJWT      AuthMethod = "jwt"
)

type PermissionLevel string

const (
	PermissionLevelAdmin  PermissionLevel = "admin"
	PermissionLevelUser   PermissionLevel = "user"
	PermissionLevelGuest  PermissionLevel = "guest"
	PermissionLevelSystem PermissionLevel = "system"
)

type ConfigKey string

const (
	ConfigKeyServerAddress  ConfigKey = "server.address"
	ConfigKeyDatabaseURL    ConfigKey = "database.url"
	ConfigKeyRedisURL       ConfigKey = "redis.url"
	ConfigKeyLogLevel       ConfigKey = "log.level"
	ConfigKeyJWTSecret      ConfigKey = "jwt.secret"
	ConfigKeyPluginDir      ConfigKey = "plugins.directory"
)

type SystemLimit int

const (
	LimitMaxUsernameLength SystemLimit = 255
	LimitMinPasswordLength SystemLimit = 8
	LimitMaxSessionPerPage SystemLimit = 100
)

type PluginType string

const (
	PluginTypeIdentityConnector PluginType = "identity_connector"
	PluginTypeMFAProvider       PluginType = "mfa_provider"
	PluginTypeProtocolAdapter   PluginType = "protocol_adapter"
	PluginTypeEventHandler      PluginType = "event_handler"
)

type EventType string

const (
	EventUserLoginSuccess EventType = "user.login.success"
	EventUserLoginFailure EventType = "user.login.failure"
	EventUserCreated      EventType = "user.created"
	EventUserDeleted      EventType = "user.deleted"
	EventPolicyUpdated    EventType = "policy.updated"
)

type CryptoAlgorithm string

const (
	CryptoAlgoBcrypt CryptoAlgorithm = "bcrypt"
	CryptoAlgoScrypt CryptoAlgorithm = "scrypt"
	CryptoAlgoArgon2 CryptoAlgorithm = "argon2"
	CryptoAlgoSHA256 CryptoAlgorithm = "sha256"
)

//Personal.AI order the ending
