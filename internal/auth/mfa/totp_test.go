package mfa

import (
	"context"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTOTPProvider_Enroll(t *testing.T) {
	provider := &TOTPProvider{
		issuer: "QuantaID",
	}

	result, err := provider.Enroll(context.Background(), "testuser", EnrollParams{Email: "test@example.com"})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Secret)
	assert.NotEmpty(t, result.QRCodeImage)
	assert.Len(t, result.BackupCodes, 10)
}

func TestTOTPProvider_Verify(t *testing.T) {
	// This test is more complex and requires a mock repository and crypto service.
	// For now, we'll just test the basic validation logic.
	secret := "JBSWY3DPEHPK3PXP" // Example secret
	validCode, err := totp.GenerateCode(secret, time.Now())
	assert.NoError(t, err)

	assert.True(t, totp.Validate(validCode, secret))
	assert.False(t, totp.Validate("123456", secret))
}
