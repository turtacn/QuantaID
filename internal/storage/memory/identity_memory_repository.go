package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
)

// IdentityMemoryRepository provides an in-memory implementation of the identity repositories.
type IdentityMemoryRepository struct {
	mu         sync.RWMutex
	users      map[string]*types.User
	groups     map[string]*types.UserGroup
	userGroups map[string]map[string]struct{} // map[userID] -> set of groupIDs
}

// NewIdentityMemoryRepository creates a new in-memory identity repository.
func NewIdentityMemoryRepository() *IdentityMemoryRepository {
	return &IdentityMemoryRepository{
		users:      make(map[string]*types.User),
		groups:     make(map[string]*types.UserGroup),
		userGroups: make(map[string]map[string]struct{}),
	}
}

// --- UserRepository implementation ---

func (r *IdentityMemoryRepository) CreateUser(ctx context.Context, user *types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Username == user.Username {
			return fmt.Errorf("username '%s' already exists", user.Username)
		}
		if u.Email == user.Email {
			return fmt.Errorf("email '%s' already exists", user.Email)
		}
	}

	user.ID = uuid.New().String()
	r.users[user.ID] = user
	return nil
}

func (r *IdentityMemoryRepository) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user with ID '%s' not found", id) // Replace with domain error later
	}
	return user, nil
}

func (r *IdentityMemoryRepository) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user with username '%s' not found", username)
}

func (r *IdentityMemoryRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user with email '%s' not found", email)
}

func (r *IdentityMemoryRepository) UpdateUser(ctx context.Context, user *types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[user.ID]; !ok {
		return fmt.Errorf("user with ID '%s' not found for update", user.ID)
	}
	r.users[user.ID] = user
	return nil
}

func (r *IdentityMemoryRepository) DeleteUser(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return fmt.Errorf("user with ID '%s' not found for deletion", id)
	}
	delete(r.users, id)
	delete(r.userGroups, id) // Also remove user's group memberships
	return nil
}

func (r *IdentityMemoryRepository) ListUsers(ctx context.Context, pq identity.PaginationQuery) ([]*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*types.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
    // Note: Simple implementation without sorting. Pagination on a map is not deterministic.
    start := pq.Offset
    end := start + pq.PageSize

    if start > len(users) {
        return []*types.User{}, nil
    }
    if end > len(users) {
        end = len(users)
    }

	return users[start:end], nil
}

func (r *IdentityMemoryRepository) FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error) {
    // This is a simplified implementation. A real-world scenario might require more complex reflection.
    r.mu.RLock()
    defer r.mu.RUnlock()

    var foundUsers []*types.User
    for _, user := range r.users {
        match := false
        switch attribute {
        case "email":
            if email, ok := value.(string); ok && user.Email == email {
                match = true
            }
        case "username":
            if username, ok := value.(string); ok && user.Username == username {
                match = true
            }
        }
        if match {
            foundUsers = append(foundUsers, user)
        }
    }
    return foundUsers, nil
}


// --- GroupRepository implementation ---

func (r *IdentityMemoryRepository) CreateGroup(ctx context.Context, group *types.UserGroup) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, g := range r.groups {
		if g.Name == group.Name {
			return fmt.Errorf("group with name '%s' already exists", group.Name)
		}
	}

	group.ID = uuid.New().String()
	r.groups[group.ID] = group
	return nil
}

func (r *IdentityMemoryRepository) GetGroupByID(ctx context.Context, id string) (*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	group, ok := r.groups[id]
	if !ok {
		return nil, fmt.Errorf("group with ID '%s' not found", id)
	}
	return group, nil
}

func (r *IdentityMemoryRepository) GetGroupByName(ctx context.Context, name string) (*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, group := range r.groups {
		if group.Name == name {
			return group, nil
		}
	}
	return nil, fmt.Errorf("group with name '%s' not found", name)
}

func (r *IdentityMemoryRepository) UpdateGroup(ctx context.Context, group *types.UserGroup) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.groups[group.ID]; !ok {
		return fmt.Errorf("group with ID '%s' not found for update", group.ID)
	}
	r.groups[group.ID] = group
	return nil
}

func (r *IdentityMemoryRepository) DeleteGroup(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.groups[id]; !ok {
		return fmt.Errorf("group with ID '%s' not found for deletion", id)
	}
	delete(r.groups, id)

	// Remove this group from all users' memberships
	for userID := range r.userGroups {
		delete(r.userGroups[userID], id)
	}
	return nil
}

func (r *IdentityMemoryRepository) ListGroups(ctx context.Context, pq identity.PaginationQuery) ([]*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	groups := make([]*types.UserGroup, 0, len(r.groups))
	for _, group := range r.groups {
		groups = append(groups, group)
	}
    // Note: Simple implementation without sorting.
    start := pq.Offset
    end := start + pq.PageSize

    if start > len(groups) {
        return []*types.UserGroup{}, nil
    }
    if end > len(groups) {
        end = len(groups)
    }
	return groups[start:end], nil
}

func (r *IdentityMemoryRepository) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[userID]; !ok {
		return fmt.Errorf("user with ID '%s' not found", userID)
	}
	if _, ok := r.groups[groupID]; !ok {
		return fmt.Errorf("group with ID '%s' not found", groupID)
	}

	if _, ok := r.userGroups[userID]; !ok {
		r.userGroups[userID] = make(map[string]struct{})
	}
	r.userGroups[userID][groupID] = struct{}{}
	return nil
}

func (r *IdentityMemoryRepository) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.userGroups[userID]; ok {
		delete(r.userGroups[userID], groupID)
	}
	return nil
}

func (r *IdentityMemoryRepository) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	groupIDs, ok := r.userGroups[userID]
	if !ok {
		return []*types.UserGroup{}, nil
	}

	groups := make([]*types.UserGroup, 0, len(groupIDs))
	for groupID := range groupIDs {
		if group, ok := r.groups[groupID]; ok {
			groups = append(groups, group)
		}
	}
	return groups, nil
}

func (r *IdentityMemoryRepository) UpsertBatch(ctx context.Context, users []*types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range users {
		var existingUser *types.User
		for _, u := range r.users {
			if u.Email == user.Email {
				existingUser = u
				break
			}
		}

		if existingUser != nil {
			existingUser.Username = user.Username
			existingUser.Attributes = user.Attributes
			existingUser.Status = user.Status
		} else {
			user.ID = uuid.New().String()
			r.users[user.ID] = user
		}
	}
	return nil
}
