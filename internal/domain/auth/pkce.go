package auth

import (
	"crypto/sha256"
	"encoding/base64"
)

// VerifyPKCE verifies the PKCE code challenge.
func VerifyPKCE(codeVerifier, codeChallenge, codeChallengeMethod string) bool {
	switch codeChallengeMethod {
	case "S256":
		h := sha256.New()
		h.Write([]byte(codeVerifier))
		hashed := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
		return hashed == codeChallenge
	case "plain":
		return codeVerifier == codeChallenge
	default:
		return false
	}
}
