package ldap

import (
	"context"
	"fmt"
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
)

// VirtualTree represents the logical directory structure
type VirtualTree struct {
	baseDN      string
	userService identity.IService
}

func NewVirtualTree(baseDN string, userService identity.IService) *VirtualTree {
	return &VirtualTree{
		baseDN:      baseDN,
		userService: userService,
	}
}

func (vt *VirtualTree) Search(ctx context.Context, baseDN string, scope int, filter *ber.Packet, attrs []string) ([]*Entry, error) {
	// Simplified Virtual Tree:
	// BaseDN: dc=example,dc=com (configurable)
	// ou=users
	//   uid=jdoe
	// ou=groups
	//   cn=admin

	// Normalize BaseDN
	baseDN = strings.ToLower(baseDN)
	serverBase := strings.ToLower(vt.baseDN)

	if !strings.HasSuffix(baseDN, serverBase) && baseDN != "" {
		// Out of scope?
		return nil, nil
	}

	var entries []*Entry

	// Check if we are searching for Users
	// Simplification: Always search users if scope includes ou=users
	// We really need to parse the filter to see what we are looking for.
	// But `ListUsers` gives all users.
	// Let's implement basic "All Users" then filter manually.
	// This is inefficient but functional for small scale.

	// Fetch all users
	// TODO: Pagination?
	users, _, err := vt.userService.ListUsers(ctx, types.UserFilter{})
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		entry := vt.ConvertUserToEntry(u)
		if entry == nil {
			continue
		}

		// Check Scope
		if !isDNInScope(entry.DN, baseDN, scope) {
			continue
		}

		// Check Filter
		if MatchesFilter(entry, filter) {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// ConvertUserToEntry maps a domain User to an LDAP Entry
func (vt *VirtualTree) ConvertUserToEntry(u *types.User) *Entry {
	dn := fmt.Sprintf("uid=%s,ou=users,%s", u.Username, vt.baseDN)

	// Fetch attributes from map if available
	sn := u.Username
	if val, ok := u.Attributes["sn"].(string); ok {
		sn = val
	}

	cn := u.Username
	if val, ok := u.Attributes["displayName"].(string); ok {
		cn = val
	}

	attrs := map[string][]string{
		"objectClass": {"top", "person", "organizationalPerson", "inetOrgPerson"},
		"uid":         {u.Username},
		"cn":          {cn},
		"sn":          {sn},
		"mail":        {string(u.Email)},
	}

	if val, ok := u.Attributes["firstName"].(string); ok {
		attrs["givenName"] = []string{val}
	}
	if val, ok := u.Attributes["lastName"].(string); ok {
		attrs["sn"] = []string{val}
	}

	return &Entry{
		DN:         dn,
		Attributes: attrs,
	}
}

func isDNInScope(dn, base string, scope int) bool {
	dn = strings.ToLower(dn)
	base = strings.ToLower(base)

	if scope == ScopeBaseObject {
		return dn == base
	}

	if !strings.HasSuffix(dn, base) {
		return false
	}

	// Remove base from dn
	remaining := strings.TrimSuffix(dn, ","+base)
	// if base was empty or root?
	if base == "" {
		remaining = dn
	}

	parts := strings.Split(remaining, ",")

	if scope == ScopeSingleLevel {
		return len(parts) == 1
	}

	return true
}
