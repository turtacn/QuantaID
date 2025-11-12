package mfa

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/pkg/plugins"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// Manager orchestrates the multi-factor authentication process.
type Manager struct {
	pluginManager  *plugins.Manager
	logger         utils.Logger
	challengeStore map[string]*types.MFAChallenge
}

// NewManager creates a new MFA manager.
func NewManager(pluginManager *plugins.Manager, logger utils.Logger) *Manager {
	return &Manager{
		pluginManager:  pluginManager,
		logger:         logger,
		challengeStore: make(map[string]*types.MFAChallenge),
	}
}

// TriggerChallenge determines the appropriate MFA method for a user and sends a challenge.
func (m *Manager) TriggerChallenge(ctx context.Context, user *types.User) (*types.MFAChallenge, error) {
	mfaMethod := m.getMfaMethodForUser(user)
	pluginName := m.getPluginNameForMethod(mfaMethod)

	plugin, err := m.pluginManager.GetPlugin(pluginName)
	if err != nil {
		m.logger.Error(ctx, "Failed to get MFA provider plugin", zap.Error(err), zap.String("plugin", pluginName))
		return nil, types.ErrPluginNotFound.WithCause(err)
	}

	provider, ok := plugin.(plugins.IMFAProvider)
	if !ok {
		err := fmt.Errorf("plugin '%s' does not implement IMFAProvider", pluginName)
		m.logger.Error(ctx, "Plugin type mismatch for MFA", zap.Error(err))
		return nil, types.ErrPluginLoadFailed.WithCause(err)
	}

	challenge, err := provider.SendChallenge(ctx, user)
	if err != nil {
		m.logger.Error(ctx, "MFA provider failed to send challenge", zap.Error(err), zap.String("plugin", pluginName))
		return nil, err
	}

	m.challengeStore[challenge.ChallengeID] = challenge
	m.logger.Info(ctx, "MFA challenge sent successfully", zap.String("challengeID", challenge.ChallengeID), zap.String("provider", string(mfaMethod)))

	return challenge, nil
}

// VerifyChallenge verifies a user's response to an MFA challenge.
func (m *Manager) VerifyChallenge(ctx context.Context, challengeID, code string) (bool, error) {
	challenge, exists := m.challengeStore[challengeID]
	if !exists {
		return false, types.ErrMfaChallengeInvalid.WithDetails(map[string]string{"reason": "not found or expired"})
	}

	pluginName := m.getPluginNameForMethod(challenge.MFAProvider)
	plugin, err := m.pluginManager.GetPlugin(pluginName)
	if err != nil {
		return false, types.ErrPluginNotFound.WithCause(err)
	}
	provider, ok := plugin.(plugins.IMFAProvider)
	if !ok {
		return false, types.ErrPluginLoadFailed
	}

	isValid, err := provider.VerifyChallenge(ctx, challengeID, code)
	if err != nil {
		return false, err
	}

	delete(m.challengeStore, challengeID)

	if !isValid {
		metrics.MFAVerificationsTotal.WithLabelValues("failure").Inc()
		m.logger.Warn(ctx, "MFA verification failed", zap.String("challengeID", challengeID))
		return false, nil
	}

	metrics.MFAVerificationsTotal.WithLabelValues("success").Inc()
	m.logger.Info(ctx, "MFA verification successful", zap.String("challengeID", challengeID))
	return true, nil
}

func (m *Manager) getMfaMethodForUser(user *types.User) types.AuthMethod {
	return types.AuthMethodTOTP
}

func (m *Manager) getPluginNameForMethod(method types.AuthMethod) string {
	switch method {
	case types.AuthMethodTOTP:
		return "totp_provider"
	case types.AuthMethodSMS:
		return "sms_provider"
	default:
		return ""
	}
}

