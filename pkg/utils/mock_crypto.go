package utils

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/mock"
	"time"
)

// MockCryptoManager is a mock implementation of the CryptoManagerInterface for testing.
type MockCryptoManager struct {
	mock.Mock
}

func (m *MockCryptoManager) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockCryptoManager) CheckPasswordHash(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}

func (m *MockCryptoManager) GenerateJWT(userID string, duration time.Duration, claims jwt.MapClaims) (string, error) {
	args := m.Called(userID, duration, claims)
	return args.String(0), args.Error(1)
}

func (m *MockCryptoManager) ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	args := m.Called(tokenString)
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

func (m *MockCryptoManager) GenerateUUID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCryptoManager) GenerateRecoveryCodes() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCryptoManager) HashRecoveryCode(code string) (string, error) {
	args := m.Called(code)
	return args.String(0), args.Error(1)
}
