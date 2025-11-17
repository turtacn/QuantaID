package oauth

import (
	"context"
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/pkg/types"
)

// DeviceFlowHandler manages the Device Authorization Grant flow.
type DeviceFlowHandler struct {
	deviceCodeRepo DeviceCodeRepository
	tokenService   TokenService
	config         DeviceFlowConfig
}

// DeviceFlowConfig holds the configuration for the Device Authorization Grant.
type DeviceFlowConfig struct {
	DeviceCodeLength   int           `yaml:"deviceCodeLength"`
	UserCodeLength     int           `yaml:"userCodeLength"`
	UserCodeCharset    string        `yaml:"userCodeCharset"`
	ExpiresIn          time.Duration `yaml:"expiresIn"`
	PollingInterval    int           `yaml:"pollingInterval"`
	VerificationURI    string        `yaml:"verificationUri"`
}

// DeviceAuthorizationResponse is the response from the device authorization endpoint.
type DeviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval,omitempty"`
}

// DeviceCodeRecord represents a device code stored in the repository.
type DeviceCodeRecord struct {
	DeviceCode string
	UserCode   string
	ClientID   string
	Scope      string
	Status     string // "pending", "authorized", "denied", "expired"
	UserID     string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

// DeviceCodeRepository defines the interface for storing and retrieving device codes.
type DeviceCodeRepository interface {
	Create(ctx context.Context, record *DeviceCodeRecord) error
	GetByDeviceCode(ctx context.Context, deviceCode string) (*DeviceCodeRecord, error)
	GetByUserCode(ctx context.Context, userCode string) (*DeviceCodeRecord, error)
	UpdateStatus(ctx context.Context, deviceCode, status string) error
	MarkUsed(ctx context.Context, deviceCode string) error
	Update(ctx context.Context, deviceCode string, updates map[string]interface{}) error
}

// TokenService defines the interface for issuing tokens.
type TokenService interface {
	IssueTokens(ctx context.Context, req types.TokenRequest) (*types.Token, error)
}

// NewDeviceFlowHandler creates a new DeviceFlowHandler.
func NewDeviceFlowHandler(
	deviceCodeRepo DeviceCodeRepository,
	tokenService TokenService,
	config DeviceFlowConfig,
) *DeviceFlowHandler {
	return &DeviceFlowHandler{
		deviceCodeRepo: deviceCodeRepo,
		tokenService:   tokenService,
		config:         config,
	}
}

// HandleDeviceAuthorizationRequest handles the initial request to the device authorization endpoint.
func (h *DeviceFlowHandler) HandleDeviceAuthorizationRequest(ctx context.Context, clientID, scope string) (*DeviceAuthorizationResponse, error) {
	deviceCode := generateSecureCode(h.config.DeviceCodeLength)
	userCode := generateUserFriendlyCode(h.config.UserCodeLength, h.config.UserCodeCharset)

	record := &DeviceCodeRecord{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		ClientID:   clientID,
		Scope:      scope,
		Status:     "pending",
		ExpiresAt:  time.Now().Add(h.config.ExpiresIn),
	}

	if err := h.deviceCodeRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to store device code: %w", err)
	}

	return &DeviceAuthorizationResponse{
		DeviceCode:              deviceCode,
		UserCode:                userCode,
		VerificationURI:         h.config.VerificationURI,
		VerificationURIComplete: fmt.Sprintf("%s?user_code=%s", h.config.VerificationURI, userCode),
		ExpiresIn:               int(h.config.ExpiresIn.Seconds()),
		Interval:                h.config.PollingInterval,
	}, nil
}

// HandleDeviceTokenRequest handles the token request from the device.
func (h *DeviceFlowHandler) HandleDeviceTokenRequest(ctx context.Context, deviceCode, clientID string) (*types.Token, error) {
	record, err := h.deviceCodeRepo.GetByDeviceCode(ctx, deviceCode)
	if err != nil {
		return nil, types.ErrInvalidGrant
	}

	if time.Now().After(record.ExpiresAt) {
		h.deviceCodeRepo.UpdateStatus(ctx, deviceCode, "expired")
		return nil, types.ErrExpiredToken
	}

	switch record.Status {
	case "pending":
		return nil, types.ErrAuthorizationPending
	case "denied":
		return nil, types.ErrAccessDenied
	case "authorized":
		tokens, err := h.tokenService.IssueTokens(ctx, types.TokenRequest{
			GrantType: "urn:ietf:params:oauth:grant-type:device_code",
			ClientID:  record.ClientID,
			UserID:    record.UserID,
			Scope:     record.Scope,
		})
		if err != nil {
			return nil, err
		}

		h.deviceCodeRepo.MarkUsed(ctx, deviceCode)
		return tokens, nil
	default:
		return nil, types.ErrInvalidGrant
	}
}

// ActivateDeviceCode is called when the user authorizes the device.
func (h *DeviceFlowHandler) ActivateDeviceCode(ctx context.Context, userCode, userID string) error {
	record, err := h.deviceCodeRepo.GetByUserCode(ctx, userCode)
	if err != nil {
		return fmt.Errorf("invalid user code: %w", err)
	}

	updates := map[string]interface{}{
		"status":  "authorized",
		"user_id": userID,
	}
	return h.deviceCodeRepo.Update(ctx, record.DeviceCode, updates)
}

// TODO: Implement these functions with proper random string generation.
func generateSecureCode(length int) string {
	return "temp_device_code"
}

func generateUserFriendlyCode(length int, charset string) string {
	return "ABCD-EFGH"
}
