package oauth

import (
	"context"
	"strings"
	"time"

	"github.com/turtacn/QuantaID/pkg/types"
)

// ClaimsMapper maps OAuth scopes to OpenID Connect claims.
type ClaimsMapper struct {
	rules      []MappingRule
	transforms map[string]TransformFunc
}

// MappingRule defines a rule for mapping a scope to a set of claims.
type MappingRule struct {
	Scope      string            `yaml:"scope"`
	Claims     []string          `yaml:"claims"`
	Conditions []Condition       `yaml:"conditions"`
	Transforms map[string]string `yaml:"transforms"`
}

// Condition defines a condition that must be met for a rule to apply.
type Condition struct {
	Field    string      `yaml:"field"`
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
}

// TransformFunc is a function that transforms a claim value.
type TransformFunc func(value interface{}) (interface{}, error)

// NewClaimsMapper creates a new ClaimsMapper.
func NewClaimsMapper(rules []MappingRule) *ClaimsMapper {
	mapper := &ClaimsMapper{
		rules:      rules,
		transforms: make(map[string]TransformFunc),
	}
	mapper.registerBuiltinTransforms()
	return mapper
}

// MapClaims maps the given scopes to a set of claims for the given user.
func (m *ClaimsMapper) MapClaims(ctx context.Context, user *types.User, scopes []string) (map[string]interface{}, error) {
	claims := make(map[string]interface{})

	for _, scope := range scopes {
		for _, rule := range m.findRulesByScope(scope) {
			if !m.evaluateConditions(user, rule.Conditions) {
				continue
			}

			for _, claimName := range rule.Claims {
				value := m.extractUserAttribute(user, claimName)

				if transformName, ok := rule.Transforms[claimName]; ok {
					if transformFunc, exists := m.transforms[transformName]; exists {
						value, _ = transformFunc(value)
					}
				}

				claims[claimName] = value
			}
		}
	}

	// Add standard claims
	claims["sub"] = user.ID
	claims["iat"] = time.Now().Unix()

	return claims, nil
}

func (m *ClaimsMapper) findRulesByScope(scope string) []MappingRule {
	var matchedRules []MappingRule
	for _, rule := range m.rules {
		if rule.Scope == scope {
			matchedRules = append(matchedRules, rule)
		}
	}
	return matchedRules
}

func (m *ClaimsMapper) evaluateConditions(user *types.User, conditions []Condition) bool {
	// TODO: Implement condition evaluation logic.
	return true
}

func (m *ClaimsMapper) extractUserAttribute(user *types.User, claimName string) interface{} {
	// TODO: Implement logic to extract user attributes, including nested ones.
	switch claimName {
	case "email":
		return user.Email
	case "username":
		return user.Username
	}
	return nil
}

func (m *ClaimsMapper) registerBuiltinTransforms() {
	m.transforms["mask_email"] = func(v interface{}) (interface{}, error) {
		var email string
		switch val := v.(type) {
		case string:
			email = val
		case types.EncryptedString:
			email = string(val)
		default:
			return v, nil
		}

		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return v, nil
		}
		return types.EncryptedString(parts[0][:1] + "***@" + parts[1]), nil
	}
}
