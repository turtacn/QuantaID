package audit

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// RetentionPolicyManager enforces data retention policies on audit logs.
type RetentionPolicyManager struct {
	repo   RetentionRepository
	config RetentionConfig
	logger *zap.Logger
}

// RetentionRepository defines the storage interface needed for retention policies.
type RetentionRepository interface {
	// ArchiveLogs moves logs older than a given cutoff time to a different tier (e.g., 'warm', 'cold').
	ArchiveLogs(ctx context.Context, cutoff time.Time, targetTier string) (int64, error)
	// DeleteLogsBefore permanently deletes logs older than a given cutoff time.
	DeleteLogsBefore(ctx context.Context, cutoff time.Time) (int64, error)
}


// RetentionConfig defines the parameters for the data lifecycle.
type RetentionConfig struct {
	HotDataRetention  time.Duration `yaml:"hot_data_retention_days"`
	WarmDataRetention time.Duration `yaml:"warm_data_retention_days"`
	ColdDataRetention time.Duration `yaml:"cold_data_retention_days"`
	EnableAutoArchive bool          `yaml:"enable_auto_archive"`
	EnableAutoDelete  bool          `yaml:"enable_auto_delete"`
}

// NewRetentionPolicyManager creates a new manager for data retention.
func NewRetentionPolicyManager(repo RetentionRepository, config RetentionConfig, logger *zap.Logger) *RetentionPolicyManager {
	return &RetentionPolicyManager{
		repo:   repo,
		config: config,
		logger: logger.Named("retention-policy"),
	}
}

// Execute applies the configured retention policies to the audit log data.
func (rpm *RetentionPolicyManager) Execute(ctx context.Context) error {
	now := time.Now().UTC()
	rpm.logger.Info("Executing data retention policy run.")

	// 1. Archive hot data to warm storage
	if rpm.config.EnableAutoArchive {
		cutoffHot := now.Add(-rpm.config.HotDataRetention)
		archivedCount, err := rpm.repo.ArchiveLogs(ctx, cutoffHot, "warm")
		if err != nil {
			rpm.logger.Error("Failed to archive hot data to warm storage.", zap.Error(err))
			// Continue execution even if one step fails
		}
		if archivedCount > 0 {
			rpm.logger.Info("Successfully archived hot data.", zap.Int64("count", archivedCount), zap.String("tier", "warm"))
		}
	}

	// 2. Archive warm data to cold storage (if applicable, not implemented in this example)
	// cutoffWarm := now.Add(-rpm.config.WarmDataRetention)
	// rpm.repo.ArchiveLogs(ctx, cutoffWarm, "cold")


	// 3. Delete data that has exceeded the final retention period
	if rpm.config.EnableAutoDelete {
		cutoffCold := now.Add(-rpm.config.ColdDataRetention)
		deletedCount, err := rpm.repo.DeleteLogsBefore(ctx, cutoffCold)
		if err != nil {
			rpm.logger.Error("Failed to delete expired data.", zap.Error(err))
			return err // Deletion is a critical step, return error
		}
		if deletedCount > 0 {
			rpm.logger.Info("Successfully deleted expired data.", zap.Int64("count", deletedCount))
		}
	}

	rpm.logger.Info("Data retention policy run completed.")
	return nil
}
