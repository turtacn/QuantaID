package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/identity/governance"
	"github.com/turtacn/QuantaID/internal/domain/identity/lifecycle"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// LifecycleJobConfig holds configuration for the lifecycle job.
type LifecycleJobConfig struct {
	Enabled             bool                          `yaml:"enabled"`
	Interval            time.Duration                 `yaml:"interval"`
	BatchSize           int                           `yaml:"batch_size"`
	LifecycleRules      []lifecycle.LifecycleRule     `yaml:"lifecycle_rules"`
	GovernanceConfig    governance.DataGovernanceConfig `yaml:"governance"`
	DryRun              bool                          `yaml:"dry_run"`
}

// LifecycleJob manages identity lifecycle and data governance.
type LifecycleJob struct {
	config          LifecycleJobConfig
	identityService identity.IService
	logger          *zap.Logger
	lifecycleEngine *lifecycle.Engine
	inspector       *governance.Inspector
}

// NewLifecycleJob creates a new lifecycle job worker.
func NewLifecycleJob(
	config LifecycleJobConfig,
	identityService identity.IService,
	logger *zap.Logger,
) *LifecycleJob {
	return &LifecycleJob{
		config:          config,
		identityService: identityService,
		logger:          logger.With(zap.String("component", "lifecycle_worker")),
		lifecycleEngine: lifecycle.NewEngine(),
		inspector:       governance.NewInspector(config.GovernanceConfig),
	}
}

// Start starts the worker loop.
func (w *LifecycleJob) Start(ctx context.Context) {
	if !w.config.Enabled {
		w.logger.Info("Lifecycle job is disabled")
		return
	}

	w.logger.Info("Starting lifecycle job", zap.Duration("interval", w.config.Interval))

	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	// Run once immediately
	w.Run(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping lifecycle job")
			return
		case <-ticker.C:
			w.Run(ctx)
		}
	}
}

// Run executes the lifecycle and governance checks.
func (w *LifecycleJob) Run(ctx context.Context) {
	w.logger.Info("Running lifecycle and governance scan")

	page := 1
	pageSize := w.config.BatchSize
	if pageSize <= 0 {
		pageSize = 100
	}

	for {
		users, total, err := w.identityService.ListUsers(ctx, types.UserFilter{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			w.logger.Error("Failed to list users", zap.Error(err))
			return
		}

		if len(users) == 0 {
			break
		}

		w.processBatch(ctx, users)

		if page*pageSize >= total {
			break
		}
		page++
	}

	w.logger.Info("Lifecycle scan completed")
}

func (w *LifecycleJob) processBatch(ctx context.Context, users []*types.User) {
	for _, user := range users {
		w.processUser(ctx, user)
	}
}

func (w *LifecycleJob) processUser(ctx context.Context, user *types.User) {
	// 1. Lifecycle Rules
	actions, err := w.lifecycleEngine.Evaluate(user, w.config.LifecycleRules)
	if err != nil {
		w.logger.Error("Failed to evaluate lifecycle rules", zap.String("userID", user.ID), zap.Error(err))
	} else if len(actions) > 0 {
		w.logger.Info("Lifecycle rules matched", zap.String("userID", user.ID), zap.Any("actions", actions))
		if !w.config.DryRun {
			if err := w.executeActions(ctx, user, actions); err != nil {
				w.logger.Error("Failed to execute lifecycle actions", zap.String("userID", user.ID), zap.Error(err))
			}
		}
	}

	// 2. Data Governance
	issues := w.inspector.Check(user)
	if len(issues) > 0 {
		w.logger.Info("Data quality issues detected", zap.String("userID", user.ID), zap.Any("issues", issues))
		// TODO: Could report these to a dashboard or metrics
	}
}

func (w *LifecycleJob) executeActions(ctx context.Context, user *types.User, actions []lifecycle.Action) error {
	for _, action := range actions {
		switch action.Type {
		case lifecycle.ActionDisable:
			if user.Status != types.UserStatusInactive {
				if err := w.identityService.ChangeUserStatus(ctx, user.ID, types.UserStatusInactive); err != nil {
					return fmt.Errorf("failed to disable user: %w", err)
				}
				w.logger.Info("User disabled by lifecycle rule", zap.String("userID", user.ID))
			}
		case lifecycle.ActionDelete:
			// Soft delete or hard delete? Assuming generic delete.
			if err := w.identityService.DeleteUser(ctx, user.ID); err != nil {
				return fmt.Errorf("failed to delete user: %w", err)
			}
			w.logger.Info("User deleted by lifecycle rule", zap.String("userID", user.ID))
		case lifecycle.ActionNotify:
			// Placeholder for notification logic
			w.logger.Info("Would notify user", zap.String("userID", user.ID), zap.Any("params", action.Params))
		case lifecycle.ActionArchive:
			// Placeholder for archive logic
			w.logger.Info("Would archive user", zap.String("userID", user.ID))
		default:
			w.logger.Warn("Unknown lifecycle action", zap.String("action", string(action.Type)))
		}
	}
	return nil
}
