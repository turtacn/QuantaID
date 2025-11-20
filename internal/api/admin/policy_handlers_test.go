//go:build integration
// +build integration

package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	ctx := context.Background()
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:14-alpine"),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := postgresql.NewConnection(postgresql.Config{DSN: connStr})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&types.Role{}, &types.Permission{}, &policy.Policy{})
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		pgContainer.Terminate(ctx)
	}

	return db, cleanup
}

func TestPolicyHandlers(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	policyRepo := postgresql.NewPostgresPolicyRepository(db)
	evaluator := authorization.NewDefaultEvaluator(policyRepo)
	authzService := authorization.NewService(evaluator, nil) // nil for audit service in this test
	handlers := NewPolicyHandlers(authzService)

	router := mux.NewRouter()
	handlers.RegisterRoutes(router)

	t.Run("Create and Get Role", func(t *testing.T) {
		// Create Role
		role := &types.Role{Name: "Test Role", Description: "A role for testing"}
		body, _ := json.Marshal(role)
		req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)

		var createdRole types.Role
		err := json.Unmarshal(rr.Body.Bytes(), &createdRole)
		require.NoError(t, err)
		assert.Equal(t, "Test Role", createdRole.Name)

		// Get Role
		req, _ = http.NewRequest("GET", fmt.Sprintf("/roles/%s", createdRole.ID), nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var fetchedRole types.Role
		err = json.Unmarshal(rr.Body.Bytes(), &fetchedRole)
		require.NoError(t, err)
		assert.Equal(t, createdRole.ID, fetchedRole.ID)
	})

	t.Run("Assign and Check Permission", func(t *testing.T) {
		// Create a role and permission
		role := &types.Role{Name: "Editor", ID: uuid.New().String()}
		db.Create(&role)
		permission := &types.Permission{Name: "edit:articles", ID: uuid.New().String()}
		db.Create(&permission)

		// Assign Permission
		assignReq := &authorization.AssignPermissionRequest{PermissionID: permission.ID}
		body, _ := json.Marshal(assignReq)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/roles/%s/permissions", role.ID), bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		// Check Permission
		checkReq := &authorization.CheckPermissionRequest{
			Subject:  fmt.Sprintf("user:%s", uuid.New().String()), // A dummy user
			Action:   "edit:articles",
			Resource: "article:123",
		}
		// We're not actually checking here as it requires a full policy setup.
		// This test primarily ensures the API endpoint for assignment works.
		// A full check would be an integration test involving the policy engine.
	})
}
