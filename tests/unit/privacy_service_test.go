package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/privacy"
	privacy_service "github.com/turtacn/QuantaID/internal/services/privacy"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// --- Helpers ---

// setupMockDB creates a dummy GORM connection for testing transaction logic.
// In a real unit test, you might use go-sqlmock to mock the DB driver,
// but for this example, we verify the service flow around the DB.
func setupMockDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&types.User{})
	return db
}

type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) RevokeAllUserSessions(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) RecordAdminAction(ctx context.Context, userID, ip, resource, action, traceID string, details map[string]interface{}) {
	m.Called(ctx, userID, ip, resource, action, traceID, details)
}

type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	args := m.Called(ctx, userID, pq)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) CreateLogEntry(ctx context.Context, entry *types.AuditLog) error {
	return nil
}

func (m *MockAuditLogRepository) GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	return nil, nil
}

// MockPrivacyRepository is a mock implementation of the privacy.Repository interface.
type MockPrivacyRepository struct {
	mock.Mock
}

func (m *MockPrivacyRepository) CreateConsentRecord(ctx context.Context, record *privacy.ConsentRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockPrivacyRepository) GetLastConsentRecord(ctx context.Context, userID, policyID string) (*privacy.ConsentRecord, error) {
	args := m.Called(ctx, userID, policyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*privacy.ConsentRecord), args.Error(1)
}

func (m *MockPrivacyRepository) CreateDSRRequest(ctx context.Context, request *privacy.DSRRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockPrivacyRepository) GetDSRRequest(ctx context.Context, requestID string) (*privacy.DSRRequest, error) {
	args := m.Called(ctx, requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*privacy.DSRRequest), args.Error(1)
}

func (m *MockPrivacyRepository) UpdateDSRRequestStatus(ctx context.Context, requestID string, status privacy.DSRRequestStatus) error {
	args := m.Called(ctx, requestID, status)
	return args.Error(0)
}

func (m *MockPrivacyRepository) GetConsentHistory(ctx context.Context, userID string) ([]*privacy.ConsentRecord, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*privacy.ConsentRecord), args.Error(1)
}

// MockIdentityRepository is a mock implementation of the identity.UserRepository interface.
type MockIdentityRepository struct {
	mock.Mock
}

func (m *MockIdentityRepository) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIdentityRepository) UpdateUser(ctx context.Context, user *types.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockIdentityRepository) CreateUser(ctx context.Context, user *types.User) error {
	return nil
}
func (m *MockIdentityRepository) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	return nil, nil
}
func (m *MockIdentityRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	return nil, nil
}
func (m *MockIdentityRepository) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockIdentityRepository) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) {
	return nil, 0, nil
}
func (m *MockIdentityRepository) ChangeUserStatus(ctx context.Context, userID string, newStatus types.UserStatus) error {
	return nil
}
func (m *MockIdentityRepository) FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error) {
	return nil, nil
}
func (m *MockIdentityRepository) GetUserByExternalID(ctx context.Context, externalID, sourceID string) (*types.User, error) {
	return nil, nil
}
func (m *MockIdentityRepository) UpsertBatch(ctx context.Context, users []*types.User) error {
	return nil
}
func (m *MockIdentityRepository) CreateBatch(ctx context.Context, users []*types.User) error {
	return nil
}
func (m *MockIdentityRepository) UpdateBatch(ctx context.Context, users []*types.User) error {
	return nil
}
func (m *MockIdentityRepository) DeleteBatch(ctx context.Context, userIDs []string) error {
	return nil
}
func (m *MockIdentityRepository) FindUsersBySource(ctx context.Context, sourceID string) ([]*types.User, error) {
	return nil, nil
}

func Test_Consent_Versioning(t *testing.T) {
	privacyRepo := new(MockPrivacyRepository)
	identityRepo := new(MockIdentityRepository)
	config := &utils.Config{
		Privacy: utils.PrivacyConfig{
			PolicyVersions: map[string]string{
				"terms_of_service": "1.1",
			},
		},
	}
	service := privacy_service.NewService(nil, nil, nil, privacyRepo, identityRepo, nil, config)

	userID := "user-123"
	policyID := "terms_of_service"

	// Mock the repository to return a consent record with version 1.0
	consentRecord := &privacy.ConsentRecord{
		PolicyVersion: "1.0",
	}
	privacyRepo.On("GetLastConsentRecord", mock.Anything, userID, policyID).Return(consentRecord, nil)

	// Check if the user has consented to the latest version (1.1)
	hasConsented, err := service.HasConsentedLatest(context.Background(), userID, policyID)

	assert.NoError(t, err)
	assert.False(t, hasConsented, "User should not have consented to the latest version")

	privacyRepo.AssertExpectations(t)
}

func Test_Erasure_Anonymization(t *testing.T) {
	privacyRepo := new(MockPrivacyRepository)
	identityRepo := new(MockIdentityRepository)
	db := setupMockDB()
	sessionManager := new(MockSessionManager)
	auditService := new(MockAuditService)

	service := privacy_service.NewService(db, sessionManager, auditService, privacyRepo, identityRepo, nil, nil)

	userID := "user-123"
	user := &types.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	identityRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	identityRepo.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *types.User) bool {
		return u.Status == "deleted"
	})).Return(nil)
	sessionManager.On("RevokeAllUserSessions", mock.Anything, userID).Return(nil)
	auditService.On("RecordAdminAction", mock.Anything, user.ID, "", "user", "account.erased", "", mock.Anything).Return(nil)

	err := service.EraseAccount(context.Background(), userID)

	assert.NoError(t, err)

	identityRepo.AssertExpectations(t)
	sessionManager.AssertExpectations(t)
	auditService.AssertExpectations(t)
}

func Test_Grant_Consent(t *testing.T) {
	privacyRepo := new(MockPrivacyRepository)
	identityRepo := new(MockIdentityRepository)
	service := privacy_service.NewService(nil, nil, nil, privacyRepo, identityRepo, nil, nil)

	userID := "user-123"
	req := privacy_service.GrantConsentRequest{
		UserID:        userID,
		PolicyID:      "terms_of_service",
		PolicyVersion: "1.1",
	}
	user := &types.User{
		ID: userID,
	}

	privacyRepo.On("CreateConsentRecord", mock.Anything, mock.Anything).Return(nil)
	identityRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	identityRepo.On("UpdateUser", mock.Anything, mock.Anything).Return(nil)

	err := service.GrantConsent(context.Background(), req)

	assert.NoError(t, err)

	privacyRepo.AssertExpectations(t)
	identityRepo.AssertExpectations(t)
}

func Test_Collect_User_Data(t *testing.T) {
	privacyRepo := new(MockPrivacyRepository)
	identityRepo := new(MockIdentityRepository)
	auditRepo := new(MockAuditLogRepository)
	service := privacy_service.NewService(nil, nil, nil, privacyRepo, identityRepo, auditRepo, nil)

	userID := "user-123"
	user := &types.User{
		ID: userID,
	}
	auditLogs := []*types.AuditLog{
		{ID: "log-1"},
	}
	consentHistory := []*privacy.ConsentRecord{
		{ID: "consent-1"},
	}

	identityRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	auditRepo.On("GetLogsForUser", mock.Anything, userID, mock.Anything).Return(auditLogs, nil)
	privacyRepo.On("GetConsentHistory", mock.Anything, userID).Return(consentHistory, nil)

	data, err := service.CollectUserData(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, user.ID, data.User.ID)
	assert.Len(t, data.AuditHistory, 1)
	assert.Len(t, data.ConsentHistory, 1)

	identityRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
	privacyRepo.AssertExpectations(t)
}
