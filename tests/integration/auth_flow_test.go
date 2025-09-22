package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/server/http"
	authservice "github.com/turtacn/QuantaID/internal/services/auth"
	authorizationservice "github.com/turtacn/QuantaID/internal/services/authorization"
	identityservice "github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTestServer(t *testing.T) *httptest.Server {
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	cryptoManager := utils.NewCryptoManager("test-secret")

	identityRepo := postgresql.NewInMemoryIdentityRepository()
	authDbRepo := postgresql.NewInMemoryAuthRepository()
	policyRepo := postgresql.NewInMemoryPolicyRepository()
	sessionRepo := redis.NewInMemorySessionRepository()
	tokenRepo := redis.NewInMemoryTokenRepository()

	hashedPassword, _ := cryptoManager.HashPassword("password123")
	testUser := &types.User{ID: "user-test-1", Username: "testuser", Email: "test@example.com", Password: hashedPassword, Status: types.UserStatusActive}
	err := identityRepo.CreateUser(context.Background(), testUser)
	require.NoError(t, err)

	identityDomainSvc := identity.NewService(identityRepo, identityRepo, cryptoManager, logger)
	authDomainSvc := auth.NewService(identityDomainSvc, sessionRepo, tokenRepo, authDbRepo, cryptoManager, logger)
	policyDomainSvc := policy.NewService(policyRepo, logger)

	identityAppSvc := identityservice.NewApplicationService(identityDomainSvc, logger)
	authAppSvc := authservice.NewApplicationService(authDomainSvc, logger, authservice.Config{
		AccessTokenDuration: time.Minute,
	})
	authzAppSvc := authorizationservice.NewApplicationService(policyDomainSvc, identityDomainSvc, logger)

	serverConfig := http.Config{Address: ":0"}
	services := http.Services{
		AuthService:     authAppSvc,
		IdentityService: identityAppSvc,
		AuthzService:    authzAppSvc,
		CryptoManager:   cryptoManager,
	}
	httpServer := http.NewServer(serverConfig, logger, services)

	return httptest.NewServer(httpServer.Router)
}

func TestAuthFlow_SuccessfulLogin(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	loginCreds := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(loginCreds)

	resp, err := server.Client().Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var loginResp authservice.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.Equal(t, "user-test-1", loginResp.User.ID)
}

func TestAuthFlow_FailedLogin(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	loginCreds := map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(loginCreds)

	resp, err := server.Client().Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errorResp struct {
		Error *types.Error `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Equal(t, types.ErrInvalidCredentials.Code, errorResp.Error.Code)
}

//Personal.AI order the ending
