package session

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// PolicyRule defines a rule for session evaluation.
type PolicyRule struct {
	ID         string
	Name       string
	Priority   int // Lower number means higher priority
	Conditions []PolicyCondition
	Action     ActionType
	Reason     string
}

// PolicyCondition defines a condition that must be met for a rule to apply.
type PolicyCondition struct {
	Field    string      // risk_level, risk_score, signal_count, inactive_minutes, signals
	Operator string      // ==, !=, >, <, >=, <=, contains
	Value    interface{}
}

// SessionPolicy manages the rules and evaluation logic.
type SessionPolicy struct {
	rules []PolicyRule
	mu    sync.RWMutex
}

// NewSessionPolicy creates a new SessionPolicy engine.
func NewSessionPolicy(rules []PolicyRule) *SessionPolicy {
	return &SessionPolicy{
		rules: rules,
	}
}

// DefaultPolicyRules returns a set of default policy rules.
func DefaultPolicyRules() []PolicyRule {
	return []PolicyRule{
		{
			ID:       "critical_risk_terminate",
			Name:     "Terminate Critical Risk Sessions",
			Priority: 1,
			Conditions: []PolicyCondition{
				{Field: "risk_level", Operator: "==", Value: "critical"},
			},
			Action: ActionTerminate,
			Reason: "Risk level reached critical threshold",
		},
		{
			ID:       "high_risk_downgrade",
			Name:     "Downgrade High Risk Sessions",
			Priority: 2,
			Conditions: []PolicyCondition{
				{Field: "risk_level", Operator: "==", Value: "high"},
			},
			Action: ActionDowngrade,
			Reason: "High risk level detected, restricting permissions",
		},
		{
			ID:       "medium_risk_mfa",
			Name:     "Require MFA for Medium Risk",
			Priority: 3,
			Conditions: []PolicyCondition{
				{Field: "risk_level", Operator: "==", Value: "medium"},
			},
			Action: ActionRequireMFA,
			Reason: "Suspicious activity detected, verification required",
		},
		{
			ID:       "geo_jump_mfa",
			Name:     "Require MFA on Geo Jump",
			Priority: 4,
			Conditions: []PolicyCondition{
				{Field: "signals", Operator: "contains", Value: "geo_jump"},
			},
			Action: ActionRequireMFA,
			Reason: "Abnormal geographical movement detected",
		},
		{
			ID:       "long_inactive_reauth",
			Name:     "Re-auth after Long Inactivity",
			Priority: 5,
			Conditions: []PolicyCondition{
				{Field: "inactive_minutes", Operator: ">", Value: 60},
			},
			Action: ActionRequireMFA,
			Reason: "Session inactive for too long",
		},
	}
}

// DetermineAction evaluates the session against the policies and returns the recommended action.
func (p *SessionPolicy) DetermineAction(ctx map[string]interface{}) (ActionType, string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Sort rules by priority
	sortedRules := make([]PolicyRule, len(p.rules))
	copy(sortedRules, p.rules)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Priority < sortedRules[j].Priority
	})

	for _, rule := range sortedRules {
		if p.matchAllConditions(ctx, rule.Conditions) {
			return rule.Action, rule.Reason
		}
	}

	return ActionNone, ""
}

func (p *SessionPolicy) matchAllConditions(ctx map[string]interface{}, conditions []PolicyCondition) bool {
	for _, cond := range conditions {
		val, exists := ctx[cond.Field]
		if !exists {
			return false
		}
		if !matchCondition(val, cond.Operator, cond.Value) {
			return false
		}
	}
	return true
}

func matchCondition(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "==":
		return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
	case "!=":
		return fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected)
	case ">":
		return toFloat(actual) > toFloat(expected)
	case "<":
		return toFloat(actual) < toFloat(expected)
	case ">=":
		return toFloat(actual) >= toFloat(expected)
	case "<=":
		return toFloat(actual) <= toFloat(expected)
	case "contains":
		// Handle slice/array contains
		if arr, ok := actual.([]string); ok {
			strExp := fmt.Sprintf("%v", expected)
			for _, v := range arr {
				if v == strExp {
					return true
				}
			}
			return false
		}
		// Handle string contains
		if str, ok := actual.(string); ok {
			return strings.Contains(str, fmt.Sprintf("%v", expected))
		}
	}
	return false
}

func toFloat(v interface{}) float64 {
	switch i := v.(type) {
	case int:
		return float64(i)
	case float64:
		return i
	case int64:
		return float64(i)
	case time.Duration:
		return i.Seconds()
	default:
		return 0
	}
}

// AddRule adds a new rule to the policy engine.
func (p *SessionPolicy) AddRule(rule PolicyRule) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rules = append(p.rules, rule)
}

// RemoveRule removes a rule by ID.
func (p *SessionPolicy) RemoveRule(ruleID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var newRules []PolicyRule
	for _, r := range p.rules {
		if r.ID != ruleID {
			newRules = append(newRules, r)
		}
	}
	p.rules = newRules
}
