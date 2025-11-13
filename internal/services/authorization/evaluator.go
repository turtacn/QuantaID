package authorization

import (
	"context"
	"net"

	"github.com/turtacn/QuantaID/internal/domain/policy"
	"golang.org/x/exp/slices"
)

// Evaluator defines the interface for the policy evaluation engine.
type Evaluator interface {
	Evaluate(ctx context.Context, evalCtx policy.EvaluationContext) (policy.Decision, error)
}

// DefaultEvaluator is the default implementation of the Evaluator interface.
// It loads rules from a configuration and evaluates them in memory.
type DefaultEvaluator struct {
	rules []Rule
}

// Rule represents a single authorization rule.
type Rule struct {
	Name        string          `yaml:"name"`
	Effect      policy.Decision `yaml:"effect"`
	Actions     []string        `yaml:"actions"`
	Subjects    []string        `yaml:"subjects"`
	IPWhitelist []string        `yaml:"ip_whitelist"`
	TimeRanges  []TimeRange     `yaml:"time_ranges"`
}

// TimeRange represents a time interval.
type TimeRange struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

// NewDefaultEvaluator creates a new DefaultEvaluator.
// In the future, this will load rules from a configuration file.
func NewDefaultEvaluator(rules []Rule) *DefaultEvaluator {
	return &DefaultEvaluator{rules: rules}
}

// Evaluate checks the evaluation context against the configured rules.
func (e *DefaultEvaluator) Evaluate(ctx context.Context, evalCtx policy.EvaluationContext) (policy.Decision, error) {
	for _, rule := range e.rules {
		if e.matches(rule, evalCtx) {
			return rule.Effect, nil
		}
	}
	// Default deny if no rule matches
	return policy.DecisionDeny, nil
}

func (e *DefaultEvaluator) matches(rule Rule, evalCtx policy.EvaluationContext) bool {
	return e.matchesAction(rule, evalCtx) &&
		e.matchesSubject(rule, evalCtx) &&
		e.matchesIP(rule, evalCtx) &&
		e.matchesTime(rule, evalCtx)
}

func (e *DefaultEvaluator) matchesAction(rule Rule, evalCtx policy.EvaluationContext) bool {
	if len(rule.Actions) == 0 {
		return true
	}
	action := string(evalCtx.Action)
	return slices.Contains(rule.Actions, "*") || slices.Contains(rule.Actions, action)
}

func (e *DefaultEvaluator) matchesSubject(rule Rule, evalCtx policy.EvaluationContext) bool {
	if len(rule.Subjects) == 0 {
		return true
	}

	if slices.Contains(rule.Subjects, "*") {
		return true
	}

	for _, subject := range rule.Subjects {
		if subject == "user:"+evalCtx.Subject.UserID {
			return true
		}
		for _, group := range evalCtx.Subject.Groups {
			if subject == "group:"+group {
				return true
			}
		}
	}
	return false
}

func (e *DefaultEvaluator) matchesIP(rule Rule, evalCtx policy.EvaluationContext) bool {
	if len(rule.IPWhitelist) == 0 {
		return true
	}

	ip := net.ParseIP(evalCtx.Environment.IP)
	if ip == nil {
		return false
	}

	for _, cidr := range rule.IPWhitelist {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			// Log this error in a real application
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

func (e *DefaultEvaluator) matchesTime(rule Rule, evalCtx policy.EvaluationContext) bool {
	if len(rule.TimeRanges) == 0 {
		return true
	}

	now := evalCtx.Environment.Time
	nowStr := now.Format("15:04")

	for _, tr := range rule.TimeRanges {
		if nowStr >= tr.Start && nowStr <= tr.End {
			return true
		}
	}
	return false
}
