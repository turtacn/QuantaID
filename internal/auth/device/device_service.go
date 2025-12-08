package device

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
)

var (
	ErrDeviceAlreadyBound = errors.New("device already bound to another user")
	ErrNotDeviceOwner     = errors.New("user does not own this device")
)

// DeviceService handles device management logic
type DeviceService struct {
	repo            DeviceRepository
	trustScorer     *TrustScorer
	anomalyDetector *AnomalyDetector
	fingerprinter   *DeviceFingerprinter
	// eventBus        events.EventPublisher // Placeholder
}

// NewDeviceService creates a new DeviceService
func NewDeviceService(repo DeviceRepository, scorer *TrustScorer, detector *AnomalyDetector, fp *DeviceFingerprinter) *DeviceService {
	return &DeviceService{
		repo:            repo,
		trustScorer:     scorer,
		anomalyDetector: detector,
		fingerprinter:   fp,
	}
}

// RegisterOrUpdate registers a new device or updates an existing one
func (s *DeviceService) RegisterOrUpdate(ctx context.Context, fingerprintData map[string]interface{}, ip, tenantID string) (*models.Device, error) {
	fpHash := s.fingerprinter.GenerateHash(fingerprintData)

	existing, err := s.repo.GetByFingerprint(ctx, fpHash)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Existing device

		// Anomaly Detection
		if s.anomalyDetector != nil {
			result := s.anomalyDetector.DetectGeoJump(existing.LastIP, existing.LastActiveAt, ip, time.Now())
			if result.Detected {
				// TODO: Publish DeviceAnomalyEvent
			}

			fpChangeResult := s.anomalyDetector.DetectFingerprintChange(map[string]interface{}(existing.FingerprintRaw), fingerprintData)
			if fpChangeResult.Detected {
				// TODO: Publish DeviceAnomalyEvent
			}
		}

		// Update fields in struct
		existing.LastIP = ip
		existing.FingerprintRaw = models.JSONMap(fingerprintData)
		existing.LastActiveAt = time.Now()

		// Recalculate Score
		existing.TrustScore = s.trustScorer.CalculateScore(existing)

		// Consolidate updates into a single Update call which saves the entire struct
		// This avoids the issue where UpdateLastActive might be overwritten or vice versa
		err = s.repo.Update(ctx, existing)
		if err != nil {
			return nil, err
		}

		return existing, nil
	}

	// New Device
	now := time.Now()
	newDevice := &models.Device{
		ID:             uuid.New().String(),
		TenantID:       tenantID,
		Fingerprint:    fpHash,
		FingerprintRaw: models.JSONMap(fingerprintData),
		LastIP:         ip,
		LastActiveAt:   now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Initial Score
	newDevice.TrustScore = s.trustScorer.CalculateScore(newDevice)

	err = s.repo.Create(ctx, newDevice)
	if err != nil {
		return nil, err
	}

	// TODO: Publish DeviceRegisteredEvent

	return newDevice, nil
}

// BindToUser binds a device to a user
func (s *DeviceService) BindToUser(ctx context.Context, deviceID, userID string) error {
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		return err
	}
	if device == nil {
		return errors.New("device not found")
	}

	if device.UserID != "" && device.UserID != userID {
		return ErrDeviceAlreadyBound
	}

	device.UserID = userID
	now := time.Now()
	device.BoundAt = &now
	device.TrustScore = s.trustScorer.CalculateScore(device)

	return s.repo.Update(ctx, device)
}

// UnbindFromUser unbinds a device from a user
func (s *DeviceService) UnbindFromUser(ctx context.Context, deviceID, userID string) error {
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		return err
	}
	if device == nil {
		return errors.New("device not found")
	}

	if device.UserID != userID {
		return ErrNotDeviceOwner
	}

	device.UserID = ""
	device.BoundAt = nil
	device.TrustScore = s.trustScorer.CalculateScore(device)

	return s.repo.Update(ctx, device)
}

// GetUserDevices returns all devices for a user
func (s *DeviceService) GetUserDevices(ctx context.Context, userID string) ([]*models.Device, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// GetTrustLevel returns the trust level for a device
func (s *DeviceService) GetTrustLevel(ctx context.Context, deviceID string) (TrustLevel, error) {
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		return "", err
	}
	if device == nil {
		return "", errors.New("device not found")
	}

	return s.trustScorer.GetTrustLevel(device.TrustScore), nil
}
