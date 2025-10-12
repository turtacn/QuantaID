package postgresql

import (
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewConnection creates a new database connection using the provided configuration.
func NewConnection(config utils.PostgresConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	duration, err := time.ParseDuration(config.ConnMaxLifetime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connMaxLifetime: %w", err)
	}
	sqlDB.SetConnMaxLifetime(duration)

	return db, nil
}

// AutoMigrate runs the GORM auto-migration, which creates or updates the database schema
// to match the provided model definitions.
func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&types.User{},
		&types.UserGroup{},
		&types.IdentityProvider{},
		&types.Policy{},
		&types.Application{},
		&types.AuditLog{},
	)
	if err != nil {
		return fmt.Errorf("gorm auto-migration failed: %w", err)
	}
	return nil
}