package utils

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// CryptoManagerInterface defines the public contract for the crypto manager.
type CryptoManagerInterface interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
	GenerateJWT(userID string, duration time.Duration, claims jwt.MapClaims) (string, error)
	ValidateJWT(tokenString string) (jwt.MapClaims, error)
	GenerateUUID() string
	GenerateRecoveryCodes() ([]string, error)
	HashRecoveryCode(code string) (string, error)
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertextHex string) (string, error)
}
