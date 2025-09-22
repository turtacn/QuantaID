package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfigManager_LoadFromBytes(t *testing.T) {
	yamlData := []byte(`
server:
  address: ":9090"
logger:
  level: "debug"
`)

	cm, err := LoadConfigFromBytes(yamlData, "yaml")
	require.NoError(t, err)
	require.NotNil(t, cm)

	// Test that a loaded value overrides a default
	assert.Equal(t, ":9090", cm.GetString("server.address"))

	// Test that a default value is still present if not overridden
	assert.Equal(t, "json", cm.GetString("logger.format"))

	// Test that a value from the loaded config is present
	assert.Equal(t, "debug", cm.GetString("logger.level"))
}

func TestConfigManager_Defaults(t *testing.T) {
	// Load an empty config to only test defaults
	cm, err := LoadConfigFromBytes([]byte{}, "yaml")
	require.NoError(t, err)
	require.NotNil(t, cm)

	assert.Equal(t, ":8080", cm.GetString("server.address"))
	assert.Equal(t, "info", cm.GetString("logger.level"))
	assert.Equal(t, "./plugins", cm.GetString("plugins.directory"))
}

type FullConfig struct {
	Server ServerConfig `mapstructure:"server"`
	Logger LoggerConfig `mapstructure:"logger"`
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
}

func TestConfigManager_Unmarshal(t *testing.T) {
	yamlData := []byte(`
server:
  address: ":1234"
logger:
  level: "warn"
`)
	cm, err := LoadConfigFromBytes(yamlData, "yaml")
	require.NoError(t, err)

	var config FullConfig
	err = cm.Unmarshal(&config)
	require.NoError(t, err)

	assert.Equal(t, ":1234", config.Server.Address)
	assert.Equal(t, "warn", config.Logger.Level)
}

//Personal.AI order the ending
