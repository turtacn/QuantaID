package redis

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
	"time"
)

// sessionWithValue stores a user session along with its expiration time for in-memory management.
type sessionWithValue struct {
	session   *types.UserSession
	expiresAt time.Time
}

// InMemorySessionRepository provides an in-memory implementation of the auth.SessionRepository interface.
// NOTE: Despite the package name 'redis', this is an IN-MEMORY implementation,
// likely used for testing or simple, non-persistent deployments. It uses a map
// with a mutex for thread-safe operations.
type InMemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]sessionWithValue
}

// NewInMemorySessionRepository creates a new, empty in-memory session repository.
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]sessionWithValue),
	}
}

// CreateSession stores a new user session in the in-memory map with a specified duration.
func (r *InMemorySessionRepository) CreateSession(ctx context.Context, session *types.UserSession, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[session.ID] = sessionWithValue{
		session:   session,
		expiresAt: time.Now().Add(duration),
	}
	return nil
}

// GetSession retrieves a session by its ID from the in-memory store.
// It returns an error if the session is not found or has expired.
// Expired sessions are lazily deleted upon access.
func (r *InMemorySessionRepository) GetSession(ctx context.Context, sessionID string) (*types.UserSession, error) {
	r.mu.RLock()
	val, exists := r.sessions[sessionID]
	r.mu.RUnlock()

	if !exists {
		return nil, types.ErrNotFound.WithDetails(map[string]string{"session_id": sessionID})
	}

	if time.Now().After(val.expiresAt) {
		// Lazily delete expired session
		r.mu.Lock()
		delete(r.sessions, sessionID)
		r.mu.Unlock()
		return nil, types.ErrNotFound.WithDetails(map[string]string{"session_id": sessionID, "reason": "expired"})
	}

	return val.session, nil
}

// DeleteSession removes a session from the in-memory store by its ID.
func (r *InMemorySessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
	return nil
}

// GetUserSessions retrieves all active (non-expired) sessions for a specific user from the in-memory store.
func (r *InMemorySessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*types.UserSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userSessions []*types.UserSession
	for _, val := range r.sessions {
		if val.session.UserID == userID {
			if time.Now().Before(val.expiresAt) {
				userSessions = append(userSessions, val.session)
			}
		}
	}
	return userSessions, nil
}
