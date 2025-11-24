package audit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/pkg/audit/events"
	"go.uber.org/zap"
)


// --- Mocks for Compliance and Retention Tests ---
type mockRetentionRepository struct {
	archivedLogs map[string][]string // Maps tier to log IDs
	deletedLogs  []string
}
func (m *mockRetentionRepository) ArchiveLogs(ctx context.Context, cutoff time.Time, targetTier string) (int64, error) {
	// Simulate archiving 5 logs
	count := int64(5)
	for i := 0; i < int(count); i++ {
		m.archivedLogs[targetTier] = append(m.archivedLogs[targetTier], fmt.Sprintf("log-%d", i))
	}
	return count, nil
}
func (m *mockRetentionRepository) DeleteLogsBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	// Simulate deleting 10 logs
	count := int64(10)
	for i := 0; i < int(count); i++ {
		m.deletedLogs = append(m.deletedLogs, fmt.Sprintf("log-%d", i))
	}
	return count, nil
}

type mockIdentityRepository struct {
	expiredAccounts []UserAccount
}
func (m *mockIdentityRepository) FindAccountsCreatedBefore(ctx context.Context, cutoff time.Time) ([]UserAccount, error) {
	if len(m.expiredAccounts) == 0 {
		return nil, ErrNotFound
	}
	return m.expiredAccounts, nil
}

// --- Compliance Checker Tests ---
func TestCheckGDPRDataRetention_Pass(t *testing.T) {
	mockIdentityRepo := &mockIdentityRepository{expiredAccounts: []UserAccount{}}
	checker := NewComplianceChecker(nil, nil, mockIdentityRepo, zap.NewNop())

	result, err := CheckGDPRDataRetention(context.Background(), checker)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestCheckGDPRDataRetention_Fail(t *testing.T) {
	mockIdentityRepo := &mockIdentityRepository{
		expiredAccounts: []UserAccount{
			{ID: "user-1", CreatedAt: time.Now().Add(-8 * 365 * 24 * time.Hour)},
		},
	}
	checker := NewComplianceChecker(nil, nil, mockIdentityRepo, zap.NewNop())

	result, err := CheckGDPRDataRetention(context.Background(), checker)
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
	assert.Contains(t, result.Details, "Found 1 user accounts")
}

func TestCheckSOC2MonitoringCoverage_Pass(t *testing.T) {
	mockAuditRepo := &mockAuditRepository{}
	mockAuditRepo.WriteSync(context.Background(), &events.AuditEvent{
		EventType: events.EventConfigChanged,
		Timestamp: time.Now().UTC(),
	})

	checker := NewComplianceChecker(nil, mockAuditRepo, nil, zap.NewNop())

	result, err := CheckSOC2MonitoringCoverage(context.Background(), checker)
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestCheckSOC2MonitoringCoverage_Fail(t *testing.T) {
	mockAuditRepo := &mockAuditRepository{} // No events in the repo
	checker := NewComplianceChecker(nil, mockAuditRepo, nil, zap.NewNop())

	result, err := CheckSOC2MonitoringCoverage(context.Background(), checker)
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
	assert.Contains(t, result.Details, "No critical system events")
}

// --- Retention Policy Tests ---

func TestRetentionPolicy_Execute(t *testing.T) {
	repo := &mockRetentionRepository{archivedLogs: make(map[string][]string)}
	config := RetentionConfig{
		HotDataRetention:  30 * 24 * time.Hour,
		ColdDataRetention: 365 * 24 * time.Hour,
		EnableAutoArchive: true,
		EnableAutoDelete:  true,
	}

	manager := NewRetentionPolicyManager(repo, config, zap.NewNop())
	err := manager.Execute(context.Background())
	require.NoError(t, err)

	assert.Len(t, repo.archivedLogs["warm"], 5)
	assert.Len(t, repo.deletedLogs, 10)
}

func TestRetentionPolicy_Disabled(t *testing.T) {
	repo := &mockRetentionRepository{archivedLogs: make(map[string][]string)}
	config := RetentionConfig{
		EnableAutoArchive: false,
		EnableAutoDelete:  false,
	}

	manager := NewRetentionPolicyManager(repo, config, zap.NewNop())
	err := manager.Execute(context.Background())
	require.NoError(t, err)

	assert.Empty(t, repo.archivedLogs)
	assert.Empty(t, repo.deletedLogs)
}
