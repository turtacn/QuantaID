package auth

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"time"
)

// SessionRepository defines the interface for a persistence layer for user sessions.
// This is typically implemented by a cache like Redis.
type SessionRepository interface {
	// CreateSession stores a new user session with a specified duration.
	CreateSession(ctx context.Context, session *types.UserSession, duration time.Duration) error
	// GetSession retrieves a user session by its ID.
	GetSession(ctx context.Context, sessionID string) (*types.UserSession, error)
	// DeleteSession removes a user session from the store.
	DeleteSession(ctx context.Context, sessionID string) error
	// GetUserSessions retrieves all active sessions for a given user.
	GetUserSessions(ctx context.Context, userID string) ([]*types.UserSession, error)
}

// TokenRepository defines the interface for managing refresh tokens and JWT deny lists.
// This is crucial for handling token revocation and refresh mechanics.
type TokenRepository interface {
	// StoreRefreshToken saves a refresh token and associates it with a user ID.
	StoreRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error
	// GetRefreshTokenUserID retrieves the user ID associated with a given refresh token.
	GetRefreshTokenUserID(ctx context.Context, token string) (string, error)
	// DeleteRefreshToken removes a refresh token from the store.
	DeleteRefreshToken(ctx context.Context, token string) error
	// AddToDenyList adds a JWT ID (jti) to a deny list for a specified duration (typically until it expires).
	AddToDenyList(ctx context.Context, jti string, duration time.Duration) error
	// IsInDenyList checks if a JWT ID (jti) is in the deny list.
	IsInDenyList(ctx context.Context, jti string) (bool, error)
}

// IdentityProviderRepository defines the CRUD interface for managing identity provider configurations in the database.
type IdentityProviderRepository interface {
	// CreateProvider adds a new identity provider configuration.
	CreateProvider(ctx context.Context, provider *types.IdentityProvider) error
	// GetProviderByID retrieves an identity provider by its unique ID.
	GetProviderByID(ctx context.Context, id string) (*types.IdentityProvider, error)
	// GetProviderByName retrieves an identity provider by its unique name.
	GetProviderByName(ctx context.Context, name string) (*types.IdentityProvider, error)
	// ListProviders returns all configured identity providers.
	ListProviders(ctx context.Context) ([]*types.IdentityProvider, error)
	// UpdateProvider modifies an existing identity provider's configuration.
	UpdateProvider(ctx context.Context, provider *types.IdentityProvider) error
	// DeleteProvider removes an identity provider configuration.
	DeleteProvider(ctx context.Context, id string) error
}

// AuditLogRepository defines the interface for a persistence layer for audit log entries.
type AuditLogRepository interface {
	// CreateLogEntry records a new audit log event.
	CreateLogEntry(ctx context.Context, entry *types.AuditLog) error
	// GetLogsForUser retrieves a paginated list of audit logs for a specific user.
	GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error)
	// GetLogsByAction retrieves a paginated list of audit logs for a specific action type.
	GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error)
}

// TokenFamily represents a chain of rotated refresh tokens.
type TokenFamily struct {
    FamilyID       string
    OriginalToken  string
    CurrentToken   string
    IssuedTokens   []string // All refresh tokens issued in this family.
    UserID         string
    ClientID       string
    RevokedAt      *time.Time
}

// TokenFamilyRepository defines the interface for managing refresh token families.
// This is essential for implementing refresh token rotation and detecting replay attacks.
type TokenFamilyRepository interface {
    // CreateFamily creates a new token family when a refresh token is first issued.
    CreateFamily(ctx context.Context, family *TokenFamily) error
    // GetFamilyByToken retrieves the token family associated with a given refresh token.
    GetFamilyByToken(ctx context.Context, token string) (*TokenFamily, error)
    // GetFamilyByID retrieves a token family by its unique ID.
    GetFamilyByID(ctx context.Context, familyID string) (*TokenFamily, error)
    // UpdateFamily updates an existing token family, typically to add a new token during rotation.
    UpdateFamily(ctx context.Context, family *TokenFamily) error
    // RevokeFamily marks an entire token family as revoked, invalidating all its tokens.
    RevokeFamily(ctx context.Context, familyID string) error
}
