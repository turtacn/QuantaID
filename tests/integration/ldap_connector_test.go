package integration

import (
	"context"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/turtacn/QuantaID/internal/services/sync"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	ldapconnector "github.com/turtacn/QuantaID/pkg/plugins/connectors/ldap"
	"github.com/turtacn/QuantaID/pkg/types"
	"os"
	"testing"
	"time"
)

// NOTE: These tests are known to be flaky due to issues with the test environment.
// If they fail, it may not be due to a code change.

var (
	ldapHost      string
	ldapPort      int
	testConnector *ldapconnector.LDAPConnector
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "rroemhild/test-openldap:latest",
		ExposedPorts: []string{"389/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("slapd starting").WithStartupTimeout(5*time.Minute),
			wait.ForListeningPort("389/tcp").WithStartupTimeout(5*time.Minute),
		),
	}

	ldapContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("could not start openldap container: %s", err)
		os.Exit(1)
	}

	host, err := ldapContainer.Host(ctx)
	if err != nil {
		fmt.Printf("could not get openldap host: %s", err)
		os.Exit(1)
	}
	ldapHost = host

	port, err := ldapContainer.MappedPort(ctx, "389")
	if err != nil {
		fmt.Printf("could not get openldap port: %s", err)
		os.Exit(1)
	}
	ldapPort = port.Int()

	config := &ldapconnector.LDAPConfig{
		Host:         ldapHost,
		Port:         ldapPort,
		UseTLS:       false,
		BindDN:       "cn=admin,dc=planetexpress,dc=com",
		BindPassword: "GoodNewsEveryone",
		BaseDN:       "ou=people,dc=planetexpress,dc=com",
		UserFilter:   "(objectClass=inetOrgPerson)",
		AttrMapping: map[string]string{
			"username": "uid",
			"email":    "mail",
		},
	}

	testConnector, err = ldapconnector.NewLDAPConnector(config)
	if err != nil {
		fmt.Printf("could not create ldap connector: %s", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	if err := ldapContainer.Terminate(ctx); err != nil {
		fmt.Printf("could not terminate openldap container: %s", err)
	}

	os.Exit(exitCode)
}

func TestLDAPAuthenticate(t *testing.T) {
	ctx := context.Background()

	// Test successful authentication
	creds := map[string]string{"username": "professor", "password": "professor"}
	authResp, err := testConnector.Authenticate(ctx, creds)
	require.NoError(t, err)
	require.NotNil(t, authResp)
	assert.True(t, authResp.Success)
	assert.NotNil(t, authResp.User)
	assert.Equal(t, "professor", authResp.User.Username)

	// Test failed authentication
	creds = map[string]string{"username": "professor", "password": "wrongpassword"}
	authResp, err = testConnector.Authenticate(ctx, creds)
	require.NoError(t, err)
	require.NotNil(t, authResp)
	assert.False(t, authResp.Success)
	assert.Equal(t, types.ErrInvalidCredentials.Code, authResp.Error.Code)
}

func TestLDAPGetUser(t *testing.T) {
	ctx := context.Background()

	user, err := testConnector.GetUser(ctx, "professor")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "professor", user.Username)
	assert.Equal(t, "professor@planetexpress.com", user.Email)
}

func TestLDAPSyncUsers(t *testing.T) {
	userRepo := postgresql.NewInMemoryIdentityRepository()
	syncService := sync.NewLDAPSyncService(testConnector, userRepo)
	ctx := context.Background()

	// First, do a full sync
	err := syncService.FullSync(ctx)
	require.NoError(t, err)

	// Check that the user was synced
	syncedUser, err := userRepo.GetUserByUsername(ctx, "professor")
	require.NoError(t, err)
	require.NotNil(t, syncedUser)
	assert.Equal(t, "professor", syncedUser.Username)
	assert.Equal(t, "professor@planetexpress.com", syncedUser.Email)
	assert.Equal(t, types.UserStatusActive, syncedUser.Status)

	// Now, delete the user from LDAP
	conn, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", ldapHost, ldapPort))
	require.NoError(t, err)
	defer conn.Close()
	err = conn.Bind("cn=admin,dc=planetexpress,dc=com", "GoodNewsEveryone")
	require.NoError(t, err)
	delReq := ldap.NewDelRequest("uid=professor,ou=people,dc=planetexpress,dc=com", []ldap.Control{})
	err = conn.Del(delReq)
	require.NoError(t, err)

	// Do another full sync
	err = syncService.FullSync(ctx)
	require.NoError(t, err)

	// Check that the user was marked as inactive
	syncedUser, err = userRepo.GetUserByUsername(ctx, "professor")
	require.NoError(t, err)
	require.NotNil(t, syncedUser)
	assert.Equal(t, types.UserStatusInactive, syncedUser.Status)
}
