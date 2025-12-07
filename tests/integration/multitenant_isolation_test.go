//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/turtacn/QuantaID/internal/multitenant"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Test_RLS_Isolation(t *testing.T) {
	ctx := context.Background()

	// 1. Start Postgres Container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// 2. Connect
	db, err := gorm.Open(gorm_postgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	// 3. Setup Schema & RLS
	// We use a simplified User model for testing
	type User struct {
		ID       int    `gorm:"primaryKey"`
		Username string
		TenantID string
	}
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	isolator := multitenant.NewTenantIsolator()
	err = isolator.EnableRowLevelSecurity(db)
	require.NoError(t, err)

	// Create a non-superuser role to test RLS
	// Postgres superusers bypass RLS, so we need a restricted user
	err = db.Exec("CREATE ROLE app_user WITH LOGIN PASSWORD 'password' NOSUPERUSER").Error
	require.NoError(t, err)

	err = db.Exec("GRANT ALL ON SCHEMA public TO app_user").Error
	require.NoError(t, err)

	err = db.Exec("GRANT ALL ON ALL TABLES IN SCHEMA public TO app_user").Error
	require.NoError(t, err)

	err = db.Exec("GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO app_user").Error
	require.NoError(t, err)

	// 4. Seed Data (admin/system level bypass for seeding?)
	// RLS is enforced on the session user. By default we are superuser (postgres).
	// Superuser bypasses RLS. So we can seed directly.

	tenantA := "tenant-a"
	tenantB := "tenant-b"

	db.Create(&User{ID: 1, Username: "user-a", TenantID: tenantA})
	db.Create(&User{ID: 2, Username: "user-b", TenantID: tenantB})

	// Helper to run query as a specific tenant role
	// We switch role inside transaction to enforce RLS

	t.Run("TenantA_CannotSeeTenantB_Users", func(t *testing.T) {
		// New session/transaction to set variable
		tx := db.Begin()
		defer tx.Rollback()

		// Switch to restricted user
		err := tx.Exec("SET ROLE app_user").Error
		require.NoError(t, err)

		err = isolator.SetTenantContext(tx, tenantA)
		require.NoError(t, err)

		var users []User
		err = tx.Find(&users).Error
		require.NoError(t, err)

		// Should only see user-a
		assert.Len(t, users, 1)
		if len(users) > 0 {
			assert.Equal(t, "user-a", users[0].Username)
		}
	})

	t.Run("TenantB_CannotSeeTenantA_Users", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()

		err := tx.Exec("SET ROLE app_user").Error
		require.NoError(t, err)

		err = isolator.SetTenantContext(tx, tenantB)
		require.NoError(t, err)

		var users []User
		err = tx.Find(&users).Error
		require.NoError(t, err)

		// Should only see user-b
		assert.Len(t, users, 1)
		if len(users) > 0 {
			assert.Equal(t, "user-b", users[0].Username)
		}
	})

	t.Run("CrossTenant_Write_Blocked", func(t *testing.T) {
		// Try to update user-b while acting as tenant-a
		tx := db.Begin()
		defer tx.Rollback()

		err := tx.Exec("SET ROLE app_user").Error
		require.NoError(t, err)

		err = isolator.SetTenantContext(tx, tenantA)
		require.NoError(t, err)

		// Try to update user-b (ID 2)
		result := tx.Model(&User{}).Where("id = ?", 2).Update("username", "hacked")
		require.NoError(t, result.Error) // DB often returns success with 0 rows affected for RLS blocks

		// Verify rows affected is 0
		assert.Equal(t, int64(0), result.RowsAffected)
	})
}
