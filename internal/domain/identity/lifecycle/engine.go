package lifecycle

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/turtacn/QuantaID/pkg/types"
)

// Engine evaluates lifecycle rules against users.
type Engine struct {
}

// NewEngine creates a new Lifecycle Engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Evaluate checks if the user matches the given rules and returns the actions to be taken.
func (e *Engine) Evaluate(user *types.User, rules []LifecycleRule) ([]Action, error) {
	var actions []Action
	seenActions := make(map[ActionType]bool)

	for _, rule := range rules {
		matched, err := e.evaluateRule(user, rule)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate rule %s: %w", rule.Name, err)
		}

		if matched {
			for _, action := range rule.Actions {
				// Simple de-duplication based on action type
				if !seenActions[action.Type] {
					actions = append(actions, action)
					seenActions[action.Type] = true
				}
			}
		}
	}

	return actions, nil
}

func (e *Engine) evaluateRule(user *types.User, rule LifecycleRule) (bool, error) {
	if len(rule.Conditions) == 0 {
		return false, nil
	}

	for _, condition := range rule.Conditions {
		matched, err := e.evaluateCondition(user, condition)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil // AND logic: if any condition fails, the rule fails
		}
	}

	return true, nil
}

func (e *Engine) evaluateCondition(user *types.User, condition Condition) (bool, error) {
	val, err := e.getAttributeValue(user, condition.Attribute)
	if err != nil {
		return false, err
	}

	return e.compare(val, condition.Operator, condition.Value)
}

func (e *Engine) getAttributeValue(user *types.User, attribute string) (interface{}, error) {
	// Handle nested attributes map
	if strings.HasPrefix(attribute, "attributes.") {
		key := strings.TrimPrefix(attribute, "attributes.")
		if val, ok := user.Attributes[key]; ok {
			return val, nil
		}
		return nil, nil // Attribute not found
	}

	// Handle struct fields
	v := reflect.ValueOf(user)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Simple case-insensitive field lookup
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// Check against JSON tag or field name
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if strings.EqualFold(field.Name, attribute) || (jsonTag != "" && strings.EqualFold(jsonTag, attribute)) {
			return v.Field(i).Interface(), nil
		}
	}

	return nil, fmt.Errorf("unknown attribute: %s", attribute)
}

func (e *Engine) compare(actual interface{}, op Operator, expected interface{}) (bool, error) {
	if actual == nil {
		// If actual is nil, only allow check for nil/empty if that's what expected implies?
		// For now, assume nil doesn't match anything except NEQ if expected is not nil?
		// Simplification: nil only matches if we are checking for existence, but here we compare values.
		return false, nil
	}

	// Handle time comparisons
	if tActual, ok := toTime(actual); ok {
		// Expect expected to be a duration string (e.g. "90d" relative to now) or specific time?
		// The prompt example says "LastLogin > 90days". Usually this means "Now - LastLogin > 90days".
		// But the Condition struct has "Attribute Op Value".
		// If Attribute is LastLoginAt, and Value is "90d", and Op is GT...
		// It might mean: LastLoginAt > 90d (makes no sense).
		// Usually "LastLogin older than 90d" means "LastLogin < Now - 90d".
		// Let's interpret the Value. If it's a duration string, we calculate the threshold time.

		// Let's try to interpret "Age of attribute".
		// If the user writes "lastLoginAt gt 90d", they mean "It has been more than 90 days since last login".
		// This translates to: Now.Sub(LastLoginAt) > 90d.

		if dExpected, ok := parseDuration(expected); ok {
			// Compare duration since actual time
			age := time.Since(tActual)
			switch op {
			case OpGt:
				return age > dExpected, nil
			case OpGte:
				return age >= dExpected, nil
			case OpLt:
				return age < dExpected, nil
			case OpLte:
				return age <= dExpected, nil
			case OpEq:
				return age == dExpected, nil // Unlikely to hit exact nanosecond
			default:
				return false, fmt.Errorf("unsupported operator for time duration: %s", op)
			}
		}

		// If not duration, maybe absolute time comparison?
		// Leaving as is for now, assuming duration is the primary use case for lifecycle.
	}

	// Handle numeric comparisons
	// Convert both to float64 for comparison
	fActual, okActual := toFloat(actual)
	fExpected, okExpected := toFloat(expected)

	if okActual && okExpected {
		switch op {
		case OpEq: return fActual == fExpected, nil
		case OpNeq: return fActual != fExpected, nil
		case OpGt: return fActual > fExpected, nil
		case OpLt: return fActual < fExpected, nil
		case OpGte: return fActual >= fExpected, nil
		case OpLte: return fActual <= fExpected, nil
		}
	}

	// String comparisons
	sActual := fmt.Sprintf("%v", actual)
	sExpected := fmt.Sprintf("%v", expected)

	switch op {
	case OpEq: return sActual == sExpected, nil
	case OpNeq: return sActual != sExpected, nil
	case OpContains: return strings.Contains(sActual, sExpected), nil
	}

	return false, nil
}

func toTime(v interface{}) (time.Time, bool) {
	if t, ok := v.(time.Time); ok {
		return t, true
	}
	if t, ok := v.(*time.Time); ok {
		if t != nil {
			return *t, true
		}
	}
	// Try parsing string if needed, but for User struct fields they are strongly typed.
	return time.Time{}, false
}

func parseDuration(v interface{}) (time.Duration, bool) {
	s, ok := v.(string)
	if !ok {
		return 0, false
	}
	// Support "d" for days which ParseDuration doesn't support directly
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		var days int
		if _, err := fmt.Sscanf(daysStr, "%d", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, true
		}
	}
	d, err := time.ParseDuration(s)
	return d, err == nil
}

func parseTimeThreshold(v interface{}) (time.Time, error) {
	// Placeholder if we need to parse absolute dates strings
	return time.Time{}, fmt.Errorf("not implemented")
}

func toFloat(v interface{}) (float64, bool) {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), true
	case reflect.Float32, reflect.Float64:
		return val.Float(), true
	}
	return 0, false
}
