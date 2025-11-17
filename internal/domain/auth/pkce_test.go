package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPKCERepository is a mock implementation of PKCERepository for testing.
type MockPKCERepository struct {
	mock.Mock
}

func (m *MockPKCERepository) GetAuthCodeChallenge(ctx context.Context, authCode string) (*PKCEChallenge, error) {
	args := m.Called(ctx, authCode)
	challenge, _ := args.Get(0).(*PKCEChallenge)
	return challenge, args.Error(1)
}

func TestPKCEValidator_S256Challenge(t *testing.T) {
	repo := new(MockPKCERepository)
	config := PKCEConfig{MinVerifierLength: 43, MaxVerifierLength: 128}
	validator := NewPKCEValidator(config, repo)

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	repo.On("GetAuthCodeChallenge", mock.Anything, "auth_code").Return(&PKCEChallenge{
		Challenge:       challenge,
		ChallengeMethod: "S256",
	}, nil)

	err := validator.VerifyCodeVerifier(context.Background(), "auth_code", verifier)
	assert.NoError(t, err)
}

func TestPKCEValidator_MismatchVerifier(t *testing.T) {
	repo := new(MockPKCERepository)
	config := PKCEConfig{MinVerifierLength: 43, MaxVerifierLength: 128}
	validator := NewPKCEValidator(config, repo)

	// This verifier has a valid length but will not match the challenge.
	verifier := "aBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	repo.On("GetAuthCodeChallenge", mock.Anything, "auth_code").Return(&PKCEChallenge{
		Challenge:       challenge,
		ChallengeMethod: "S256",
	}, nil)

	err := validator.VerifyCodeVerifier(context.Background(), "auth_code", verifier)
	assert.EqualError(t, err, ErrInvalidCodeVerifier.Error())
}
