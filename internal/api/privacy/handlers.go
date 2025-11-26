package privacy

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/services/privacy"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/internal/server/middleware"
	"github.com/turtacn/QuantaID/pkg/types"
)

func getUserIDFromCtx(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		return "", types.ErrInternal.WithDetails(map[string]string{"reason": "User ID not found in context."})
	}
	return userID, nil
}

// Handlers provides HTTP handlers for privacy-related features.
type Handlers struct {
	privacyService *privacy.Service
}

// NewHandlers creates new privacy handlers.
func NewHandlers(privacyService *privacy.Service) *Handlers {
	return &Handlers{
		privacyService: privacyService,
	}
}

// RegisterRoutes registers the privacy routes to the router.
func (h *Handlers) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/privacy/consent", h.handleConsent).Methods("POST")
	r.HandleFunc("/privacy/export", h.handleExport).Methods("POST")
	r.HandleFunc("/privacy/erasure", h.handleErasure).Methods("POST")
}

func (h *Handlers) handleConsent(w http.ResponseWriter, r *http.Request) {
	var req privacy.GrantConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteJSONError(w, types.ErrInvalidRequest.WithCause(err), http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromCtx(r)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			handlers.WriteJSONError(w, appErr, http.StatusInternalServerError)
		} else {
			handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}
	req.UserID = userID

	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()

	if err := h.privacyService.GrantConsent(r.Context(), req); err != nil {
		if appErr, ok := err.(*types.Error); ok {
			handlers.WriteJSONError(w, appErr, http.StatusInternalServerError)
		} else {
			handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	handlers.WriteJSON(w, http.StatusOK, nil)
}

func (h *Handlers) handleExport(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromCtx(r)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			handlers.WriteJSONError(w, appErr, http.StatusInternalServerError)
		} else {
			handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	data, err := h.privacyService.CollectUserData(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			handlers.WriteJSONError(w, appErr, http.StatusInternalServerError)
		} else {
			handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	jsonData, err := h.privacyService.ExportToJSON(data)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			handlers.WriteJSONError(w, appErr, http.StatusInternalServerError)
		} else {
			handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\"export.json\"")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *Handlers) handleErasure(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromCtx(r)
	if err != nil {
		if appErr, ok := err.(*types.Error); ok {
			handlers.WriteJSONError(w, appErr, http.StatusInternalServerError)
		} else {
			handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	if err := h.privacyService.EraseAccount(r.Context(), userID); err != nil {
		if appErr, ok := err.(*types.Error); ok {
			handlers.WriteJSONError(w, appErr, http.StatusInternalServerError)
		} else {
			handlers.WriteJSONError(w, types.ErrInternal.WithCause(err), http.StatusInternalServerError)
		}
		return
	}

	handlers.WriteJSON(w, http.StatusOK, nil)
}
