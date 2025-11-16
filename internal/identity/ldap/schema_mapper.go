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
	default:
		user.Attributes[field] = value
	}
}

func (sm *SchemaMapper) isMapped(attrName string) bool {
	for _, mapping := range sm.config.Mappings {
		if mapping.LDAPAttr == attrName {
			return true
		}
	}
	return false
}
