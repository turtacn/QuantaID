package ldap

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/go-ldap/ldap/v3"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
	"testing"
)

type MockLDAPClient struct {
	mock.Mock
}

func (m *MockLDAPClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLDAPClient) Search(ctx context.Context, baseDN string, filter string, attributes []string) ([]*types.Entry, error) {
	args := m.Called(ctx, baseDN, filter, attributes)
	return args.Get(0).([]*types.Entry), args.Error(1)
}

func (m *MockLDAPClient) SearchPaged(ctx context.Context, baseDN string, filter string, pageSize uint32, cookie string) ([]*types.Entry, string, error) {
	args := m.Called(ctx, baseDN, filter, pageSize, cookie)
	return args.Get(0).([]*types.Entry), args.String(1), args.Error(2)
}

func (m *MockLDAPClient) PersistentSearch(ctx context.Context, baseDN, filter string, attributes []string) (*types.SearchResult, error) {
	args := m.Called(ctx, baseDN, filter, attributes)
	return args.Get(0).(*types.SearchResult), args.Error(1)
}

func (m *MockLDAPClient) Close() {
	m.Called()
}

func TestSyncEngine_FullSync_CompareAndSwap(t *testing.T) {
	// Arrange
	mockLDAP := new(MockLDAPClient)
	mockIdentityRepo := new(identity.MockIdentityRepository)
	mockSyncStateRepo := new(identity.MockSyncStateRepository)
	mockMetrics := metrics.NewSyncMetrics("test")
	logger, _ := zap.NewDevelopment()

	schemaMapper := NewSchemaMapper(SchemaMapConfig{
		Mappings: []AttributeMapping{
			{LDAPAttr: "uid", QuantaField: "username"},
			{LDAPAttr: "mail", QuantaField: "email"},
			{LDAPAttr: "userAccountControl", QuantaField: "userAccountControl"},
		},
	})
	deduplicator := NewDeduplicator([]DeduplicationRule{}, &ConflictManager{})

	engine := NewSyncEngine(
		mockLDAP,
		mockIdentityRepo,
		schemaMapper,
		deduplicator,
		mockSyncStateRepo,
		mockMetrics,
		SyncConfig{BatchSize: 10},
		logger,
	)

	localUsers := []*types.User{
		{ID: "1", Username: "alice", Email: "alice@example.com", Status: types.UserStatusInactive}, // Status changed
		{ID: "3", Username: "charlie", Email: "charlie@example.com", Status: types.UserStatusActive}, // To be deleted
	}

	mockLDAP.On("SearchPaged", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*ldap.Entry{
		{Attributes: []*ldap.EntryAttribute{{Name: "uid", Values: []string{"alice"}}, {Name: "mail", Values: []string{"alice@example.com"}}}},
		{Attributes: []*ldap.EntryAttribute{{Name: "uid", Values: []string{"bob"}}, {Name: "mail", Values: []string{"bob@example.com"}}}},
	}, "", nil)
	mockIdentityRepo.On("FindUsersBySource", mock.Anything, "test").Return(localUsers, nil)
	mockIdentityRepo.On("CreateBatch", mock.Anything, mock.Anything).Return(nil)
	mockIdentityRepo.On("UpdateBatch", mock.Anything, mock.MatchedBy(func(users []*types.User) bool {
		return len(users) == 1 && users[0].Username == "alice"
	})).Return(nil)
	mockIdentityRepo.On("DeleteBatch", mock.Anything, mock.Anything).Return(nil)
	mockSyncStateRepo.On("MarkCompleted", mock.Anything, "test", mock.Anything).Return(nil)

	// Act
	err := engine.StartFullSync(context.Background(), "test", "dc=example,dc=com", "(objectClass=*)")

	// Assert
	assert.NoError(t, err)
	mockLDAP.AssertExpectations(t)
	mockIdentityRepo.AssertExpectations(t)
	mockSyncStateRepo.AssertExpectations(t)
}
