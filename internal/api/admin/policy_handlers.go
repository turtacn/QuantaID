package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	domain_policy "github.com/turtacn/QuantaID/internal/domain/policy"
	policy_service "github.com/turtacn/QuantaID/internal/services/policy"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/pkg/types"
)

// PolicyHandlers provides HTTP handlers for managing policies.
type PolicyHandlers struct {
	service policy_service.PolicyService
}

// NewPolicyHandlers creates a new PolicyHandlers.
func NewPolicyHandlers(service policy_service.PolicyService) *PolicyHandlers {
	return &PolicyHandlers{service: service}
}

// RegisterRoutes registers the policy management routes on the given router.
func (h *PolicyHandlers) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/roles", h.createRole).Methods("POST")
	router.HandleFunc("/roles", h.listRoles).Methods("GET")
	router.HandleFunc("/roles/{roleID}", h.updateRole).Methods("PUT")
	router.HandleFunc("/roles/{roleID}", h.deleteRole).Methods("DELETE")

	router.HandleFunc("/permissions", h.createPermission).Methods("POST")
	router.HandleFunc("/permissions", h.listPermissions).Methods("GET")
	router.HandleFunc("/roles/{roleID}/permissions", h.addPermissionToRole).Methods("POST")

	router.HandleFunc("/users/{userID}/roles", h.assignRoleToUser).Methods("POST")
	router.HandleFunc("/users/{userID}/roles/{roleID}", h.unassignRoleFromUser).Methods("DELETE")
	// ... other routes
}

func (h *PolicyHandlers) createRole(w http.ResponseWriter, r *http.Request) {
	var role domain_policy.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusBadRequest, Message: "Invalid request body"}, http.StatusBadRequest)
		return
	}

	if err := h.service.CreateRole(r.Context(), &role); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Failed to create role"}, http.StatusInternalServerError)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, role)
}

func (h *PolicyHandlers) listRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListRoles(r.Context())
	if err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Failed to list roles"}, http.StatusInternalServerError)
		return
	}
	handlers.WriteJSON(w, http.StatusOK, roles)
}

func (h *PolicyHandlers) updateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.ParseUint(vars["roleID"], 10, 32)
	if err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusBadRequest, Message: "Invalid role ID"}, http.StatusBadRequest)
		return
	}

	var role domain_policy.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusBadRequest, Message: "Invalid request body"}, http.StatusBadRequest)
		return
	}
	role.ID = uint(roleID)

	if err := h.service.UpdateRole(r.Context(), &role); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Failed to update role"}, http.StatusInternalServerError)
		return
	}
	handlers.WriteJSON(w, http.StatusOK, role)
}

func (h *PolicyHandlers) deleteRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.ParseUint(vars["roleID"], 10, 32)
	if err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusBadRequest, Message: "Invalid role ID"}, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteRole(r.Context(), uint(roleID)); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Failed to delete role"}, http.StatusInternalServerError)
		return
	}
	handlers.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *PolicyHandlers) assignRoleToUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	var body struct {
		RoleID uint `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusBadRequest, Message: "Invalid request body"}, http.StatusBadRequest)
		return
	}

	if err := h.service.AssignRoleToUser(r.Context(), userID, body.RoleID); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Failed to assign role to user"}, http.StatusInternalServerError)
		return
	}
	handlers.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *PolicyHandlers) unassignRoleFromUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	roleID, err := strconv.ParseUint(vars["roleID"], 10, 32)
	if err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusBadRequest, Message: "Invalid role ID"}, http.StatusBadRequest)
		return
	}

	if err := h.service.UnassignRoleFromUser(r.Context(), userID, uint(roleID)); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Failed to unassign role from user"}, http.StatusInternalServerError)
		return
	}
	handlers.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *PolicyHandlers) createPermission(w http.ResponseWriter, r *http.Request) {
	var permission domain_policy.Permission
	if err := json.NewDecoder(r.Body).Decode(&permission); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusBadRequest, Message: "Invalid request body"}, http.StatusBadRequest)
		return
	}

	if err := h.service.CreatePermission(r.Context(), &permission); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Failed to create permission"}, http.StatusInternalServerError)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, permission)
}

func (h *PolicyHandlers) listPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.service.ListPermissions(r.Context())
	if err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Failed to list permissions"}, http.StatusInternalServerError)
		return
	}
	handlers.WriteJSON(w, http.StatusOK, permissions)
}

func (h *PolicyHandlers) addPermissionToRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID, err := strconv.ParseUint(vars["roleID"], 10, 32)
	if err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusBadRequest, Message: "Invalid role ID"}, http.StatusBadRequest)
		return
	}

	var body struct {
		PermissionID uint `json:"permission_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusBadRequest, Message: "Invalid request body"}, http.StatusBadRequest)
		return
	}

	if err := h.service.AddPermissionToRole(r.Context(), uint(roleID), body.PermissionID); err != nil {
		handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Failed to add permission to role"}, http.StatusInternalServerError)
		return
	}
	handlers.WriteJSON(w, http.StatusNoContent, nil)
}
