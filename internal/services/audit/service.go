package audit

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/services/webhook"
	"github.com/turtacn/QuantaID/pkg/audit/events"
	"github.com/turtacn/QuantaID/pkg/types"
	"time"
)

// Service provides a standardized way to record audit events.
// It uses the audit pipeline to dispatch events to configured sinks.
type Service struct {
	pipeline          *audit.Pipeline
	webhookDispatcher *webhook.Dispatcher
	repo              audit.AuditRepository // Add repo access
}

// NewService creates a new audit service.
func NewService(p *audit.Pipeline, wd *webhook.Dispatcher) *Service {
	return &Service{pipeline: p, webhookDispatcher: wd}
}

// WithRepository attaches a repository to the service for querying.
func (s *Service) WithRepository(repo audit.AuditRepository) *Service {
	s.repo = repo
	return s
}

func (s *Service) dispatchWebhook(ctx context.Context, eventType string, payload interface{}) {
	if s.webhookDispatcher != nil {
		// Use a detached context to ensure webhook delivery isn't cancelled by the request context
		go s.webhookDispatcher.Dispatch(context.Background(), eventType, payload)
	}
}

// generateAuditID creates a new unique ID for an audit event.
func generateAuditID() string {
	return uuid.New().String()
}

// RecordLoginSuccess records a successful user login event.
func (s *Service) RecordLoginSuccess(ctx context.Context, userID, ip, traceID string, details map[string]any) {
	event := &events.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "auth",
		Action:    "login_success",
		UserID:    userID,
		IP:        ip,
		Result:    events.ResultSuccess,
		TraceID:   traceID,
		Details:   details,
	}
	s.pipeline.Emit(ctx, event)
	s.dispatchWebhook(ctx, "login.success", event)
}

// RecordLoginFailed records a failed user login attempt.
func (s *Service) RecordLoginFailed(ctx context.Context, userID, ip, traceID string, reason string, details map[string]any) {
	if details == nil {
		details = make(map[string]any)
	}
	details["reason"] = reason

	event := &events.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "auth",
		Action:    "login_failed",
		UserID:    userID,
		IP:        ip,
		Result:    events.ResultFailure,
		TraceID:   traceID,
		Details:   details,
	}
	s.pipeline.Emit(ctx, event)
	s.dispatchWebhook(ctx, "login.failed", event)
}

// RecordUserCreated records a user creation event.
func (s *Service) RecordUserCreated(ctx context.Context, user *types.User, ip, traceID string) {
	event := &events.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "identity",
		Action:    "user.created",
		UserID:    user.ID,
		IP:        ip,
		Result:    events.ResultSuccess,
		TraceID:   traceID,
		Details: map[string]any{
			"username": user.Username,
			"email":    string(user.Email),
		},
	}
	s.pipeline.Emit(ctx, event)
	s.dispatchWebhook(ctx, "user.created", event)
}

// RecordPolicyDecision records the outcome of a policy evaluation.
func (s *Service) RecordPolicyDecision(ctx context.Context, userID, ip, resource, traceID string, result string, details map[string]any) {
	event := &events.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "policy",
		Action:    "policy_evaluated",
		UserID:    userID,
		IP:        ip,
		Resource:  resource,
		Result:    events.Result(result), // "success", "fail", "deny"
		TraceID:   traceID,
		Details:   details,
	}
	s.pipeline.Emit(ctx, event)
}

// GetLogsForUser retrieves audit logs for a specific user.
func (s *Service) GetLogsForUser(ctx context.Context, userID string, limit int) ([]*events.AuditEvent, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("audit repository not configured")
	}

	// Use QueryFilter to filter by UserID
	// Assuming limit needs to be handled by repository or we slice the result
	// The interface doesn't have Limit in QueryFilter, checking internal/audit/repository.go
	// It has PaginationQuery in MockAuditLogRepository (but that's a mock).
	// Real interface has `Query(ctx, filter QueryFilter)`.
	// QueryFilter has Start/EndTimestamp, EventTypes, ActorID, TargetID.
	// It doesn't seem to have Pagination.
	// Wait, MockAuditLogRepository used `types.PaginationQuery`.
	// Let's check `internal/audit/repository.go` again.
	// It says `Query(ctx context.Context, filter QueryFilter)`.
	// And `QueryFilter` struct definition.

	// I will just use Query for now and assume it returns recent logs or I filter manually?
	// The interface might need update for limit.
	// For now, let's just query by ActorID (UserID).

	filter := audit.QueryFilter{
		ActorID: userID,
		// Should probably set a time range, e.g., last 30 days
		StartTimestamp: time.Now().AddDate(0, 0, -30),
		EndTimestamp:   time.Now(),
	}

	logs, err := s.repo.Query(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Apply limit if needed (in memory for now if repo doesn't support it)
	if limit > 0 && len(logs) > limit {
		return logs[:limit], nil
	}

	return logs, nil
}

// RecordAdminAction records an action performed by an administrator.
func (s *Service) RecordAdminAction(ctx context.Context, userID, ip, resource, action, traceID string, details map[string]any) {
	event := &events.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "admin",
		Action:    action,
		UserID:    userID,
		IP:        ip,
		Resource:  resource,
		Result:    events.ResultSuccess, // Assume admin actions are successful unless otherwise specified
		TraceID:   traceID,
		Details:   details,
	}
	s.pipeline.Emit(ctx, event)
}

// RecordHighRiskLogin records a login attempt that was flagged as high-risk.
func (s *Service) RecordHighRiskLogin(ctx context.Context, userID, ip, traceID string, score float64, factors []string, details map[string]any) {
	if details == nil {
		details = make(map[string]any)
	}
	details["risk_score"] = score
	details["risk_factors"] = factors

	event := &events.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "risk",
		Action:    "high_risk_login",
		UserID:    userID,
		IP:        ip,
		Result:    events.ResultSuccess, // The event is about flagging, not the login outcome itself
		TraceID:   traceID,
		Details:   details,
	}
	s.pipeline.Emit(ctx, event)
}
