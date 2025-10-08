package workflows

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/internal/orchestrator"
	"github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/pkg/types"
)

// AuthWorkflow defines the standard authentication workflow by registering a sequence of steps
// with the orchestrator engine. This modularizes the authentication process.
type AuthWorkflow struct {
	engine      *orchestrator.Engine
	authService *auth.ApplicationService
}

// NewAuthWorkflow creates a new AuthWorkflow instance and registers the authentication
// workflow with the provided engine.
//
// Parameters:
//   - engine: The orchestrator engine to register the workflow with.
//   - authService: The application service for handling authentication logic.
//
// Returns:
//   A new instance of AuthWorkflow.
func NewAuthWorkflow(engine *orchestrator.Engine, authService *auth.ApplicationService) *AuthWorkflow {
	awf := &AuthWorkflow{
		engine:      engine,
		authService: authService,
	}
	awf.register()
	return awf
}

// register defines the sequence of steps for the "standard_auth_flow" and
// registers it with the orchestrator engine.
func (awf *AuthWorkflow) register() {
	wf := &orchestrator.Workflow{
		Name: "standard_auth_flow",
		Steps: []orchestrator.Step{
			{Name: "validate_input", Func: awf.validateInput},
			{Name: "authenticate_primary", Func: awf.authenticatePrimary},
			{Name: "check_mfa_required", Func: awf.checkMfaRequired},
			{Name: "issue_tokens", Func: awf.issueTokens},
		},
	}
	awf.engine.RegisterWorkflow(wf)
}

// validateInput is a workflow step that checks for the presence of required
// credentials (username and password) in the workflow's state.
func (awf *AuthWorkflow) validateInput(ctx context.Context, state orchestrator.State) error {
	_, userOk := state["username"]
	_, passOk := state["password"]
	if !userOk || !passOk {
		return fmt.Errorf("missing username or password in workflow state")
	}
	return nil
}

// authenticatePrimary is a workflow step that performs the primary authentication
// check (e.g., validating a password). In this dummy implementation, it simulates a successful authentication.
func (awf *AuthWorkflow) authenticatePrimary(ctx context.Context, state orchestrator.State) error {
	username := state["username"].(string)
	fmt.Printf("Authenticating user: %s\n", username)
	state["user"] = &types.User{ID: "user-123", Username: username}
	return nil
}

// checkMfaRequired is a workflow step that determines if a multi-factor authentication
// step is necessary. This is a placeholder for more complex logic.
func (awf *AuthWorkflow) checkMfaRequired(ctx context.Context, state orchestrator.State) error {
	user := state["user"].(*types.User)
	isMfaRequired := false
	fmt.Printf("Checking MFA status for user: %s. Required: %t\n", user.Username, isMfaRequired)
	state["mfa_required"] = isMfaRequired
	return nil
}

// issueTokens is the final step in a successful workflow, responsible for generating
// and adding the authentication tokens to the state. This is a placeholder implementation.
func (awf *AuthWorkflow) issueTokens(ctx context.Context, state orchestrator.State) error {
	user := state["user"].(*types.User)
	fmt.Printf("Issuing tokens for user: %s\n", user.ID)
	state["tokens"] = &types.Token{AccessToken: "dummy-access-token", TokenType: "Bearer"}
	return nil
}
