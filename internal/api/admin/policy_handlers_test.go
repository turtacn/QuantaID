package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	domain_policy "github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/services/policy"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	})

	host, _ := postgresContainer.Host(ctx)
	port, _ := postgresContainer.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=test sslmode=disable", host, port.Port())

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Run migrations
	db.AutoMigrate(&domain_policy.Role{}, &domain_policy.Permission{}, &domain_policy.UserRole{})

	return db
}

func TestPolicyHandlers(t *testing.T) {
	db := setupTestDB(t)

	repo := postgresql.NewRBACRepository(db)
	service := policy.NewService(repo)
	handlers := NewPolicyHandlers(service)

	router := mux.NewRouter()
	handlers.RegisterRoutes(router)

	t.Run("Create and List Roles", func(t *testing.T) {
		// Create a new role
		role := &domain_policy.Role{Code: "test-role", Description: "A test role"}
		body, _ := json.Marshal(role)
		req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)

		// List roles and check if the new role is there
		req, _ = http.NewRequest("GET", "/roles", nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		var roles []*domain_policy.Role
		json.Unmarshal(rr.Body.Bytes(), &roles)
		assert.NotEmpty(t, roles)
		assert.Equal(t, "test-role", roles[0].Code)

		// Update the role
		updatedRole := roles[0]
		updatedRole.Description = "An updated test role"
		body, _ = json.Marshal(updatedRole)
		req, _ = http.NewRequest("PUT", fmt.Sprintf("/roles/%d", updatedRole.ID), bytes.NewBuffer(body))
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		// Delete the role
		req, _ = http.NewRequest("DELETE", fmt.Sprintf("/roles/%d", updatedRole.ID), nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("Manage Permissions", func(t *testing.T) {
		// Create a new permission
		permission := &domain_policy.Permission{Resource: "test-resource", Action: "test-action"}
		body, _ := json.Marshal(permission)
		req, _ := http.NewRequest("POST", "/permissions", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)

		// List permissions
		req, _ = http.NewRequest("GET", "/permissions", nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		var permissions []*domain_policy.Permission
		json.Unmarshal(rr.Body.Bytes(), &permissions)
		assert.NotEmpty(t, permissions)
		assert.Equal(t, "test-resource", permissions[0].Resource)

		// Create a role to assign the permission to
		role := &domain_policy.Role{Code: "permission-role", Description: "A test role for permissions"}
		body, _ = json.Marshal(role)
		req, _ = http.NewRequest("POST", "/roles", bytes.NewBuffer(body))
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)
		var createdRole domain_policy.Role
		json.Unmarshal(rr.Body.Bytes(), &createdRole)

		// Assign the permission to the role
		assignBody := struct {
			PermissionID uint `json:"permission_id"`
		}{
			PermissionID: permissions[0].ID,
		}
		body, _ = json.Marshal(assignBody)
		req, _ = http.NewRequest("POST", fmt.Sprintf("/roles/%d/permissions", createdRole.ID), bytes.NewBuffer(body))
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("Assign Roles to User", func(t *testing.T) {
		// Create a role to assign to the user
		role := &domain_policy.Role{Code: "user-role", Description: "A test role for users"}
		body, _ := json.Marshal(role)
		req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)
		var createdRole domain_policy.Role
		json.Unmarshal(rr.Body.Bytes(), &createdRole)

		// Assign the role to the user
		userID := "test-user"
		assignBody := struct {
			RoleID uint `json:"role_id"`
		}{
			RoleID: createdRole.ID,
		}
		body, _ = json.Marshal(assignBody)
		req, _ = http.NewRequest("POST", fmt.Sprintf("/users/%s/roles", userID), bytes.NewBuffer(body))
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Unassign the role from the user
		req, _ = http.NewRequest("DELETE", fmt.Sprintf("/users/%s/roles/%d", userID, createdRole.ID), nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}
