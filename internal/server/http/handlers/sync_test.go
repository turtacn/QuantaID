package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/internal/services/sync"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// Mocks
type MockLDAPConnector struct {
	mock.Mock
}

func (m *MockLDAPConnector) SyncUsers(ctx context.Context) ([]*types.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*types.User), args.Error(1)
}

func (m *MockLDAPConnector) SearchUsers(ctx context.Context, filter string) ([]*types.User, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*types.User), args.Error(1)
}

type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) RecordAdminAction(ctx context.Context, userID, ip, resource, action, traceID string, details map[string]any) {
	m.Called(ctx, userID, ip, resource, action, traceID, details)
}

func setupHandlerTest(t *testing.T) (*httptest.Server, *memory.IdentityMemoryRepository, *MockLDAPConnector, *MockAuditService) {
	userRepo := memory.NewIdentityMemoryRepository()
	ldapConnector := new(MockLDAPConnector)
	auditService := new(MockAuditService)
	logger := zap.NewNop()

	config := &sync.LDAPSyncConfig{
		ConflictStrategy: sync.ConflictPreferRemote,
	}

	syncService := sync.NewLDAPSyncService(ldapConnector, userRepo, config, auditService, logger)
	syncHandler := handlers.NewSyncHandler(syncService, logger)

	router := mux.NewRouter()
	syncHandler.RegisterRoutes(router)

	server := httptest.NewServer(router)
	return server, userRepo, ldapConnector, auditService
}

func Test_HandleFullSync(t *testing.T) {
	server, _, ldapConnector, auditService := setupHandlerTest(t)
	defer server.Close()

	ldapUsers := []*types.User{
		{Username: "testuser", Email: "test@test.com"},
	}

	ldapConnector.On("SyncUsers", mock.Anything).Return(ldapUsers, nil)
	auditService.On("RecordAdminAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	req, _ := http.NewRequest("POST", server.URL+"/admin/sync/ldap/full", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats sync.SyncStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats.Created)
}

func Test_HandleIncrementalSync(t *testing.T) {
	server, _, ldapConnector, auditService := setupHandlerTest(t)
	defer server.Close()

	since := time.Now().UTC().Add(-5 * time.Minute)
	sinceStr := since.Format(time.RFC3339)

	ldapConnector.On("SearchUsers", mock.Anything, mock.Anything).Return([]*types.User{}, nil)
	auditService.On("RecordAdminAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	req, _ := http.NewRequest("POST", server.URL+"/admin/sync/ldap/incremental?since="+sinceStr, nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func Test_HandleGetStatus(t *testing.T) {
	server, _, ldapConnector, auditService := setupHandlerTest(t)
	defer server.Close()

	// Perform a sync to populate the status
	ldapConnector.On("SyncUsers", mock.Anything).Return([]*types.User{}, nil)
	auditService.On("RecordAdminAction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	req, _ := http.NewRequest("POST", server.URL+"/admin/sync/ldap/full", nil)
	_, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	// Now get the status
	req, _ = http.NewRequest("GET", server.URL+"/admin/sync/ldap/status", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats sync.SyncStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	assert.NoError(t, err)
	assert.NotZero(t, stats.StartTime)
}
