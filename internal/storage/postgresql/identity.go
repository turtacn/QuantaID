package postgresql

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
)

// InMemoryIdentityRepository provides an in-memory implementation of the identity-related repositories,
// NOTE: Despite the package name 'postgresql', this is an IN-MEMORY implementation,
type InMemoryIdentityRepository struct {
	mu               sync.RWMutex
	users            map[string]*types.User
	groups           map[string]*types.UserGroup
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

func (r *InMemoryIdentityRepository) CreateUser(ctx context.Context, user *types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; exists {
		return types.ErrConflict.WithDetails(map[string]string{"id": user.ID})
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryIdentityRepository) GetUserByID(ctx context.Context, id string) (*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[id]
	if !exists {
		return nil, types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	return user, nil
}
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

func (r *InMemoryIdentityRepository) UpdateUser(ctx context.Context, user *types.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": user.ID})
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryIdentityRepository) DeleteUser(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[id]; !exists {
		return types.ErrNotFound.WithDetails(map[string]string{"id": id})
	}
	delete(r.users, id)
	return nil
}

func (r *InMemoryIdentityRepository) ListUsers(ctx context.Context, pq identity.PaginationQuery) ([]*types.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*types.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
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

func (r *InMemoryIdentityRepository) CreateGroup(ctx context.Context, group *types.UserGroup) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[group.ID]; exists {
		return types.ErrConflict
	}
	r.groups[group.ID] = group
	return nil
}
func (r *InMemoryIdentityRepository) GetGroupByID(ctx context.Context, id string) (*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	group, exists := r.groups[id]
	if !exists {
		return nil, types.ErrNotFound
	}
	return group, nil
}
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
func (r *InMemoryIdentityRepository) UpdateGroup(ctx context.Context, group *types.UserGroup) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[group.ID]; !exists {
		return types.ErrNotFound
	}
	r.groups[group.ID] = group
	return nil
}
func (r *InMemoryIdentityRepository) DeleteGroup(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[id]; !exists {
		return types.ErrNotFound
	}
	delete(r.groups, id)
	delete(r.groupMemberships, id)
	return nil
}
func (r *InMemoryIdentityRepository) ListGroups(ctx context.Context, pq identity.PaginationQuery) ([]*types.UserGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	groups := make([]*types.UserGroup, 0, len(r.groups))
	for _, group := range r.groups {
		groups = append(groups, group)
	}
	return groups, nil
}
func (r *InMemoryIdentityRepository) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[userID]; !ok {
		return fmt.Errorf("user not found")
	}
	if _, ok := r.groups[groupID]; !ok {
		return fmt.Errorf("group not found")
	}

	members := r.groupMemberships[groupID]
	for _, memberID := range members {
		if memberID == userID {
			return nil
		} // already a member
	}
	r.groupMemberships[groupID] = append(members, userID)
	return nil
}

func (r *InMemoryIdentityRepository) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	members, ok := r.groupMemberships[groupID]
	if !ok {
		return nil
	}

	var newMembers []string
	for _, memberID := range members {
		if memberID != userID {
			newMembers = append(newMembers, memberID)
		}
	}
	r.groupMemberships[groupID] = newMembers
	return nil
}
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
