//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/auth/device"
	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
	"github.com/turtacn/QuantaID/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Using SQLite for easier integration testing setup without full Docker overhead if possible,
// or use the project's standard DB setup.
// Given instructions mention "PostgresDeviceRepository", it expects GORM.
// I'll use SQLite memory mode to simulate DB for this integration test
// to ensure logic works with a DB, as `testcontainers` might be heavy/slow and user mentioned "tests/testutils" or similar.
// But user requirements said "Test_Device_Register_And_Retrieve: 注册设备后可正确查询".

func setupDeviceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.Device{})
	require.NoError(t, err)

	return db
}

func Test_Device_Register_And_Retrieve(t *testing.T) {
	db := setupDeviceTestDB(t)
	repo := device.NewPostgresDeviceRepository(db)
	fp := device.NewDeviceFingerprinter()
	scorer := device.NewTrustScorer(utils.DeviceTrustConfig{BaseScore: 20})
	// No geo service for this test
	service := device.NewDeviceService(repo, scorer, nil, fp)

	ctx := context.Background()
	fpData := map[string]interface{}{
		"user_agent": "Mozilla/5.0 ...",
		"platform":   "Linux",
	}

	// Register
	dev, err := service.RegisterOrUpdate(ctx, fpData, "127.0.0.1", "tenant-1")
	require.NoError(t, err)
	assert.NotEmpty(t, dev.ID)
	assert.Equal(t, 20, dev.TrustScore)

	// Retrieve
	found, err := repo.GetByID(ctx, dev.ID)
	require.NoError(t, err)
	assert.Equal(t, dev.ID, found.ID)
	assert.Equal(t, dev.Fingerprint, found.Fingerprint)
}

func Test_Device_BindUser_UpdatesTrust(t *testing.T) {
	db := setupDeviceTestDB(t)
	repo := device.NewPostgresDeviceRepository(db)
	fp := device.NewDeviceFingerprinter()
	scorer := device.NewTrustScorer(utils.DeviceTrustConfig{BaseScore: 20, BoundBonus: 10})
	service := device.NewDeviceService(repo, scorer, nil, fp)

	ctx := context.Background()
	fpData := map[string]interface{}{"ua": "test"}

	// Create device
	dev, err := service.RegisterOrUpdate(ctx, fpData, "1.2.3.4", "tenant-1")
	require.NoError(t, err)
	assert.Equal(t, 20, dev.TrustScore)

	// Bind User
	err = service.BindToUser(ctx, dev.ID, "user-123")
	require.NoError(t, err)

	// Check updated score
	updated, err := repo.GetByID(ctx, dev.ID)
	require.NoError(t, err)
	assert.Equal(t, "user-123", updated.UserID)
	assert.Equal(t, 30, updated.TrustScore) // 20 + 10
}

func Test_Device_MultipleUsers_Isolation(t *testing.T) {
	db := setupDeviceTestDB(t)
	repo := device.NewPostgresDeviceRepository(db)

	ctx := context.Background()

	dev1 := &models.Device{ID: "d1", UserID: "u1", Fingerprint: "f1"}
	dev2 := &models.Device{ID: "d2", UserID: "u2", Fingerprint: "f2"}

	require.NoError(t, repo.Create(ctx, dev1))
	require.NoError(t, repo.Create(ctx, dev2))

	u1Devices, err := repo.GetByUserID(ctx, "u1")
	require.NoError(t, err)
	assert.Len(t, u1Devices, 1)
	assert.Equal(t, "d1", u1Devices[0].ID)

	u2Devices, err := repo.GetByUserID(ctx, "u2")
	require.NoError(t, err)
	assert.Len(t, u2Devices, 1)
	assert.Equal(t, "d2", u2Devices[0].ID)
}
