package ldap

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/identity"
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

func TestSyncEngine_FullSync(t *testing.T) {
	mockLDAP := new(MockLDAPClient)
	mockIdentityRepo := new(identity.MockIdentityRepository)
	mockSyncStateRepo := new(identity.MockSyncStateRepository)
	mockMetrics := metrics.NewSyncMetrics("test")
	logger := zap.NewNop()

	schemaMapper := NewSchemaMapper(SchemaMapConfig{})
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

	mockLDAP.On("SearchPaged", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*types.Entry{}, "", nil)
	mockIdentityRepo.On("UpsertBatch", mock.Anything, mock.Anything).Return(nil)
	mockSyncStateRepo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockSyncStateRepo.On("MarkCompleted", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := engine.StartFullSync(context.Background(), "test", "dc=example,dc=com", "(objectClass=*)")
	assert.NoError(t, err)

	mockLDAP.AssertExpectations(t)
	mockIdentityRepo.AssertExpectations(t)
	mockSyncStateRepo.AssertExpectations(t)
}
