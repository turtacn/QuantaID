//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/auth/session"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// Mock Repo for integration test
type InMemorySessionRepo struct {
	sessions map[string]*session.Session
}

func NewInMemoryRepo() *InMemorySessionRepo {
	return &InMemorySessionRepo{sessions: make(map[string]*session.Session)}
}
func (r *InMemorySessionRepo) GetByID(ctx context.Context, id string) (*session.Session, error) {
	if s, ok := r.sessions[id]; ok {
		return s, nil
	}
	return nil, types.ErrNotFound
}
func (r *InMemorySessionRepo) Update(ctx context.Context, s *session.Session) error {
	r.sessions[s.ID] = s
	return nil
}
func (r *InMemorySessionRepo) Delete(ctx context.Context, id string) error {
	delete(r.sessions, id)
	return nil
}
func (r *InMemorySessionRepo) DeleteFromCache(ctx context.Context, id string) error {
	return nil
}
func (r *InMemorySessionRepo) FindByRiskLevel(ctx context.Context, level string, limit int) ([]*session.Session, error) {
	var result []*session.Session
	for _, s := range r.sessions {
		if s.RiskLevel == level {
			result = append(result, s)
		}
	}
	return result, nil
}
func (r *InMemorySessionRepo) FindActiveSessions(ctx context.Context, limit int) ([]*session.Session, error) {
	var result []*session.Session
	for _, s := range r.sessions {
		if s.Status == session.SessionStatusActive {
			result = append(result, s)
		}
	}
	return result, nil
}

// Implement Create for EventRepo
type MockEventRepo struct{}
func (r *MockEventRepo) Create(ctx context.Context, event interface{}) error { return nil }

func TestSession_ContinuousEvaluation_Periodic(t *testing.T) {
	// Setup
	repo := NewInMemoryRepo()
	logger := zap.NewNop()

	monitorConfig := session.MonitorConfig{
		SuspiciousIPThreshold: 30,
	}
	monitor := session.NewRiskMonitor(nil, nil, nil, nil, monitorConfig)
	policy := session.NewSessionPolicy(session.DefaultPolicyRules())
	evalConfig := session.EvaluationConfig{
		LowRiskThreshold: 25,
		MediumRiskThreshold: 50,
		HighRiskThreshold: 75,
	}
	evaluator := session.NewSessionEvaluator(nil, monitor, policy, nil, evalConfig)
	actions := session.NewSessionActions(repo, nil, nil)

	schedulerConfig := session.SchedulerConfig{
		DefaultInterval: 100 * time.Millisecond,
		BatchSize: 10,
	}

	scheduler := session.NewEvaluationScheduler(evaluator, actions, repo, schedulerConfig, logger)

	// Create Session
	sess := &session.Session{
		ID: "sess_integration_1",
		Status: session.SessionStatusActive,
		RiskLevel: "low",
		LastEvaluatedAt: time.Now().Add(-1 * time.Hour), // Old evaluation
		LastActivityAt: time.Now(),
		Permissions: []string{"read"},
	}
	repo.Update(context.Background(), sess)

	// Start Scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for evaluation cycle
	time.Sleep(300 * time.Millisecond)

	// Check if LastEvaluatedAt updated
	updatedSess, _ := repo.GetByID(ctx, "sess_integration_1")
	assert.True(t, updatedSess.LastEvaluatedAt.After(sess.LastEvaluatedAt))

	scheduler.Stop(ctx)
}
