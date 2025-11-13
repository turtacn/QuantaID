package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/services/authorization"
)

func TestDefaultEvaluator(t *testing.T) {
	rules := []authorization.Rule{
		{
			Name:     "allow-admins-to-read-dashboard",
			Effect:   policy.DecisionAllow,
			Actions:  []string{"dashboard.read"},
			Subjects: []string{"group:admins"},
		},
		{
			Name:        "allow-access-from-whitelist",
			Effect:      policy.DecisionAllow,
			Actions:     []string{"api.read"},
			Subjects:    []string{"user:123"},
			IPWhitelist: []string{"192.168.1.10/32"},
		},
		{
			Name:       "allow-access-during-business-hours",
			Effect:     policy.DecisionAllow,
			Actions:    []string{"api.write"},
			Subjects:   []string{"user:123"},
			TimeRanges: []authorization.TimeRange{{Start: "09:00", End: "17:00"}},
		},
		{
			Name:     "deny-all-by-default",
			Effect:   policy.DecisionDeny,
			Actions:  []string{"*"},
			Subjects: []string{"*"},
		},
	}
	evaluator := authorization.NewDefaultEvaluator(rules)
	ctx := context.Background()

	testCases := []struct {
		Name          string
		EvalContext   policy.EvaluationContext
		Expected      policy.Decision
	}{
		{
			Name: "admin can read dashboard",
			EvalContext: policy.EvaluationContext{
				Subject: policy.Subject{UserID: "456", Groups: []string{"admins"}},
				Action:  "dashboard.read",
			},
			Expected: policy.DecisionAllow,
		},
		{
			Name: "non-admin cannot read dashboard",
			EvalContext: policy.EvaluationContext{
				Subject: policy.Subject{UserID: "123", Groups: []string{"users"}},
				Action:  "dashboard.read",
			},
			Expected: policy.DecisionDeny,
		},
		{
			Name: "user with whitelisted IP can read api",
			EvalContext: policy.EvaluationContext{
				Subject:     policy.Subject{UserID: "123"},
				Action:      "api.read",
				Environment: policy.Environment{IP: "192.168.1.10"},
			},
			Expected: policy.DecisionAllow,
		},
		{
			Name: "user with non-whitelisted IP cannot read api",
			EvalContext: policy.EvaluationContext{
				Subject:     policy.Subject{UserID: "123"},
				Action:      "api.read",
				Environment: policy.Environment{IP: "10.0.0.5"},
			},
			Expected: policy.DecisionDeny,
		},
		{
			Name: "user can write to api during business hours",
			EvalContext: policy.EvaluationContext{
				Subject:     policy.Subject{UserID: "123"},
				Action:      "api.write",
				Environment: policy.Environment{Time: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
			},
			Expected: policy.DecisionAllow,
		},
		{
			Name: "user cannot write to api outside of business hours",
			EvalContext: policy.EvaluationContext{
				Subject:     policy.Subject{UserID: "123"},
				Action:      "api.write",
				Environment: policy.Environment{Time: time.Date(2023, 1, 1, 20, 0, 0, 0, time.UTC)},
			},
			Expected: policy.DecisionDeny,
		},
		{
			Name: "wildcard subject denies any other action",
			EvalContext: policy.EvaluationContext{
				Subject: policy.Subject{UserID: "789", Groups: []string{"auditors"}},
				Action:  "some.other.action",
			},
			Expected: policy.DecisionDeny,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			decision, err := evaluator.Evaluate(ctx, tc.EvalContext)
			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, decision)
		})
	}
}
