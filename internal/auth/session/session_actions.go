package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
	"gorm.io/gorm"
)

// ActionType defines the type of action to take on a session.
type ActionType string

const (
	ActionNone          ActionType = "none"
	ActionMonitor       ActionType = "monitor"
	ActionRequireMFA    ActionType = "require_mfa"
	ActionDowngrade     ActionType = "downgrade"
	ActionTerminate     ActionType = "terminate"
)

// SessionRepository defines the interface for session storage.
// This allows decoupling from specific implementations.
type SessionRepository interface {
	GetByID(ctx context.Context, id string) (*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
	DeleteFromCache(ctx context.Context, id string) error
}

// SessionEventRepository defines the interface for storing session events.
type SessionEventRepository interface {
	Create(ctx context.Context, event *models.SessionEvent) error
}

// NotificationService defines the interface for sending notifications.
type NotificationService interface {
	SendStepUpRequired(userID, reason string) error
	SendSessionDowngraded(userID, reason string) error
	SendSessionTerminated(userID, reason string) error
}

// SessionActions executes actions on sessions.
type SessionActions struct {
	sessionRepo   SessionRepository
	eventRepo     SessionEventRepository
	notifyService NotificationService
}

// NewSessionActions creates a new SessionActions executor.
func NewSessionActions(sessionRepo SessionRepository, eventRepo SessionEventRepository, notifyService NotificationService) *SessionActions {
	return &SessionActions{
		sessionRepo:   sessionRepo,
		eventRepo:     eventRepo,
		notifyService: notifyService,
	}
}

// Execute performs the requested action on the session.
func (a *SessionActions) Execute(ctx context.Context, session *Session, action ActionType, reason string) error {
	previousState := session.GetState()
	var err error

	switch action {
	case ActionNone:
		return nil
	case ActionMonitor:
		err = a.monitor(ctx, session)
	case ActionRequireMFA:
		err = a.requireMFA(ctx, session, reason)
	case ActionDowngrade:
		err = a.downgrade(ctx, session, reason)
	case ActionTerminate:
		err = a.terminate(ctx, session, reason)
	}

	if err != nil {
		return err
	}

	// Record event
	return a.recordEvent(ctx, session, action, previousState, reason)
}

func (a *SessionActions) monitor(ctx context.Context, session *Session) error {
	// For now, just a placeholder for enhanced logging or similar
	return nil
}

func (a *SessionActions) requireMFA(ctx context.Context, session *Session, reason string) error {
	session.Status = SessionStatusStepUpRequired
	session.StepUpReason = reason
	session.StepUpRequiredAt = time.Now()

	if err := a.sessionRepo.Update(ctx, session); err != nil {
		return err
	}

	if a.notifyService != nil {
		_ = a.notifyService.SendStepUpRequired(session.UserID, reason)
	}
	return nil
}

func (a *SessionActions) downgrade(ctx context.Context, session *Session, reason string) error {
	session.Status = SessionStatusDowngraded
	session.DowngradeReason = reason
	session.DowngradedAt = time.Now()

	// Remove sensitive permissions logic would go here
	// session.Permissions = filterSensitivePermissions(session.Permissions)

	if session.AuthLevel > 1 {
		session.AuthLevel = session.AuthLevel - 1
	} else {
		session.AuthLevel = 1
	}

	if err := a.sessionRepo.Update(ctx, session); err != nil {
		return err
	}

	if a.notifyService != nil {
		_ = a.notifyService.SendSessionDowngraded(session.UserID, reason)
	}
	return nil
}

func (a *SessionActions) terminate(ctx context.Context, session *Session, reason string) error {
	session.Status = SessionStatusTerminated
	session.TerminatedAt = time.Now()
	session.TerminateReason = reason

	if err := a.sessionRepo.Update(ctx, session); err != nil {
		return err
	}

	// Ensure it is removed from cache/active store
	_ = a.sessionRepo.DeleteFromCache(ctx, session.ID)

	if a.notifyService != nil {
		_ = a.notifyService.SendSessionTerminated(session.UserID, reason)
	}
	return nil
}

// Upgrade upgrades the session after successful verification (e.g. MFA).
func (a *SessionActions) Upgrade(ctx context.Context, session *Session) error {
	previousState := session.GetState()

	session.Status = SessionStatusActive
	session.StepUpReason = ""
	session.AuthLevel = min(session.AuthLevel+1, 4)
	session.LastStepUpAt = time.Now()

	// Restore permissions would happen here

	if err := a.sessionRepo.Update(ctx, session); err != nil {
		return err
	}

	return a.recordEvent(ctx, session, "upgraded", previousState, "MFA Verification Successful")
}

func (a *SessionActions) recordEvent(ctx context.Context, session *Session, action ActionType, prevState models.SessionState, reason string) error {
	if a.eventRepo == nil {
		return nil
	}

	eventType := models.SessionEventType(string(action))
	if action == "upgraded" {
		eventType = models.SessionEventUpgraded
	}

	event := &models.SessionEvent{
		ID:            uuid.New().String(),
		SessionID:     session.ID,
		TenantID:      session.TenantID,
		UserID:        session.UserID,
		EventType:     eventType,
		PreviousState: prevState,
		NewState:      session.GetState(),
		TriggerReason: reason,
		RiskScore:     session.CurrentRiskScore,
		IPAddress:     session.CurrentIP,
		DeviceID:      session.DeviceID,
		CreatedAt:     time.Now(),
	}

	return a.eventRepo.Create(ctx, event)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PostgresSessionEventRepository implements SessionEventRepository
type PostgresSessionEventRepository struct {
	db *gorm.DB
}

func NewPostgresSessionEventRepository(db *gorm.DB) *PostgresSessionEventRepository {
	return &PostgresSessionEventRepository{db: db}
}

func (r *PostgresSessionEventRepository) Create(ctx context.Context, event *models.SessionEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}
