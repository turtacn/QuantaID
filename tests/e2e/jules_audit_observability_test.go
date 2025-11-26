package e2e_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// This E2E test assumes the server is running and accessible at SERVER_URL.
// It also assumes the audit log is being written to a known location AUDIT_LOG_PATH.

const (
	serverURL    = "http://localhost:8080" // Default, can be overridden by env var
	auditLogPath = "./logs/audit_jules.log" // Default, can be overridden by env var
)

func TestE2E_Audit_LoginFailure(t *testing.T) {
	if os.Getenv("E2E_TESTS") == "" {
		t.Skip("Skipping E2E test; set E2E_TESTS to run.")
	}

	targetURL := os.Getenv("SERVER_URL")
	if targetURL == "" {
		targetURL = serverURL
	}

	targetLogPath := os.Getenv("AUDIT_LOG_PATH")
	if targetLogPath == "" {
		targetLogPath = auditLogPath
	}

	// Reset log file to ensure a clean state
	_ = os.Truncate(targetLogPath, 0)

	// Act: Perform a failed login request
	loginReq := map[string]string{"username": "e2e_user", "password": "e2e_password"}
	reqBody, _ := json.Marshal(loginReq)
	resp, err := http.Post(targetURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Assert: Check the audit log file for the login_failed event
	time.Sleep(200 * time.Millisecond) // Allow time for file write

	logData, err := os.ReadFile(targetLogPath)
	require.NoError(t, err)

	assert.Contains(t, string(logData), `"action":"login_failed"`)
	assert.Contains(t, string(logData), `"user_id":"e2e_user"`)
}

func TestE2E_Metrics_Endpoint(t *testing.T) {
	if os.Getenv("E2E_TESTS") == "" {
		t.Skip("Skipping E2E test; set E2E_TESTS to run.")
	}

	targetURL := os.Getenv("SERVER_URL")
	if targetURL == "" {
		targetURL = serverURL
	}

	// Act: Make a request to a known endpoint to generate metrics
	_, err := http.Get(targetURL + "/health")
	require.NoError(t, err)

	// Assert: Check the /metrics endpoint
	resp, err := http.Get(targetURL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Check for a specific HTTP metric that should now exist
	bodyStr := string(body)
	assert.True(t, strings.Contains(bodyStr, `quantaid_http_requests_total{method="GET",path="/health",status="200"}`),
		"Metric for /health endpoint not found")
}
