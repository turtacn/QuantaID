package handlers

import (
	"encoding/json"
	"github.com/turtacn/QuantaID/internal/services/application"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"net/http"
)

// ApplicationHandlers provides HTTP handlers for application-related endpoints.
type ApplicationHandlers struct {
	appService *application.ApplicationService
	logger     utils.Logger
}

// NewApplicationHandlers creates a new set of application handlers.
func NewApplicationHandlers(appService *application.ApplicationService, logger utils.Logger) *ApplicationHandlers {
	return &ApplicationHandlers{
		appService: appService,
		logger:     logger,
	}
}

// CreateApplication is the HTTP handler for creating a new application.
func (h *ApplicationHandlers) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var req application.CreateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, types.ErrBadRequest.WithCause(err), http.StatusBadRequest)
		return
	}

	app, appErr := h.appService.CreateApplication(r.Context(), req)
	if appErr != nil {
		WriteJSONError(w, appErr, appErr.HttpStatus)
		return
	}

	WriteJSON(w, http.StatusCreated, app)
}