package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/pkg/types"
)

func TestAuthMemoryRepository_SessionManagement(t *testing.T) {
	repo := memory.NewAuthMemoryRepository()
	ctx := context.Background()
	session := &types.UserSession{
		ID:     "session123",
		UserID: "user123",
	}

	// Create
	err := repo.CreateSession(ctx, session, time.Hour)
	assert.NoError(t, err)

	// Get
	retrievedSession, err := repo.GetSession(ctx, "session123")
	assert.NoError(t, err)
	assert.Equal(t, "user123", retrievedSession.UserID)

	// Get User Sessions
	userSessions, err := repo.GetUserSessions(ctx, "user123")
	assert.NoError(t, err)
	assert.Len(t, userSessions, 1)
	assert.Equal(t, "session123", userSessions[0].ID)

	// Delete
	err = repo.DeleteSession(ctx, "session123")
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetSession(ctx, "session123")
	assert.Error(t, err)
}

func TestAuthMemoryRepository_TokenManagement(t *testing.T) {
	repo := memory.NewAuthMemoryRepository()
	ctx := context.Background()
	refreshToken := "refresh123"
	userID := "user123"
	jti := "jwt_id_123"

	// Store Refresh Token
	err := repo.StoreRefreshToken(ctx, refreshToken, userID, time.Hour)
	assert.NoError(t, err)

	// Get Refresh Token User ID
	retrievedUserID, err := repo.GetRefreshTokenUserID(ctx, refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, retrievedUserID)

	// Delete Refresh Token
	err = repo.DeleteRefreshToken(ctx, refreshToken)
	assert.NoError(t, err)
	_, err = repo.GetRefreshTokenUserID(ctx, refreshToken)
	assert.Error(t, err)

	// Add to Deny List
	err = repo.AddToDenyList(ctx, jti, time.Minute)
	assert.NoError(t, err)

	// Check Deny List
	isDenied, err := repo.IsInDenyList(ctx, jti)
	assert.NoError(t, err)
	assert.True(t, isDenied)

	// Check non-existent JTI
	isDenied, err = repo.IsInDenyList(ctx, "other_jti")
	assert.NoError(t, err)
	assert.False(t, isDenied)
}

func TestAuthMemoryRepository_DenyListExpiration(t *testing.T) {
	repo := memory.NewAuthMemoryRepository()
	ctx := context.Background()
	jti := "jwt_id_exp"

	// Add with short duration
	err := repo.AddToDenyList(ctx, jti, 1*time.Millisecond)
	assert.NoError(t, err)

	// Wait for it to expire
	time.Sleep(5 * time.Millisecond)

	isDenied, err := repo.IsInDenyList(ctx, jti)
	assert.NoError(t, err)
	assert.False(t, isDenied, "JTI should have expired from the deny list")
}
