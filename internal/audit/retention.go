package audit

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// RetentionManager handles the automatic deletion of old audit logs.
type RetentionManager struct {
	repo   AuditRepository
	logger *zap.Logger
	cron   *cron.Cron
}

// NewRetentionManager creates a new RetentionManager.
func NewRetentionManager(repo AuditRepository, logger *zap.Logger) *RetentionManager {
	return &RetentionManager{
		repo:   repo,
		logger: logger.Named("retention-manager"),
		cron:   cron.New(),
	}
}

// Start starts the retention manager's cron job.
func (rm *RetentionManager) Start(ctx context.Context, retentionDays int) {
	if retentionDays <= 0 {
		rm.logger.Info("Audit log retention is disabled.")
		return
	}

	schedule := "0 1 * * *" // Run once a day at 1 AM

	rm.cron.AddFunc(schedule, func() {
		rm.logger.Info("Running audit log retention job...")
		cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
		if err := rm.repo.DeleteBefore(ctx, cutoff); err != nil {
			rm.logger.Error("Failed to delete old audit logs", zap.Error(err))
		}
	})

	rm.cron.Start()
	rm.logger.Info("Audit log retention job scheduled.", zap.Int("retention_days", retentionDays))
}

// Stop stops the retention manager's cron job.
func (rm *RetentionManager) Stop() {
	rm.cron.Stop()
	rm.logger.Info("Audit log retention job stopped.")
}
