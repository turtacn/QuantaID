package oauth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/pkg/types"
)

func TestClaimsMapper_BasicMapping(t *testing.T) {
	rules := []MappingRule{
		{
			Scope:  "profile",
			Claims: []string{"username", "email"},
		},
	}
	mapper := NewClaimsMapper(rules)
	user := &types.User{
		ID:       "user123",
		Username: "testuser",
		Email:    "test@example.com",
	}
	scopes := []string{"profile"}

	claims, err := mapper.MapClaims(context.Background(), user, scopes)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", claims["username"])
	assert.Equal(t, "test@example.com", claims["email"])
	assert.Equal(t, "user123", claims["sub"])
}

func TestClaimsMapper_TransformFunction(t *testing.T) {
	rules := []MappingRule{
		{
			Scope:      "email",
			Claims:     []string{"email"},
			Transforms: map[string]string{"email": "mask_email"},
		},
	}
	mapper := NewClaimsMapper(rules)
	user := &types.User{
		ID:    "user123",
		Email: "test@example.com",
	}
	scopes := []string{"email"}

	claims, err := mapper.MapClaims(context.Background(), user, scopes)
	assert.NoError(t, err)
	assert.Equal(t, "t***@example.com", claims["email"])
}
