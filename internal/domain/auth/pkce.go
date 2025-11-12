package auth

import (
	"crypto/sha256"
	"encoding/base64"
)

// VerifyPKCE verifies the code verifier against the code challenge.
func VerifyPKCE(codeVerifier, codeChallenge, method string) bool {
	if method != "S256" {
		return false // Only S256 is supported.
	}

	hash := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])

	return computed == codeChallenge
}
