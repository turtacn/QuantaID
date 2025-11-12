package ldap

import (
	"github.com/go-ldap/ldap/v3"
	"github.com/turtacn/QuantaID/pkg/types"
)

type Mapper struct {
	attrMapping map[string]string
}

func NewMapper(attrMapping map[string]string) *Mapper {
	return &Mapper{attrMapping: attrMapping}
}

func (m *Mapper) GetAttributeList() []string {
	attrs := make([]string, 0, len(m.attrMapping))
	for _, ldapAttr := range m.attrMapping {
		attrs = append(attrs, ldapAttr)
	}
	return attrs
}

func (m *Mapper) MapEntryToUser(entry *ldap.Entry) (*types.User, error) {
	user := &types.User{
		Username:   entry.GetAttributeValue(m.attrMapping["username"]),
		Email:      entry.GetAttributeValue(m.attrMapping["email"]),
		Attributes: make(map[string]interface{}),
	}

	if val, ok := m.attrMapping["display_name"]; ok {
		user.Attributes["displayName"] = entry.GetAttributeValue(val)
	}
	if val, ok := m.attrMapping["phone"]; ok {
		user.Phone = entry.GetAttributeValue(val)
	}

	for _, attr := range entry.Attributes {
		user.Attributes[attr.Name] = attr.Values
	}

	return user, nil
}
