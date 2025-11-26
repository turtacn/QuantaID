package postgresql

import (
	"context"
	"gorm.io/gorm"
	"github.com/turtacn/QuantaID/internal/domain/privacy"
	"github.com/turtacn/QuantaID/pkg/types"
)

type PostgresPrivacyRepository struct {
	db *gorm.DB
}

func NewPostgresPrivacyRepository(db *gorm.DB) *PostgresPrivacyRepository {
	return &PostgresPrivacyRepository{db: db}
}

func (r *PostgresPrivacyRepository) CreateConsentRecord(ctx context.Context, record *privacy.ConsentRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *PostgresPrivacyRepository) GetLastConsentRecord(ctx context.Context, userID, policyID string) (*privacy.ConsentRecord, error) {
	var record privacy.ConsentRecord
	err := r.db.WithContext(ctx).Where("user_id = ? AND policy_id = ?", userID, policyID).Order("created_at desc").First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, types.ErrNotFound
	}
	return &record, err
}

func (r *PostgresPrivacyRepository) GetConsentHistory(ctx context.Context, userID string) ([]*privacy.ConsentRecord, error) {
	var records []*privacy.ConsentRecord
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&records).Error
	return records, err
}

func (r *PostgresPrivacyRepository) CreateDSRRequest(ctx context.Context, request *privacy.DSRRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *PostgresPrivacyRepository) GetDSRRequest(ctx context.Context, requestID string) (*privacy.DSRRequest, error) {
	var request privacy.DSRRequest
	err := r.db.WithContext(ctx).Where("id = ?", requestID).First(&request).Error
	if err == gorm.ErrRecordNotFound {
		return nil, types.ErrNotFound
	}
	return &request, err
}

func (r *PostgresPrivacyRepository) UpdateDSRRequestStatus(ctx context.Context, requestID string, status privacy.DSRRequestStatus) error {
	return r.db.WithContext(ctx).Model(&privacy.DSRRequest{}).Where("id = ?", requestID).Update("status", status).Error
}
