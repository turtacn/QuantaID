package automator

import (
	"context"
)

// SecurityAction defines the interface for a security action that can be executed by the automator engine.
type SecurityAction interface {
	// ID returns the unique identifier of the action.
	ID() string
	// Execute performs the security action.
	Execute(ctx context.Context, input ActionInput) error
}

// ActionInput provides the context for a security action.
type ActionInput struct {
	UserID    string
	IP        string
	SessionID string
	RiskScore float64
	Metadata  map[string]any
}
