package scim

import (
	"time"

	"github.com/turtacn/QuantaID/pkg/scim"
	pkg_types "github.com/turtacn/QuantaID/pkg/types"
)

// ToDomainUser converts a SCIM User to a domain User
func ToDomainUser(sUser *scim.User) *pkg_types.User {
	u := &pkg_types.User{
		Username: sUser.UserName,
		Status:   pkg_types.UserStatusActive,
		Attributes: make(map[string]interface{}),
	}

	if !sUser.Active {
		u.Status = pkg_types.UserStatusInactive
	}

	// Map ExternalID
	if sUser.ExternalID != "" {
		u.Attributes["externalId"] = sUser.ExternalID
	}

	// Map Email (primary or first)
	for _, email := range sUser.Emails {
		if email.Primary || u.Email == "" {
			u.Email = pkg_types.EncryptedString(email.Value)
		}
	}

	// Map Phone
	for _, phone := range sUser.PhoneNumbers {
		if phone.Primary || u.Phone == "" {
			u.Phone = pkg_types.EncryptedString(phone.Value)
		}
	}

	// Map Name to attributes since User struct doesn't have Name fields
	if sUser.Name != nil {
		nameMap := make(map[string]interface{})
		if sUser.Name.GivenName != "" {
			nameMap["givenName"] = sUser.Name.GivenName
		}
		if sUser.Name.FamilyName != "" {
			nameMap["familyName"] = sUser.Name.FamilyName
		}
		if sUser.Name.Formatted != "" {
			nameMap["formatted"] = sUser.Name.Formatted
		}
		u.Attributes["name"] = nameMap
	}

	return u
}

// ToSCIMUser converts a domain User to a SCIM User
func ToSCIMUser(user *pkg_types.User) *scim.User {
	sUser := &scim.User{
		Resource: scim.Resource{
			Schemas: []string{scim.SchemaUser},
			ID:      user.ID,
			Meta: &scim.Meta{
				ResourceType: "User",
				Created:      user.CreatedAt.Format(time.RFC3339),
				LastModified: user.UpdatedAt.Format(time.RFC3339),
				// Location should be set by the handler or a URL builder
			},
		},
		UserName: user.Username,
		Active:   user.Status == pkg_types.UserStatusActive,
	}

	// Map Attributes back
	if val, ok := user.Attributes["externalId"]; ok {
		if strVal, ok := val.(string); ok {
			sUser.ExternalID = strVal
		}
	}

	// Map Email
	if user.Email != "" {
		sUser.Emails = []scim.Email{
			{
				Value:   string(user.Email),
				Type:    "work",
				Primary: true,
			},
		}
	}

	// Map Phone
	if user.Phone != "" {
		sUser.PhoneNumbers = []scim.Phone{
			{
				Value:   string(user.Phone),
				Type:    "work",
				Primary: true,
			},
		}
	}

	// Map Name
	if val, ok := user.Attributes["name"]; ok {
		if nameMap, ok := val.(map[string]interface{}); ok {
			name := &scim.Name{}
			if v, ok := nameMap["givenName"].(string); ok {
				name.GivenName = v
			}
			if v, ok := nameMap["familyName"].(string); ok {
				name.FamilyName = v
			}
			if v, ok := nameMap["formatted"].(string); ok {
				name.Formatted = v
			}
			sUser.Name = name
		}
	}

	return sUser
}

// ToDomainGroup converts a SCIM Group to a domain UserGroup
func ToDomainGroup(sGroup *scim.Group) *pkg_types.UserGroup {
	g := &pkg_types.UserGroup{
		Name: sGroup.DisplayName,
		Metadata: make(map[string]interface{}),
	}

	if sGroup.ExternalID != "" {
		g.Metadata["externalId"] = sGroup.ExternalID
	}

	// Note: Members are usually handled separately via AddUserToGroup or sync logic
	// because they refer to User IDs.

	return g
}

// ToSCIMGroup converts a domain UserGroup to a SCIM Group
func ToSCIMGroup(group *pkg_types.UserGroup) *scim.Group {
	sGroup := &scim.Group{
		Resource: scim.Resource{
			Schemas: []string{scim.SchemaGroup},
			ID:      group.ID,
			Meta: &scim.Meta{
				ResourceType: "Group",
				Created:      group.CreatedAt.Format(time.RFC3339),
				LastModified: group.UpdatedAt.Format(time.RFC3339),
			},
		},
		DisplayName: group.Name,
	}

	if val, ok := group.Metadata["externalId"]; ok {
		if strVal, ok := val.(string); ok {
			sGroup.ExternalID = strVal
		}
	}

	// Members - Populate if available in the group struct (which has Users []User)
	if len(group.Users) > 0 {
		members := make([]scim.Member, len(group.Users))
		for i, u := range group.Users {
			members[i] = scim.Member{
				Value:   u.ID,
				Display: u.Username,
				Ref:     "/scim/v2/Users/" + u.ID, // Ideally construct full URL
			}
		}
		sGroup.Members = members
	}

	return sGroup
}
