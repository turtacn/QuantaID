package engine

import (
	"context"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/turtacn/QuantaID/internal/domain/policy"
)

// DBRBACProvider is an implementation of RBACProvider that uses a database repository.
type DBRBACProvider struct {
	repo  policy.RBACRepository
	cache *cache.Cache
}

// NewDBRBACProvider creates a new DBRBACProvider.
func NewDBRBACProvider(repo policy.RBACRepository) *DBRBACProvider {
	return &DBRBACProvider{
		repo:  repo,
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

// IsAllowed checks if a subject has permission to perform an action on a resource.
func (p *DBRBACProvider) IsAllowed(ctx context.Context, subjectID, action, resource string) (bool, error) {
	permissions, err := p.getPermissionsForUser(ctx, subjectID)
	if err != nil {
		return false, err
	}

	for perm := range permissions {
		if p.match(perm, action, resource) {
			return true, nil
		}
	}

	return false, nil
}

// getPermissionsForUser retrieves a user's permissions, using a cache to avoid database lookups.
func (p *DBRBACProvider) getPermissionsForUser(ctx context.Context, subjectID string) (map[string]struct{}, error) {
	if perms, found := p.cache.Get(subjectID); found {
		return perms.(map[string]struct{}), nil
	}

	roles, err := p.repo.GetRolesForUser(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	permissions := make(map[string]struct{})
	for _, role := range roles {
		for _, perm := range role.Permissions {
			permissions[perm.Resource+":"+perm.Action] = struct{}{}
		}
	}

	p.cache.Set(subjectID, permissions, cache.DefaultExpiration)
	return permissions, nil
}

// match checks if the requested action and resource match a permission.
// This could be extended to support wildcards.
func (p *DBRBACProvider) match(permission, action, resource string) bool {
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return false
	}
	// Simple string matching for now
	return parts[0] == resource && parts[1] == action
}
