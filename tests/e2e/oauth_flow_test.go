package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/pkg/auth/protocols"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

type mockUserRepo struct{}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	return &types.User{ID: id, Status: types.UserStatusActive, Attributes: map[string]interface{}{"name": "test"}, Email: "test@test.com"}, nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	return &types.User{Email: email}, nil
}

func (m *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	return &types.User{ID: "test_user", Username: username, Status: types.UserStatusActive, Attributes: map[string]interface{}{"name": "test"}, Email: "test@test.com"}, nil
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *types.User) error {
	return nil
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, user *types.User) error {
	return nil
}

func (m *mockUserRepo) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int64, error) {
	return []*types.User{}, 0, nil
}

type mockAppRepo struct{}

func (m *mockAppRepo) GetApplicationByClientID(ctx context.Context, clientID string) (*types.Application, error) {
	return &types.Application{
		ProtocolConfig: map[string]interface{}{
			"client_id":     clientID,
			"redirect_uris": []string{"http://localhost:3000/callback"},
		},
	}, nil
}

type mockRedisClient struct {
	data map[string]string
}

func (m *mockRedisClient) Client() *redis.Client {
	return nil
}

func (m *mockRedisClient) Close() error {
	return nil
}

func (m *mockRedisClient) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.data[key] = string(value.([]byte))
	return nil
}

func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	val, ok := m.data[key]
	if !ok {
		return "", redis.Nil
	}
	return val, nil
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func readBody(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	return buf.String()
}

func TestOAuthAuthorizationCodeFlow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	redisClient := &mockRedisClient{data: make(map[string]string)}
	oauthAdapter := protocols.NewOAuthAdapter().(*protocols.OAuthAdapter)
	oidcAdapter := protocols.NewOIDCAdapter().(*protocols.OIDCAdapter)
	oauthAdapter.SetUserRepo(&mockUserRepo{})
	oauthAdapter.SetAppRepo(&mockAppRepo{})
	oauthAdapter.SetRedis(redisClient)
	oauthAdapter.SetOIDCAdapter(oidcAdapter)
	oidcAdapter.SetUserRepo(&mockUserRepo{})
	oidcAdapter.SetPrivateKey(privateKey)
	oauthHandler := handlers.NewOAuthHandler(oauthAdapter, utils.NewZapLoggerWrapper(logger))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oidcHandler := handlers.NewOIDCHandler(oidcAdapter, utils.NewZapLoggerWrapper(logger), r.Host)
		switch r.URL.Path {
		case "/oauth/authorize":
			oauthHandler.Authorize(w, r)
		case "/oauth/token":
			oauthHandler.Token(w, r)
		case "/oauth/userinfo":
			oidcHandler.UserInfo(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// 1. Authorization Request
	codeVerifier := "test_code_verifier"
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	authURL := server.URL + "/oauth/authorize" +
		"?response_type=code" +
		"&client_id=test_client_id" +
		"&redirect_uri=http://localhost:3000/callback" +
		"&scope=openid profile email" +
		"&state=test_state" +
		"&code_challenge=" + codeChallenge +
		"&code_challenge_method=S256"

	req, _ := http.NewRequest("GET", authURL, nil)
	// We need a client that doesn't follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, _ := client.Do(req)

	if resp.StatusCode != http.StatusFound {
		body := readBody(resp.Body)
		t.Logf("Response body: %s", body)
		t.Fatalf("expected 302 redirect, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		t.Fatalf("missing Location header; got status %d", resp.StatusCode)
	}

	redirectURL, err := url.Parse(location)
	require.NoError(t, err)
	assert.NotEmpty(t, redirectURL.Query().Get("code"))

	// 2. Token Exchange
	tokenReq := url.Values{}
	tokenReq.Set("grant_type", "authorization_code")
	tokenReq.Set("code", redirectURL.Query().Get("code"))
	tokenReq.Set("redirect_uri", "http://localhost:3000/callback")
	tokenReq.Set("client_id", "test_client_id")
	tokenReq.Set("code_verifier", codeVerifier)

	req, _ = http.NewRequest("POST", server.URL+"/oauth/token", strings.NewReader(tokenReq.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, _ = http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp types.TokenResponse
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.NotEmpty(t, tokenResp.AccessToken)
	assert.NotEmpty(t, tokenResp.IDToken)

	// 3. UserInfo Request
	req, _ = http.NewRequest("GET", server.URL+"/oauth/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	resp, _ = http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userInfo types.UserInfo
	json.NewDecoder(resp.Body).Decode(&userInfo)
	assert.NotEmpty(t, userInfo.Subject)
}
