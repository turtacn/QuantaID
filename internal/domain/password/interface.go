package password

import (
	"context"
)

// IService defines the contract for password operations.
type IService interface {
	// Verify checks if the provided password matches the stored password for the user.
	Verify(ctx context.Context, userID, password string) (bool, error)
	// Hash hashes a password.
	Hash(password string) (string, error)
}
