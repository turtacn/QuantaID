package audit

import (
	"context"
	"time"
)

// UserAccount represents a user account from the identity store.
// This is a simplified representation for the compliance checker's needs.
type UserAccount struct {
	ID        string
	CreatedAt time.Time
}

// IdentityRepository defines the interface for accessing user data.
// This allows the compliance checker to remain decoupled from the concrete
// identity storage implementation.
type IdentityRepository interface {
	// FindAccountsCreatedBefore finds user accounts created before a certain timestamp.
	// This is used to check for accounts that may have exceeded their retention period.
	FindAccountsCreatedBefore(ctx context.Context, cutoff time.Time) ([]UserAccount, error)
}
