package adaptive

import (
	"context"
	"sort"
	"time"
)

type PolicyEngine struct {
	policies  []Policy
	evaluator *ConditionEvaluator
}

type Policy struct {
	Name       string      `yaml:"name"`
	Priority   int         `yaml:"priority"`
	Conditions []Condition `yaml:"conditions"`
	Actions    []Action    `yaml:"actions"`
}

type Condition struct {
	Type     string      `yaml:"type"`
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
	Values   []string    `yaml:"values"`
}

type Action struct {
	RequireMFA     []string `yaml:"require_mfa"`
	DenyMFA        []string `yaml:"deny_mfa"`
	SkipMFA        bool     `yaml:"skip_mfa"`
	RequireApproval bool    `yaml:"require_approval"`
	NotifyUser     bool     `yaml:"notify_user"`
	AlertAdmin     bool     `yaml:"alert_admin"`
	TemporaryBlock string   `yaml:"temporary_block"`
}

type AuthContext struct {
	UserRoles     []string
	RiskScore     *RiskScore
	Timestamp     time.Time
	DeviceTrusted bool
}

type PolicyDecision struct {
	MatchedPolicies []string
	Actions         []Action
}

func (pe *PolicyEngine) Evaluate(ctx context.Context, authCtx *AuthContext) (*PolicyDecision, error) {
	sort.Slice(pe.policies, func(i, j int) bool {
		return pe.policies[i].Priority < pe.policies[j].Priority
	})

	decision := &PolicyDecision{
		Actions: []Action{},
	}

	for _, policy := range pe.policies {
		matched := true
		for _, condition := range policy.Conditions {
			if !pe.evaluator.EvaluateCondition(condition, authCtx) {
				matched = false
				break
			}
		}

		if matched {
			decision.MatchedPolicies = append(decision.MatchedPolicies, policy.Name)
			decision.Actions = append(decision.Actions, policy.Actions...)
		}
	}

	return decision, nil
}

type ConditionEvaluator struct{}

func (ce *ConditionEvaluator) EvaluateCondition(cond Condition, authCtx *AuthContext) bool {
	switch cond.Type {
	case "role":
		// return ce.evaluateRoleCondition(cond, authCtx.UserRoles)
	case "risk_score":
		// return ce.evaluateNumericCondition(cond, authCtx.RiskScore.TotalScore)
	case "time":
		// return ce.evaluateTimeCondition(cond, authCtx.Timestamp)
	case "device_trusted":
		// return ce.evaluateBoolCondition(cond, authCtx.DeviceTrusted)
	}
	return false
}
