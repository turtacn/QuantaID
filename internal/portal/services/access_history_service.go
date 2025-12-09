package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/turtacn/QuantaID/internal/auth/adaptive"
	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
)

// AccessHistoryService handles user access history logic.
type AccessHistoryService struct {
	db         *gorm.DB
	geoService adaptive.GeoIPProvider
}

// NewAccessHistoryService creates a new instance of AccessHistoryService.
func NewAccessHistoryService(db *gorm.DB, geoService adaptive.GeoIPProvider) *AccessHistoryService {
	return &AccessHistoryService{
		db:         db,
		geoService: geoService,
	}
}

// RecordAccess creates a new access log entry.
func (s *AccessHistoryService) RecordAccess(ctx context.Context, log *models.AccessLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}

	if s.geoService != nil && log.IPAddress != "" {
		location, err := s.geoService.GetLocation(ctx, log.IPAddress)
		if err == nil && location != nil {
			log.Location = location.City
		}
	}

	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	return s.db.WithContext(ctx).Create(log).Error
}

// Query retrieves access logs based on the provided filter.
func (s *AccessHistoryService) Query(ctx context.Context, filter models.AccessLogFilter) ([]models.AccessLog, int64, error) {
	var logs []models.AccessLog
	var total int64

	query := s.db.WithContext(ctx).Model(&models.AccessLog{})

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.DeviceID != "" {
		query = query.Where("device_id = ?", filter.DeviceID)
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", filter.EndTime)
	}
	if filter.Success != nil {
		query = query.Where("success = ?", filter.Success)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("created_at DESC")

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetRecentByUser retrieves the most recent access logs for a user.
func (s *AccessHistoryService) GetRecentByUser(ctx context.Context, userID string, limit int) ([]models.AccessLog, error) {
	filter := models.AccessLogFilter{
		UserID: userID,
		Limit:  limit,
	}
	logs, _, err := s.Query(ctx, filter)
	return logs, err
}

// CleanupOldLogs deletes access logs older than the specified retention period.
func (s *AccessHistoryService) CleanupOldLogs(ctx context.Context, tenantID string, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := s.db.WithContext(ctx).Where("tenant_id = ? AND created_at < ?", tenantID, cutoff).Delete(&models.AccessLog{})
	return result.RowsAffected, result.Error
}
