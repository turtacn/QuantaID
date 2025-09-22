package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Mock Service ---

type MockIdentityService struct {
	mock.Mock
}
func (m *MockIdentityService) CreateUser(ctx context.Context, req identity.CreateUserRequest) (*types.User, *types.Error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil { return nil, args.Get(1).(*types.Error) }
	return args.Get(0).(*types.User), nil
}
func (m *MockIdentityService) GetUserByID(ctx context.Context, userID string) (*types.User, *types.Error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil { return nil, args.Get(1).(*types.Error) }
	return args.Get(0).(*types.User), nil
}
func (m *MockIdentityService) AddUserToGroup(ctx context.Context, req identity.AddUserToGroupRequest) *types.Error { return nil }

// --- Tests ---

func TestIdentityHandlers_CreateUser(t *testing.T) {
	mockIdentitySvc := new(MockIdentityService)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	identityHandler := NewIdentityHandlers(mockIdentitySvc, logger)

	createReq := identity.CreateUserRequest{Username: "test", Email: "test@test.com", Password: "password"}
	userResp := &types.User{ID: "123", Username: "test"}

	mockIdentitySvc.On("CreateUser", mock.Anything, createReq).Return(userResp, nil).Once()

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	identityHandler.CreateUser(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var respBody types.User
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "123", respBody.ID)
	mockIdentitySvc.AssertExpectations(t)
}

func TestIdentityHandlers_GetUser(t *testing.T) {
	mockIdentitySvc := new(MockIdentityService)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	identityHandler := NewIdentityHandlers(mockIdentitySvc, logger)

	userID := "user-456"
	userResp := &types.User{ID: userID, Username: "found_user"}

	mockIdentitySvc.On("GetUserByID", mock.Anything, userID).Return(userResp, nil).Once()

	req := httptest.NewRequest("GET", fmt.Sprintf("/users/%s", userID), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/users/{id}", identityHandler.GetUser)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody types.User
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, userID, respBody.ID)
	mockIdentitySvc.AssertExpectations(t)
}

//Personal.AI order the ending
