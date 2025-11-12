package ldap

// LDAPConfig holds the configuration for the LDAP connector.
type LDAPConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	UseTLS       bool   `mapstructure:"use_tls"`
	BindDN       string `mapstructure:"bind_dn"`
	BindPassword string `mapstructure:"bind_password"`
	BaseDN       string `mapstructure:"base_dn"`
	UserFilter   string `mapstructure:"user_filter"`
	AttrMapping  map[string]string `mapstructure:"attribute_mapping"`
	Sync         SyncConfig `mapstructure:"sync"`
}

// SyncConfig holds the configuration for the user synchronization.
type SyncConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Interval     string `mapstructure:"interval"`
	FullSyncCron string `mapstructure:"full_sync_cron"`
}
