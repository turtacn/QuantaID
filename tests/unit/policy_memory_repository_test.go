package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/storage/memory"
	"github.com/turtacn/QuantaID/pkg/types"
)

func TestPolicyMemoryRepository_CRUD(t *testing.T) {
	repo := memory.NewPolicyMemoryRepository()
	ctx := context.Background()
	policy := &types.Policy{
		Description: "Test Policy",
		Subjects:    []string{"user:123"},
		Actions:     []string{"read"},
		Resources:   []string{"/data/resource1"},
	}

	// Create
	err := repo.CreatePolicy(ctx, policy)
	assert.NoError(t, err)
	assert.NotEmpty(t, policy.ID)

	// Get
	retrievedPolicy, err := repo.GetPolicyByID(ctx, policy.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Test Policy", retrievedPolicy.Description)

	// Update
	policy.Description = "Updated Policy Description"
	err = repo.UpdatePolicy(ctx, policy)
	assert.NoError(t, err)
	retrievedPolicy, _ = repo.GetPolicyByID(ctx, policy.ID)
	assert.Equal(t, "Updated Policy Description", retrievedPolicy.Description)

	// Delete
	err = repo.DeletePolicy(ctx, policy.ID)
	assert.NoError(t, err)
	_, err = repo.GetPolicyByID(ctx, policy.ID)
	assert.Error(t, err)
}

func TestPolicyMemoryRepository_FindPolicies(t *testing.T) {
	repo := memory.NewPolicyMemoryRepository()
	ctx := context.Background()

	p1 := &types.Policy{Subjects: []string{"user:1"}, Actions: []string{"read"}, Resources: []string{"res:1"}}
	p2 := &types.Policy{Subjects: []string{"user:2"}, Actions: []string{"write"}, Resources: []string{"res:1"}}
	p3 := &types.Policy{Subjects: []string{"user:1", "user:2"}, Actions: []string{"read"}, Resources: []string{"res:2"}}
	repo.CreatePolicy(ctx, p1)
	repo.CreatePolicy(ctx, p2)
	repo.CreatePolicy(ctx, p3)

	// Find by Subject
	policies, err := repo.FindPoliciesForSubject(ctx, "user:1")
	assert.NoError(t, err)
	assert.Len(t, policies, 2)

	// Find by Resource
	policies, err = repo.FindPoliciesForResource(ctx, "res:1")
	assert.NoError(t, err)
	assert.Len(t, policies, 2)

	// Find by Action
	policies, err = repo.FindPoliciesForAction(ctx, "write")
	assert.NoError(t, err)
	assert.Len(t, policies, 1)
	assert.Equal(t, p2.ID, policies[0].ID)
}
