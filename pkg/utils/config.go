package utils

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
)

// PostgresConfig holds all configuration for the PostgreSQL database connection.
type PostgresConfig struct {
	DSN             string `mapstructure:"dsn"`
	MaxIdleConns    int    `mapstructure:"maxIdleConns"`
	MaxOpenConns    int    `mapstructure:"maxOpenConns"`
	ConnMaxLifetime string `mapstructure:"connMaxLifetime"`
}

// StorageConfig holds the configuration for the storage layer.
type StorageConfig struct {
	Mode string `mapstructure:"mode"` // "postgres" / "memory"
}

// Config holds all configuration for the application.
type Config struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

// ConfigManager is a wrapper around the Viper library that handles loading,
// accessing, and watching application configuration.
type ConfigManager struct {
	v *viper.Viper
}

// NewConfigManager creates a new ConfigManager instance by reading from a configuration file.
// It sets up default values, reads from environment variables, and watches the config file for changes.
//
// Parameters:
//   - configPath: The directory where the configuration file is located.
//   - configName: The name of the configuration file (without extension).
//   - configType: The type of the configuration file (e.g., "yaml", "json").
//   - logger: A logger instance for logging messages.
//
// Returns:
//   A new ConfigManager instance or an error if the configuration file cannot be read.
func NewConfigManager(configPath, configName, configType string, logger Logger) (*ConfigManager, error) {
	v := viper.New()
	v.AddConfigPath(configPath)
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	setDefaults(v)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Warn(nil, "Config file not found, using defaults and environment variables", zap.String("path", configPath))
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	cm := &ConfigManager{v: v}
	cm.watchConfig(logger)

	return cm, nil
}

// Get retrieves a generic configuration value by key.
func (cm *ConfigManager) Get(key string) interface{} { return cm.v.Get(key) }
// GetString retrieves a string configuration value by key.
func (cm *ConfigManager) GetString(key string) string { return cm.v.GetString(key) }
// GetInt retrieves an integer configuration value by key.
func (cm *ConfigManager) GetInt(key string) int { return cm.v.GetInt(key) }
// GetBool retrieves a boolean configuration value by key.
func (cm *ConfigManager) GetBool(key string) bool { return cm.v.GetBool(key) }
// Unmarshal decodes the entire configuration into a struct.
func (cm *ConfigManager) Unmarshal(rawVal interface{}) error { return cm.v.Unmarshal(rawVal) }

// watchConfig sets up a file watcher that reloads the configuration when the file changes.
func (cm *ConfigManager) watchConfig(logger Logger) {
	cm.v.OnConfigChange(func(e fsnotify.Event) {
		logger.Info(nil, "Configuration file changed, reloading.", zap.String("file", e.Name))
	})
	cm.v.WatchConfig()
}

// setDefaults establishes default values for essential configuration keys.
// This ensures that the application can run with a minimal configuration.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.address", ":8080")
	v.SetDefault("server.readTimeout", "15s")
	v.SetDefault("server.writeTimeout", "15s")
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.console.enabled", true)
	v.SetDefault("logger.file.enabled", false)
	v.SetDefault("logger.file.path", "/var/log/quantid.log")
	v.SetDefault("logger.file.maxSizeMB", 100)
	v.SetDefault("logger.file.maxBackups", 3)
	v.SetDefault("logger.file.maxAgeDays", 7)
	v.SetDefault("logger.file.compress", false)
	v.SetDefault("postgres.dsn", "postgres://user:password@localhost:5432/quantid?sslmode=disable")
	v.SetDefault("postgres.maxIdleConns", 10)
	v.SetDefault("postgres.maxOpenConns", 100)
	v.SetDefault("postgres.connMaxLifetime", "1h")
	v.SetDefault("redis.url", "redis://localhost:6379/0")
	v.SetDefault("plugins.directory", "./plugins")
}

// LoadConfigFromBytes creates a new ConfigManager by reading configuration data from a byte slice.
// This is particularly useful for tests or for loading configuration from non-file sources.
//
// Parameters:
//   - configData: The byte slice containing the configuration data.
//   - configType: The format of the configuration data (e.g., "yaml").
//
// Returns:
//   A new ConfigManager instance or an error if the data cannot be parsed.
func LoadConfigFromBytes(configData []byte, configType string) (*ConfigManager, error) {
	v := viper.New()
	v.SetConfigType(configType)
	setDefaults(v)
	if err := v.ReadConfig(bytes.NewBuffer(configData)); err != nil {
		return nil, fmt.Errorf("failed to read config from bytes: %w", err)
	}
	return &ConfigManager{v: v}, nil
}
