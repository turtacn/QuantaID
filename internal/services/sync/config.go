package sync

import "time"

type ConflictStrategy string

const (
	ConflictPreferLocal  ConflictStrategy = "prefer_local"
	ConflictPreferRemote ConflictStrategy = "prefer_remote"
)

type LifecycleRule struct {
	SourceAttr   string `yaml:"source_attr"`
	MatchValue   string `yaml:"match_value"`
	TargetStatus string `yaml:"target_status"`
}

type LDAPSyncConfig struct {
	FullSyncInterval    time.Duration    `yaml:"full_sync_interval"`
	IncrementalInterval time.Duration    `yaml:"incremental_interval"`
	ConflictStrategy    ConflictStrategy `yaml:"conflict_strategy"`
	LifecycleRules      []LifecycleRule  `yaml:"lifecycle_rules"`
}
