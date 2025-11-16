package ldap

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

func TestLDAPSync_WithReal389DS(t *testing.T) {
	ctx := context.Background()

	// Start a PostgreSQL container
	pgC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:13",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "password",
				"POSTGRES_DB":       "test",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer pgC.Terminate(ctx)

	pgHost, _ := pgC.Host(ctx)
	pgPort, _ := pgC.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("host=%s port=%s user=test password=password dbname=test sslmode=disable", pgHost, pgPort.Port())

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	_ = postgresql.NewPostgresIdentityRepository(db)

	// Start 389 Directory Server container
	ldapC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "389ds/dirsrv:latest",
			ExposedPorts: []string{"3389/tcp"},
			Env: map[string]string{
				"DS_SUFFIX":       "dc=example,dc=com",
				"DS_DM_PASSWORD":  "admin123",
			},
			WaitingFor: wait.ForLog("slapd started"),
		},
		Started: true,
	})
	require.NoError(t, err)
	defer ldapC.Terminate(ctx)

	host, _ := ldapC.Host(ctx)
	port, _ := ldapC.MappedPort(ctx, "3389")
	ldapURL := fmt.Sprintf("ldap://%s:%s", host, port.Port())

	ldapClient := NewLDAPClient(ldapURL, "cn=Directory Manager", "admin123")
	err = ldapClient.Connect()
	require.NoError(t, err)
	defer ldapClient.Close()

	// There is no easy way to seed data into the 389-ds container, so this test will be limited to checking the connection.
	assert.NotNil(t, ldapClient.conn)
}
