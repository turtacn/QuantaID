package actions

import (
	"context"
	"fmt"

	"github.com/turtacn/QuantaID/internal/security/automator"
	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// LockUserAction implements the SecurityAction interface to lock a user's account.
type LockUserAction struct {
	db *gorm.DB
}

// NewLockUserAction creates a new instance of LockUserAction.
func NewLockUserAction(db *gorm.DB) *LockUserAction {
	return &LockUserAction{db: db}
}

// ID returns the unique identifier for the action.
func (a *LockUserAction) ID() string {
	return "lock_user"
}

// Execute updates the user's status to 'locked'.
func (a *LockUserAction) Execute(ctx context.Context, input automator.ActionInput) error {
	if input.UserID == "" {
		return fmt.Errorf("UserID is required for LockUserAction")
	}

	result := a.db.WithContext(ctx).Model(&types.User{}).Where("id = ?", input.UserID).Update("status", types.UserStatusLocked)
	if result.Error != nil {
		return fmt.Errorf("failed to lock user %s: %w", input.UserID, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user %s not found for locking", input.UserID)
	}

	return nil
}
