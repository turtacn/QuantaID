package postgresql

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
)

// InMemoryAuthRepository provides an in-memory implementation of the auth-related repositories,
// specifically the IdentityProviderRepository and AuditLogRepository.
// NOTE: Despite the package name 'postgresql', this is an IN-MEMORY implementation,
// likely used for testing or simple, non-persistent deployments. It uses maps and slices
// with a mutex for thread-safe operations.
type InMemoryAuthRepository struct {
	mu        sync.RWMutex
	providers map[string]*types.IdentityProvider
	auditLogs []*types.AuditLog
}

// NewInMemoryAuthRepository creates a new, empty in-memory auth repository.
func NewInMemoryAuthRepository() *InMemoryAuthRepository {
	return &InMemoryAuthRepository{
		providers: make(map[string]*types.IdentityProvider),
		auditLogs: make([]*types.AuditLog, 0),
	}
}

// --- IdentityProviderRepository Implementation ---

// CreateProvider adds a new identity provider to the in-memory store.
func (r *InMemoryAuthRepository) CreateProvider(ctx context.Context, provider *types.IdentityProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[provider.ID]; exists {
		return types.ErrConflict.WithDetails(map[string]string{"id": provider.ID})
	}
	r.providers[provider.ID] = provider
	return nil
}

// GetProviderByID retrieves an identity provider by its ID from the in-memory store.
func (r *InMemoryAuthRepository) GetProviderByID(ctx context.Context, id string) (*types.IdentityProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, exists := r.providers[id]
	if !exists {
		return nil, types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	return provider, nil
}

// GetProviderByName searches for an identity provider by its name in the in-memory store.
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

// ListProviders returns all identity providers from the in-memory store.
func (r *InMemoryAuthRepository) ListProviders(ctx context.Context) ([]*types.IdentityProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	providers := make([]*types.IdentityProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers, nil
}

// UpdateProvider updates an existing identity provider in the in-memory store.
func (r *InMemoryAuthRepository) UpdateProvider(ctx context.Context, provider *types.IdentityProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[provider.ID]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": provider.ID})
	}
	r.providers[provider.ID] = provider
	return nil
}

// DeleteProvider removes an identity provider from the in-memory store by its ID.
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

// CreateLogEntry adds a new audit log entry to the in-memory slice.
func (r *InMemoryAuthRepository) CreateLogEntry(ctx context.Context, entry *types.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.auditLogs = append(r.auditLogs, entry)
	return nil
}

// GetLogsForUser retrieves a paginated list of audit logs for a specific user from the in-memory store.
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

// GetLogsByAction retrieves a paginated list of audit logs for a specific action from the in-memory store.
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
