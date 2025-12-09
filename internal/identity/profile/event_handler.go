package profile

import (
	"context"
	"time"
)

// EventBus defines the interface for subscribing to events
type EventBus interface {
	Subscribe(topic string, handler func(context.Context, interface{}) error)
}

// ProfileEventHandler handles events related to user profiles
type ProfileEventHandler struct {
	profileService *ProfileService
	profileBuilder *ProfileBuilder
}

// NewProfileEventHandler creates a new ProfileEventHandler
func NewProfileEventHandler(profileService *ProfileService, profileBuilder *ProfileBuilder) *ProfileEventHandler {
	return &ProfileEventHandler{
		profileService: profileService,
		profileBuilder: profileBuilder,
	}
}

// RegisterHandlers registers event handlers
func (h *ProfileEventHandler) RegisterHandlers(eventBus EventBus) {
	// These topics are placeholders. Real topic names depend on the event system implementation.
	eventBus.Subscribe("user.login", h.HandleLoginEvent)
	eventBus.Subscribe("device.anomaly", h.HandleAnomalyEvent)
	eventBus.Subscribe("device.registered", h.HandleNewDeviceEvent)
	eventBus.Subscribe("user.password.changed", h.HandlePasswordChangeEvent)
	eventBus.Subscribe("user.mfa.verified", h.HandleMFAEvent)
}

// HandleLoginEvent handles user login events
func (h *ProfileEventHandler) HandleLoginEvent(ctx context.Context, event interface{}) error {
	// Assuming event is a map or struct we can parse
	// For simplicity, we assume the event payload is passed as map[string]interface{}
	evt, ok := event.(map[string]interface{})
	if !ok {
		return nil
	}

	userID, _ := evt["user_id"].(string)
	if userID == "" {
		return nil
	}

	return h.profileBuilder.IncrementalUpdate(ctx, userID, map[string]interface{}{
		"type": "login",
		"success": evt["success"],
	})
}

// HandleAnomalyEvent handles anomaly events
func (h *ProfileEventHandler) HandleAnomalyEvent(ctx context.Context, event interface{}) error {
	evt, ok := event.(map[string]interface{})
	if !ok {
		return nil
	}

	userID, _ := evt["user_id"].(string)
	if userID == "" {
		return nil
	}

	anomalyType, _ := evt["anomaly_type"].(string)

	anomaly := AnomalyEvent{
		Type:      anomalyType,
		Timestamp: time.Now(),
		Details:   evt,
	}

	return h.profileService.HandleAnomalyEvent(ctx, userID, anomaly)
}

// HandleNewDeviceEvent handles new device registration events
func (h *ProfileEventHandler) HandleNewDeviceEvent(ctx context.Context, event interface{}) error {
	evt, ok := event.(map[string]interface{})
	if !ok {
		return nil
	}

	userID, _ := evt["user_id"].(string)
	if userID == "" {
		return nil
	}

	// We don't fetch and modify profile locally to avoid race condition and overwrite.
	// We delegate to HandleAnomalyEvent with type "new_device" which is handled by RiskScorer.
	return h.profileService.HandleAnomalyEvent(ctx, userID, AnomalyEvent{Type: "new_device", Timestamp: time.Now()})
}

// HandlePasswordChangeEvent handles password change events
func (h *ProfileEventHandler) HandlePasswordChangeEvent(ctx context.Context, event interface{}) error {
	evt, ok := event.(map[string]interface{})
	if !ok {
		return nil
	}

	userID, _ := evt["user_id"].(string)
	if userID == "" {
		return nil
	}

	profile, err := h.profileService.GetProfile(ctx, userID)
	if err != nil {
		return err
	}

	profile.Behavior.PasswordChangeCount++
	now := time.Now()
	profile.Behavior.LastPasswordChange = &now

	// Update behavior
	// Assuming Repository is accessible via Service or directly.
	// Ideally Service should expose UpdateBehavior or generic Update
	// For now, trigger refresh
	_, err = h.profileService.RefreshProfile(ctx, userID)
	return err
}

// HandleMFAEvent handles MFA verification events
func (h *ProfileEventHandler) HandleMFAEvent(ctx context.Context, event interface{}) error {
	evt, ok := event.(map[string]interface{})
	if !ok {
		return nil
	}

	userID, _ := evt["user_id"].(string)
	success, _ := evt["success"].(bool)

	if !success {
		return h.profileService.HandleAnomalyEvent(ctx, userID, AnomalyEvent{Type: "mfa_failure", Timestamp: time.Now()})
	}

	return h.profileBuilder.IncrementalUpdate(ctx, userID, map[string]interface{}{"type": "mfa_verified"})
}
