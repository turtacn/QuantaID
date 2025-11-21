package postgresql

import (
	"context"

	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// PostgresApplicationRepository provides a GORM-based implementation of the application-related repositories.
type PostgresApplicationRepository struct {
	db *gorm.DB
}

// NewPostgresApplicationRepository creates a new PostgreSQL application repository.
func NewPostgresApplicationRepository(db *gorm.DB) *PostgresApplicationRepository {
	return &PostgresApplicationRepository{db: db}
}

// --- ApplicationRepository Implementation ---

// CreateApplication adds a new application to the database.
func (r *PostgresApplicationRepository) CreateApplication(ctx context.Context, app *types.Application) error {
	return r.db.WithContext(ctx).Create(app).Error
}

// GetApplicationByID retrieves an application by its ID from the database.
func (r *PostgresApplicationRepository) GetApplicationByID(ctx context.Context, id string) (*types.Application, error) {
	var app types.Application
	err := r.db.WithContext(ctx).First(&app, "id = ?", id).Error
	return &app, err
}

// GetApplicationByName searches for an application by its name in the database.
func (r *PostgresApplicationRepository) GetApplicationByName(ctx context.Context, name string) (*types.Application, error) {
	var app types.Application
	err := r.db.WithContext(ctx).First(&app, "name = ?", name).Error
	return &app, err
}

// UpdateApplication updates an existing application in the database.
func (r *PostgresApplicationRepository) UpdateApplication(ctx context.Context, app *types.Application) error {
	return r.db.WithContext(ctx).Save(app).Error
}

// DeleteApplication removes an application from the database.
func (r *PostgresApplicationRepository) DeleteApplication(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.Application{}, "id = ?", id).Error
}

// ListApplications returns a paginated list of all applications from the database.
func (r *PostgresApplicationRepository) ListApplications(ctx context.Context, pq types.PaginationQuery) ([]*types.Application, error) {
	var apps []*types.Application
	err := r.db.WithContext(ctx).Offset(pq.Offset).Limit(pq.PageSize).Find(&apps).Error
	return apps, err
}

// GetApplicationByClientID retrieves an application by its client ID from the database.
func (r *PostgresApplicationRepository) GetApplicationByClientID(ctx context.Context, clientID string) (*types.Application, error) {
	var app types.Application
	err := r.db.WithContext(ctx).First(&app, "id = ?", clientID).Error
	return &app, err
}