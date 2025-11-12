package postgresql

import (
	"context"
	"fmt"
	"strings"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
)

// InMemoryIdentityRepository provides an in-memory implementation of the identity-related repositories,
// specifically the UserRepository and GroupRepository.
// NOTE: Despite the package name 'postgresql', this is an IN-MEMORY implementation,
// likely used for testing or simple, non-persistent deployments. It uses maps and slices
// with a mutex for thread-safe operations.
type InMemoryIdentityRepository struct {
	mu     sync.RWMutex
	users  map[string]*types.User
	groups map[string]*types.UserGroup
	// groupMemberships is an in-memory representation of the many-to-many relationship.
	groupMemberships map[string][]string // groupID -> []userID
}

// NewInMemoryIdentityRepository creates a new, empty in-memory identity repository.
func NewInMemoryIdentityRepository() *InMemoryIdentityRepository {
	return &InMemoryIdentityRepository{
		users:            make(map[string]*types.User),
		groups:           make(map[string]*types.UserGroup),
		groupMemberships: make(map[string][]string),
	}
}

// --- UserRepository Implementation ---

// CreateUser adds a new user to the in-memory store.
func (r *InMemoryIdentityRepository) CreateUser(ctx context.Context, user *types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; exists {
		return types.ErrConflict.WithDetails(map[string]string{"id": user.ID})
	}
	r.users[user.ID] = user
	return nil
}

// GetUserByID retrieves a user by their ID from the in-memory store.
func (r *InMemoryIdentityRepository) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[id]
	if !exists {
		return nil, types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	return user, nil
}

// GetUserByUsername searches for a user by their username in the in-memory store.
func (r *InMemoryIdentityRepository) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, types.ErrUserNotFound
}

// GetUserByEmail searches for a user by their email in the in-memory store.
func (r *InMemoryIdentityRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, types.ErrNotFound.WithDetails(map[string]string{"email": email})
}

// UpdateUser updates an existing user in the in-memory store.
func (r *InMemoryIdentityRepository) UpdateUser(ctx context.Context, user *types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": user.ID})
	}
	r.users[user.ID] = user
	return nil
}

// DeleteUser removes a user from the in-memory store.
func (r *InMemoryIdentityRepository) DeleteUser(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[id]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	delete(r.users, id)
	return nil
}


func (r *InMemoryIdentityRepository) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filteredUsers []*types.User
	for _, user := range r.users {
		if filter.Search == "" || (strings.Contains(user.Username, filter.Search) || strings.Contains(user.Email, filter.Search)) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	total := int64(len(filteredUsers))

	start := filter.Offset
	if start > len(filteredUsers) {
		start = len(filteredUsers)
	}

	end := start + filter.Limit
	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}

	return filteredUsers[start:end], total, nil
}

// FindUsersByAttribute searches for users with a matching attribute value in the in-memory store.
func (r *InMemoryIdentityRepository) FindUsersByAttribute(ctx context.Context, attribute string, value interface{}) ([]*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var foundUsers []*types.User
	for _, user := range r.users {
		if val, ok := user.Attributes[attribute]; ok && val == value {
			foundUsers = append(foundUsers, user)
		}
	}
	return foundUsers, nil
}

// --- GroupRepository Implementation ---

// CreateGroup adds a new group to the in-memory store.
func (r *InMemoryIdentityRepository) CreateGroup(ctx context.Context, group *types.UserGroup) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[group.ID]; exists { return types.ErrConflict }
	r.groups[group.ID] = group
	return nil
}

// GetGroupByID retrieves a group by its ID from the in-memory store.
func (r *InMemoryIdentityRepository) GetGroupByID(ctx context.Context, id string) (*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	group, exists := r.groups[id]
	if !exists { return nil, types.ErrNotFound }
	return group, nil
}

// GetGroupByName searches for a group by its name in the in-memory store.
func (r *InMemoryIdentityRepository) GetGroupByName(ctx context.Context, name string) (*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, group := range r.groups {
		if group.Name == name {
			return group, nil
		}
	}
	return nil, types.ErrNotFound
}

// UpdateGroup updates an existing group in the in-memory store.
func (r *InMemoryIdentityRepository) UpdateGroup(ctx context.Context, group *types.UserGroup) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[group.ID]; !exists { return types.ErrNotFound }
	r.groups[group.ID] = group
	return nil
}

// DeleteGroup removes a group and its memberships from the in-memory store.
func (r *InMemoryIdentityRepository) DeleteGroup(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[id]; !exists { return types.ErrNotFound }
	delete(r.groups, id)
	delete(r.groupMemberships, id)
	return nil
}

// ListGroups returns a paginated list of all groups from the in-memory store.
func (r *InMemoryIdentityRepository) ListGroups(ctx context.Context) ([]*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	groups := make([]*types.UserGroup, 0, len(r.groups))
	for _, group := range r.groups {
		groups = append(groups, group)
	}
	return groups, nil
}

// AddUserToGroup creates a membership link between a user and a group in the in-memory store.
func (r *InMemoryIdentityRepository) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[userID]; !ok { return fmt.Errorf("user not found") }
	if _, ok := r.groups[groupID]; !ok { return fmt.Errorf("group not found") }

	members := r.groupMemberships[groupID]
	for _, memberID := range members {
		if memberID == userID { return nil } // already a member
	}
	r.groupMemberships[groupID] = append(members, userID)
	return nil
}

// RemoveUserFromGroup removes a membership link between a user and a group in the in-memory store.
func (r *InMemoryIdentityRepository) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	members, ok := r.groupMemberships[groupID]
	if !ok { return nil }

	var newMembers []string
	for _, memberID := range members {
		if memberID != userID {
			newMembers = append(newMembers, memberID)
		}
	}
	r.groupMemberships[groupID] = newMembers
	return nil
}

// GetUserGroups retrieves all groups a user is a member of from the in--memory store.
func (r *InMemoryIdentityRepository) GetUserGroups(ctx context.Context, userID string) ([]*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var userGroups []*types.UserGroup
	for groupID, members := range r.groupMemberships {
		for _, memberID := range members {
			if memberID == userID {
				if group, ok := r.groups[groupID]; ok {
					userGroups = append(userGroups, group)
				}
			}
		}
	}
	return userGroups, nil
}
