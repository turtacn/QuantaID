package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyPKCE(t *testing.T) {
	// Test case for S256 method
	codeVerifier := "test_verifier"
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	assert.True(t, VerifyPKCE(codeVerifier, codeChallenge, "S256"))

	// Test case for plain method
	codeVerifier = "test_verifier"
	codeChallenge = "test_verifier"
	assert.True(t, VerifyPKCE(codeVerifier, codeChallenge, "plain"))

	// Test case for invalid method
	codeVerifier = "test_verifier"
	codeChallenge = "test_verifier"
	assert.False(t, VerifyPKCE(codeVerifier, codeChallenge, "invalid"))
}
