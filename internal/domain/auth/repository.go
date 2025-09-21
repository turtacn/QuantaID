package auth

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"time"
)

// SessionRepository defines the interface for storing and retrieving user sessions.
type SessionRepository interface {
	CreateSession(ctx context.Context, session *types.UserSession, duration time.Duration) error
	GetSession(ctx context.Context, sessionID string) (*types.UserSession, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserSessions(ctx context.Context, userID string) ([]*types.UserSession, error)
}

// TokenRepository defines the interface for managing tokens.
type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error
	GetRefreshTokenUserID(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	AddToDenyList(ctx context.Context, jti string, duration time.Duration) error
	IsInDenyList(ctx context.Context, jti string) (bool, error)
}

// IdentityProviderRepository defines the interface for managing identity provider configurations.
type IdentityProviderRepository interface {
	CreateProvider(ctx context.Context, provider *types.IdentityProvider) error
	GetProviderByID(ctx context.Context, id string) (*types.IdentityProvider, error)
	GetProviderByName(ctx context.Context, name string) (*types.IdentityProvider, error)
	ListProviders(ctx context.Context) ([]*types.IdentityProvider, error)
	UpdateProvider(ctx context.Context, provider *types.IdentityProvider) error
	DeleteProvider(ctx context.Context, id string) error
}

// AuditLogRepository defines the interface for storing audit logs.
type AuditLogRepository interface {
	CreateLogEntry(ctx context.Context, entry *types.AuditLog) error
	GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error)
	GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error)
}

//Personal.AI order the ending
