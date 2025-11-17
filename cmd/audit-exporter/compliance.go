package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/turtacn/QuantaID/internal/audit"
	"go.uber.org/zap"
)

// mockIdentityRepository is a simple mock for the CLI tool.
// In a real application, this CLI might have a gRPC client to the main server
// or direct read-only access to the user database.
type mockIdentityRepository struct{}

func (m *mockIdentityRepository) FindAccountsCreatedBefore(ctx context.Context, cutoff time.Time) ([]audit.UserAccount, error) {
	// Return an empty slice to simulate a "pass" condition for the GDPR check.
	return []audit.UserAccount{}, nil
}

func generateComplianceReport(standard audit.ComplianceStandard, format audit.ReportFormat, repo audit.AuditRepository) (*audit.Report, error) {
	// 1. Set up the dependencies for the compliance checker
	logger := zap.NewNop()
	identityRepo := &mockIdentityRepository{}

	// 2. Define the rules that the checker will use
	rules := []audit.ComplianceRule{
		{
			ID: "GDPR-05", Standard: audit.StandardGDPR,
			CheckFunc: audit.CheckGDPRDataRetention,
		},
		{
			ID: "SOC2-CC7.2", Standard: audit.StandardSOC2,
			CheckFunc: audit.CheckSOC2MonitoringCoverage,
		},
	}

	checker := audit.NewComplianceChecker(rules, repo, identityRepo, logger)

	// 3. Run the compliance check for the specified standard
	complianceReport, err := checker.RunChecksForStandard(context.Background(), standard)
	if err != nil {
		return nil, err
	}

	// 4. Format the report into the desired output format
	var reportContent []byte
	var mimeType string

	switch format {
	case audit.FormatJSON:
		reportContent, err = json.MarshalIndent(complianceReport, "", "  ")
		if err != nil {
			return nil, err
		}
		mimeType = "application/json"
	default: // Default to a simple text representation
		reportContent = []byte(formatComplianceReportAsText(complianceReport))
		mimeType = "text/plain"
	}

	report := &audit.Report{
		Title:       string(standard) + " Compliance Report",
		GeneratedAt: time.Now().UTC(),
		Format:      format,
		Content:     reportContent,
		MimeType:    mimeType,
	}

	return report, nil
}

func formatComplianceReportAsText(report *audit.ComplianceReport) string {
	var builder strings.Builder
	builder.WriteString("========================================\n")
	builder.WriteString("Compliance Report\n")
	builder.WriteString("========================================\n")
	builder.WriteString("Standard: " + string(report.Standard) + "\n")
	builder.WriteString("Generated At: " + report.Timestamp.Format(time.RFC1123) + "\n")
	builder.WriteString("Pass Rate: " + fmt.Sprintf("%.2f%%\n", report.PassRate*100))
	builder.WriteString("----------------------------------------\n")

	for _, result := range report.Results {
		builder.WriteString("Rule ID: " + result.RuleID + "\n")
		builder.WriteString("Status: " + strings.ToUpper(result.Status) + "\n")
		if result.Details != "" {
			builder.WriteString("Details: " + result.Details + "\n")
		}
		builder.WriteString("----------------------------------------\n")
	}

	return builder.String()
}
