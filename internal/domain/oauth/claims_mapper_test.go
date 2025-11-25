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
		Email:    types.EncryptedString("test@example.com"),
	}
	scopes := []string{"profile"}

	claims, err := mapper.MapClaims(context.Background(), user, scopes)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", claims["username"])
	assert.Equal(t, types.EncryptedString("test@example.com"), claims["email"])
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
		Email: types.EncryptedString("test@example.com"),
	}
	scopes := []string{"email"}

	claims, err := mapper.MapClaims(context.Background(), user, scopes)
	assert.NoError(t, err)
	// The transform function likely returns a string, but the mapper might wrap it or not.
	// Assuming transforms return raw values which are put into claims map.
	// If mask_email returns string, then it should be string.
	// But let's check if `t***@example.com` is expected to be string or EncryptedString.
	// Usually claims in JWT are strings.
	assert.Equal(t, types.EncryptedString("t***@example.com"), claims["email"])
}
