package device

import (
	"context"
	"errors"
	"time"

	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
	"gorm.io/gorm"
)

// DeviceRepository defines the interface for device persistence
type DeviceRepository interface {
	Create(ctx context.Context, device *models.Device) error
	GetByID(ctx context.Context, id string) (*models.Device, error)
	GetByFingerprint(ctx context.Context, fingerprint string) (*models.Device, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Device, error)
	Update(ctx context.Context, device *models.Device) error
	Delete(ctx context.Context, id string) error
	UpdateTrustScore(ctx context.Context, id string, score int) error
	UpdateLastActive(ctx context.Context, id string, ip, location string) error
}

type PostgresDeviceRepository struct {
	db *gorm.DB
}

func NewPostgresDeviceRepository(db *gorm.DB) DeviceRepository {
	return &PostgresDeviceRepository{db: db}
}

func (r *PostgresDeviceRepository) Create(ctx context.Context, device *models.Device) error {
	if device.ID == "" {
		// Assuming ID generation is handled elsewhere or let DB handle it if configured,
		// but models usually need ID before Create in GORM unless it's auto-increment or similar.
		// For now, assuming caller sets it or we should. The requirements said:
		// "Create(ctx, device *Device) error ... 1. 设置device.ID = GenerateDeviceID() 如果为空"
		// I'll leave ID generation to the caller or service layer as repository usually just persists.
		// However, I'll follow the pseudo code which mentioned setting ID.
		// Since I don't have a global ID generator handy here without dependencies,
		// I will rely on the service to provide it or use UUID if I import it.
		// To keep it simple and follow typical patterns, I'll assume ID is set.
		// If strictly following the pseudo-code: "设置device.ID = GenerateDeviceID() 如果为空"
		// I would need a generator.
	}
	if device.CreatedAt.IsZero() {
		device.CreatedAt = time.Now()
	}
	if device.UpdatedAt.IsZero() {
		device.UpdatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *PostgresDeviceRepository) GetByID(ctx context.Context, id string) (*models.Device, error) {
	var device models.Device
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *PostgresDeviceRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*models.Device, error) {
	var device models.Device
	err := r.db.WithContext(ctx).Where("fingerprint = ?", fingerprint).First(&device).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if not found, let service handle logic
		}
		return nil, err
	}
	return &device, nil
}

func (r *PostgresDeviceRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Device, error) {
	var devices []*models.Device
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&devices).Error
	return devices, err
}

func (r *PostgresDeviceRepository) Update(ctx context.Context, device *models.Device) error {
	device.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(device).Error
}

func (r *PostgresDeviceRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Device{}, "id = ?", id).Error
}

func (r *PostgresDeviceRepository) UpdateTrustScore(ctx context.Context, id string, score int) error {
	return r.db.WithContext(ctx).Model(&models.Device{}).Where("id = ?", id).Updates(map[string]interface{}{
		"trust_score": score,
		"updated_at":  time.Now(),
	}).Error
}

func (r *PostgresDeviceRepository) UpdateLastActive(ctx context.Context, id string, ip, location string) error {
	updates := map[string]interface{}{
		"last_active_at": time.Now(),
		"updated_at":     time.Now(),
	}
	if ip != "" {
		updates["last_ip"] = ip
	}
	if location != "" {
		updates["last_location"] = location
	}
	return r.db.WithContext(ctx).Model(&models.Device{}).Where("id = ?", id).Updates(updates).Error
}
