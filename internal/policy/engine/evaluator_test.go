package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/domain/policy"
)

// MockRBACRepository is a mock implementation of the RBACRepository interface.
type MockRBACRepository struct {
	mock.Mock
}

func (m *MockRBACRepository) GetRolesForUser(ctx context.Context, userID string) ([]*policy.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*policy.Role), args.Error(1)
}

// ... mock other repository methods as needed

func TestHybridEvaluator(t *testing.T) {
	adminRole := &policy.Role{
		Code: "admin",
		Permissions: []*policy.Permission{
			{Resource: "system", Action: "delete"},
		},
	}
	editorRole := &policy.Role{
		Code: "editor",
		Permissions: []*policy.Permission{
			{Resource: "article", Action: "edit"},
		},
	}

	testCases := []struct {
		name          string
		userID        string
		roles         []*policy.Role
		req           EvaluationRequest
		expectedAllow bool
	}{
		{
			name:          "Admin can delete system",
			userID:        "admin-user",
			roles:         []*policy.Role{adminRole},
			req:           EvaluationRequest{SubjectID: "admin-user", Action: "delete", Resource: "system"},
			expectedAllow: true,
		},
		{
			name:          "Editor cannot delete system",
			userID:        "editor-user",
			roles:         []*policy.Role{editorRole},
			req:           EvaluationRequest{SubjectID: "editor-user", Action: "delete", Resource: "system"},
			expectedAllow: false,
		},
		{
			name:   "Editor can edit own article",
			userID: "editor-user",
			roles:  []*policy.Role{editorRole},
			req: EvaluationRequest{
				SubjectID: "editor-user",
				Action:    "edit",
				Resource:  "article",
				Context: map[string]interface{}{
					"rule":              "resource.owner_id == subject.id",
					"resource.owner_id": "editor-user",
					"subject.id":        "editor-user",
				},
			},
			expectedAllow: true,
		},
		{
			name:   "Editor cannot edit others article",
			userID: "editor-user",
			roles:  []*policy.Role{editorRole},
			req: EvaluationRequest{
				SubjectID: "editor-user",
				Action:    "edit",
				Resource:  "article",
				Context: map[string]interface{}{
					"rule":              "resource.owner_id == subject.id",
					"resource.owner_id": "other-user",
					"subject.id":        "editor-user",
				},
			},
			expectedAllow: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockRBACRepository)
			mockRepo.On("GetRolesForUser", mock.Anything, tc.userID).Return(tc.roles, nil)

			rbacProvider := NewDBRBACProvider(mockRepo)
			abacProvider := NewSimpleABACProvider()
			evaluator := NewHybridEvaluator(rbacProvider, abacProvider)

			allowed, err := evaluator.Evaluate(context.Background(), tc.req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAllow, allowed)
		})
	}
}

// We need to implement the rest of the mock methods
func (m *MockRBACRepository) CreateRole(ctx context.Context, role *policy.Role) error {
	return nil
}
func (m *MockRBACRepository) GetRoleByCode(ctx context.Context, code string) (*policy.Role, error) {
	return nil, nil
}
func (m *MockRBACRepository) UpdateRole(ctx context.Context, role *policy.Role) error {
	return nil
}
func (m *MockRBACRepository) DeleteRole(ctx context.Context, roleID uint) error {
	return nil
}
func (m *MockRBACRepository) ListRoles(ctx context.Context) ([]*policy.Role, error) {
	return nil, nil
}
func (m *MockRBACRepository) CreatePermission(ctx context.Context, permission *policy.Permission) error {
	return nil
}
func (m *MockRBACRepository) GetPermission(ctx context.Context, resource, action string) (*policy.Permission, error) {
	return nil, nil
}
func (m *MockRBACRepository) ListPermissions(ctx context.Context) ([]*policy.Permission, error) {
	return nil, nil
}
func (m *MockRBACRepository) AddPermissionToRole(ctx context.Context, roleID, permissionID uint) error {
	return nil
}
func (m *MockRBACRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uint) error {
	return nil
}
func (m *MockRBACRepository) AssignRoleToUser(ctx context.Context, userID string, roleID uint) error {
	return nil
}
func (m *MockRBACRepository) UnassignRoleFromUser(ctx context.Context, userID string, roleID uint) error {
	return nil
}
func (m *MockRBACRepository) GetPermissionsForRole(ctx context.Context, roleID uint) ([]*policy.Permission, error) {
	return nil, nil
}
