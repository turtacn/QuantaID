package audit

import (
	"context"
	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/services/webhook"
	"github.com/turtacn/QuantaID/pkg/types"
	"time"
)

// Service provides a standardized way to record audit events.
// It uses the audit pipeline to dispatch events to configured sinks.
type Service struct {
	pipeline          *audit.Pipeline
	webhookDispatcher *webhook.Dispatcher
}

// NewService creates a new audit service.
func NewService(p *audit.Pipeline, wd *webhook.Dispatcher) *Service {
	return &Service{pipeline: p, webhookDispatcher: wd}
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
	event := &audit.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "auth",
		Action:    "login_success",
		UserID:    userID,
		IP:        ip,
		Result:    "success",
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

	event := &audit.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "auth",
		Action:    "login_failed",
		UserID:    userID,
		IP:        ip,
		Result:    "fail",
		TraceID:   traceID,
		Details:   details,
	}
	s.pipeline.Emit(ctx, event)
	s.dispatchWebhook(ctx, "login.failed", event)
}

// RecordUserCreated records a user creation event.
func (s *Service) RecordUserCreated(ctx context.Context, user *types.User, ip, traceID string) {
	event := &audit.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "identity",
		Action:    "user.created",
		UserID:    user.ID,
		IP:        ip,
		Result:    "success",
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
	event := &audit.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "policy",
		Action:    "policy_evaluated",
		UserID:    userID,
		IP:        ip,
		Resource:  resource,
		Result:    audit.Result(result), // "success", "fail", "deny"
		TraceID:   traceID,
		Details:   details,
	}
	s.pipeline.Emit(ctx, event)
}

// RecordAdminAction records an action performed by an administrator.
func (s *Service) RecordAdminAction(ctx context.Context, userID, ip, resource, action, traceID string, details map[string]any) {
	event := &audit.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "admin",
		Action:    action,
		UserID:    userID,
		IP:        ip,
		Resource:  resource,
		Result:    "success", // Assume admin actions are successful unless otherwise specified
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

	event := &audit.AuditEvent{
		ID:        generateAuditID(),
		Timestamp: time.Now().UTC(),
		Category:  "risk",
		Action:    "high_risk_login",
		UserID:    userID,
		IP:        ip,
		Result:    "success", // The event is about flagging, not the login outcome itself
		TraceID:   traceID,
		Details:   details,
	}
	s.pipeline.Emit(ctx, event)
}
