package admin

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/audit/events"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	"github.com/turtacn/QuantaID/pkg/types"
)

// AdminUserHandler handles user management operations for administrators.
type AdminUserHandler struct {
	userService identity.IService
	auditLogger *audit.AuditLogger
}

// NewAdminUserHandler creates a new AdminUserHandler.
func NewAdminUserHandler(userService identity.IService, auditLogger *audit.AuditLogger) *AdminUserHandler {
	return &AdminUserHandler{
		userService: userService,
		auditLogger: auditLogger,
	}
}

// ListUsers handles the retrieval of a paginated and searchable list of users.
func (h *AdminUserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	if pageSize == 0 {
		pageSize = 20
	}

	filter := types.UserFilter{
		Query:     query.Get("q"),
		Status:    []types.UserStatus{types.UserStatus(query.Get("status"))},
		Page:      page,
		PageSize:  pageSize,
		SortBy:    query.Get("sortBy"),
		SortOrder: query.Get("sortOrder"),
	}

	users, total, err := h.userService.ListUsers(r.Context(), filter)
	if err != nil {
		handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		return
	}

	response := struct {
		Users []*types.User `json:"users"`
		Total int           `json:"total"`
	}{
		Users: users,
		Total: total,
	}

	handlers.WriteJSON(w, http.StatusOK, response)
}

// BanUser handles banning a user.
func (h *AdminUserHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	err := h.userService.ChangeUserStatus(r.Context(), userID, types.UserStatusLocked)
	if err != nil {
		handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		return
	}

	h.auditLogger.Record(r.Context(), &events.AuditEvent{
		EventType: events.EventDataModified,
		Actor:     events.Actor{ID: r.Context().Value(middleware.UserIDContextKey).(string), Type: "user"},
		Target:    events.Target{ID: userID, Type: "user"},
		Result:    events.ResultSuccess,
	})

	handlers.WriteJSON(w, http.StatusOK, nil)
}

// UnbanUser handles unbanning a user.
func (h *AdminUserHandler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	err := h.userService.ChangeUserStatus(r.Context(), userID, types.UserStatusActive)
	if err != nil {
		handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		return
	}

	h.auditLogger.Record(r.Context(), &events.AuditEvent{
		EventType: events.EventDataModified,
		Actor:     events.Actor{ID: r.Context().Value(middleware.UserIDContextKey).(string), Type: "user"},
		Target:    events.Target{ID: userID, Type: "user"},
		Result:    events.ResultSuccess,
	})

	handlers.WriteJSON(w, http.StatusOK, nil)
}
