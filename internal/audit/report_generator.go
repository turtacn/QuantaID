package audit

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/turtacn/QuantaID/pkg/audit/events"
)

// ReportFormat defines the output format for a generated report.
type ReportFormat string

const (
	FormatJSON ReportFormat = "json"
	FormatCSV  ReportFormat = "csv"
	FormatPDF  ReportFormat = "pdf" // PDF generation can be complex and might require a dedicated library.
)

// ReportRequest defines the parameters for generating a report.
type ReportRequest struct {
	Format ReportFormat
	Filter QueryFilter
	// Could also include TemplateID for predefined reports
}

// Report represents a generated audit report.
type Report struct {
	Title       string
	GeneratedAt time.Time
	Format      ReportFormat
	Content     []byte
	MimeType    string
}

// ReportGenerator creates audit reports from the audit log data.
type ReportGenerator struct {
	repo AuditRepository
}

// NewReportGenerator creates a new report generator.
func NewReportGenerator(repo AuditRepository) *ReportGenerator {
	return &ReportGenerator{repo: repo}
}

// GenerateReport creates a report based on the provided request.
func (rg *ReportGenerator) GenerateReport(ctx context.Context, req *ReportRequest) (*Report, error) {
	// 1. Fetch data from the repository
	events, err := rg.repo.Query(ctx, req.Filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query audit events for report")
	}

	// 2. Generate content based on the requested format
	var reportContent []byte
	var mimeType string

	switch req.Format {
	case FormatJSON:
		reportContent, err = json.MarshalIndent(events, "", "  ")
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal report data to JSON")
		}
		mimeType = "application/json"
	case FormatCSV:
		reportContent, err = rg.exportCSV(events)
		if err != nil {
			return nil, errors.Wrap(err, "failed to export report data to CSV")
		}
		mimeType = "text/csv"
	case FormatPDF:
		// PDF generation is complex. As a placeholder, we generate a formatted
		// text summary and label it as a PDF. A real implementation would use
		// a library like go-fpdf.
		reportContent = rg.exportTextSummary(events)
		mimeType = "application/pdf" // Still claim PDF for file extension purposes
	default:
		return nil, fmt.Errorf("unsupported report format: %s", req.Format)
	}

	report := &Report{
		Title:       "Audit Log Report", // Title could be more dynamic
		GeneratedAt: time.Now().UTC(),
		Format:      req.Format,
		Content:     reportContent,
		MimeType:    mimeType,
	}

	return report, nil
}

// exportCSV converts a slice of audit events to a CSV byte slice.
func (rg *ReportGenerator) exportCSV(logEvents []*events.AuditEvent) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header row
	header := []string{
		"id", "timestamp", "event_type", "actor_id", "actor_type",
		"target_id", "target_type", "action", "result", "ip_address", "user_agent",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write data rows
	for _, event := range logEvents {
		record := []string{
			event.ID,
			event.Timestamp.Format(time.RFC3339),
			string(event.EventType),
			event.Actor.ID,
			event.Actor.Type,
			event.Target.ID,
			event.Target.Type,
			event.Action,
			string(event.Result),
			event.IPAddress,
			event.UserAgent,
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// exportTextSummary creates a human-readable text summary of audit events.
func (rg *ReportGenerator) exportTextSummary(logEvents []*events.AuditEvent) []byte {
	var buf bytes.Buffer

	buf.WriteString("========================================\n")
	buf.WriteString("       Audit Log Report Summary\n")
	buf.WriteString("========================================\n")
	buf.WriteString(fmt.Sprintf("Report Generated: %s\n", time.Now().UTC().Format(time.RFC1123)))
	buf.WriteString(fmt.Sprintf("Total Events: %d\n", len(logEvents)))
	buf.WriteString("----------------------------------------\n\n")

	for i, event := range logEvents {
		buf.WriteString(fmt.Sprintf("Event #%d\n", i+1))
		buf.WriteString(fmt.Sprintf("  ID: %s\n", event.ID))
		buf.WriteString(fmt.Sprintf("  Timestamp: %s\n", event.Timestamp.Format(time.RFC3339)))
		buf.WriteString(fmt.Sprintf("  Type: %s\n", event.EventType))
		buf.WriteString(fmt.Sprintf("  Actor: %s (%s)\n", event.Actor.ID, event.Actor.Type))
		buf.WriteString(fmt.Sprintf("  Action: %s\n", event.Action))
		buf.WriteString(fmt.Sprintf("  Result: %s\n", event.Result))
		if event.IPAddress != "" {
			buf.WriteString(fmt.Sprintf("  IP Address: %s\n", event.IPAddress))
		}
		buf.WriteString("\n")
	}

	buf.WriteString("--- End of Report ---\n")

	return buf.Bytes()
}
