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

// ApplicationService provides application-level use cases for auditing.
type ApplicationService struct {
	auditRepo auth.AuditLogRepository
	logger    utils.Logger
}

// NewApplicationService creates a new audit application service.
func NewApplicationService(auditRepo auth.AuditLogRepository, logger utils.Logger) *ApplicationService {
	return &ApplicationService{
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// RecordEvent records a new audit event.
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

// GetUserHistory retrieves the audit history for a specific user.
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

//Personal.AI order the ending
