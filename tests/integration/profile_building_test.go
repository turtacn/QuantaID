//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/identity/profile"
	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Mock dependencies for integration test
type MockAccessLogRepo struct {
	Logs []models.AccessLog
}

func (m *MockAccessLogRepo) GetLogsForUser(ctx context.Context, userID string, since time.Time) ([]models.AccessLog, error) {
	return m.Logs, nil
}

type MockDeviceRepo struct{}

func (m *MockDeviceRepo) GetDevicesByUserID(ctx context.Context, userID string) ([]models.Device, error) {
	return []models.Device{}, nil
}

type MockIdentityService struct {
	User *types.User
}

func (m *MockIdentityService) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	return m.User, nil
}

// Implement other methods of identity.IService stub if needed, or use a proper mock
func (m *MockIdentityService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) { return nil, nil }
func (m *MockIdentityService) GetUserByEmail(ctx context.Context, email string) (*types.User, error) { return nil, nil }
// func (m *MockIdentityService) CreateUser(ctx context.Context, user *types.User) error { return nil } // Removed incorrect one
func (m *MockIdentityService) UpdateUser(ctx context.Context, user *types.User) error { return nil }
func (m *MockIdentityService) DeleteUser(ctx context.Context, userID string) error { return nil }
func (m *MockIdentityService) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) { return nil, 0, nil }
func (m *MockIdentityService) GetUserByExternalID(ctx context.Context, externalID, sourceType string) (*types.User, error) { return nil, nil }
func (m *MockIdentityService) AddUserToGroup(ctx context.Context, userID, groupID string) error { return nil }
func (m *MockIdentityService) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error { return nil }
// func (m *MockIdentityService) GetUserGroups(ctx context.Context, userID string) ([]types.UserGroup, error) { return nil, nil } // Removed incorrect one
func (m *MockIdentityService) ChangeUserStatus(ctx context.Context, userID string, status types.UserStatus) error { return nil }
func (m *MockIdentityService) GetUser(ctx context.Context, userID string) (*types.User, error) { return m.User, nil }
func (m *MockIdentityService) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) { return nil, nil } // Correct one
func (m *MockIdentityService) GetUserRepo() identity.UserRepository { return nil }
func (m *MockIdentityService) CreateUser(ctx context.Context, username, email, password string) (*types.User, error) { return nil, nil } // Correct one

// Group Management Stubs
func (m *MockIdentityService) CreateGroup(ctx context.Context, group *types.UserGroup) error { return nil }
func (m *MockIdentityService) GetGroup(ctx context.Context, groupID string) (*types.UserGroup, error) { return nil, nil }
func (m *MockIdentityService) UpdateGroup(ctx context.Context, group *types.UserGroup) error { return nil }
func (m *MockIdentityService) DeleteGroup(ctx context.Context, groupID string) error { return nil }
func (m *MockIdentityService) ListGroups(ctx context.Context, offset, limit int) ([]*types.UserGroup, error) { return nil, nil }

type MockMFAService struct{}

func (m *MockMFAService) HasAnyMethod(ctx context.Context, userID string) (bool, error) {
	return true, nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&profile.UserProfile{})
	require.NoError(t, err)

	return db
}

func TestProfile_BuildFromEvents(t *testing.T) {
	db := setupTestDB(t)
	repo := profile.NewPostgresProfileRepository(db)

	userID := "user_integration_1"

	// Mock Access Logs
	logs := []models.AccessLog{
		{Action: "login", Success: true, UserID: userID, Timestamp: time.Now()},
		{Action: "login", Success: true, UserID: userID, Timestamp: time.Now()},
		{Action: "login", Success: true, UserID: userID, Timestamp: time.Now()},
		{Action: "login", Success: true, UserID: userID, Timestamp: time.Now()},
		{Action: "login", Success: true, UserID: userID, Timestamp: time.Now()},
		{Action: "login", Success: true, UserID: userID, Timestamp: time.Now(), MFAVerified: true},
	}

	accessRepo := &MockAccessLogRepo{Logs: logs}
	deviceRepo := &MockDeviceRepo{}
	user := &types.User{
		ID: userID,
		Attributes: map[string]interface{}{
			"tenant_id": "tenant1",
			"email_verified": true,
		},
	}
	userRepo := &MockIdentityService{User: user}
	mfaService := &MockMFAService{}

	builder := profile.NewProfileBuilder(repo, accessRepo, deviceRepo, userRepo, mfaService)

	// Act
	prof, err := builder.BuildOrUpdate(context.Background(), userID)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, int64(6), prof.Behavior.TotalLogins)
	assert.Greater(t, prof.Behavior.MFAUsageRate, 0.0)

	// Test Persistence
	loadedProfile, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.NotNil(t, loadedProfile)
	assert.Equal(t, prof.ID, loadedProfile.ID)
	assert.Equal(t, int64(6), loadedProfile.Behavior.TotalLogins)
}

func TestProfile_RiskUpdate_OnAnomaly(t *testing.T) {
	db := setupTestDB(t)
	repo := profile.NewPostgresProfileRepository(db)
	userID := "user_risk_1"

	// Initialize profile
	initialProfile := &profile.UserProfile{
		ID:       "prof_1",
		UserID:   userID,
		TenantID: "tenant1",
	}
	repo.Create(context.Background(), initialProfile)

	scorer := profile.NewRiskScorer(utils.RiskScorerConfig{GeoJumpWeight: 20}, nil)
	service := profile.NewProfileService(repo, nil, scorer, nil, nil, nil)

	// Act: Trigger Anomaly
	anomaly := profile.AnomalyEvent{Type: "geo_jump", Timestamp: time.Now()}
	err := service.HandleAnomalyEvent(context.Background(), userID, anomaly)
	require.NoError(t, err)

	// Assert
	updatedProfile, err := repo.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 1, updatedProfile.Risk.GeoJumpCount)
	assert.Equal(t, 20, updatedProfile.RiskScore)
	assert.Equal(t, "low", updatedProfile.RiskLevel) // 20 < 25
}
