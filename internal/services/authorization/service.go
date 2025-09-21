package authorization

import (
	"context"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// ApplicationService provides application-level use cases for authorization.
type ApplicationService struct {
	policyDomain   *policy.Service
	identityDomain identity.IService
	logger         utils.Logger
}

// NewApplicationService creates a new authorization application service.
func NewApplicationService(policyDomain *policy.Service, identityDomain identity.IService, logger utils.Logger) *ApplicationService {
	return &ApplicationService{
		policyDomain:   policyDomain,
		identityDomain: identityDomain,
		logger:         logger,
	}
}

// CheckPermissionRequest defines the DTO for an authorization check.
type CheckPermissionRequest struct {
	UserID     string                 `json:"userId"`
	Action     string                 `json:"action"`
	ResourceID string                 `json:"resourceId"`
	Context    map[string]interface{} `json:"context"`
}

// CheckPermission is the main method for checking if a user has permission to perform an action.
func (s *ApplicationService) CheckPermission(ctx context.Context, req CheckPermissionRequest) (bool, *types.Error) {
	user, err := s.identityDomain.GetUser(ctx, req.UserID)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return false, appErr
		}
		return false, types.ErrInternal.WithCause(err)
	}

	userGroups, err := s.identityDomain.GetUserGroups(ctx, req.UserID)
	if err != nil {
		s.logger.Warn(ctx, "Could not fetch user groups for authz check", zap.Error(err), zap.String("userID", req.UserID))
	}

	groupIDs := make([]string, len(userGroups))
	for i, g := range userGroups {
		groupIDs[i] = g.ID
	}

	evalCtx := &types.PolicyEvaluationContext{
		Subject: map[string]interface{}{
			"id":     user.ID,
			"groups": groupIDs,
		},
		Action: req.Action,
		Resource: map[string]interface{}{
			"id": req.ResourceID,
		},
		Context: req.Context,
	}

	decision, domainErr := s.policyDomain.Evaluate(ctx, evalCtx)
	if domainErr != nil {
		if appErr, ok := domainErr.(*types.Error); ok {
			return false, appErr
		}
		return false, types.ErrInternal.WithCause(domainErr)
	}

	if !decision.Allowed {
		s.logger.Info(ctx, "Authorization denied", zap.String("userID", req.UserID), zap.String("action", req.Action), zap.String("resource", req.ResourceID), zap.String("reason", decision.Reason))
		return false, types.ErrForbidden
	}

	s.logger.Info(ctx, "Authorization granted", zap.String("userID", req.UserID), zap.String("action", req.Action), zap.String("resource", req.ResourceID))
	return true, nil
}

//Personal.AI order the ending
