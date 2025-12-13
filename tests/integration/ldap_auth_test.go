//go:build integration
// +build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/domain/password"
	protocol_ldap "github.com/turtacn/QuantaID/internal/protocol/ldap"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Mock Identity Service for Integration
type MockIdentity struct {
	identity.IService
	users map[string]*types.User
}

func (m *MockIdentity) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockIdentity) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) {
	var list []*types.User
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, len(list), nil
}

// Mock Password Service for Integration
type MockPassword struct {
	password.IService
}

func (m *MockPassword) Verify(ctx context.Context, userID, pwd string) (bool, error) {
	// Accept any password "secret"
	return pwd == "secret", nil
}

func TestLDAPIntegration(t *testing.T) {
	// Setup Server
	mockID := &MockIdentity{
		users: map[string]*types.User{
			"user1": {ID: "1", Username: "user1", Email: "user1@example.com"},
		},
	}
	mockPwd := &MockPassword{}

	port := 3389
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	logger, _ := zap.NewDevelopment()
	server := protocol_ldap.NewServer(addr, "dc=example,dc=com", nil, mockID, mockPwd, logger)

	err := server.Start()
	require.NoError(t, err)
	defer server.Stop()

	// Wait for start
	time.Sleep(100 * time.Millisecond)

	// Client Connect
	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s", addr))
	require.NoError(t, err)
	defer l.Close()

	// 1. Test Bind (Success)
	err = l.Bind("uid=user1,ou=users,dc=example,dc=com", "secret")
	assert.NoError(t, err, "Bind should succeed")

	// 2. Test Bind (Fail)
	err = l.Bind("uid=user1,ou=users,dc=example,dc=com", "wrong")
	assert.Error(t, err, "Bind should fail")

	// Reconnect for anonymous/search if bind failed closed connection?
	// Usually Bind failure does not close connection in LDAPv3 but some servers might.
	// Our server implementation continues loop.

	// 3. Test Search
	// Bind again
	_ = l.Bind("uid=user1,ou=users,dc=example,dc=com", "secret")

	searchReq := ldap.NewSearchRequest(
		"dc=example,dc=com",
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		"(uid=user1)",
		[]string{"cn", "mail"},
		nil,
	)

	sr, err := l.Search(searchReq)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(sr.Entries))
	assert.Equal(t, "uid=user1,ou=users,dc=example,dc=com", sr.Entries[0].DN)
	assert.Equal(t, "user1@example.com", sr.Entries[0].GetAttributeValue("mail"))
}
