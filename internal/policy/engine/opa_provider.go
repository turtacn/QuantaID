package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/rego"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// OPAProvider handles communication with the Open Policy Agent.
// It supports both in-process (SDK) and remote (Sidecar) modes.
type OPAProvider struct {
	config utils.OPAConfig
	// For SDK mode
	query rego.PreparedEvalQuery
	mu    sync.RWMutex
	// For Sidecar mode
	httpClient *http.Client
}

// Config returns the configuration.
func (p *OPAProvider) Config() utils.OPAConfig {
	return p.config
}

// NewOPAProvider creates a new OPAProvider based on the configuration.
func NewOPAProvider(cfg utils.OPAConfig) (*OPAProvider, error) {
	p := &OPAProvider{
		config: cfg,
	}

	if !cfg.Enabled {
		return p, nil
	}

	if cfg.Mode == "sidecar" {
		p.httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
		return p, nil
	}

	// Default to SDK mode
	if cfg.PolicyFile == "" {
		return nil, fmt.Errorf("policy file path is required for OPA SDK mode")
	}

	ctx := context.Background()
	if err := p.loadPolicy(ctx); err != nil {
		return nil, fmt.Errorf("failed to load OPA policy: %w", err)
	}

	return p, nil
}

// loadPolicy loads or reloads the Rego policy for SDK mode.
func (p *OPAProvider) loadPolicy(ctx context.Context) error {
	policyContent, err := os.ReadFile(p.config.PolicyFile)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	// We assume the query is "data.quantaid.authz.allow" based on the provided rego file.
	// In a more generic implementation, the query string might be configurable.
	queryStr := "data.quantaid.authz.allow"

	r := rego.New(
		rego.Query(queryStr),
		rego.Module(p.config.PolicyFile, string(policyContent)),
	)

	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare rego query: %w", err)
	}

	p.mu.Lock()
	p.query = query
	p.mu.Unlock()
	return nil
}

// Reload reloads the policy from the file.
func (p *OPAProvider) Reload(ctx context.Context) error {
	return p.loadPolicy(ctx)
}

// Evaluate checks if the request is allowed by the OPA policy.
func (p *OPAProvider) Evaluate(ctx context.Context, req EvaluationRequest) (bool, error) {
	if !p.config.Enabled {
		return true, nil // If OPA is disabled, we don't block
	}

	input := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    req.SubjectID,
			"roles": req.Context["roles"], // Assuming roles are passed in context or we need to fetch them
			// Add other user attributes as needed
		},
		"resource": map[string]interface{}{
			"id":   req.Resource, // This might need parsing if resource is structured
			"type": req.Context["resource_type"],
		},
		"action": req.Action,
		"env":    req.Context["env"], // Time, IP, etc.
	}

	// Handle SDK Mode
	if p.config.Mode == "sdk" || p.config.Mode == "" {
		p.mu.RLock()
		query := p.query
		p.mu.RUnlock()

		rs, err := query.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			return false, fmt.Errorf("opa eval error: %w", err)
		}

		if len(rs) == 0 || len(rs[0].Expressions) == 0 {
			return false, nil // Undefined result usually means deny
		}

		allowed, ok := rs[0].Expressions[0].Value.(bool)
		if !ok {
			return false, fmt.Errorf("unexpected result type from opa")
		}

		return allowed, nil
	}

	// Handle Sidecar Mode
	if p.config.Mode == "sidecar" {
		return p.evaluateSidecar(ctx, input)
	}

	return false, fmt.Errorf("unknown opa mode: %s", p.config.Mode)
}

// evaluateSidecar queries an external OPA server.
func (p *OPAProvider) evaluateSidecar(ctx context.Context, input interface{}) (bool, error) {
	reqBody := map[string]interface{}{
		"input": input,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("opa sidecar request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("opa sidecar returned status: %d", resp.StatusCode)
	}

	// OPA returns {"result": true/false} for a boolean query
	var result struct {
		Result bool `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Result, nil
}
