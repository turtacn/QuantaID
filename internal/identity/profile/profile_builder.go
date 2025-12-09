package profile

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
)

// AccessLogRepository defines the interface for reading access logs
type AccessLogRepository interface {
	GetLogsForUser(ctx context.Context, userID string, since time.Time) ([]models.AccessLog, error)
}

// DeviceRepository defines the interface for reading device data
type DeviceRepository interface {
	GetDevicesByUserID(ctx context.Context, userID string) ([]models.Device, error)
}

// MFAService defines the interface for checking MFA status
type MFAService interface {
	HasAnyMethod(ctx context.Context, userID string) (bool, error)
}

// ProfileBuilder aggregates data to build or update user profiles
type ProfileBuilder struct {
	profileRepo   ProfileRepository
	accessLogRepo AccessLogRepository
	deviceRepo    DeviceRepository
	userRepo      identity.IService
	mfaService    MFAService
}

// NewProfileBuilder creates a new ProfileBuilder
func NewProfileBuilder(
	profileRepo ProfileRepository,
	accessLogRepo AccessLogRepository,
	deviceRepo DeviceRepository,
	userRepo identity.IService,
	mfaService MFAService,
) *ProfileBuilder {
	return &ProfileBuilder{
		profileRepo:   profileRepo,
		accessLogRepo: accessLogRepo,
		deviceRepo:    deviceRepo,
		userRepo:      userRepo,
		mfaService:    mfaService,
	}
}

// BuildOrUpdate builds or updates the profile for a given user
func (b *ProfileBuilder) BuildOrUpdate(ctx context.Context, userID string) (*UserProfile, error) {
	profile, err := b.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		user, err := b.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		// Assuming user has TenantID, if not, we might need another way to get it
		// types.User does not have TenantID field, using a safe fallback or extracting from attributes if available
		tenantID := "default"
		if tid, ok := user.Attributes["tenant_id"].(string); ok {
			tenantID = tid
		}

		profile = &UserProfile{
			ID:             uuid.New().String(),
			UserID:         userID,
			TenantID:       tenantID,
			CreatedAt:      time.Now(),
			AutoTags:       []string{},
			ManualTags:     []string{},
			Behavior:       BehaviorMetrics{},
			Risk:           RiskIndicators{},
			QualityDetails: QualityDetails{},
		}
		if err := b.profileRepo.Create(ctx, profile); err != nil {
			return nil, err
		}
	}

	behavior, err := b.buildBehaviorMetrics(ctx, userID)
	if err == nil {
		profile.Behavior = behavior
	}

	quality, err := b.buildQualityDetails(ctx, userID)
	if err == nil {
		profile.QualityDetails = quality
	}

	now := time.Now()
	profile.LastActivityAt = &now
	profile.UpdatedAt = now

	if err := b.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (b *ProfileBuilder) buildBehaviorMetrics(ctx context.Context, userID string) (BehaviorMetrics, error) {
	since := time.Now().AddDate(0, 0, -90) // Last 90 days
	logs, err := b.accessLogRepo.GetLogsForUser(ctx, userID, since)
	if err != nil {
		return BehaviorMetrics{}, err
	}

	metrics := BehaviorMetrics{}
	deviceSet := make(map[string]struct{})
	locationSet := make(map[string]struct{})
	ipSet := make(map[string]struct{})
	hourlyActivity := make([]int, 24)

	var successfulLogins int64
	var mfaVerifiedCount int64

	for _, log := range logs {
		if log.Action == "login" {
			if log.Success {
				metrics.TotalLogins++
				successfulLogins++
			} else {
				metrics.FailedLogins++
			}
		}

		if log.DeviceID != "" {
			deviceSet[log.DeviceID] = struct{}{}
		}
		if log.Location != "" {
			locationSet[log.Location] = struct{}{}
		}
		if log.IPAddress != "" {
			ipSet[log.IPAddress] = struct{}{}
		}

		// Use CreatedAt as Timestamp is transient
		hour := log.CreatedAt.Hour()
		hourlyActivity[hour]++

		if log.MFAVerified {
			mfaVerifiedCount++
		}
	}

	metrics.UniqueDevices = len(deviceSet)
	metrics.UniqueLocations = len(locationSet)
	metrics.UniqueIPs = len(ipSet)
	metrics.LoginFrequency = float64(metrics.TotalLogins) / (90.0 / 7.0) // Per week
	if metrics.TotalLogins > 0 {
		metrics.MFAUsageRate = float64(mfaVerifiedCount) / float64(metrics.TotalLogins)
	}

	// Calculate peak hours (top 3)
	// Simplified: store counts for now, or just the indices
	// The requirement is []int active hours. Let's filter hours with significant activity.
	for h, count := range hourlyActivity {
		if count > 0 {
			metrics.PeakActivityHours = append(metrics.PeakActivityHours, h)
		}
	}

	return metrics, nil
}

func (b *ProfileBuilder) buildQualityDetails(ctx context.Context, userID string) (QualityDetails, error) {
	user, err := b.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return QualityDetails{}, err
	}

	hasMFA, _ := b.mfaService.HasAnyMethod(ctx, userID)

	// Check user verification status if available in Attributes, or assume false if fields missing
	emailVerified := false
	if v, ok := user.Attributes["email_verified"].(bool); ok {
		emailVerified = v
	}

	phoneVerified := false
	if v, ok := user.Attributes["phone_verified"].(bool); ok {
		phoneVerified = v
	}

	details := QualityDetails{
		HasEmail:         string(user.Email) != "", // EncryptedString cast to string
		EmailVerified:    emailVerified,
		HasPhone:         string(user.Phone) != "", // EncryptedString cast to string
		PhoneVerified:    phoneVerified,
		HasMFA:           hasMFA,
		HasRecoveryEmail: false, // Not in standard User struct yet, placeholder
	}

	// Calculate completeness
	fields := 0.0
	filled := 0.0

	fields++
	if details.HasEmail { filled++ }
	fields++
	if details.EmailVerified { filled++ }
	fields++
	if details.HasPhone { filled++ }
	fields++
	if details.PhoneVerified { filled++ }

	// Add other fields from user.Attributes if needed

	if fields > 0 {
		details.ProfileComplete = filled / fields
	}

	return details, nil
}

// IncrementalUpdate updates metrics incrementally based on a single event
func (b *ProfileBuilder) IncrementalUpdate(ctx context.Context, userID string, event map[string]interface{}) error {
	profile, err := b.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if profile == nil {
		// If profile doesn't exist, trigger full build
		_, err := b.BuildOrUpdate(ctx, userID)
		return err
	}

	eventType, _ := event["type"].(string)

	switch eventType {
	case "login":
		profile.Behavior.TotalLogins++
		// Update other metrics like unique IP/Location if details present
	case "mfa_verified":
		// Recalculate rate? Or just track raw counts in a better model
	}

	now := time.Now()
	profile.LastActivityAt = &now
	return b.profileRepo.UpdateBehavior(ctx, userID, profile.Behavior)
}
