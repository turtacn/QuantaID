package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/services/application"
	"github.com/turtacn/QuantaID/internal/services/platform"
	"github.com/turtacn/QuantaID/pkg/types"
)

type MockApplicationService struct {
	mock.Mock
}

func (m *MockApplicationService) ListApplications(ctx context.Context) ([]*types.Application, *types.Error) {
	args := m.Called(ctx)
	if args.Get(1) == nil {
		return args.Get(0).([]*types.Application), nil
	}
	return args.Get(0).([]*types.Application), args.Get(1).(*types.Error)
}

func (m *MockApplicationService) CreateApplication(ctx context.Context, req application.CreateApplicationRequest) (*types.Application, *types.Error) {
	args := m.Called(ctx, req)
	if args.Get(1) == nil {
		return args.Get(0).(*types.Application), nil
	}
	return args.Get(0).(*types.Application), args.Get(1).(*types.Error)
}

func (m *MockApplicationService) GetApplicationByID(ctx context.Context, id string) (*types.Application, *types.Error) {
	args := m.Called(ctx, id)
	if args.Get(1) == nil {
		return args.Get(0).(*types.Application), nil
	}
	return args.Get(0).(*types.Application), args.Get(1).(*types.Error)
}

func TestDevCenterService_ListApps(t *testing.T) {
	mockAppSvc := new(MockApplicationService)
	devCenterSvc := platform.NewDevCenterService(mockAppSvc, nil, nil, nil)

	expectedApps := []*types.Application{
		{ID: "app-1", Name: "App 1", Protocol: "oidc", Status: "active"},
		{ID: "app-2", Name: "App 2", Protocol: "saml", Status: "inactive"},
	}

	mockAppSvc.On("ListApplications", mock.Anything).Return(expectedApps, nil)

	dtos, err := devCenterSvc.ListApps(context.Background())

	assert.NoError(t, err)
	assert.Len(t, dtos, 2)
	assert.Equal(t, "App 1", dtos[0].Name)
	assert.True(t, dtos[0].Enabled)
	assert.Equal(t, "App 2", dtos[1].Name)
	assert.False(t, dtos[1].Enabled)

	mockAppSvc.AssertExpectations(t)
}
