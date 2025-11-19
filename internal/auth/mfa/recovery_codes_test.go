package mfa

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"testing"
)

func TestRecoveryCodeProvider_GenerateAndStore(t *testing.T) {
	// Arrange
	repo := memory.NewMFAFactorMemoryRepository()
	crypto := utils.NewCryptoManager("test-secret")
	provider := NewRecoveryCodeProvider(repo, crypto)
	factor := &types.MFAFactor{ID: uuid.New(), UserID: uuid.New()}
	repo.CreateMFAFactor(context.Background(), factor)

	// Act
	codes, err := provider.GenerateAndStore(context.Background(), factor)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, codes, 10)
	assert.NotEmpty(t, factor.BackupCodes)
}

func TestRecoveryCodeProvider_Verify(t *testing.T) {
	// Arrange
	repo := memory.NewMFAFactorMemoryRepository()
	crypto := utils.NewCryptoManager("test-secret")
	provider := NewRecoveryCodeProvider(repo, crypto)
	userID := uuid.New()
	user := &types.User{ID: userID.String()}
	factor := &types.MFAFactor{ID: uuid.New(), UserID: userID}
	repo.CreateMFAFactor(context.Background(), factor)
	codes, err := provider.GenerateAndStore(context.Background(), factor)
	assert.NoError(t, err)

	validCode := codes[0]
	invalidCode := "invalid-code"

	// Act & Assert

	// 1. Test a valid code
	valid, err := provider.Verify(context.Background(), user, validCode)
	assert.NoError(t, err)
	assert.True(t, valid)

	// 2. Test the same code again (should be invalid)
	invalidAfterUse, err := provider.Verify(context.Background(), user, validCode)
	assert.NoError(t, err)
	assert.False(t, invalidAfterUse)

	// 3. Test an invalid code
	invalidAttempt, err := provider.Verify(context.Background(), user, invalidCode)
	assert.NoError(t, err)
	assert.False(t, invalidAttempt)
}
