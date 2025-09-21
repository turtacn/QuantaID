package postgresql

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
)

// InMemoryAuthRepository is an in-memory implementation of the auth repositories.
type InMemoryAuthRepository struct {
	mu        sync.RWMutex
	providers map[string]*types.IdentityProvider
	auditLogs []*types.AuditLog
}

// NewInMemoryAuthRepository creates a new in-memory auth repository.
func NewInMemoryAuthRepository() *InMemoryAuthRepository {
	return &InMemoryAuthRepository{
		providers: make(map[string]*types.IdentityProvider),
		auditLogs: make([]*types.AuditLog, 0),
	}
}

// --- IdentityProviderRepository Implementation ---

func (r *InMemoryAuthRepository) CreateProvider(ctx context.Context, provider *types.IdentityProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[provider.ID]; exists {
		return types.ErrConflict.WithDetails(map[string]string{"id": provider.ID})
	}
	r.providers[provider.ID] = provider
	return nil
}

func (r *InMemoryAuthRepository) GetProviderByID(ctx context.Context, id string) (*types.IdentityProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, exists := r.providers[id]
	if !exists {
		return nil, types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	return provider, nil
}

func (r *InMemoryAuthRepository) GetProviderByName(ctx context.Context, name string) (*types.IdentityProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.providers {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, types.ErrNotFound.WithDetails(map[string]string{"name": name})
}

func (r *InMemoryAuthRepository) ListProviders(ctx context.Context) ([]*types.IdentityProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	providers := make([]*types.IdentityProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers, nil
}

func (r *InMemoryAuthRepository) UpdateProvider(ctx context.Context, provider *types.IdentityProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[provider.ID]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": provider.ID})
	}
	r.providers[provider.ID] = provider
	return nil
}

func (r *InMemoryAuthRepository) DeleteProvider(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[id]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	delete(r.providers, id)
	return nil
}

// --- AuditLogRepository Implementation ---

func (r *InMemoryAuthRepository) CreateLogEntry(ctx context.Context, entry *types.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.auditLogs = append(r.auditLogs, entry)
	return nil
}

func (r *InMemoryAuthRepository) GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var userLogs []*types.AuditLog
	for _, log := range r.auditLogs {
		if log.ActorID == userID {
			userLogs = append(userLogs, log)
		}
	}
	start, end := pq.Offset, pq.Offset+pq.PageSize
	if start > len(userLogs) { return []*types.AuditLog{}, nil }
	if end > len(userLogs) { end = len(userLogs) }
	return userLogs[start:end], nil
}

func (r *InMemoryAuthRepository) GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var actionLogs []*types.AuditLog
	for _, log := range r.auditLogs {
		if log.Action == action {
			actionLogs = append(actionLogs, log)
		}
	}
	start, end := pq.Offset, pq.Offset+pq.PageSize
	if start > len(actionLogs) { return []*types.AuditLog{}, nil }
	if end > len(actionLogs) { end = len(actionLogs) }
	return actionLogs[start:end], nil
}

//Personal.AI order the ending
