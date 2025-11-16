package ldap

import (
	"context"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"time"
)

type SyncScheduler struct {
	engine *SyncEngine
	cron   *cron.Cron
	config SchedulerConfig
	logger *zap.Logger
}

type SchedulerConfig struct {
	FullSyncSchedule string
	EnableAutoRetry  bool
	MaxRetries       int
	RetryBackoff     time.Duration
}

func NewSyncScheduler(engine *SyncEngine, config SchedulerConfig, logger *zap.Logger) *SyncScheduler {
	return &SyncScheduler{
		engine: engine,
		cron:   cron.New(),
		config: config,
		logger: logger,
	}
}

func (s *SyncScheduler) Start(ctx context.Context, sourceID, baseDN, filter string) {
	s.cron.AddFunc(s.config.FullSyncSchedule, func() {
		s.logger.Info("triggering scheduled full sync", zap.String("sourceID", sourceID))
		if err := s.engine.StartFullSync(ctx, sourceID, baseDN, filter); err != nil {
			s.logger.Error("full sync failed", zap.Error(err))
			if s.config.EnableAutoRetry {
				s.scheduleRetry(ctx, "full_sync", sourceID, baseDN, filter)
			}
		}
	})

	if s.engine.config.IncrementalEnable {
		go func() {
			for {
				s.logger.Info("starting incremental sync listener", zap.String("sourceID", sourceID))
				err := s.engine.StartIncrementalSync(ctx, sourceID, baseDN, filter)
				if err != nil {
					s.logger.Error("incremental sync listener failed, restarting after backoff", zap.Error(err))
					time.Sleep(s.config.RetryBackoff)
				}
				if ctx.Err() != nil {
					s.logger.Info("context cancelled, stopping incremental sync listener")
					return
				}
			}
		}()
	}

	s.cron.Start()
	s.logger.Info("sync scheduler started")
}

func (s *SyncScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("sync scheduler stopped")
}

func (s *SyncScheduler) scheduleRetry(ctx context.Context, taskType, sourceID, baseDN, filter string) {
	backoff := s.config.RetryBackoff
	for attempt := 1; attempt <= s.config.MaxRetries; attempt++ {
		s.logger.Info("scheduling retry", zap.Int("attempt", attempt), zap.Duration("backoff", backoff))
		time.Sleep(backoff)

		var err error
		if taskType == "full_sync" {
			err = s.engine.StartFullSync(ctx, sourceID, baseDN, filter)
		}

		if err == nil {
			s.logger.Info("retry successful", zap.String("task", taskType), zap.Int("attempt", attempt))
			return
		}

		s.logger.Warn("retry failed", zap.String("task", taskType), zap.Int("attempt", attempt), zap.Error(err))
		backoff *= 2
	}
	s.logger.Error("all retries failed for task", zap.String("task", taskType), zap.Int("max_retries", s.config.MaxRetries))
}
