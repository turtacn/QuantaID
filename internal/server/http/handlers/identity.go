package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/services/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"net/http"
)

// IdentityHandlers provides HTTP handlers for identity-related endpoints,
// such as user and group management.
type IdentityHandlers struct {
	identityService *identity.ApplicationService
	logger          utils.Logger
}

// NewIdentityHandlers creates a new set of identity-related handlers.
//
// Parameters:
//   - identityService: The application service for identity management logic.
//   - logger: The logger for handler-specific messages.
//
// Returns:
//   A new IdentityHandlers instance.
func NewIdentityHandlers(identityService *identity.ApplicationService, logger utils.Logger) *IdentityHandlers {
	return &IdentityHandlers{
		identityService: identityService,
		logger:          logger,
	}
}

// CreateUser is the HTTP handler for creating a new user.
// It decodes the request body, calls the identity service, and writes the
// newly created user object or an error as a JSON response.
func (h *IdentityHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req identity.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, types.ErrBadRequest.WithCause(err), http.StatusBadRequest)
		return
	}

	user, err := h.identityService.CreateUser(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			WriteJSONError(w, appErr, appErr.HttpStatus)
		} else {
			WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	WriteJSON(w, http.StatusCreated, user)
}

// GetUser is the HTTP handler for retrieving a user by their ID from the URL path.
// It extracts the user ID, calls the identity service, and writes the
// user object or an error as a JSON response.
func (h *IdentityHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, ok := vars["id"]
	if !ok {
		WriteJSONError(w, types.ErrBadRequest.WithDetails(map[string]string{"error": "missing user ID"}), http.StatusBadRequest)
		return
	}

	user, err := h.identityService.GetUserByID(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			WriteJSONError(w, appErr, appErr.HttpStatus)
		} else {
			WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	WriteJSON(w, http.StatusOK, user)
}
