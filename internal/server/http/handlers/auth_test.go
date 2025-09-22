package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Mock Service ---

type MockAuthService struct {
	mock.Mock
}
func (m *MockAuthService) Login(ctx context.Context, req auth.LoginRequest) (*auth.LoginResponse, *types.Error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*types.Error)
	}
	return args.Get(0).(*auth.LoginResponse), nil
}
func (m *MockAuthService) Logout(ctx context.Context, req auth.LogoutRequest) *types.Error {
	return nil
}

// --- Tests ---

func TestAuthHandlers_Login(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	// We need to update NewAuthHandlers to accept an interface
	// For now, let's assume this change is made.
	authHandler := NewAuthHandlers(mockAuthSvc, logger)

	t.Run("Successful Login", func(t *testing.T) {
		loginReq := auth.LoginRequest{Username: "test", Password: "password"}
		loginResp := &auth.LoginResponse{AccessToken: "some-token", User: &auth.UserDTO{ID: "123"}}

		mockAuthSvc.On("Login", mock.Anything, loginReq).Return(loginResp, nil).Once()

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		authHandler.Login(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var respBody auth.LoginResponse
		err := json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, "some-token", respBody.AccessToken)
		mockAuthSvc.AssertExpectations(t)
	})

	t.Run("Failed Login", func(t *testing.T) {
		loginReq := auth.LoginRequest{Username: "test", Password: "wrong"}

		mockAuthSvc.On("Login", mock.Anything, loginReq).Return(nil, types.ErrInvalidCredentials).Once()

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		authHandler.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var errBody map[string]map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &errBody)
		assert.NoError(t, err)
		assert.Equal(t, types.ErrInvalidCredentials.Code, errBody["error"]["code"])
		mockAuthSvc.AssertExpectations(t)
	})
}

//Personal.AI order the ending
