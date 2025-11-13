package config

import (
	"fmt"
	"github.com/turtacn/QuantaID/internal/audit"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

// AuditSinkConfig defines the configuration for a single audit sink.
type AuditSinkConfig struct {
	Type string                 `yaml:"type"`
	Path string                 `yaml:"path,omitempty"` // For file sink
	// Add other sink-specific fields here, e.g., for Kafka, Syslog
}

// AuditConfig defines the configuration for the entire audit pipeline.
type AuditConfig struct {
	Sinks []AuditSinkConfig `yaml:"sinks"`
}

// LoadAuditConfig loads the audit configuration from a YAML file.
func LoadAuditConfig(path string) (*AuditConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit config file: %w", err)
	}

	var cfg AuditConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit config: %w", err)
	}

	return &cfg, nil
}

// NewPipelineFromConfig creates a new audit pipeline from a configuration.
func NewPipelineFromConfig(cfg *AuditConfig, logger *zap.Logger) (*audit.Pipeline, error) {
	var sinks []audit.Sink
	for _, sinkCfg := range cfg.Sinks {
		switch sinkCfg.Type {
		case "file":
			if sinkCfg.Path == "" {
				return nil, fmt.Errorf("file sink requires a path")
			}
			sink, err := audit.NewFileSink(sinkCfg.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to create file sink: %w", err)
			}
			sinks = append(sinks, sink)
		case "stdout":
			sink := audit.NewStdoutSink()
			sinks = append(sinks, sink)
		// In a real scenario, you would add cases for "kafka", "syslog", etc.
		// These are omitted here as they are not required for the Jules environment.
		default:
			return nil, fmt.Errorf("unsupported audit sink type: %s", sinkCfg.Type)
		}
	}

	return audit.NewPipeline(logger, sinks...), nil
}
