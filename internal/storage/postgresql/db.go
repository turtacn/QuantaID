package postgresql

import (
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewConnection creates a new database connection using the provided configuration.
func NewConnection(config utils.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.Host, config.User, config.Password, config.DbName, config.Port, config.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Use(&PrometheusPlugin{}); err != nil {
		return nil, fmt.Errorf("failed to use prometheus plugin: %w", err)
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
		&types.MFAFactor{},
		&types.MFAVerificationLog{},
		&policy.Role{},
		&policy.Permission{},
		&policy.UserRole{},
	)
	if err != nil {
		return fmt.Errorf("gorm auto-migration failed: %w", err)
	}

	// Add custom indexes
	// User table indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC)").Error; err != nil {
		return err
	}

	// AuditLog table indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action ON audit_logs(user_id, action, created_at DESC)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_ip ON audit_logs(ip_address, created_at DESC)").Error; err != nil {
		return err
	}

	return nil
}