package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/multitenant"
	"github.com/turtacn/QuantaID/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_QuotaManager_CheckUserQuota(t *testing.T) {
	// Setup SQLite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create a dummy users table
	type User struct {
		ID       int
		TenantID string
	}
	err = db.AutoMigrate(&User{})
	assert.NoError(t, err)

	// Seed data
	tenantA := "tenant-a"
	tenantB := "tenant-b"

	db.Create(&User{ID: 1, TenantID: tenantA})
	db.Create(&User{ID: 2, TenantID: tenantA})
	db.Create(&User{ID: 3, TenantID: tenantB})

	// Config
	config := map[string]utils.TenantQuotas{
		tenantA: {MaxUsers: 3}, // Allowed
		tenantB: {MaxUsers: 1}, // Exceeded if we add more? Wait, check is against current count?
		// "CheckUserQuota" usually checks if we CAN add a user, or if we ARE currently okay?
		// Implementation: `if count >= max { return error }`
		// So if count is 2 and max is 3, it returns nil (OK to add 1 more? or just state is OK?)
		// Usually "Check" is called before "Add".
		// If current count is 2, max is 3. 2 >= 3 is False. Returns nil. OK.
		// If current count is 2, max is 2. 2 >= 2 is True. Returns Error. (Cannot add more).
	}

	qm := multitenant.NewQuotaManager(db, nil, config)

	t.Run("Allowed", func(t *testing.T) {
		// Tenant A has 2 users, max 3. Should allow.
		err := qm.CheckUserQuota(context.Background(), tenantA)
		assert.NoError(t, err)
	})

	t.Run("Exceeded", func(t *testing.T) {
		// Tenant B has 1 user, max 1. Should fail (cannot add more).
		// Wait, if max is 1, and we have 1. Check says >= Max (1 >= 1). Error.
		// This implies we are at capacity.
		err := qm.CheckUserQuota(context.Background(), tenantB)
		assert.ErrorIs(t, err, multitenant.ErrQuotaExceeded)
	})
}

func Test_TenantContext_RoundTrip(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-123"

	// Set
	ctxWithTenant := multitenant.WithTenantID(ctx, tenantID)

	// Get
	got, ok := multitenant.GetTenantID(ctxWithTenant)
	assert.True(t, ok)
	assert.Equal(t, tenantID, got)

	// MustGet
	assert.NotPanics(t, func() {
		val := multitenant.MustGetTenantID(ctxWithTenant)
		assert.Equal(t, tenantID, val)
	})

	// Missing
	_, ok = multitenant.GetTenantID(ctx)
	assert.False(t, ok)

	assert.Panics(t, func() {
		multitenant.MustGetTenantID(ctx)
	})
}
