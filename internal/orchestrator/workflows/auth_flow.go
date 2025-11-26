package workflows

import (
	"context"
	"fmt"
	"time"

	auth_domain "github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/orchestrator"
	"github.com/turtacn/QuantaID/internal/security/automator"
	auth_service "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/pkg/types"
)

// AuthWorkflow defines the standard authentication workflow by registering a sequence of steps
// with the orchestrator engine. This modularizes the authentication process.
type AuthWorkflow struct {
	engine      *orchestrator.Engine
	authService *auth_service.ApplicationService
	riskEngine  auth_domain.RiskEngine
	automator   *automator.Engine
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
func NewAuthWorkflow(engine *orchestrator.Engine, authService *auth_service.ApplicationService, riskEngine auth_domain.RiskEngine, automator *automator.Engine) *AuthWorkflow {
	awf := &AuthWorkflow{
		engine:      engine,
		authService: authService,
		riskEngine:  riskEngine,
		automator:   automator,
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
			{Name: "assess_risk", Func: awf.assessRisk},
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

// assessRisk is a workflow step that evaluates the risk of the login attempt.
func (awf *AuthWorkflow) assessRisk(ctx context.Context, state orchestrator.State) error {
	user := state["user"].(*types.User)
	clientIP, _ := state["client_ip"].(string)
	userAgent, _ := state["user_agent"].(string)

	// For testing, we can override the last login details from the state.
	lastLoginIP, _ := state["last_login_ip"].(string)
	if lastLoginIP == "" {
		lastLoginIP = "192.168.1.1" // Default dummy data
	}
	lastLoginCountry, _ := state["last_login_country"].(string)
	if lastLoginCountry == "" {
		lastLoginCountry = "US" // Default dummy data
	}

	now, _ := state["now"].(time.Time)
	if now.IsZero() {
		now = time.Now().UTC()
	}

	authCtx := auth_domain.AuthContext{
		UserID:    user.ID,
		IPAddress: clientIP,
		UserAgent: userAgent,
		Timestamp: now,
	}

	score, level, err := awf.riskEngine.Evaluate(ctx, authCtx)
	if err != nil {
		return err
	}

	state["risk_score"] = score
	state["risk_level"] = level

	// Execute automated security responses.
	actionInput := automator.ActionInput{
		UserID:    user.ID,
		IP:        clientIP,
		RiskScore: float64(score),
		Metadata:  map[string]any{"user_agent": userAgent},
	}
	isBlocking, err := awf.automator.Execute(ctx, actionInput)
	if err != nil {
		// Log the error but don't fail the workflow, as the response actions
		// might not be critical to the login flow itself.
	}

	if isBlocking {
		return fmt.Errorf("login blocked by security policy")
	}

	if level == auth_domain.RiskLevelHigh {
		return fmt.Errorf("login blocked due to high risk")
	}

	return nil
}

// checkMfaRequired is a workflow step that determines if a multi-factor authentication
// step is necessary. This is a placeholder for more complex logic.
func (awf *AuthWorkflow) checkMfaRequired(ctx context.Context, state orchestrator.State) error {
	user := state["user"].(*types.User)
	level, ok := state["risk_level"].(auth_domain.RiskLevel)
	if !ok {
		// Default to MFA required if risk assessment is missing for some reason.
		state["mfa_required"] = true
		fmt.Printf("Risk assessment not found for user: %s. Defaulting to MFA required.\n", user.Username)
		return nil
	}

	isMfaRequired := level == auth_domain.RiskLevelMedium || level == auth_domain.RiskLevelHigh

	fmt.Printf("Checking MFA status for user: %s. Risk Level: %s, MFA Required: %t\n", user.Username, level, isMfaRequired)
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
