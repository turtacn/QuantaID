package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/internal/storage/postgresql"
	"github.com/turtacn/QuantaID/pkg/audit/events"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	app := &cli.App{
		Name:  "audit-exporter",
		Usage: "A CLI tool to export audit logs and generate compliance reports.",
		Commands: []*cli.Command{
			{
				Name:  "export",
				Usage: "Export audit logs",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "start-date", Usage: "Start date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "end-date", Usage: "End date (YYYY-MM-DD)"},
					&cli.StringFlag{Name: "event-type", Usage: "Filter by event type"},
					&cli.StringFlag{Name: "format", Value: "json", Usage: "Output format (json, csv, pdf)"},
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output file path"},
				},
				Action: exportAction,
			},
			{
				Name:  "compliance-report",
				Usage: "Generate a compliance report",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "standard", Required: true, Usage: "Compliance standard (GDPR, SOC2)"},
					&cli.StringFlag{Name: "format", Value: "pdf", Usage: "Output format (pdf, json)"},
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output file path"},
				},
				Action: complianceReportAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// connectDB initializes the database connection.
func connectDB() (*gorm.DB, error) {
	// A simplified config loader for the CLI tool.
	// In a real app, this would be more robust.
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", dbHost, dbUser, dbPassword, dbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}


func exportAction(c *cli.Context) error {
	db, err := connectDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := postgresql.NewPostgresAuditLogRepository(db)
	generator := audit.NewReportGenerator(repo)

	filter := audit.QueryFilter{}
	if c.IsSet("start-date") {
		parsedTime, err := time.Parse("2006-01-02", c.String("start-date"))
		if err != nil {
			return fmt.Errorf("invalid start-date format: %w", err)
		}
		filter.StartTimestamp = parsedTime
	}
	if c.IsSet("end-date") {
		parsedTime, err := time.Parse("2006-01-02", c.String("end-date"))
		if err != nil {
			return fmt.Errorf("invalid end-date format: %w", err)
		}
		filter.EndTimestamp = parsedTime
	}
	if c.IsSet("event-type") {
		filter.EventTypes = []events.EventType{events.EventType(c.String("event-type"))}
	}

	req := &audit.ReportRequest{
		Format: audit.ReportFormat(c.String("format")),
		Filter: filter,
	}

	report, err := generator.GenerateReport(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	if c.IsSet("output") {
		return ioutil.WriteFile(c.String("output"), report.Content, 0644)
	}

	fmt.Println(string(report.Content))
	return nil
}

func complianceReportAction(c *cli.Context) error {
	db, err := connectDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := postgresql.NewPostgresAuditLogRepository(db)
	standard := audit.ComplianceStandard(c.String("standard"))
	format := audit.ReportFormat(c.String("format"))

	report, err := generateComplianceReport(standard, format, repo)
	if err != nil {
		return fmt.Errorf("failed to generate compliance report: %w", err)
	}

	if c.IsSet("output") {
		return ioutil.WriteFile(c.String("output"), report.Content, 0644)
	}

	fmt.Println(string(report.Content))
	return nil
}
