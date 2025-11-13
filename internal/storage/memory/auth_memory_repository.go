package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/pkg/types"
)

// AuthMemoryRepository provides an in-memory implementation of the auth repositories.
type AuthMemoryRepository struct {
	mu            sync.RWMutex
	sessions      map[string]*types.UserSession
	refreshTokens map[string]string // map[token] -> userID
	denyList      map[string]time.Time // map[jti] -> expiry time
	auditLogs     []*types.AuditLog
}

// NewAuthMemoryRepository creates a new in-memory auth repository.
func NewAuthMemoryRepository() *AuthMemoryRepository {
	return &AuthMemoryRepository{
		sessions:      make(map[string]*types.UserSession),
		refreshTokens: make(map[string]string),
		denyList:      make(map[string]time.Time),
		auditLogs:     make([]*types.AuditLog, 0),
	}
}

// --- SessionRepository implementation ---

func (r *AuthMemoryRepository) CreateSession(ctx context.Context, session *types.UserSession, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID] = session
	return nil
}

func (r *AuthMemoryRepository) GetSession(ctx context.Context, sessionID string) (*types.UserSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session with ID '%s' not found", sessionID)
	}
	return session, nil
}

func (r *AuthMemoryRepository) DeleteSession(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
	return nil
}

func (r *AuthMemoryRepository) GetUserSessions(ctx context.Context, userID string) ([]*types.UserSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var userSessions []*types.UserSession
	for _, session := range r.sessions {
		if session.UserID == userID {
			userSessions = append(userSessions, session)
		}
	}
	return userSessions, nil
}

// --- TokenRepository implementation ---

func (r *AuthMemoryRepository) StoreRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshTokens[token] = userID
	return nil
}

func (r *AuthMemoryRepository) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	userID, ok := r.refreshTokens[token]
	if !ok {
		return "", fmt.Errorf("refresh token not found")
	}
	return userID, nil
}

func (r *AuthMemoryRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.refreshTokens, token)
	return nil
}

func (r *AuthMemoryRepository) AddToDenyList(ctx context.Context, jti string, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.denyList[jti] = time.Now().Add(duration)
	return nil
}

func (r *AuthMemoryRepository) IsInDenyList(ctx context.Context, jti string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	expiry, exists := r.denyList[jti]
	if !exists || time.Now().After(expiry) {
		return false, nil
	}
	return true, nil
}

// --- AuditLogRepository implementation ---

func (r *AuthMemoryRepository) CreateLogEntry(ctx context.Context, entry *types.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	r.auditLogs = append(r.auditLogs, entry)
	return nil
}

func (r *AuthMemoryRepository) GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userLogs []*types.AuditLog
	for _, log := range r.auditLogs {
		if log.ActorID == userID {
			userLogs = append(userLogs, log)
		}
	}

	start := pq.Offset
	if start > len(userLogs) {
		start = len(userLogs)
	}

	end := start + pq.PageSize
	if end > len(userLogs) {
		end = len(userLogs)
	}

	return userLogs[start:end], nil
}

func (r *AuthMemoryRepository) GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var actionLogs []*types.AuditLog
	for _, log := range r.auditLogs {
		if log.Action == action {
			actionLogs = append(actionLogs, log)
		}
	}

	start := pq.Offset
	if start > len(actionLogs) {
		start = len(actionLogs)
	}

	end := start + pq.PageSize
	if end > len(actionLogs) {
		end = len(actionLogs)
	}

	return actionLogs[start:end], nil
}
