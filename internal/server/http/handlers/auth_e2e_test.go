package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// MockAuthService is a mock implementation of the AuthServiceInterface for testing.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) LoginWithPassword(ctx context.Context, req auth.AuthnRequest, serviceConfig auth.Config) (*types.AuthResult, error) {
	args := m.Called(ctx, req, serviceConfig)
	return args.Get(0).(*types.AuthResult), args.Error(1)
}

func (m *MockAuthService) VerifyMFAChallenge(ctx context.Context, req *types.VerifyMFARequest, serviceConfig auth.Config) (*types.AuthResult, error) {
	args := m.Called(ctx, req, serviceConfig)
	return args.Get(0).(*types.AuthResult), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, sessionID, accessToken string) error {
	args := m.Called(ctx, sessionID, accessToken)
	return args.Error(0)
}

func TestAuthFlow_LowRisk_NoMFA(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handlers := NewAuthHandlers(mockAuthService, &utils.MockLogger{})

	authResult := &types.AuthResult{
		Token: &types.Token{AccessToken: "test-token"},
	}
	mockAuthService.On("LoginWithPassword", mock.Anything, mock.Anything, mock.Anything).Return(authResult, nil)

	reqBody, _ := json.Marshal(auth.AuthnRequest{Username: "test", Password: "password"})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	// Act
	handlers.Login(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	var token types.Token
	json.Unmarshal(rr.Body.Bytes(), &token)
	assert.Equal(t, "test-token", token.AccessToken)
}

func TestAuthFlow_MediumRisk_WithTOTP(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	handlers := NewAuthHandlers(mockAuthService, &utils.MockLogger{})

	authResult := &types.AuthResult{
		IsMfaRequired: true,
		MFAChallenge:  &types.MFAChallenge{ChallengeID: "test-challenge"},
	}
	mockAuthService.On("LoginWithPassword", mock.Anything, mock.Anything, mock.Anything).Return(authResult, nil)

	reqBody, _ := json.Marshal(auth.AuthnRequest{Username: "test", Password: "password"})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	// Act
	handlers.Login(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	var mfaChallenge types.MFAChallenge
	json.Unmarshal(rr.Body.Bytes(), &mfaChallenge)
	assert.Equal(t, "test-challenge", mfaChallenge.ChallengeID)
}
