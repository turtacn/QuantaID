package session

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// SchedulerConfig holds configuration for the EvaluationScheduler.
type SchedulerConfig struct {
	DefaultInterval  time.Duration // Default 5m
	HighRiskInterval time.Duration // Default 1m
	CriticalInterval time.Duration // Default 30s
	BatchSize        int           // Default 100
	WorkerCount      int           // Default 4
}

// EvaluationScheduler manages periodic session risk evaluation.
type EvaluationScheduler struct {
	evaluator   *SessionEvaluator
	actions     *SessionActions
	sessionRepo SessionRepository // Need specific methods for finding sessions
	riskStore   interface{}       // Placeholder if needed directly
	config      SchedulerConfig
	stopCh      chan struct{}
	wg          sync.WaitGroup
	running     atomic.Bool
	logger      *zap.Logger
}

// ExtendedSessionRepository interface adding search capabilities needed by scheduler
type ExtendedSessionRepository interface {
	SessionRepository
	FindByRiskLevel(ctx context.Context, level string, limit int) ([]*Session, error)
	FindActiveSessions(ctx context.Context, limit int) ([]*Session, error)
}

// NewEvaluationScheduler creates a new EvaluationScheduler.
func NewEvaluationScheduler(evaluator *SessionEvaluator, actions *SessionActions, sessionRepo ExtendedSessionRepository, config SchedulerConfig, logger *zap.Logger) *EvaluationScheduler {
	if config.DefaultInterval == 0 {
		config.DefaultInterval = 5 * time.Minute
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.WorkerCount == 0 {
		config.WorkerCount = 4
	}

	return &EvaluationScheduler{
		evaluator:   evaluator,
		actions:     actions,
		sessionRepo: sessionRepo,
		config:      config,
		logger:      logger,
	}
}

// Start begins the scheduling loop.
func (s *EvaluationScheduler) Start(ctx context.Context) error {
	if !s.running.CompareAndSwap(false, true) {
		return nil // Already running
	}
	s.stopCh = make(chan struct{})

	// Start workers?
	// For simplicity in this implementation, the scheduler loop handles dispatching or we just process sequentially in batch.
	// To strictly follow the "worker" pattern:
	// We would need a channel to feed sessions to workers.

	s.wg.Add(1)
	go s.scheduleLoop(ctx)

	s.logger.Info("Session evaluation scheduler started")
	return nil
}

// Stop stops the scheduler.
func (s *EvaluationScheduler) Stop(ctx context.Context) error {
	if !s.running.CompareAndSwap(true, false) {
		return nil
	}
	close(s.stopCh)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *EvaluationScheduler) scheduleLoop(ctx context.Context) {
	defer s.wg.Done()
	// Check more frequently than the smallest interval to ensure timely evaluation
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scheduleBatch(ctx)
		}
	}
}

func (s *EvaluationScheduler) scheduleBatch(ctx context.Context) {
	// Implementation note: In a real distributed system, we would need locking or partitioning
	// to avoid multiple instances processing the same sessions.

	sessions := s.getSessionsNeedingEvaluation(ctx)
	for _, session := range sessions {
		// In a real worker pool, send to channel. Here we process directly for simplicity or spawn goroutine.
		s.wg.Add(1)
		go func(sess *Session) {
			defer s.wg.Done()
			s.evaluateSession(ctx, sess)
		}(session)
	}
}

func (s *EvaluationScheduler) getSessionsNeedingEvaluation(ctx context.Context) []*Session {
	repo, ok := s.sessionRepo.(ExtendedSessionRepository)
	if !ok {
		return nil
	}

	var result []*Session

	// 1. Critical sessions
	critical, _ := repo.FindByRiskLevel(ctx, "critical", s.config.BatchSize/4)
	for _, sess := range critical {
		if time.Since(sess.LastEvaluatedAt) > s.config.CriticalInterval {
			result = append(result, sess)
		}
	}

	// 2. High risk sessions
	if len(result) < s.config.BatchSize {
		high, _ := repo.FindByRiskLevel(ctx, "high", s.config.BatchSize/4)
		for _, sess := range high {
			if time.Since(sess.LastEvaluatedAt) > s.config.HighRiskInterval {
				result = append(result, sess)
			}
		}
	}

	// 3. Normal sessions
	if len(result) < s.config.BatchSize {
		remaining := s.config.BatchSize - len(result)
		normal, _ := repo.FindActiveSessions(ctx, remaining)
		for _, sess := range normal {
			if time.Since(sess.LastEvaluatedAt) > s.config.DefaultInterval {
				result = append(result, sess)
			}
		}
	}

	return result
}

func (s *EvaluationScheduler) evaluateSession(ctx context.Context, session *Session) {
	result, err := s.evaluator.Evaluate(ctx, session)
	if err != nil {
		s.logger.Error("Evaluation failed", zap.String("sessionID", session.ID), zap.Error(err))
		return
	}

	if result.RecommendedAction != ActionNone {
		err = s.actions.Execute(ctx, session, result.RecommendedAction, result.Reason)
		if err != nil {
			s.logger.Error("Action execution failed", zap.Error(err))
		}
	}

	session.LastEvaluatedAt = time.Now()
	// Update session last evaluated time
	_ = s.sessionRepo.Update(ctx, session)
}

// TriggerImmediate triggers an immediate evaluation for a session.
func (s *EvaluationScheduler) TriggerImmediate(ctx context.Context, sessionID string) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.evaluateSession(ctx, session)
	}()
	return nil
}
