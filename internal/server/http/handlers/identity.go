package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"net/http"
)

// IdentityHandlers provides HTTP handlers for identity-related endpoints.
type IdentityHandlers struct {
	identityService *identity.ApplicationService
	logger          utils.Logger
}

// NewIdentityHandlers creates a new set of identity handlers.
func NewIdentityHandlers(identityService *identity.ApplicationService, logger utils.Logger) *IdentityHandlers {
	return &IdentityHandlers{
		identityService: identityService,
		logger:          logger,
	}
}

// CreateUser is the HTTP handler for creating a new user.
func (h *IdentityHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req identity.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, types.ErrBadRequest.WithCause(err), http.StatusBadRequest)
		return
	}

	user, appErr := h.identityService.CreateUser(r.Context(), req)
	if appErr != nil {
		WriteJSONError(w, appErr, appErr.HttpStatus)
		return
	}

	WriteJSON(w, http.StatusCreated, user)
}

// GetUser is the HTTP handler for retrieving a user by their ID.
func (h *IdentityHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, ok := vars["id"]
	if !ok {
		WriteJSONError(w, types.ErrBadRequest.WithDetails(map[string]string{"error": "missing user ID"}), http.StatusBadRequest)
		return
	}

	user, appErr := h.identityService.GetUserByID(r.Context(), userID)
	if appErr != nil {
		WriteJSONError(w, appErr, appErr.HttpStatus)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

//Personal.AI order the ending
