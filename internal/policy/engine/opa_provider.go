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

// LoadPolicy loads or reloads the Rego policy for SDK mode from the specified path.
func (p *OPAProvider) LoadPolicy(ctx context.Context, path string) error {
	policyContent, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	// Query the whole package to get both 'allow' and 'deny' rules
	queryStr := "data.quantaid.authz"

	r := rego.New(
		rego.Query(queryStr),
		rego.Module(path, string(policyContent)),
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

// loadPolicy is an internal helper that uses the config path
func (p *OPAProvider) loadPolicy(ctx context.Context) error {
	return p.LoadPolicy(ctx, p.config.PolicyFile)
}

// Reload reloads the policy from the file.
func (p *OPAProvider) Reload(ctx context.Context) error {
	return p.loadPolicy(ctx)
}

// OPAResult holds the result of OPA evaluation
type OPAResult struct {
	Allow bool
	Deny  bool
}

// Evaluate checks if the request is allowed by the OPA policy.
// It returns a boolean for 'allow' (explicitly allowed), a boolean for 'deny' (explicitly denied), and an error.
func (p *OPAProvider) Evaluate(ctx context.Context, req EvaluationRequest) (bool, bool, error) {
	if !p.config.Enabled {
		return false, false, nil // If OPA is disabled, it neither allows nor denies
	}

	input := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    req.SubjectID,
			"roles": req.Context["roles"],
			// Add other user attributes as needed
		},
		"resource": map[string]interface{}{
			"id":   req.Resource,
			"type": req.Context["resource_type"],
		},
		"action": req.Action,
		"context": req.Context, // Pass raw context as well for flexibility
		"env":    req.Context["env"], // Preserve legacy env field
	}

	// Handle SDK Mode
	if p.config.Mode == "sdk" || p.config.Mode == "" {
		p.mu.RLock()
		query := p.query
		p.mu.RUnlock()

		rs, err := query.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			return false, false, fmt.Errorf("opa eval error: %w", err)
		}

		if len(rs) == 0 || len(rs[0].Expressions) == 0 {
			return false, false, nil
		}

		resultMap, ok := rs[0].Expressions[0].Value.(map[string]interface{})
		if !ok {
			return false, false, fmt.Errorf("unexpected result type from opa: expected map")
		}

		allow, _ := resultMap["allow"].(bool)
		deny, _ := resultMap["deny"].(bool)

		return allow, deny, nil
	}

	// Handle Sidecar Mode
	if p.config.Mode == "sidecar" {
		return p.evaluateSidecar(ctx, input)
	}

	return false, false, fmt.Errorf("unknown opa mode: %s", p.config.Mode)
}

// evaluateSidecar queries an external OPA server.
func (p *OPAProvider) evaluateSidecar(ctx context.Context, input interface{}) (bool, bool, error) {
	reqBody := map[string]interface{}{
		"input": input,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return false, false, fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false, false, fmt.Errorf("opa sidecar request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, false, fmt.Errorf("opa sidecar returned status: %d", resp.StatusCode)
	}

	// Expecting result to be a map with allow/deny
	var result struct {
		Result map[string]interface{} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, false, fmt.Errorf("failed to decode response: %w", err)
	}

	allow, _ := result.Result["allow"].(bool)
	deny, _ := result.Result["deny"].(bool)

	return allow, deny, nil
}
