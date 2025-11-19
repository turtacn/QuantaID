package handlers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/pkg/auth/protocols"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
)

var testPrivateKey string

func init() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	testPrivateKey = string(privateKeyPEM)
}

func TestOIDCHandler_Discovery(t *testing.T) {
	// Arrange
	logger := utils.NewNoopLogger()
	adapter := protocols.NewOIDCAdapter()
	adapter.Initialize(context.Background(), types.ConnectorConfig{
		PrivateKey: testPrivateKey,
	}, logger)
	handler := NewOIDCHandler(adapter.(*protocols.OIDCAdapter), logger, "http://localhost:8080")
	req, err := http.NewRequest("GET", "/.well-known/openid-configuration", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	// Act
	handler.Discovery(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	var discovery map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &discovery)
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", discovery["issuer"])
	assert.Equal(t, "http://localhost:8080/oauth/authorize", discovery["authorization_endpoint"])
}

func TestOIDCHandler_JWKS(t *testing.T) {
	// Arrange
	logger := utils.NewNoopLogger()
	adapter := protocols.NewOIDCAdapter()
	adapter.Initialize(context.Background(), types.ConnectorConfig{
		PrivateKey: testPrivateKey,
	}, logger)
	handler := NewOIDCHandler(adapter.(*protocols.OIDCAdapter), logger, "http://localhost:8080")
	req, err := http.NewRequest("GET", "/.well-known/jwks.json", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	// Act
	handler.JWKS(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	var jwks map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &jwks)
	assert.NoError(t, err)
	assert.Contains(t, jwks, "keys")
}
