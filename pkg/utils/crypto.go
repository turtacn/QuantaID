package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// CryptoManager provides cryptographic functions.
type CryptoManager struct {
	jwtSecret []byte
}

// NewCryptoManager creates a new CryptoManager.
func NewCryptoManager(jwtSecret string) *CryptoManager {
	return &CryptoManager{jwtSecret: []byte(jwtSecret)}
}

// HashPassword generates a bcrypt hash of the password.
func (cm *CryptoManager) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a password with a hash.
func (cm *CryptoManager) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates a new JWT for a given user ID and custom claims.
func (cm *CryptoManager) GenerateJWT(userID string, duration time.Duration, claims jwt.MapClaims) (string, error) {
	if claims == nil {
		claims = jwt.MapClaims{}
	}

	claims["sub"] = userID
	claims["iss"] = "QuantaID"
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["jti"] = uuid.New().String() // Add a unique identifier for the token

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(cm.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	return signedToken, nil
}

// ValidateJWT validates a JWT and returns the claims.
func (cm *CryptoManager) ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return cm.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateUUID generates a new UUID string.
func (cm *CryptoManager) GenerateUUID() string {
	return uuid.New().String()
}

//Personal.AI order the ending
