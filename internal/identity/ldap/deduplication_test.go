package ldap

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/pkg/types"
	"testing"
)

func TestDeduplicator_EmailMatch(t *testing.T) {
	dedup := NewDeduplicator([]DeduplicationRule{
		{MatchFields: []string{"email"}, Priority: 1},
	}, &ConflictManager{})

	users := []*types.User{
		{Username: "user1", Email: "test@example.com"},
		{Username: "user2", Email: "test@example.com"},
	}

	result, err := dedup.Process(context.Background(), users)
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, "user1", result[0].Username)
}
