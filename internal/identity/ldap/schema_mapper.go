package ldap

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/turtacn/QuantaID/pkg/types"
	"strings"
)

type SchemaMapper struct {
	config SchemaMapConfig
}

type SchemaMapConfig struct {
	Mappings []AttributeMapping `yaml:"mappings"`
	Defaults map[string]string  `yaml:"defaults"`
}

type AttributeMapping struct {
	LDAPAttr     string `yaml:"ldap_attr"`
	QuantaField  string `yaml:"quanta_field"`
	Required     bool   `yaml:"required"`
	Transform    string `yaml:"transform"`
	FallbackAttr string `yaml:"fallback_attr"`
}

func NewSchemaMapper(config SchemaMapConfig) *SchemaMapper {
	return &SchemaMapper{config: config}
}

func (sm *SchemaMapper) MapEntry(entry *ldap.Entry) (*types.User, error) {
	user := &types.User{
		Attributes: make(map[string]interface{}),
		Status:     types.UserStatusActive, // Default to active for new users from LDAP
	}

	for _, mapping := range sm.config.Mappings {
		value := entry.GetAttributeValue(mapping.LDAPAttr)
		if value == "" && mapping.FallbackAttr != "" {
			value = entry.GetAttributeValue(mapping.FallbackAttr)
		}

		if value == "" && mapping.Required {
			return nil, fmt.Errorf("required field %s is missing", mapping.LDAPAttr)
		}

		value = sm.applyTransform(value, mapping.Transform)
		sm.setField(user, mapping.QuantaField, value)
	}

	for _, attr := range entry.Attributes {
		if !sm.isMapped(attr.Name) {
			user.Attributes[attr.Name] = attr.Values
		}
	}

	return user, nil
}

func (sm *SchemaMapper) applyTransform(value, transform string) string {
	switch transform {
	case "lowercase":
		return strings.ToLower(value)
	case "trim":
		return strings.TrimSpace(value)
	default:
		return value
	}
}

func (sm *SchemaMapper) setField(user *types.User, field, value string) {
	switch field {
	case "username":
		user.Username = value
	case "email":
		user.Email = value
	case "userAccountControl":
		// Handle the UserAccountControl attribute to map to user status
		// This is a simplified example. A real implementation would be more robust.
		if val, err := sm.parseUserAccountControl(value); err == nil {
			if (val & 2) != 0 { // ADS_UF_ACCOUNTDISABLE
				user.Status = types.UserStatusInactive
			}
		}
	default:
		user.Attributes[field] = value
	}
}

func (sm *SchemaMapper) parseUserAccountControl(value string) (int, error) {
	// The value is a string, so we need to parse it.
	// This is a placeholder for a more robust implementation.
	var val int
	_, err := fmt.Sscanf(value, "%d", &val)
	return val, err
}

func (sm *SchemaMapper) isMapped(attrName string) bool {
	for _, mapping := range sm.config.Mappings {
		if mapping.LDAPAttr == attrName {
			return true
		}
	}
	return false
}
