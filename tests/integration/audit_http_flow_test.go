package integration

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/server/http"
	"github.com/turtacn/QuantaID/pkg/audit/events"
	"github.com/turtacn/QuantaID/pkg/utils"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogin_AuditEventsEmitted(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test; set INTEGRATION_TESTS to run.")
	}
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	// Setup: Create a server with a file-based audit sink
	cfg := utils.Config{
		Storage: utils.StorageConfig{Mode: "memory"},
	}
	httpCfg := http.Config{Address: ":0"} // Use a random free port
	logger := utils.NewNoopLogger()
	crypto := utils.NewCryptoManager("secret")

	// Point the audit config to our temp log file
	originalJulesConfig := "configs/audit/pipeline.jules.yaml"
	tempJulesConfig := filepath.Join(tmpDir, "pipeline.jules.yaml")
	configData, err := os.ReadFile(originalJulesConfig)
	require.NoError(t, err)

	newConfigData := bytes.Replace(configData, []byte("./logs/audit_jules.log"), []byte(logPath), 1)
	err = os.WriteFile(tempJulesConfig, newConfigData, 0644)
	require.NoError(t, err)

	// We need to temporarily move the config file so the server can find it
	err = os.Rename(originalJulesConfig, originalJulesConfig+".bak")
	require.NoError(t, err)
	err = os.Rename(tempJulesConfig, originalJulesConfig)
	require.NoError(t, err)

	defer func() {
		os.Remove(originalJulesConfig)
		os.Rename(originalJulesConfig+".bak", originalJulesConfig)
	}()

	server, err := http.NewServerWithConfig(httpCfg, &cfg, logger, crypto)
	require.NoError(t, err)
	testServer := httptest.NewServer(server.Router)
	defer testServer.Close()

	// Act: Perform a failed login request
	loginReq := map[string]string{"username": "testuser", "password": "wrongpassword"}
	reqBody, _ := json.Marshal(loginReq)
	resp, err := testServer.Client().Post(testServer.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	// Assert: Check the audit log file for the login_failed event
	// We need to give the file sink a moment to write the event
	time.Sleep(100 * time.Millisecond)

	logFile, err := os.Open(logPath)
	require.NoError(t, err)
	defer logFile.Close()

	decoder := json.NewDecoder(logFile)
	var foundEvent bool
	for {
		var event events.AuditEvent
		if err := decoder.Decode(&event); err == io.EOF {
			break
		}
		require.NoError(t, err)

		if event.Action == "login_failed" {
			foundEvent = true
			assert.Equal(t, "auth", event.Category)
			assert.Equal(t, "testuser", event.UserID)
			assert.Equal(t, events.ResultFailure, event.Result)
			break
		}
	}

	assert.True(t, foundEvent, "Expected login_failed event was not found in audit log")
}
