package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/services/platform"
	"github.com/turtacn/QuantaID/pkg/types"
)

type DevCenterHandler struct {
	svc platform.Service
}

func NewDevCenterHandler(svc platform.Service) *DevCenterHandler {
	return &DevCenterHandler{svc: svc}
}

func (h *DevCenterHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/apps", h.ListApps).Methods("GET")
	r.HandleFunc("/apps", h.CreateApp).Methods("POST")
	r.HandleFunc("/connectors", h.ListConnectors).Methods("GET")
	r.HandleFunc("/connectors/{id}/enable", h.EnableConnector).Methods("POST")
	r.HandleFunc("/diagnostics", h.Diagnostics).Methods("GET")
}

func (h *DevCenterHandler) ListApps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apps, err := h.svc.ListApps(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(apps)
}

func (h *DevCenterHandler) CreateApp(w http.ResponseWriter, r *http.Request) {
	var req types.CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	app, err := h.svc.CreateApp(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

func (h *DevCenterHandler) ListConnectors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	connectors, err := h.svc.ListConnectors(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(connectors)
}

func (h *DevCenterHandler) EnableConnector(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	if err := h.svc.EnableConnector(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *DevCenterHandler) Diagnostics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	diagnostics, err := h.svc.Diagnostics(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(diagnostics)
}
