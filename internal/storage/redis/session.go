package redis

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
	"time"
)

type sessionWithValue struct {
	session   *types.UserSession
	expiresAt time.Time
}

// InMemorySessionRepository is an in-memory implementation of the SessionRepository.
type InMemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]sessionWithValue
}

// NewInMemorySessionRepository creates a new in-memory session repository.
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]sessionWithValue),
	}
}

func (r *InMemorySessionRepository) CreateSession(ctx context.Context, session *types.UserSession, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[session.ID] = sessionWithValue{
		session:   session,
		expiresAt: time.Now().Add(duration),
	}
	return nil
}

func (r *InMemorySessionRepository) GetSession(ctx context.Context, sessionID string) (*types.UserSession, error) {
	r.mu.RLock()
	val, exists := r.sessions[sessionID]
	r.mu.RUnlock()

	if !exists {
		return nil, types.ErrNotFound.WithDetails(map[string]string{"session_id": sessionID})
	}

	if time.Now().After(val.expiresAt) {
		r.mu.Lock()
		delete(r.sessions, sessionID)
		r.mu.Unlock()
		return nil, types.ErrNotFound.WithDetails(map[string]string{"session_id": sessionID, "reason": "expired"})
	}

	return val.session, nil
}

func (r *InMemorySessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
	return nil
}

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

//Personal.AI order the ending
