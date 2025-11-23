package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/services/webhook"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

type WebhookHandler struct {
	service *webhook.Service
	logger  utils.Logger
}

func NewWebhookHandler(service *webhook.Service, logger utils.Logger) *WebhookHandler {
	return &WebhookHandler{
		service: service,
		logger:  logger,
	}
}

type CreateSubscriptionRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

func (h *WebhookHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sub, err := h.service.CreateSubscription(req.URL, req.Events)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to create subscription", zap.Error(err))
		http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

func (h *WebhookHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs, err := h.service.ListSubscriptions()
	if err != nil {
		h.logger.Error(r.Context(), "Failed to list subscriptions", zap.Error(err))
		http.Error(w, "Failed to list subscriptions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}

func (h *WebhookHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteSubscription(id); err != nil {
		h.logger.Error(r.Context(), "Failed to delete subscription", zap.Error(err))
		http.Error(w, "Failed to delete subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WebhookHandler) RotateSecret(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sub, err := h.service.RotateSecret(id)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to rotate secret", zap.Error(err))
		http.Error(w, "Failed to rotate secret", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}
