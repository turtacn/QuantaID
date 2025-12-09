package session

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// SessionStatus defines the status of a session.
type SessionStatus string

const (
	SessionStatusActive         SessionStatus = "active"
	SessionStatusDowngraded     SessionStatus = "downgraded"
	SessionStatusStepUpRequired SessionStatus = "step_up_required"
	SessionStatusTerminated     SessionStatus = "terminated"
)

// Session represents the rich domain model for a user session.
type Session struct {
	ID                string           `json:"id"`
	UserID            string           `json:"user_id"`
	TenantID          string           `json:"tenant_id"`
	DeviceID          string           `json:"device_id"`
	Status            SessionStatus    `json:"status"`
	AuthLevel         int              `json:"auth_level"`
	RiskLevel         string           `json:"risk_level"`
	CurrentRiskScore  int              `json:"current_risk_score"`
	Permissions       []string         `json:"permissions"`
	CurrentIP         string           `json:"current_ip"`
	PreviousIP        string           `json:"previous_ip"`
	LastIPChangeAt    time.Time        `json:"last_ip_change_at"`
	PreviousDeviceID  string           `json:"previous_device_id"`
	CreatedAt         time.Time        `json:"created_at"`
	ExpiresAt         time.Time        `json:"expires_at"`
	LastActivityAt    time.Time        `json:"last_activity_at"`
	LastEvaluatedAt   time.Time        `json:"last_evaluated_at"`
	LastStepUpAt      time.Time        `json:"last_step_up_at"`
	StepUpReason      string           `json:"step_up_reason"`
	StepUpRequiredAt  time.Time        `json:"step_up_required_at"`
	DowngradeReason   string           `json:"downgrade_reason"`
	DowngradedAt      time.Time        `json:"downgraded_at"`
	TerminateReason   string           `json:"terminate_reason"`
	TerminatedAt      time.Time        `json:"terminated_at"`
	UserAgent         string           `json:"user_agent"`
	DeviceFingerprint string           `json:"device_fingerprint"`
}

// GetState converts Session to models.SessionState.
func (s *Session) GetState() models.SessionState {
	return models.SessionState{
		Status:          string(s.Status),
		RiskLevel:       s.RiskLevel,
		Permissions:     s.Permissions,
		AuthLevel:       s.AuthLevel,
		LastEvaluatedAt: s.LastEvaluatedAt,
	}
}

// IsExpired checks if the session is expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SessionManager is the entry point for session management with continuous evaluation.
type SessionManager struct {
	// Wrapped Redis SessionManager for low-level storage operations if needed
	// or we access the repository directly.
	// For this implementation, we assume we use a repository pattern which might wrap Redis.

	repo            ExtendedSessionRepository
	riskStore       *redis.SessionRiskStore
	evaluator       *SessionEvaluator
	actions         *SessionActions
	scheduler       *EvaluationScheduler
	eventRepo       SessionEventRepository
	logger          *zap.Logger
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager(
	repo ExtendedSessionRepository,
	riskStore *redis.SessionRiskStore,
	evaluator *SessionEvaluator,
	actions *SessionActions,
	scheduler *EvaluationScheduler,
	eventRepo SessionEventRepository,
	logger *zap.Logger,
) *SessionManager {
	return &SessionManager{
		repo:      repo,
		riskStore: riskStore,
		evaluator: evaluator,
		actions:   actions,
		scheduler: scheduler,
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// CreateSession creates a new session with initial risk evaluation.
func (m *SessionManager) CreateSession(ctx context.Context, user *types.User, deviceID string, authLevel int, ip, userAgent string) (*Session, error) {
	sessionID := uuid.New().String()

	session := &Session{
		ID:             sessionID,
		UserID:         user.ID,
		// TenantID:       user.TenantID, // Assuming User has TenantID, if not, empty
		DeviceID:       deviceID,
		Status:         SessionStatusActive,
		AuthLevel:      authLevel,
		CurrentIP:      ip,
		UserAgent:      userAgent,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour), // Default TTL
	}

	// Initial Evaluation
	result, err := m.evaluator.Evaluate(ctx, session)
	if err == nil {
		session.CurrentRiskScore = result.RiskScore
		session.RiskLevel = result.RiskLevel
		session.LastEvaluatedAt = time.Now()
	} else {
		m.logger.Error("Initial session evaluation failed", zap.Error(err))
		session.RiskLevel = "unknown"
	}

	// Persist
	if err := m.repo.Update(ctx, session); err != nil { // Assuming Update can also Create/Upsert or we add Create method
		return nil, err
	}

	// Record Event
	m.recordSessionEvent(ctx, session, models.SessionEventCreated, "Session Created")

	return session, nil
}

// ValidateSession retrieves and validates a session.
func (m *SessionManager) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	session, err := m.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.Status == SessionStatusTerminated {
		return nil, errors.New("session terminated")
	}
	if session.IsExpired() {
		return nil, errors.New("session expired")
	}

	// Quick Risk Check
	safe, err := m.evaluator.QuickCheck(ctx, session)
	if err != nil {
		m.logger.Warn("Quick risk check failed", zap.Error(err))
	} else if !safe {
		return nil, errors.New("session risk too high")
	}

	if session.Status == SessionStatusStepUpRequired {
		return nil, errors.New("step-up authentication required")
	}

	// Update Activity
	session.LastActivityAt = time.Now()
	// Optimistic update (maybe async or sampled)
	_ = m.repo.Update(ctx, session)

	return session, nil
}

// TerminateSession terminates a session.
func (m *SessionManager) TerminateSession(ctx context.Context, sessionID, reason string) error {
	session, err := m.repo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	return m.actions.Execute(ctx, session, ActionTerminate, reason)
}

func (m *SessionManager) recordSessionEvent(ctx context.Context, session *Session, eventType models.SessionEventType, reason string) {
	if m.eventRepo == nil {
		return
	}
	event := &models.SessionEvent{
		ID:            uuid.New().String(),
		SessionID:     session.ID,
		TenantID:      session.TenantID,
		UserID:        session.UserID,
		EventType:     eventType,
		NewState:      session.GetState(),
		TriggerReason: reason,
		RiskScore:     session.CurrentRiskScore,
		IPAddress:     session.CurrentIP,
		DeviceID:      session.DeviceID,
		CreatedAt:     time.Now(),
	}
	_ = m.eventRepo.Create(ctx, event)
}

// StartScheduler starts the evaluation scheduler.
func (m *SessionManager) StartScheduler(ctx context.Context) {
	if m.scheduler != nil {
		_ = m.scheduler.Start(ctx)
	}
}

// StopScheduler stops the evaluation scheduler.
func (m *SessionManager) StopScheduler(ctx context.Context) {
	if m.scheduler != nil {
		_ = m.scheduler.Stop(ctx)
	}
}
