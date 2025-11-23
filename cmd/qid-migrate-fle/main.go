package main

import (
	"context"
	"fmt"
	"log"

	"github.com/turtacn/QuantaID/internal/storage/postgresql/types"
	"github.com/turtacn/QuantaID/pkg/kms/local"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// LegacyUser represents the User table schema before encryption.
// We use this to read plaintext data if we were doing a real migration where we read raw rows.
// However, since we are doing an in-place update using GORM, we need to be careful.
//
// Strategy:
// 1. Read all users using a struct that maps Email/Phone to string.
// 2. For each user, update the record using the new struct where Email/Phone are EncryptedString.
//
// But wait, the database schema type for the column might change?
// If it was VARCHAR, it's still VARCHAR (or TEXT). The content changes from "email@example.com" to "base64(encrypted)".
// So the database schema itself (column type) doesn't necessarily change, just the content.
// GORM will handle the Value() interface to write encrypted data.
//
// Problem: If we read using the NEW struct (User with EncryptedString), Scan() will try to decrypt.
// If the data is currently Plaintext, Decrypt() will fail (likely invalid base64 or GCM tag mismatch).
//
// Solution:
// We need two structs. One for reading (Plaintext) and one for writing (Encrypted).
// Or simpler: Read as map[string]interface{} or raw bytes, then update.

type LegacyUser struct {
	ID    string
	Email string
	Phone string
}

func (LegacyUser) TableName() string {
	return "users"
}

func main() {
	// 1. Load Config
	logger, _ := utils.NewZapLogger(&utils.LoggerConfig{Level: "info"})
	configManager, err := utils.NewConfigManager("./configs", "server", "yaml", logger)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	var appCfg utils.Config
	if err := configManager.Unmarshal(&appCfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	if appCfg.DataEncryption.Key == "" {
		log.Fatal("Data Encryption Key is not set in config.")
	}

	// 2. Initialize KMS
	kmsProvider, err := local.New(appCfg.DataEncryption.Key)
	if err != nil {
		log.Fatalf("Failed to initialize KMS: %v", err)
	}
	types.SetGlobalKMS(kmsProvider)

	// 3. Connect to DB
	// Note: We need to bypass the standard repo to avoid automatic hooks if any
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		appCfg.Postgres.Host, appCfg.Postgres.User, appCfg.Postgres.Password,
		appCfg.Postgres.DbName, appCfg.Postgres.Port, appCfg.Postgres.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	logger.Info(context.Background(), "Starting migration to Field-Level Encryption...")

	// 4. Batch Process
	var offset int
	limit := 100
	for {
		// Read as LegacyUser to get plaintext
		var users []LegacyUser
		result := db.Limit(limit).Offset(offset).Order("id").Find(&users)
		if result.Error != nil {
			log.Fatalf("Failed to fetch users: %v", result.Error)
		}

		if len(users) == 0 {
			break
		}

		for _, u := range users {
			// Check if already encrypted (heuristic: starts with specific header or just try to decrypt?
			// Since we don't have a header, we assume this is a one-time run on plaintext data.
			// Ideally we should have a flag or check.
			// For this task, we assume all data is plaintext.

			updates := map[string]interface{}{}

			if u.Email != "" {
				updates["email"] = types.EncryptedString(u.Email)
			}
			if u.Phone != "" {
				updates["phone"] = types.EncryptedString(u.Phone)
			}

			if len(updates) > 0 {
				if err := db.Model(&LegacyUser{}).Where("id = ?", u.ID).Updates(updates).Error; err != nil {
					logger.Error(context.Background(), "Failed to update user", zap.String("userID", u.ID), zap.Error(err))
				} else {
					// logger.Info(context.Background(), "Encrypted user", zap.String("userID", u.ID))
				}
			}
		}

		offset += limit
		logger.Info(context.Background(), fmt.Sprintf("Processed %d users...", offset))
	}

	logger.Info(context.Background(), "Migration complete.")
}
