package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"io"
	"time"
)

// CryptoManager provides cryptographic utility functions, such as password hashing,
// JWT generation/validation, and UUID creation.
type CryptoManager struct {
	jwtSecret []byte
	aesKey    []byte
}

// NewCryptoManager creates a new CryptoManager with the given JWT secret.
// The secret is used for signing and verifying JWTs and deriving the AES key.
//
// Parameters:
//   - jwtSecret: The secret key for JWT operations.
//
// Returns:
//   A new CryptoManager instance.
func NewCryptoManager(jwtSecret string) *CryptoManager {
	hash := sha256.Sum256([]byte(jwtSecret))
	return &CryptoManager{
		jwtSecret: []byte(jwtSecret),
		aesKey:    hash[:], // Use the 32-byte hash as the AES key
	}
}

// Encrypt encrypts plaintext using AES-GCM and returns it as a hex-encoded string.
func (cm *CryptoManager) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(cm.aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a hex-encoded AES-GCM ciphertext.
func (cm *CryptoManager) Decrypt(ciphertextHex string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex ciphertext: %w", err)
	}

	block, err := aes.NewCipher(cm.aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext is too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
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

// GenerateRandomString generates a cryptographically secure random string of specified length.
// It uses a predefined set of characters including uppercase, lowercase, digits, and special symbols.
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes), nil
}

// GenerateUUID generates a new version 4 UUID as a string.
// Used when CryptoManager instance is not available (e.g., simple helpers)
func GenerateUUID() string {
	return uuid.New().String()
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

// GenerateRecoveryCodes creates a slice of 10 unique, 8-character alphanumeric recovery codes.
//
// Returns:
//  A slice of recovery codes.
func (cm *CryptoManager) GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code, err := cm.generateSingleRecoveryCode()
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	return codes, nil
}

// generateSingleRecoveryCode generates a single 8-character alphanumeric recovery code.
func (cm *CryptoManager) generateSingleRecoveryCode() (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

// HashRecoveryCode generates a bcrypt hash of a recovery code.
func (cm *CryptoManager) HashRecoveryCode(code string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash recovery code: %w", err)
	}
	return string(bytes), nil
}
