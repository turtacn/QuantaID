package profile

import (
	"context"
	"reflect"
	"strings"
	"time"
)

// TagCondition defines the condition for an auto-tag rule
type TagCondition struct {
	Field    string      `mapstructure:"field"`    // "behavior.unique_locations", "risk_score" etc.
	Operator string      `mapstructure:"operator"` // ">", "<", ">=", "==", "contains"
	Value    interface{} `mapstructure:"value"`
}

// AutoTagRule defines a rule for automatic tagging
type AutoTagRule struct {
	Tag       string       `mapstructure:"tag"`
	Condition TagCondition `mapstructure:"condition"`
}

// TagManager manages user tags
type TagManager struct {
	autoRules   []AutoTagRule
	profileRepo ProfileRepository
}

// NewTagManager creates a new TagManager
func NewTagManager(rules []map[string]interface{}, profileRepo ProfileRepository) *TagManager {
	parsedRules := make([]AutoTagRule, len(rules))

	for i, r := range rules {
		rule := AutoTagRule{}
		if t, ok := r["tag"].(string); ok {
			rule.Tag = t
		}
		if c, ok := r["condition"].(map[string]interface{}); ok {
			rule.Condition = TagCondition{}
			if f, ok := c["field"].(string); ok {
				rule.Condition.Field = f
			}
			if op, ok := c["operator"].(string); ok {
				rule.Condition.Operator = op
			}
			rule.Condition.Value = c["value"]
		}
		parsedRules[i] = rule
	}

	return &TagManager{
		autoRules:   parsedRules,
		profileRepo: profileRepo,
	}
}

// DefaultAutoTagRules returns the default set of auto-tag rules
func DefaultAutoTagRules() []AutoTagRule {
	return []AutoTagRule{
		{Tag: TagFrequentTraveler, Condition: TagCondition{Field: "behavior.unique_locations", Operator: ">=", Value: 5}},
		{Tag: TagHighValueUser, Condition: TagCondition{Field: "behavior.login_frequency", Operator: ">=", Value: 10.0}},
		{Tag: TagDormantUser, Condition: TagCondition{Field: "last_activity_days", Operator: ">", Value: 30}},
		{Tag: TagNewUser, Condition: TagCondition{Field: "account_age_days", Operator: "<", Value: 7}},
		{Tag: TagSecurityConscious, Condition: TagCondition{Field: "behavior.mfa_usage_rate", Operator: ">=", Value: 0.9}},
		{Tag: TagHighRisk, Condition: TagCondition{Field: "risk_score", Operator: ">=", Value: 75}},
	}
}

// EvaluateAutoTags evaluates the rules against a profile and returns applicable tags
func (m *TagManager) EvaluateAutoTags(profile *UserProfile) []string {
	var tags []string
	for _, rule := range m.autoRules {
		if m.evaluateCondition(profile, rule.Condition) {
			tags = append(tags, rule.Tag)
		}
	}
	return tags
}

func (m *TagManager) evaluateCondition(profile *UserProfile, cond TagCondition) bool {
	// Handle virtual fields
	var value interface{}
	found := false

	switch cond.Field {
	case "last_activity_days":
		if profile.LastActivityAt != nil {
			value = time.Since(*profile.LastActivityAt).Hours() / 24.0
			found = true
		}
	case "account_age_days":
		value = time.Since(profile.CreatedAt).Hours() / 24.0
		found = true
	default:
		value, found = getFieldValue(profile, cond.Field)
	}

	if !found {
		return false
	}

	return compareValues(value, cond.Operator, cond.Value)
}

// compareValues compares two values based on the operator
func compareValues(actual interface{}, operator string, target interface{}) bool {
	// Simple type coercion logic for comparison
	v1 := reflect.ValueOf(actual)
	v2 := reflect.ValueOf(target)

	// Handle numeric comparisons
	if isNumber(v1.Kind()) && isNumber(v2.Kind()) {
		f1 := toFloat(v1)
		f2 := toFloat(v2)
		switch operator {
		case ">":
			return f1 > f2
		case ">=":
			return f1 >= f2
		case "<":
			return f1 < f2
		case "<=":
			return f1 <= f2
		case "==":
			return f1 == f2
		}
	}

	// Handle generic equality
	if operator == "==" {
		return reflect.DeepEqual(actual, target)
	}

	// Handle contains (string)
	if operator == "contains" {
		if s1, ok1 := actual.(string); ok1 {
			if s2, ok2 := target.(string); ok2 {
				return strings.Contains(s1, s2)
			}
		}
	}

	return false
}

func isNumber(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Float64
}

func toFloat(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	}
	return 0
}

// getFieldValue retrieves a field value by dot-notation path (e.g., "behavior.total_logins")
func getFieldValue(obj interface{}, path string) (interface{}, bool) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	parts := strings.Split(path, ".")
	for _, part := range parts {
		if val.Kind() == reflect.Struct {
			// Find field by name (case-insensitive simple match or json tag match could be better,
			// but here we assume struct field names are mapped or we match by exact name/CamelCase)
			// Actually, let's try to find by name directly.
			f := val.FieldByName(toCamelCase(part))
			if !f.IsValid() {
				// Try direct name match if CamelCase conversion fails or isn't used
				f = val.FieldByName(part)
			}

			if !f.IsValid() {
				// Try case-insensitive search
				found := false
				for i := 0; i < val.NumField(); i++ {
					typeField := val.Type().Field(i)
					if strings.EqualFold(typeField.Name, part) {
						f = val.Field(i)
						found = true
						break
					}
					// Check json tag
					tag := typeField.Tag.Get("json")
					if tag != "" {
						tagName := strings.Split(tag, ",")[0]
						if tagName == part {
							f = val.Field(i)
							found = true
							break
						}
					}
				}
				if !found {
					return nil, false
				}
			}
			val = f
		} else {
			return nil, false
		}
	}

	return val.Interface(), true
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}


// AddManualTag adds a manual tag to a user
func (m *TagManager) AddManualTag(ctx context.Context, userID, tag string) error {
	profile, err := m.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return nil // Or error?
	}

	// Check if exists
	for _, t := range profile.ManualTags {
		if t == tag {
			return nil
		}
	}

	profile.ManualTags = append(profile.ManualTags, tag)
	return m.profileRepo.UpdateTags(ctx, userID, profile.AutoTags, profile.ManualTags)
}

// RemoveManualTag removes a manual tag from a user
func (m *TagManager) RemoveManualTag(ctx context.Context, userID, tag string) error {
	profile, err := m.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return nil
	}

	newTags := make(StringSlice, 0)
	for _, t := range profile.ManualTags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	profile.ManualTags = newTags
	return m.profileRepo.UpdateTags(ctx, userID, profile.AutoTags, profile.ManualTags)
}
