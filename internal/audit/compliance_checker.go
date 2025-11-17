package audit

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ComplianceChecker runs automated checks against predefined compliance rules.
type ComplianceChecker struct {
	rules        []ComplianceRule
	auditRepo    AuditRepository
	identityRepo IdentityRepository
	logger       *zap.Logger
}

// ComplianceStandard defines the regulatory standard, e.g., GDPR, SOC2.
type ComplianceStandard string

const (
	StandardGDPR     ComplianceStandard = "GDPR"
	StandardSOC2     ComplianceStandard = "SOC2"
	StandardISO27001 ComplianceStandard = "ISO27001"
)

// ComplianceRule defines a single, verifiable compliance control.
type ComplianceRule struct {
	ID          string
	Name        string
	Standard    ComplianceStandard
	Description string
	CheckFunc   func(ctx context.Context, checker *ComplianceChecker) (*ComplianceResult, error)
}

// ComplianceResult is the outcome of a single compliance rule check.
type ComplianceResult struct {
	RuleID  string `json:"rule_id"`
	Status  string `json:"status"` // "pass", "fail", "error"
	Details string `json:"details,omitempty"`
	Error   error  `json:"-"`
}

// ComplianceReport is a collection of results for a specific compliance standard.
type ComplianceReport struct {
	Standard  ComplianceStandard  `json:"standard"`
	Timestamp time.Time           `json:"timestamp"`
	Results   []*ComplianceResult `json:"results"`
	PassRate  float64             `json:"pass_rate"`
}

// NewComplianceChecker creates a new checker with a set of rules.
func NewComplianceChecker(rules []ComplianceRule, auditRepo AuditRepository, identityRepo IdentityRepository, logger *zap.Logger) *ComplianceChecker {
	return &ComplianceChecker{
		rules:        rules,
		auditRepo:    auditRepo,
		identityRepo: identityRepo,
		logger:       logger.Named("compliance-checker"),
	}
}

// RunChecksForStandard executes all compliance checks for a given standard.
func (cc *ComplianceChecker) RunChecksForStandard(ctx context.Context, standard ComplianceStandard) (*ComplianceReport, error) {
	var results []*ComplianceResult
	var passedCount int

	rulesForStandard := cc.getRulesByStandard(standard)
	if len(rulesForStandard) == 0 {
		return nil, fmt.Errorf("no compliance rules defined for standard: %s", standard)
	}

	for _, rule := range rulesForStandard {
		result, err := rule.CheckFunc(ctx, cc)
		if err != nil {
			result = &ComplianceResult{RuleID: rule.ID, Status: "error", Details: err.Error(), Error: err}
		}
		results = append(results, result)
		if result.Status == "pass" {
			passedCount++
		}
	}

	report := &ComplianceReport{
		Standard:  standard,
		Timestamp: time.Now().UTC(),
		Results:   results,
		PassRate:  float64(passedCount) / float64(len(rulesForStandard)),
	}

	return report, nil
}

func (cc *ComplianceChecker) getRulesByStandard(standard ComplianceStandard) []ComplianceRule {
	var filteredRules []ComplianceRule
	for _, rule := range cc.rules {
		if rule.Standard == standard {
			filteredRules = append(filteredRules, rule)
		}
	}
	return filteredRules
}

// --- Check Functions ---

var (
	// ErrNotFound is returned when a query returns no results. This is not always an error.
	ErrNotFound = errors.New("not found")
)


// CheckGDPRDataRetention verifies that user accounts are not held longer than the configured retention period.
func CheckGDPRDataRetention(ctx context.Context, checker *ComplianceChecker) (*ComplianceResult, error) {
	// This check assumes a max retention period of 7 years.
	// In a real system, this would be configurable.
	cutoff := time.Now().UTC().Add(-7 * 365 * 24 * time.Hour)

	expiredUsers, err := checker.identityRepo.FindAccountsCreatedBefore(ctx, cutoff)
	if err != nil {
		if err == ErrNotFound {
			return &ComplianceResult{RuleID: "GDPR-05", Status: "pass", Details: "No user accounts found beyond the retention period."}, nil
		}
		return nil, err
	}

	if len(expiredUsers) > 0 {
		return &ComplianceResult{
			RuleID:  "GDPR-05",
			Status:  "fail",
			Details: fmt.Sprintf("Found %d user accounts with creation dates older than %s.", len(expiredUsers), cutoff.Format("2006-01-02")),
		}, nil
	}

	return &ComplianceResult{RuleID: "GDPR-05", Status: "pass"}, nil
}

// CheckSOC2MonitoringCoverage verifies that critical system events are being audited.
func CheckSOC2MonitoringCoverage(ctx context.Context, checker *ComplianceChecker) (*ComplianceResult, error) {
	// Check for a critical system event (e.g., config change) in the last 7 days.
	filter := QueryFilter{
		StartTimestamp: time.Now().UTC().Add(-7 * 24 * time.Hour),
		EventTypes:     []EventType{EventConfigChanged, EventKeyRotated},
	}

	events, err := checker.auditRepo.Query(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return &ComplianceResult{
			RuleID:  "SOC2-CC7.2",
			Status:  "fail",
			Details: "No critical system events (e.g., config change, key rotation) were audited in the last 7 days.",
		}, nil
	}

	return &ComplianceResult{RuleID: "SOC2-CC7.2", Status: "pass", Details: fmt.Sprintf("Found %d critical system events in the last 7 days.", len(events))}, nil
}
