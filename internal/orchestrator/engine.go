package orchestrator

import (
	"context"
	"fmt"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// State is a map used to pass data between workflow steps.
// It acts as a shared memory space for a single workflow execution.
type State map[string]interface{}

// StepFunc defines the function signature for a single workflow step.
// Each step receives the current context and state, and can modify the state.
// It returns an error to halt the workflow execution.
type StepFunc func(ctx context.Context, state State) error

// Workflow defines a sequence of steps to be executed in order.
type Workflow struct {
	// Name is the unique identifier for the workflow.
	Name string
	// Steps is the ordered list of steps that make up the workflow.
	Steps []Step
}

// Step is a single, named unit of work within a workflow.
type Step struct {
	// Name is the identifier for the step, used for logging and debugging.
	Name string
	// Func is the function that will be executed for this step.
	Func StepFunc
}

// Engine is responsible for registering and executing workflows.
// It manages the lifecycle of a workflow, executing its steps sequentially
// and handling state passing and error propagation.
type Engine struct {
	workflows map[string]*Workflow
	logger    utils.Logger
}

// NewEngine creates a new, empty workflow engine.
//
// Parameters:
//   - logger: The logger for engine-level messages.
//
// Returns:
//   A new workflow engine instance.
func NewEngine(logger utils.Logger) *Engine {
	return &Engine{
		workflows: make(map[string]*Workflow),
		logger:    logger,
	}
}

// RegisterWorkflow adds a new workflow to the engine's registry, making it available for execution.
//
// Parameters:
//   - wf: The workflow to register.
//
// Returns:
//   An error if a workflow with the same name is already registered.
func (e *Engine) RegisterWorkflow(wf *Workflow) error {
	if _, exists := e.workflows[wf.Name]; exists {
		return fmt.Errorf("workflow with name '%s' is already registered", wf.Name)
	}
	e.workflows[wf.Name] = wf
	e.logger.Info(context.Background(), "Workflow registered", zap.String("workflow_name", wf.Name))
	return nil
}

// Execute runs a registered workflow by its name with a given initial state.
// It processes each step in sequence, passing the state from one step to the next.
// If any step returns an error, the execution is halted and the error is returned.
//
// Parameters:
//   - ctx: The context for the entire workflow execution.
//   - workflowName: The name of the workflow to execute.
//   - initialState: The initial data to be used by the workflow.
//
// Returns:
//   The final state after the last successful step, or an error if the workflow fails.
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
