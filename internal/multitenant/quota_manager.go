package multitenant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrQuotaExceeded = errors.New("quota exceeded")
)

// QuotaManager manages tenant resource quotas.
type QuotaManager struct {
	db          *gorm.DB
	redisClient redis.RedisClientInterface
	config      map[string]utils.TenantQuotas
}

// NewQuotaManager creates a new QuotaManager.
func NewQuotaManager(db *gorm.DB, redisClient redis.RedisClientInterface, config map[string]utils.TenantQuotas) *QuotaManager {
	return &QuotaManager{
		db:          db,
		redisClient: redisClient,
		config:      config,
	}
}

// CheckUserQuota checks if the tenant has exceeded their user quota.
func (m *QuotaManager) CheckUserQuota(ctx context.Context, tenantID string) error {
	quotas, ok := m.config[tenantID]
	if !ok {
		// If no specific quota defined, assume unlimited or default?
		// For now, let's assume no check if not configured, or we could have a default.
		// Given the task, let's assume valid config or skip.
		return nil
	}

	if quotas.MaxUsers <= 0 {
		return nil // Unlimited
	}

	var count int64
	// Assumes "users" table has "tenant_id" column.
	// Using raw SQL or a model if available. Since this package might not import types to avoid cycles,
	// we assume "users" table. However, relying on GORM model would be safer if no cycle.
	// For now using Table("users").
	if err := m.db.WithContext(ctx).Table("users").Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if count >= int64(quotas.MaxUsers) {
		return fmt.Errorf("%w: max users %d reached", ErrQuotaExceeded, quotas.MaxUsers)
	}

	return nil
}

// CheckApplicationQuota checks if the tenant has exceeded their application quota.
func (m *QuotaManager) CheckApplicationQuota(ctx context.Context, tenantID string) error {
	quotas, ok := m.config[tenantID]
	if !ok {
		return nil
	}

	if quotas.MaxApplications <= 0 {
		return nil // Unlimited
	}

	var count int64
	if err := m.db.WithContext(ctx).Table("applications").Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count applications: %w", err)
	}

	if count >= int64(quotas.MaxApplications) {
		return fmt.Errorf("%w: max applications %d reached", ErrQuotaExceeded, quotas.MaxApplications)
	}

	return nil
}

// CheckAPICallQuota checks if the tenant has exceeded their daily API call quota.
// This doesn't increment the counter, just checks.
func (m *QuotaManager) CheckAPICallQuota(ctx context.Context, tenantID string) error {
	quotas, ok := m.config[tenantID]
	if !ok {
		return nil
	}

	if quotas.MaxAPICallsPerDay <= 0 {
		return nil // Unlimited
	}

	key := m.getAPIQuotaKey(tenantID)
	countStr, err := m.redisClient.Get(ctx, key)
	if err != nil {
		// go-redis V9 returns redis.Nil error when key doesn't exist.
		// Since we use the interface which returns (string, error), we check if error is "redis: nil"
		if err.Error() == "redis: nil" {
			return nil // No usage yet
		}
		return fmt.Errorf("failed to get api usage: %w", err)
	}

	var count int64
	fmt.Sscanf(countStr, "%d", &count)

	if count >= quotas.MaxAPICallsPerDay {
		return fmt.Errorf("%w: max api calls %d reached", ErrQuotaExceeded, quotas.MaxAPICallsPerDay)
	}

	return nil
}

// IncrementAPICall increments the API call counter for the tenant.
func (m *QuotaManager) IncrementAPICall(ctx context.Context, tenantID string) error {
	quotas, ok := m.config[tenantID]
	if !ok || quotas.MaxAPICallsPerDay <= 0 {
		return nil
	}

	key := m.getAPIQuotaKey(tenantID)

	// Increment atomically
	// The interface doesn't have Incr. We should add it or use raw client?
	// The interface has Client() which returns *redis.Client.
	newVal, err := m.redisClient.Client().Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment api usage: %w", err)
	}

	// Set expiration if it's the first increment (approx logic)
	// Or just always set expiration to 24h from now?
	// Better: Set expiration to end of current day (UTC).
	if newVal == 1 {
		now := time.Now().UTC()
		nextDay := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
		ttl := nextDay.Sub(now)
		m.redisClient.Expire(ctx, key, ttl)
	}

	// Re-check quota after increment?
	// The prompt distinguishes Check and Increment.
	// Often we want Check then Increment or Increment then Check.
	// Since `CheckAPICallQuota` exists separately, we assume the caller handles flow.
	// But strictly, if we increment and now it's exceeded, should we return error?
	// If `newVal > quotas.MaxAPICallsPerDay`, strictly we exceeded it.

	if newVal > quotas.MaxAPICallsPerDay {
		return fmt.Errorf("%w: max api calls %d reached", ErrQuotaExceeded, quotas.MaxAPICallsPerDay)
	}

	return nil
}

func (m *QuotaManager) getAPIQuotaKey(tenantID string) string {
	date := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf("quota:api:%s:%s", tenantID, date)
}
