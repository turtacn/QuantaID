package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/pkg/types"
)

type MockDevCenterService struct {
	mock.Mock
}

func (m *MockDevCenterService) ListApps(ctx context.Context) ([]*types.DevCenterAppDTO, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*types.DevCenterAppDTO), args.Error(1)
}

func (m *MockDevCenterService) CreateApp(ctx context.Context, req types.CreateAppRequest) (*types.DevCenterAppDTO, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*types.DevCenterAppDTO), args.Error(1)
}

func (m *MockDevCenterService) ListConnectors(ctx context.Context) ([]*types.DevCenterConnectorDTO, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*types.DevCenterConnectorDTO), args.Error(1)
}

func (m *MockDevCenterService) EnableConnector(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDevCenterService) Diagnostics(ctx context.Context) (*types.DiagnosticsDTO, error) {
	args := m.Called(ctx)
	return args.Get(0).(*types.DiagnosticsDTO), args.Error(1)
}

func TestDevCenterAPI_AdminCanManageApps(t *testing.T) {
	mockSvc := new(MockDevCenterService)
	handler := handlers.NewDevCenterHandler(mockSvc)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	expectedApps := []*types.DevCenterAppDTO{
		{ID: "app-1", Name: "App 1", Protocol: "oidc", Enabled: true},
	}

	mockSvc.On("ListApps", mock.Anything).Return(expectedApps, nil)

	req, _ := http.NewRequest("GET", "/apps", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apps []*types.DevCenterAppDTO
	json.Unmarshal(rr.Body.Bytes(), &apps)

	assert.Equal(t, expectedApps, apps)

	mockSvc.AssertExpectations(t)
}
