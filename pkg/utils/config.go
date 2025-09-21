package utils

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
)

// ConfigManager handles application configuration.
type ConfigManager struct {
	v *viper.Viper
}

// NewConfigManager creates a new ConfigManager instance.
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

func (cm *ConfigManager) Get(key string) interface{} { return cm.v.Get(key) }
func (cm *ConfigManager) GetString(key string) string { return cm.v.GetString(key) }
func (cm *ConfigManager) GetInt(key string) int { return cm.v.GetInt(key) }
func (cm *ConfigManager) GetBool(key string) bool { return cm.v.GetBool(key) }
func (cm *ConfigManager) Unmarshal(rawVal interface{}) error { return cm.v.Unmarshal(rawVal) }

func (cm *ConfigManager) watchConfig(logger Logger) {
	cm.v.OnConfigChange(func(e fsnotify.Event) {
		logger.Info(nil, "Configuration file changed, reloading.", zap.String("file", e.Name))
	})
	cm.v.WatchConfig()
}

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
	v.SetDefault("database.url", "postgres://user:password@localhost:5432/quantid?sslmode=disable")
	v.SetDefault("redis.url", "redis://localhost:6379/0")
	v.SetDefault("plugins.directory", "./plugins")
}

func LoadConfigFromBytes(configData []byte, configType string) (*ConfigManager, error) {
	v := viper.New()
	v.SetConfigType(configType)
	setDefaults(v)
	if err := v.ReadConfig(bytes.NewBuffer(configData)); err != nil {
		return nil, fmt.Errorf("failed to read config from bytes: %w", err)
	}
	return &ConfigManager{v: v}, nil
}

//Personal.AI order the ending
