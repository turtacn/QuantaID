package ldap

import (
	"context"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// MockLDAPClient is a mock implementation of LDAPClientInterface
type MockLDAPClient struct {
	mock.Mock
}

func (m *MockLDAPClient) SearchPaged(ctx context.Context, baseDN, filter string, pageSize uint32, cookie string) ([]*ldap.Entry, string, error) {
	args := m.Called(ctx, baseDN, filter, pageSize, cookie)
	return args.Get(0).([]*ldap.Entry), args.String(1), args.Error(2)
}

func (m *MockLDAPClient) PersistentSearch(ctx context.Context, baseDN, filter string, attributes []string) (*ldap.SearchResult, error) {
	args := m.Called(ctx, baseDN, filter, attributes)
	return args.Get(0).(*ldap.SearchResult), args.Error(1)
}

func (m *MockLDAPClient) Close() {
	m.Called()
}

func (m *MockLDAPClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLDAPClient) Search(ctx context.Context, baseDN string, filter string, attributes []string) ([]*ldap.Entry, error) {
	args := m.Called(ctx, baseDN, filter, attributes)
	return args.Get(0).([]*ldap.Entry), args.Error(1)
}

// MockIdentityRepo is a partial mock for IdentityRepository
type MockIdentityRepo struct {
	identity.IdentityRepository // Embed the interface to satisfy remaining methods
	mock.Mock
}

func (m *MockIdentityRepo) UpsertBatch(ctx context.Context, users []*types.User) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

func (m *MockIdentityRepo) GetUserByExternalID(ctx context.Context, externalID, sourceID string) (*types.User, error) {
	args := m.Called(ctx, externalID, sourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIdentityRepo) FindUsersBySource(ctx context.Context, sourceID string) ([]*types.User, error) {
	args := m.Called(ctx, sourceID)
	return args.Get(0).([]*types.User), args.Error(1)
}

// MockSyncStateRepo mock
type MockSyncStateRepo struct {
	identity.SyncStateRepository
	mock.Mock
}

func (m *MockSyncStateRepo) MarkCompleted(ctx context.Context, sourceID string, completedAt time.Time) error {
	args := m.Called(ctx, sourceID, completedAt)
	return args.Error(0)
}

func TestSyncEngine_StartFullSync_Batching(t *testing.T) {
	// Setup
	mockLDAP := new(MockLDAPClient)
	mockRepo := new(MockIdentityRepo)
	mockStateRepo := new(MockSyncStateRepo)
	logger := zap.NewNop()
	metrics := &metrics.SyncMetrics{} // Should be mocked properly if used heavily
	schemaMapper := NewSchemaMapper(SchemaMapConfig{}) // Mock mapping config if needed

	// Config
	config := SyncConfig{
		BatchSize:        10,
		ConcurrencyLimit: 1,
		ConflictStrategy: "RemoteWins",
	}

	engine := NewSyncEngine(mockLDAP, mockRepo, schemaMapper, nil, mockStateRepo, metrics, config, logger)

	// Mock LDAP Data - 25 users
	var entries []*ldap.Entry
	for i := 0; i < 25; i++ {
		entries = append(entries, &ldap.Entry{
			DN: "uid=user" + string(rune(i)),
			Attributes: []*ldap.EntryAttribute{
				{Name: "uid", Values: []string{"user" + string(rune(i))}},
				{Name: "mail", Values: []string{"user" + string(rune(i)) + "@example.com"}},
			},
		})
	}

	// Paging Logic Mock
	// Page 1: 0-9 (10 items), cookie="cookie1"
	mockLDAP.On("SearchPaged", mock.Anything, "dc=example,dc=com", "(objectClass=person)", uint32(10), "").Return(entries[0:10], "cookie1", nil)
	// Page 2: 10-19 (10 items), cookie="cookie2"
	mockLDAP.On("SearchPaged", mock.Anything, "dc=example,dc=com", "(objectClass=person)", uint32(10), "cookie1").Return(entries[10:20], "cookie2", nil)
	// Page 3: 20-24 (5 items), cookie=""
	mockLDAP.On("SearchPaged", mock.Anything, "dc=example,dc=com", "(objectClass=person)", uint32(10), "cookie2").Return(entries[20:25], "", nil)

	// Expect UpsertBatch calls
	// Batch 1: 10 users
	mockRepo.On("UpsertBatch", mock.Anything, mock.MatchedBy(func(users []*types.User) bool {
		return len(users) == 10
	})).Return(nil).Once()
	// Batch 2: 10 users
	mockRepo.On("UpsertBatch", mock.Anything, mock.MatchedBy(func(users []*types.User) bool {
		return len(users) == 10
	})).Return(nil).Once()
	// Batch 3: 5 users
	mockRepo.On("UpsertBatch", mock.Anything, mock.MatchedBy(func(users []*types.User) bool {
		return len(users) == 5
	})).Return(nil).Once()

	mockStateRepo.On("MarkCompleted", mock.Anything, "ldap-1", mock.Anything).Return(nil)

	// Execute
	err := engine.StartFullSync(context.Background(), "ldap-1", "dc=example,dc=com", "(objectClass=person)")

	// Verify
	assert.NoError(t, err)
	mockLDAP.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockStateRepo.AssertExpectations(t)
}

func TestSyncEngine_ConflictResolution_Merge(t *testing.T) {
	// Setup
	mockLDAP := new(MockLDAPClient)
	mockRepo := new(MockIdentityRepo)
	mockStateRepo := new(MockSyncStateRepo)
	logger := zap.NewNop()
	metrics := &metrics.SyncMetrics{}
	schemaMapper := NewSchemaMapper(SchemaMapConfig{
		Mappings: []AttributeMapping{
			{LDAPAttr: "uid", QuantaField: "username", Required: true},
			{LDAPAttr: "mail", QuantaField: "email", Required: true},
			{LDAPAttr: "mobile", QuantaField: "phone"},
		},
	})

	config := SyncConfig{
		BatchSize:        1, // Process one by one to test logic easier
		ConflictStrategy: "Merge",
	}

	engine := NewSyncEngine(mockLDAP, mockRepo, schemaMapper, nil, mockStateRepo, metrics, config, logger)

	// Remote User
	remoteEntry := &ldap.Entry{
		DN: "uid=alice,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{Name: "uid", Values: []string{"alice"}},
			{Name: "mail", Values: []string{"alice@new.com"}}, // Changed
			{Name: "mobile", Values: []string{"123456"}}, // New
		},
	}

	// Local User
	localUser := &types.User{
		ID:         "local-id-1",
		Username:   "alice",
		Email:      "alice@old.com", // Should NOT be overwritten if Merge strategy protects non-empty?
		// Actually Merge strategy implementation: "if localUser.Email == "" { ... }"
		// So it preserves local email.
		Phone:      "", // Should be updated
		ExternalID: "uid=alice,dc=example,dc=com", // Assume this is the key
	}

	// Mock LDAP
	mockLDAP.On("SearchPaged", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]*ldap.Entry{remoteEntry}, "", nil)

	// Mock Repo GetUserByExternalID
	mockRepo.On("GetUserByExternalID", mock.Anything, mock.Anything, "ldap-1").Return(localUser, nil)

	mockRepo.On("UpsertBatch", mock.Anything, mock.MatchedBy(func(users []*types.User) bool {
		if len(users) != 1 {
			return false
		}
		u := users[0]
		// Verify Merge Logic
		// Email should be preserved (local wins for non-empty)
		if u.Email != "alice@old.com" {
			t.Logf("Mismatch Email: expected alice@old.com, got %s", u.Email)
			return false
		}
		// Phone should be updated (remote wins for empty)
		if u.Phone != "123456" {
			t.Logf("Mismatch Phone: expected 123456, got %s", u.Phone)
			return false
		}
		return true
	})).Return(nil)

	mockStateRepo.On("MarkCompleted", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := engine.StartFullSync(context.Background(), "ldap-1", "dc=example", "filter")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
