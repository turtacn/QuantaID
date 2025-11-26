package automator

import (
	"context"
	"golang.org/x/sync/errgroup"

	"github.com/turtacn/QuantaID/internal/config"
	"go.uber.org/zap"
)

// Engine orchestrates the execution of security actions based on risk scores.
type Engine struct {
	config   *config.SecurityConfig
	registry map[string]SecurityAction
	logger   *zap.Logger
}

// NewEngine creates a new AutomatorEngine.
func NewEngine(cfg *config.SecurityConfig, logger *zap.Logger) *Engine {
	return &Engine{
		config:   cfg,
		registry: make(map[string]SecurityAction),
		logger:   logger,
	}
}

// RegisterAction adds a security action to the engine's registry.
func (e *Engine) RegisterAction(action SecurityAction) {
	e.registry[action.ID()] = action
}

// Execute evaluates the risk score against the configured policies and executes the appropriate actions.
func (e *Engine) Execute(ctx context.Context, input ActionInput) (bool, error) {
	var actionsToRun []SecurityAction
	var isBlocking bool

	for _, policy := range e.config.ResponsePolicies {
		if input.RiskScore >= policy.RiskThreshold {
			e.logger.Info("triggering security policy",
				zap.String("policy_name", policy.Name),
				zap.Float64("risk_score", input.RiskScore),
				zap.Float64("threshold", policy.RiskThreshold),
			)
			if policy.IsBlocking {
				isBlocking = true
			}
			for _, actionID := range policy.ActionIDs {
				if action, ok := e.registry[actionID]; ok {
					actionsToRun = append(actionsToRun, action)
				} else {
					e.logger.Warn("security action not found in registry", zap.String("action_id", actionID))
				}
			}
		}
	}

	if len(actionsToRun) == 0 {
		return false, nil
	}

	// Use an errgroup to execute actions concurrently.
	g, gCtx := errgroup.WithContext(ctx)
	for _, action := range actionsToRun {
		act := action // a's' closure
		g.Go(func() error {
			err := act.Execute(gCtx, input)
			if err != nil {
				e.logger.Error("failed to execute security action",
					zap.String("action_id", act.ID()),
					zap.Error(err),
				)
				// Continue execution even if one action fails.
			}
			// TODO: Add audit logging for each executed action.
			return nil
		})
	}

	// We don't return the error from the errgroup because we want to continue
	// even if some actions fail. The errors are logged above.
	_ = g.Wait()

	return isBlocking, nil
}
