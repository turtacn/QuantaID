package orchestrator

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// State is a map used to pass data between workflow steps.
type State map[string]interface{}

// StepFunc defines the function signature for a single workflow step.
type StepFunc func(ctx context.Context, state State) error

// Workflow defines a sequence of steps to be executed.
type Workflow struct {
	Name  string
	Steps []Step
}

// Step is a single, named step within a workflow.
type Step struct {
	Name string
	Func StepFunc
}

// Engine executes workflows.
type Engine struct {
	workflows map[string]*Workflow
	logger    utils.Logger
}

// NewEngine creates a new workflow engine.
func NewEngine(logger utils.Logger) *Engine {
	return &Engine{
		workflows: make(map[string]*Workflow),
		logger:    logger,
	}
}

// RegisterWorkflow makes a workflow available to the engine.
func (e *Engine) RegisterWorkflow(wf *Workflow) error {
	if _, exists := e.workflows[wf.Name]; exists {
		return fmt.Errorf("workflow with name '%s' is already registered", wf.Name)
	}
	e.workflows[wf.Name] = wf
	e.logger.Info(context.Background(), "Workflow registered", zap.String("workflow_name", wf.Name))
	return nil
}

// Execute runs a registered workflow with an initial state.
func (e *Engine) Execute(ctx context.Context, workflowName string, initialState State) (State, error) {
	wf, exists := e.workflows[workflowName]
	if !exists {
		return nil, fmt.Errorf("workflow '%s' not found", workflowName)
	}

	e.logger.Info(ctx, "Executing workflow", zap.String("workflow_name", workflowName))
	currentState := initialState

	for _, step := range wf.Steps {
		e.logger.Debug(ctx, "Executing step", zap.String("workflow_name", workflowName), zap.String("step_name", step.Name))
		err := step.Func(ctx, currentState)
		if err != nil {
			e.logger.Error(ctx, "Workflow step failed", zap.String("workflow_name", workflowName), zap.String("step_name", step.Name), zap.Error(err))
			return nil, fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}
	}

	e.logger.Info(ctx, "Workflow executed successfully", zap.String("workflow_name", workflowName))
	return currentState, nil
}

//Personal.AI order the ending
