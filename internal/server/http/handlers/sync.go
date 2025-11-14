package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/services/sync"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

type SyncHandler struct {
	syncService *sync.LDAPSyncService
	logger      *zap.Logger
}

func NewSyncHandler(syncService *sync.LDAPSyncService, logger *zap.Logger) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		logger:      logger.Named("SyncHandler"),
	}
}

func (h *SyncHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/admin/sync/ldap/full", h.handleFullSync).Methods("POST")
	router.HandleFunc("/admin/sync/ldap/incremental", h.handleIncrementalSync).Methods("POST")
	router.HandleFunc("/admin/sync/ldap/status", h.handleGetStatus).Methods("GET")
}

func (h *SyncHandler) handleFullSync(w http.ResponseWriter, r *http.Request) {
	stats, err := h.syncService.FullSync(r.Context())
	if err != nil {
		h.logger.Error("Full sync failed", zap.Error(err))
		WriteJSONError(w, &types.Error{Message: "Failed to perform full sync"}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

func (h *SyncHandler) handleIncrementalSync(w http.ResponseWriter, r *http.Request) {
	sinceStr := r.URL.Query().Get("since")
	var since time.Time
	var err error

	if sinceStr == "" {
		since = time.Now().UTC().Add(-24 * time.Hour) // Default to last 24 hours
	} else {
		since, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			WriteJSONError(w, &types.Error{Message: "Invalid 'since' parameter"}, http.StatusBadRequest)
			return
		}
	}

	stats, err := h.syncService.IncrementalSync(r.Context(), since)
	if err != nil {
		h.logger.Error("Incremental sync failed", zap.Error(err))
		WriteJSONError(w, &types.Error{Message: "Failed to perform incremental sync"}, http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

func (h *SyncHandler) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	stats := h.syncService.GetLastSyncStatus()
	WriteJSON(w, http.StatusOK, stats)
}
