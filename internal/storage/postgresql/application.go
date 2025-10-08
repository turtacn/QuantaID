package postgresql

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
)

// InMemoryApplicationRepository provides an in-memory implementation of the application repository.
// NOTE: Despite the package name 'postgresql', this is an IN-MEMORY implementation,
// likely used for testing or simple, non-persistent deployments.
type InMemoryApplicationRepository struct {
	mu           sync.RWMutex
	applications map[string]*types.Application
}

// NewInMemoryApplicationRepository creates a new, empty in-memory application repository.
func NewInMemoryApplicationRepository() *InMemoryApplicationRepository {
	return &InMemoryApplicationRepository{
		applications: make(map[string]*types.Application),
	}
}

// CreateApplication adds a new application to the in-memory store.
func (r *InMemoryApplicationRepository) CreateApplication(ctx context.Context, app *types.Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.applications[app.ID]; exists {
		return types.ErrConflict.WithDetails(map[string]string{"id": app.ID})
	}
	r.applications[app.ID] = app
	return nil
}

// GetApplicationByID retrieves an application by its ID from the in-memory store.
func (r *InMemoryApplicationRepository) GetApplicationByID(ctx context.Context, id string) (*types.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	app, exists := r.applications[id]
	if !exists {
		return nil, types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	return app, nil
}

// GetApplicationByName retrieves an application by its name from the in-memory store.
func (r *InMemoryApplicationRepository) GetApplicationByName(ctx context.Context, name string) (*types.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, app := range r.applications {
		if app.Name == name {
			return app, nil
		}
	}
	return nil, types.ErrNotFound.WithDetails(map[string]string{"name": name})
}

// UpdateApplication updates an existing application in the in-memory store.
func (r *InMemoryApplicationRepository) UpdateApplication(ctx context.Context, app *types.Application) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.applications[app.ID]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": app.ID})
	}
	r.applications[app.ID] = app
	return nil
}

// DeleteApplication removes an application from the in-memory store.
func (r *InMemoryApplicationRepository) DeleteApplication(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.applications[id]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	delete(r.applications, id)
	return nil
}

// ListApplications returns a paginated list of all applications from the in-memory store.
func (r *InMemoryApplicationRepository) ListApplications(ctx context.Context, pq types.PaginationQuery) ([]*types.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	apps := make([]*types.Application, 0, len(r.applications))
	for _, app := range r.applications {
		apps = append(apps, app)
	}
	start, end := pq.Offset, pq.Offset+pq.PageSize
	if start > len(apps) {
		return []*types.Application{}, nil
	}
	if end > len(apps) {
		end = len(apps)
	}
	return apps[start:end], nil
}