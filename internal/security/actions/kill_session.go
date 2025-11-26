package actions

import (
	"context"
	"fmt"

	"github.com/turtacn/QuantaID/internal/security/automator"
	"github.com/turtacn/QuantaID/internal/storage/redis"
)

// KillSessionAction implements the SecurityAction interface to revoke all of a user's sessions.
type KillSessionAction struct {
	sessionManager *redis.SessionManager
}

// NewKillSessionAction creates a new instance of KillSessionAction.
func NewKillSessionAction(sessionManager *redis.SessionManager) *KillSessionAction {
	return &KillSessionAction{sessionManager: sessionManager}
}

// ID returns the unique identifier for the action.
func (a *KillSessionAction) ID() string {
	return "kill_session"
}

// Execute revokes all active sessions for the given user.
func (a *KillSessionAction) Execute(ctx context.Context, input automator.ActionInput) error {
	if input.UserID == "" {
		return fmt.Errorf("UserID is required for KillSessionAction")
	}

	err := a.sessionManager.RevokeAllUserSessions(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to kill sessions for user %s: %w", input.UserID, err)
	}

	return nil
}
