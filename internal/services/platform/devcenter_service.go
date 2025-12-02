package platform

import (
	"context"

	"github.com/turtacn/QuantaID/internal/services/application"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	"github.com/turtacn/QuantaID/pkg/auth/mfa"
	"github.com/turtacn/QuantaID/pkg/types"
)

type DevCenterService struct {
	appSvc      application.Service
	apiKeySvc   *APIKeyService
	policySvc   *authorization.Service
	mfaSvc      *mfa.Manager
}

func NewDevCenterService(
	appSvc application.Service,
	apiKeySvc *APIKeyService,
	policySvc *authorization.Service,
	mfaSvc *mfa.Manager,
) *DevCenterService {
	return &DevCenterService{
		appSvc:      appSvc,
		apiKeySvc:   apiKeySvc,
		policySvc:   policySvc,
		mfaSvc:      mfaSvc,
	}
}

// ListApps lists all applications.
func (s *DevCenterService) ListApps(ctx context.Context) ([]*types.DevCenterAppDTO, error) {
	apps, err := s.appSvc.ListApplications(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*types.DevCenterAppDTO, len(apps))
	for i, app := range apps {
		dtos[i] = &types.DevCenterAppDTO{
			ID:       app.ID,
			Name:     app.Name,
			Protocol: string(app.Protocol),
			Enabled:  app.Status == types.ApplicationStatusActive,
		}
	}

	return dtos, nil
}

// CreateApp creates a new application.
func (s *DevCenterService) CreateApp(ctx context.Context, req types.CreateAppRequest) (*types.DevCenterAppDTO, error) {
	// For now, we only support OIDC applications.
	protocolConfig := types.JSONB{
		"redirect_uris": []string{req.RedirectURI},
	}

	app, err := s.appSvc.CreateApplication(ctx, application.CreateApplicationRequest{
		Name:           req.Name,
		Protocol:       types.ProtocolType(req.Protocol),
		ProtocolConfig: protocolConfig,
	})
	if err != nil {
		return nil, err
	}

	return &types.DevCenterAppDTO{
		ID:          app.ID,
		Name:        app.Name,
		Protocol:    string(app.Protocol),
		RedirectURI: req.RedirectURI,
		Enabled:     app.Status == types.ApplicationStatusActive,
	}, nil
}

// ListConnectors lists all connectors.
func (s *DevCenterService) ListConnectors(ctx context.Context) ([]*types.DevCenterConnectorDTO, error) {
	// For now, we'll return a hardcoded list of connectors.
	return []*types.DevCenterConnectorDTO{
		{
			ID:      "ldap-1",
			Type:    "ldap",
			Name:    "LDAP Connector",
			Enabled: true,
		},
	}, nil
}

// EnableConnector enables a connector.
func (s *DevCenterService) EnableConnector(ctx context.Context, id string) error {
	// For now, we'll just return nil.
	return nil
}

// Diagnostics returns diagnostics information.
func (s *DevCenterService) Diagnostics(ctx context.Context) (*types.DiagnosticsDTO, error) {
	// For now, we'll return some mock data.
	return &types.DiagnosticsDTO{
		Version:   "1.0.0",
		BuildTime: "2021-01-01T00:00:00Z",
		GoVersion: "go1.17",
		ConfigInfo: map[string]string{
			"environment": "development",
		},
	}, nil
}
