package application

import (
	"context"
	"github.com/turtacn/QuantaID/internal/domain/application"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// ApplicationService provides application-level use cases for managing applications.
// It acts as a facade over the application domain repository.
type ApplicationService struct {
	appRepo application.Repository
	logger  utils.Logger
	crypto  *utils.CryptoManager
}

// NewApplicationService creates a new application service.
func NewApplicationService(appRepo application.Repository, logger utils.Logger, crypto *utils.CryptoManager) *ApplicationService {
	return &ApplicationService{
		appRepo: appRepo,
		logger:  logger,
		crypto:  crypto,
	}
}

// CreateApplicationRequest defines the DTO for a request to create a new application.
type CreateApplicationRequest struct {
	Name           string                `json:"name"`
	Description    string                `json:"description,omitempty"`
	Protocol       types.ProtocolType    `json:"protocol"`
	ProtocolConfig types.JSONB           `json:"protocolConfig"`
}

// GetApplicationByID retrieves an application by its ID.
func (s *ApplicationService) GetApplicationByID(ctx context.Context, id string) (*types.Application, *types.Error) {
	app, err := s.appRepo.GetApplicationByID(ctx, id)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}
	return app, nil
}

// CreateApplication handles the creation of a new application.
func (s *ApplicationService) CreateApplication(ctx context.Context, req CreateApplicationRequest) (*types.Application, *types.Error) {
	app := &types.Application{
		ID:             s.crypto.GenerateUUID(),
		Name:           req.Name,
		Description:    req.Description,
		Status:         types.ApplicationStatusActive,
		Protocol:       req.Protocol,
		ProtocolConfig: req.ProtocolConfig,
	}

	if err := s.appRepo.CreateApplication(ctx, app); err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}

	return app, nil
}

// ListApplications retrieves a list of all applications.
func (s *ApplicationService) ListApplications(ctx context.Context) ([]*types.Application, *types.Error) {
	apps, err := s.appRepo.ListApplications(ctx, types.PaginationQuery{})
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			return nil, appErr
		}
		return nil, types.ErrInternal.WithCause(err)
	}
	return apps, nil
}