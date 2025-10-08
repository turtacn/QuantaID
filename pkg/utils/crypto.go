package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// CryptoManager provides cryptographic utility functions, such as password hashing,
// JWT generation/validation, and UUID creation.
type CryptoManager struct {
	jwtSecret []byte
}

// NewCryptoManager creates a new CryptoManager with the given JWT secret.
// The secret is used for signing and verifying JWTs.
//
// Parameters:
//   - jwtSecret: The secret key for JWT operations.
//
// Returns:
//   A new CryptoManager instance.
func NewCryptoManager(jwtSecret string) *CryptoManager {
	return &CryptoManager{jwtSecret: []byte(jwtSecret)}
}

// HashPassword generates a bcrypt hash of a given password.
//
// Parameters:
//   - password: The plain-text password to hash.
//
// Returns:
//   The hashed password as a string, or an error if hashing fails.
func (cm *CryptoManager) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plain-text password with a bcrypt hash to see if they match.
//
// Parameters:
//   - password: The plain-text password.
//   - hash: The bcrypt hash to compare against.
//
// Returns:
//   True if the password matches the hash, false otherwise.
func (cm *CryptoManager) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT creates and signs a new JSON Web Token (JWT).
// It includes standard claims (sub, iss, iat, exp, jti) and any custom claims provided.
//
// Parameters:
//   - userID: The subject of the token.
//   - duration: The token's validity period.
//   - claims: A map of custom claims to include in the token.
//
// Returns:
//   The signed JWT as a string, or an error if signing fails.
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

// ValidateJWT parses and validates a JWT string.
// It checks the signature and standard claims (like expiration).
//
// Parameters:
//   - tokenString: The JWT string to validate.
//
// Returns:
//   The claims from the token as a map if the token is valid, or an error otherwise.
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

// GenerateUUID generates a new version 4 UUID as a string.
//
// Returns:
//   A new UUID string.
func (cm *CryptoManager) GenerateUUID() string {
	return uuid.New().String()
}
