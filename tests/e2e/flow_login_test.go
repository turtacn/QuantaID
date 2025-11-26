//go:build integration
package e2e_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	server "github.com/turtacn/QuantaID/internal/server/http"
	"github.com/turtacn/QuantaID/pkg/utils"
)

func TestFullOIDCLoginFlow(t *testing.T) {
	// --- Test Setup ---
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	configManager, err := utils.NewConfigManager("../../configs", "server", "yaml", logger)
	require.NoError(t, err, "Failed to load configuration")
	var appCfg utils.Config
	err = configManager.Unmarshal(&appCfg)
	require.NoError(t, err, "Failed to unmarshal configuration")

	pgPort, _ := strconv.Atoi(os.Getenv("QID_POSTGRES_PORT"))
	redisPort, _ := strconv.Atoi(os.Getenv("QID_REDIS_PORT"))

	// Override config with test container details from env vars
	appCfg.Postgres.Host = os.Getenv("QID_POSTGRES_HOST")
	appCfg.Postgres.Port = pgPort
	appCfg.Postgres.User = os.Getenv("QID_POSTGRES_USER")
	appCfg.Postgres.Password = os.Getenv("QID_POSTGRES_PASSWORD")
	appCfg.Postgres.DbName = os.Getenv("QID_POSTGRES_DBNAME")
	appCfg.Redis.Host = os.Getenv("QID_REDIS_HOST")
	appCfg.Redis.Port = redisPort
	appCfg.Storage.Mode = "postgres"

	cryptoManager := utils.NewCryptoManager("test-secret")

	httpServer, err := server.NewServerWithConfig(server.Config{
		Address:      ":8081", // Use a different port for E2E tests
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}, &appCfg, logger, cryptoManager)
	require.NoError(t, err, "Failed to create server")

	go httpServer.Start()
	defer httpServer.Stop(context.Background())

	err = waitForServer("http://localhost:8081/healthz", 5*time.Second)
	require.NoError(t, err, "Server did not start in time")

	// Create a test user directly via the service
	testUser, err := httpServer.Services.IdentityDomainService.CreateUser(context.Background(), "alice", "alice@example.com", "password123")
	require.NoError(t, err, "Failed to create test user")

	// --- Test Execution ---

	// Use a client with a cookie jar to simulate a browser
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		// Don't follow redirects automatically
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Step 1: Client requests /oauth/authorize
	authURL := "http://localhost:8081/oauth/authorize?client_id=test-client&redirect_uri=http://localhost:8080/callback&response_type=code&scope=openid"
	resp, err := client.Get(authURL)
	require.NoError(t, err, "Failed on initial authorize request")
	defer resp.Body.Close()

	// Assert 1: Received 302 redirect to the login page
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	location, err := resp.Location()
	require.NoError(t, err)
	assert.Contains(t, location.Path, "/auth/login")

	// Step 2: Extract CSRF token and submit login form
	loginPageResp, err := client.Get(location.String())
	require.NoError(t, err)
	defer loginPageResp.Body.Close()

	bodyBytes, _ := io.ReadAll(loginPageResp.Body)
	csrfToken := extractCSRFToken(string(bodyBytes))
	require.NotEmpty(t, csrfToken, "Could not find CSRF token on login page")

	loginValues := url.Values{}
	loginValues.Set("username", "alice")
	loginValues.Set("password", "password123")
	loginValues.Set("_csrf", csrfToken)

	// Action 2: Simulate submitting the login form
	postResp, err := client.PostForm("http://localhost:8081"+location.Path, loginValues)
	require.NoError(t, err)
	defer postResp.Body.Close()

	// Assert 2: Received 302 redirect back to the callback URL with an authorization code
	assert.Equal(t, http.StatusFound, postResp.StatusCode, "Expected redirect after successful login")
	callbackLocation, err := postResp.Location()
	require.NoError(t, err)
	assert.Equal(t, "/callback", callbackLocation.Path)
	authCode := callbackLocation.Query().Get("code")
	require.NotEmpty(t, authCode, "Authorization code was missing from callback URL")

	// Step 3: Exchange authorization code for a token
	tokenValues := url.Values{}
	tokenValues.Set("grant_type", "authorization_code")
	tokenValues.Set("code", authCode)
	tokenValues.Set("client_id", "test-client")
	tokenValues.Set("client_secret", "test-secret")
	tokenValues.Set("redirect_uri", "http://localhost:8080/callback")

	// This request doesn't need the cookie jar
	tokenResp, err := http.PostForm("http://localhost:8081/oauth/token", tokenValues)
	require.NoError(t, err)
	defer tokenResp.Body.Close()

	// Assert 3: Response contains the tokens
	require.Equal(t, http.StatusOK, tokenResp.StatusCode, "Token exchange failed")
	var tokenResponse map[string]interface{}
	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	require.NoError(t, err)
	assert.Contains(t, tokenResponse, "access_token")
	assert.Contains(t, tokenResponse, "id_token")
	assert.Contains(t, tokenResponse, "refresh_token")

	// Step 4: Parse the ID token and verify its claims
	idTokenString, ok := tokenResponse["id_token"].(string)
	require.True(t, ok, "id_token was not a string")

	claims, err := cryptoManager.ValidateJWT(idTokenString)
	require.NoError(t, err)
	assert.Equal(t, testUser.ID, claims["sub"])
}

func extractCSRFToken(body string) string {
	// Super simple CSRF extraction for test purposes
	// In a real app, this might be more robust (e.g., using goquery)
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if strings.Contains(line, "name=\"_csrf\"") {
			parts := strings.Split(line, "\"")
			for i, part := range parts {
				if part == "value=" && i+1 < len(parts) {
					return parts[i+1]
				}
			}
		}
	}
	return ""
}
