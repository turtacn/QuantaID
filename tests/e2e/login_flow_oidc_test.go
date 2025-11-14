package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/orchestrator"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	audit_service "github.com/turtacn/QuantaID/internal/services/audit"
	auth_service "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/internal/workflows"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func setupTestServer(t *testing.T) (*httptest.Server, *memory.IdentityMemoryRepository, *memory.InMemorySink) {
	logger := utils.NewNoopLogger()
	cryptoManager := utils.NewCryptoManager("test-secret")

	identityRepo := memory.NewIdentityMemoryRepository()
	identityService := identity.NewService(identityRepo, identityRepo, cryptoManager, logger)

	auditSink := memory.NewInMemorySink()
	auditPipeline := audit.NewPipeline(zap.NewNop(), auditSink)
	auditService := audit_service.NewService(auditPipeline)

	authRepo := memory.NewAuthMemoryRepository()
	riskEngine := auth_service.NewSimpleRiskEngine(auth_service.SimpleRiskConfig{MfaThreshold: 100}, auditService)
	authDomainService := auth.NewService(identityService, authRepo, authRepo, authRepo, cryptoManager, logger, riskEngine)
	authAppService := auth_service.NewApplicationService(authDomainService, auditService, logger, auth_service.Config{AccessTokenDuration: time.Minute * 15}, trace.NewNoopTracerProvider().Tracer("test"))

	engine := orchestrator.NewEngine(logger)
	workflows.RegisterLoginWorkflow(engine, workflows.LoginDeps{
		RiskEngine:   riskEngine,
		AuthService:  authAppService,
		AuditService: auditService,
		Logger:       zap.NewNop(),
	})

	authHandlers := handlers.NewAuthHandlers(authAppService, engine, logger)

	r := mux.NewRouter()
	r.HandleFunc("/auth/login", authHandlers.Login).Methods("POST")
	r.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")

	return httptest.NewServer(r), identityRepo, auditSink
}

func TestLoginFlow_Success(t *testing.T) {
	server, identityRepo, auditSink := setupTestServer(t)
	defer server.Close()

	// Create a user for testing
	cryptoManager := utils.NewCryptoManager("test-secret")
	hashedPassword, _ := cryptoManager.HashPassword("password")
	user := &types.User{
		ID:       "test-user",
		Username: "testuser",
		Password: hashedPassword,
		Status:   types.UserStatusActive,
	}
	err := identityRepo.CreateUser(context.Background(), user)
	assert.NoError(t, err)

	loginReq := auth_service.LoginRequest{
		Username: "testuser",
		Password: "password",
	}
	reqBody, _ := json.Marshal(loginReq)

	resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp auth_service.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)

	// Check audit log
	events := auditSink.GetEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "login_success", events[0].Action)
	assert.Equal(t, user.ID, events[0].UserID)

	// Check protected route
	req, _ := http.NewRequest("GET", server.URL+"/protected", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
