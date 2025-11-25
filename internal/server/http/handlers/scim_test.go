package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	scim_pkg "github.com/turtacn/QuantaID/pkg/scim"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// MockIdentityService for testing
type MockIdentityService struct {
	identity.IService
	createFunc      func(ctx context.Context, username, email, password string) (*types.User, error)
	getUserByIDFunc func(ctx context.Context, userID string) (*types.User, error)
	listUsersFunc   func(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error)
	repo            identity.UserRepository
}

func (m *MockIdentityService) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) {
	return m.createFunc(ctx, username, email, password)
}

func (m *MockIdentityService) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	return m.getUserByIDFunc(ctx, userID)
}

func (m *MockIdentityService) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) {
	return m.listUsersFunc(ctx, filter)
}

func (m *MockIdentityService) GetUserRepo() identity.UserRepository {
	return m.repo
}

// MockRepo for updates
type MockRepo struct {
	identity.UserRepository
	updateUserFunc func(ctx context.Context, user *types.User) error
}

func (m *MockRepo) UpdateUser(ctx context.Context, user *types.User) error {
	if m.updateUserFunc != nil {
		return m.updateUserFunc(ctx, user)
	}
	return nil
}

func TestSCIMHandler_CreateUser(t *testing.T) {
	mockRepo := &MockRepo{}
	mockSvc := &MockIdentityService{
		createFunc: func(ctx context.Context, username, email, password string) (*types.User, error) {
			return &types.User{
				ID:       "123",
				Username: username,
				Email:    types.EncryptedString(email),
				Status:   types.UserStatusActive,
			}, nil
		},
		repo: mockRepo,
	}

	logger := &utils.ZapLogger{} // Simplified or mock
	handler := NewSCIMHandler(mockSvc, logger)

	sUser := scim_pkg.User{
		UserName: "bjensen",
		Emails: []scim_pkg.Email{{Value: "bjensen@example.com", Primary: true}},
	}
	body, _ := json.Marshal(sUser)
	req := httptest.NewRequest("POST", "/Users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != scim_pkg.ContentType {
		t.Errorf("Expected Content-Type %s, got %s", scim_pkg.ContentType, resp.Header.Get("Content-Type"))
	}
}

func TestSCIMHandler_GetUser(t *testing.T) {
	mockSvc := &MockIdentityService{
		getUserByIDFunc: func(ctx context.Context, userID string) (*types.User, error) {
			if userID == "123" {
				return &types.User{
					ID:       "123",
					Username: "bjensen",
					Status:   types.UserStatusActive,
				}, nil
			}
			return nil, types.ErrNotFound
		},
	}
	logger := &utils.ZapLogger{}
	handler := NewSCIMHandler(mockSvc, logger)

	// Test Found
	req := httptest.NewRequest("GET", "/Users/123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "123"})
	w := httptest.NewRecorder()
	handler.GetUser(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}

	// Test Not Found
	req = httptest.NewRequest("GET", "/Users/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w = httptest.NewRecorder()
	handler.GetUser(w, req)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Result().StatusCode)
	}
}
