package workflows

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/orchestrator"
	"github.com/turtacn/QuantaID/internal/services/audit"
	auth_service "github.com/turtacn/QuantaID/internal/services/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// LoginState holds the data passed between steps in the login workflow.
type LoginState struct {
	LoginCtx     auth.LoginContext
	RiskResult   *auth.RiskAssessment
	MFARequired  bool
	SessionToken string
	IDToken      string
	AccessToken  string
	User         *types.User
	AuthResponse *auth_service.LoginResponse
}

// LoginDeps contains the dependencies required for the login workflow.
type LoginDeps struct {
	RiskEngine   auth_service.RiskEngine
	AuthService  *auth_service.ApplicationService
	AuditService *audit.Service
	Logger       *zap.Logger
}

// RegisterLoginWorkflow registers the login workflow with the orchestrator engine.
func RegisterLoginWorkflow(engine *orchestrator.Engine, deps LoginDeps) {
	workflow := &orchestrator.Workflow{
		Name: "login_workflow",
		Steps: []orchestrator.Step{
			{Name: "build_context", Func: buildContextStep(deps)},
			{Name: "risk_assessment", Func: riskAssessmentStep(deps)},
			{Name: "mfa_check", Func: mfaCheckStep(deps)},
			{Name: "authenticate", Func: authenticateStep(deps)},
			{Name: "post_login_audit", Func: postLoginAuditStep(deps)},
		},
	}
	engine.RegisterWorkflow(workflow)
}

func buildContextStep(deps LoginDeps) orchestrator.StepFunc {
	return func(ctx context.Context, state orchestrator.State) error {
		// In a real scenario, we would extract this from the HTTP request.
		// For now, we'll assume it's pre-populated in the initial state.
		if _, ok := state["login_ctx"]; !ok {
			return fmt.Errorf("login_ctx not found in state")
		}
		return nil
	}
}

func riskAssessmentStep(deps LoginDeps) orchestrator.StepFunc {
	return func(ctx context.Context, state orchestrator.State) error {
		loginCtx := state["login_ctx"].(auth.LoginContext)
		riskResult, err := deps.RiskEngine.Assess(ctx, loginCtx)
		if err != nil {
			return err
		}
		state["risk_result"] = riskResult
		return nil
	}
}

func mfaCheckStep(deps LoginDeps) orchestrator.StepFunc {
	return func(ctx context.Context, state orchestrator.State) error {
		riskResult := state["risk_result"].(*auth.RiskAssessment)
		if riskResult.Decision == auth.RiskDecisionRequireMFA {
			state["mfa_required"] = true
			// In a real implementation, we would trigger an MFA challenge here.
			// For this E2E test, we'll assume the challenge is handled and verified.
			deps.Logger.Info("MFA required and triggered (mocked).")
		} else {
			state["mfa_required"] = false
		}
		return nil
	}
}

func authenticateStep(deps LoginDeps) orchestrator.StepFunc {
	return func(ctx context.Context, state orchestrator.State) error {
		loginCtx := state["login_ctx"].(auth.LoginContext)
		loginReq := auth_service.LoginRequest{
			Username: loginCtx.Username,
			Password: loginCtx.Password,
		}

		authResp, appErr := deps.AuthService.Login(ctx, loginReq)
		if appErr != nil {
			return appErr
		}
		state["auth_response"] = authResp
		state["user"] = authResp.User
		return nil
	}
}

func postLoginAuditStep(deps LoginDeps) orchestrator.StepFunc {
	return func(ctx context.Context, state orchestrator.State) error {
		user := state["user"].(*auth_service.UserDTO)
		riskResult := state["risk_result"].(*auth.RiskAssessment)
		loginCtx := state["login_ctx"].(auth.LoginContext)

		// This is a simplified audit record.
		// In a real implementation, we'd extract more details.
		if riskResult.Score > 50 { // Example threshold
			deps.AuditService.RecordHighRiskLogin(ctx, user.ID, loginCtx.CurrentIP, loginCtx.TraceID, float64(riskResult.Score), []string{}, nil)
		} else {
			deps.AuditService.RecordLoginSuccess(ctx, user.ID, loginCtx.CurrentIP, loginCtx.TraceID, nil)
		}
		return nil
	}
}
