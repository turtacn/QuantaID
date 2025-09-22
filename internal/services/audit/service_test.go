package audit

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"sync"
	"testing"
	"time"
)

// --- Mock Repository ---

type MockAuditRepository struct{ mock.Mock }
func (m *MockAuditRepository) CreateLogEntry(ctx context.Context, entry *types.AuditLog) error {
	return m.Called(ctx, entry).Error(0)
}
func (m *MockAuditRepository) GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error) { return nil, nil }
func (m *MockAuditRepository) GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error) { return nil, nil }

// --- Tests ---

func TestAuditApplicationService_RecordEvent(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "error"})
	appSvc := NewApplicationService(mockRepo, logger)
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)

	mockRepo.On("CreateLogEntry", mock.Anything, mock.MatchedBy(func(log *types.AuditLog) bool {
		defer wg.Done()
		return log.ActorID == "user1" && log.Action == ActionLoginSuccess
	})).Return(nil).Once()

	appSvc.RecordEvent(ctx, "user1", ActionLoginSuccess, "session", StatusSuccess, nil)

	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		mockRepo.AssertExpectations(t)
	case <-time.After(1 * time.Second):
		t.Fatal("Test timed out waiting for mock to be called")
	}
}

//Personal.AI order the ending
