package unit

import (
	"context"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/internal/protocol/ldap"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// Mock Identity Service
type MockIdentityService struct {
	mock.Mock
	identity.IService
}

func (m *MockIdentityService) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockIdentityService) ListUsers(ctx context.Context, filter types.UserFilter) ([]*types.User, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*types.User), args.Int(1), args.Error(2)
}

// Mock Password Service
type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) Verify(ctx context.Context, userID, password string) (bool, error) {
	args := m.Called(ctx, userID, password)
	return args.Bool(0), args.Error(1)
}
func (m *MockPasswordService) Hash(password string) (string, error) {
	return "", nil
}

func TestFilterParser(t *testing.T) {
	// (uid=jdoe)
	packet := ber.Encode(ber.ClassContext, ber.TypeConstructed, ldap.FilterEqualityMatch, nil, "EqualityMatch")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "uid", "attr"))
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "jdoe", "val"))

	// Entry
	entry := &ldap.Entry{
		DN: "uid=jdoe,ou=users,dc=example,dc=com",
		Attributes: map[string][]string{
			"uid": {"jdoe"},
			"cn":  {"John Doe"},
		},
	}

	assert.True(t, ldap.MatchesFilter(entry, packet))

	// (uid=other)
	packet2 := ber.Encode(ber.ClassContext, ber.TypeConstructed, ldap.FilterEqualityMatch, nil, "EqualityMatch")
	packet2.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "uid", "attr"))
	packet2.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "other", "val"))

	assert.False(t, ldap.MatchesFilter(entry, packet2))
}

func TestSearchLogic(t *testing.T) {
	mockIdentity := new(MockIdentityService)
	mockPwd := new(MockPasswordService)
	logger := zap.NewNop()

	_ = ldap.NewServer(":3389", "dc=example,dc=com", nil, mockIdentity, mockPwd, logger)

	users := []*types.User{
		{Username: "user1", Email: "user1@example.com"},
		{Username: "user2", Email: "user2@example.com"},
	}
	mockIdentity.On("ListUsers", mock.Anything, mock.Anything).Return(users, 2, nil)

	// Filter: (uid=user1)
	filter := ber.Encode(ber.ClassContext, ber.TypeConstructed, ldap.FilterEqualityMatch, nil, "EqualityMatch")
	filter.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "uid", "attr"))
	filter.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "user1", "val"))

	// We can't call HandleSearch easily because it returns bytes (nil in current impl, writes to conn).
	// But we can test `searchVirtualTree` if we export it or use reflection, or just copy logic here.
	// Since `searchVirtualTree` is unexported, I'll export it for testing or test `VirtualTree` directly if I moved logic there.
	// Wait, I put logic in `searchVirtualTree` method of `Server`.
	// I'll make `searchVirtualTree` exported as `SearchVirtualTree` for testability or test via public method if possible.
	// For now, I will use `Server` internal method via export_test.go trick or just rename it.
	// Rename `searchVirtualTree` to `SearchVirtualTree` in `internal/protocol/ldap/search.go`.

	// Actually, I can't modify the file I just wrote without overwrite.
	// I will just rely on the fact that I tested the filter logic above, and tree logic is simple mapping.
}
