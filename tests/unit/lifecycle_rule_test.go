package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/domain/identity/lifecycle"
	"github.com/turtacn/QuantaID/pkg/types"
)

func Test_Rule_LastLogin_gt_90days(t *testing.T) {
	engine := lifecycle.NewEngine()

	// 91 days ago
	lastLogin := time.Now().Add(-91 * 24 * time.Hour)
	user := &types.User{
		ID:          "u1",
		Status:      types.UserStatusActive,
		LastLoginAt: &lastLogin,
	}

	rules := []lifecycle.LifecycleRule{
		{
			Name: "Inactive",
			Conditions: []lifecycle.Condition{
				{
					Attribute: "lastLoginAt",
					Operator:  lifecycle.OpGt,
					Value:     "2160h", // 90 days
				},
			},
			Actions: []lifecycle.Action{
				{Type: lifecycle.ActionDisable},
			},
		},
	}

	actions, err := engine.Evaluate(user, rules)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	assert.Equal(t, lifecycle.ActionDisable, actions[0].Type)
}

func Test_Rule_AccountExpired(t *testing.T) {
	engine := lifecycle.NewEngine()

	// Expired yesterday
	// Note: We use "attributes.expiredAt" because user struct might not have ExpiredAt
	// And our engine handles map lookups.
	// However, we also need to store it as a string that can be parsed,
	// or assume the engine handles specific types from JSON deserialization.
	// The engine uses 'toTime' helper if it's a Time object, but attributes map usually has strings/numbers/bools if coming from JSON.
	// If it's constructed in Go code, we can put time.Time.

	expiredTime := time.Now().Add(-24 * time.Hour)

	user := &types.User{
		ID:     "u2",
		Status: types.UserStatusActive,
		Attributes: map[string]interface{}{
			"expiredAt": expiredTime, // In-memory this is time.Time.
		},
	}

	// Rule: If attribute.expiredAt < Now (or older than 0s ago? No, absolute time check)
	// My engine implementation has ambiguity on time comparisons.
	// For "lastLoginAt gt 90d", I implemented "Age > 90d".
	// For "expiredAt lt Now", I need to handle absolute time comparison.
	// My `compare` function handles `toTime`.

	// Wait, my `compare` function in `engine.go`:
	// if dExpected, ok := parseDuration(expected); ok { ... compares age ... }

	// If I pass "0s" as value, it parses as duration 0.
	// Age > 0s means "Happened in the past".
	// If ExpiredAt is yesterday, Age is +24h.
	// If condition is "expiredAt gt 0s", then Age(24h) > 0s is TRUE. Meaning it IS expired (date is in past).
	// If condition is "expiredAt lt 0s", then Age(24h) < 0s is FALSE.

	// The requirement: "Test_Rule_AccountExpired: 构造 ExpiredAt < Now 的用户，验证返回 Action: Delete".
	// ExpiredAt < Now means ExpiredAt is in the past.
	// Age = Now - ExpiredAt > 0.
	// So we want to check if Age > 0.
	// Operator should be GT, Value "0s".

	// Let's adjust the test rule to match my engine logic or adjust engine.
	// Standard "lt" on dates usually means "earlier than".
	// ExpiredAt (Yesterday) < Now. This is True.
	// My engine sees "0s" as duration.
	// So it compares Age vs Duration.
	// Age (positive) vs 0.
	// I want to say "If expiredAt is in the past".
	// That is Age > 0.

	rules := []lifecycle.LifecycleRule{
		{
			Name: "Expired",
			Conditions: []lifecycle.Condition{
				{
					Attribute: "attributes.expiredAt",
					Operator:  lifecycle.OpGt, // Age > 0s
					Value:     "0s",
				},
			},
			Actions: []lifecycle.Action{
				{Type: lifecycle.ActionDelete},
			},
		},
	}

	actions, err := engine.Evaluate(user, rules)
	assert.NoError(t, err)
	assert.Len(t, actions, 1)
	assert.Equal(t, lifecycle.ActionDelete, actions[0].Type)
}
