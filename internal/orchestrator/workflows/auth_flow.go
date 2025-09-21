package workflows

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/internal/orchestrator"
	"github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/pkg/types"
)

// AuthWorkflow defines and registers the standard authentication workflow.
type AuthWorkflow struct {
	engine      *orchestrator.Engine
	authService *auth.ApplicationService
}

// NewAuthWorkflow creates and registers the authentication workflow.
func NewAuthWorkflow(engine *orchestrator.Engine, authService *auth.ApplicationService) *AuthWorkflow {
	awf := &AuthWorkflow{
		engine:      engine,
		authService: authService,
	}
	awf.register()
	return awf
}

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

func (awf *AuthWorkflow) validateInput(ctx context.Context, state orchestrator.State) error {
	_, userOk := state["username"]
	_, passOk := state["password"]
	if !userOk || !passOk {
		return fmt.Errorf("missing username or password in workflow state")
	}
	return nil
}

func (awf *AuthWorkflow) authenticatePrimary(ctx context.Context, state orchestrator.State) error {
	username := state["username"].(string)
	fmt.Printf("Authenticating user: %s\n", username)
	state["user"] = &types.User{ID: "user-123", Username: username}
	return nil
}

func (awf *AuthWorkflow) checkMfaRequired(ctx context.Context, state orchestrator.State) error {
	user := state["user"].(*types.User)
	isMfaRequired := false
	fmt.Printf("Checking MFA status for user: %s. Required: %t\n", user.Username, isMfaRequired)
	state["mfa_required"] = isMfaRequired
	return nil
}

func (awf *AuthWorkflow) issueTokens(ctx context.Context, state orchestrator.State) error {
	user := state["user"].(*types.User)
	fmt.Printf("Issuing tokens for user: %s\n", user.ID)
	state["tokens"] = &types.Token{AccessToken: "dummy-access-token", TokenType: "Bearer"}
	return nil
}

//Personal.AI order the ending
