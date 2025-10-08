package audit

import (
	"context"
	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"time"
)

// ApplicationService provides application-level use cases for auditing. It acts as an
// intermediary between the transport layer (e.g., HTTP handlers) and the domain
// layer (repositories), encapsulating the logic for recording and retrieving audit events.
type ApplicationService struct {
	auditRepo auth.AuditLogRepository
	logger    utils.Logger
}

// NewApplicationService creates a new audit application service.
//
// Parameters:
//   - auditRepo: The repository for persisting audit log entries.
//   - logger: The logger for service-level messages.
//
// Returns:
//   A new instance of ApplicationService.
func NewApplicationService(auditRepo auth.AuditLogRepository, logger utils.Logger) *ApplicationService {
	return &ApplicationService{
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// RecordEvent asynchronously records a new audit event. It constructs an AuditLog
// entry and saves it to the repository in a separate goroutine to avoid blocking the caller.
//
// Parameters:
//   - ctx: The context of the request that triggered the event.
//   - actorID: The ID of the user or system that performed the action.
//   - action: A string describing the action (e.g., "user.login").
//   - resource: The resource that was affected (e.g., "user:123").
//   - status: The outcome of the action ("success" or "failure").
//   - eventContext: Additional contextual data about the event.
func (s *ApplicationService) RecordEvent(ctx context.Context, actorID, action, resource, status string, eventContext map[string]interface{}) {
	logEntry := &types.AuditLog{
		ID:        uuid.New().String(),
		ActorID:   actorID,
		Action:    action,
		Resource:  resource,
		Status:    status,
		Context:   eventContext,
		Timestamp: time.Now().UTC(),
	}

	go func() {
		err := s.auditRepo.CreateLogEntry(context.Background(), logEntry)
		if err != nil {
			s.logger.Error(context.Background(), "Failed to record audit event", zap.Error(err), zap.String("actorID", actorID), zap.String("action", action))
		}
	}()
}

// GetUserHistory retrieves a paginated list of audit events for a specific user.
// It handles input validation for pagination parameters.
//
// Parameters:
//   - ctx: The context for the request.
//   - userID: The ID of the user whose history is being requested.
//   - page: The page number to retrieve.
//   - pageSize: The number of items per page.
//
// Returns:
//   A slice of audit log entries and an application error if one occurs.
func (s *ApplicationService) GetUserHistory(ctx context.Context, userID string, page, pageSize int) ([]*types.AuditLog, *types.Error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	pagination := types.PaginationQuery{
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}

	logs, err := s.auditRepo.GetLogsForUser(ctx, userID, pagination)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}

	return logs, nil
}

