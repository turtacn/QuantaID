package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/pkg/types"
)

func TestIdentityMemoryRepository_CreateAndGetUser(t *testing.T) {
	repo := memory.NewIdentityMemoryRepository()
	ctx := context.Background()
	user := &types.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	err := repo.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)

	retrievedUser, err := repo.GetUserByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Username, retrievedUser.Username)

	retrievedUser, err = repo.GetUserByUsername(ctx, user.Username)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUser.ID)

	retrievedUser, err = repo.GetUserByEmail(ctx, string(user.Email))
    assert.NoError(t, err)
    assert.Equal(t, user.ID, retrievedUser.ID)
}

func TestIdentityMemoryRepository_UpdateUser(t *testing.T) {
	repo := memory.NewIdentityMemoryRepository()
	ctx := context.Background()
	user := &types.User{Username: "testuser", Email: "test@example.com"}
	repo.CreateUser(ctx, user)

	user.Email = "updated@example.com"
	err := repo.UpdateUser(ctx, user)
	assert.NoError(t, err)

	retrievedUser, _ := repo.GetUserByID(ctx, user.ID)
	assert.Equal(t, types.EncryptedString("updated@example.com"), retrievedUser.Email)
}

func TestIdentityMemoryRepository_DeleteUser(t *testing.T) {
	repo := memory.NewIdentityMemoryRepository()
	ctx := context.Background()
	user := &types.User{Username: "testuser", Email: "test@example.com"}
	repo.CreateUser(ctx, user)

	err := repo.DeleteUser(ctx, user.ID)
	assert.NoError(t, err)

	_, err = repo.GetUserByID(ctx, user.ID)
	assert.Error(t, err)
}

func TestIdentityMemoryRepository_CreateAndGetGroup(t *testing.T) {
    repo := memory.NewIdentityMemoryRepository()
    ctx := context.Background()
    group := &types.UserGroup{
        Name: "testgroup",
    }

    err := repo.CreateGroup(ctx, group)
    assert.NoError(t, err)
    assert.NotEmpty(t, group.ID)

    retrievedGroup, err := repo.GetGroupByID(ctx, group.ID)
    assert.NoError(t, err)
    assert.Equal(t, group.Name, retrievedGroup.Name)

    retrievedGroup, err = repo.GetGroupByName(ctx, group.Name)
    assert.NoError(t, err)
    assert.Equal(t, group.ID, retrievedGroup.ID)
}

func TestIdentityMemoryRepository_AddAndRemoveUserFromGroup(t *testing.T) {
    repo := memory.NewIdentityMemoryRepository()
    ctx := context.Background()
    user := &types.User{Username: "testuser", Email: "test@example.com"}
    repo.CreateUser(ctx, user)
    group := &types.UserGroup{Name: "testgroup"}
    repo.CreateGroup(ctx, group)

    err := repo.AddUserToGroup(ctx, user.ID, group.ID)
    assert.NoError(t, err)

    userGroups, err := repo.GetUserGroups(ctx, user.ID)
    assert.NoError(t, err)
    assert.Len(t, userGroups, 1)
    assert.Equal(t, group.ID, userGroups[0].ID)

    err = repo.RemoveUserFromGroup(ctx, user.ID, group.ID)
    assert.NoError(t, err)

    userGroups, err = repo.GetUserGroups(ctx, user.ID)
    assert.NoError(t, err)
    assert.Len(t, userGroups, 0)
}
