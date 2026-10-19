package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/service"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func runHistoricalMediaReadOnlyInventoryCommand(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("audit-historical-media", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "optional database app config file; storage identity uses runtime environment")
	batchSize := flags.Int("batch-size", 256, "rows loaded per read-only batch, 1..1000")
	maxRecords := flags.Int("max-records-per-source", 100000, "maximum rows per registration table or business field, 1..1000000")
	maxIssues := flags.Int("max-issues", 1000, "maximum issue details retained, 1..100000; totals still include all scanned issues")
	format := flags.String("format", "text", "report format: text or json")
	timeout := flags.Duration("timeout", 5*time.Minute, "positive maximum audit duration")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("audit-historical-media does not accept positional arguments")
	}
	if *batchSize < 1 || *batchSize > 1000 || *maxRecords < 1 || *maxRecords > 1000000 || *maxIssues < 1 || *maxIssues > 100000 || *timeout <= 0 {
		return errors.New("invalid historical media audit limits")
	}
	if *format != "text" && *format != "json" {
		return errors.New("format must be text or json")
	}
	if strings.TrimSpace(os.Getenv("STORAGE_BASE_URL")) == "" && strings.TrimSpace(*configPath) != "" {
		return errors.New("STORAGE_BASE_URL is required when auditing with -config")
	}
	databaseConfig, _, err := readRuntimeConfig(*configPath)
	if err != nil {
		return err
	}
	var dialector gorm.Dialector
	switch databaseConfig.Driver {
	case "postgres":
		dialector = postgres.Open(databaseConfig.GetDSN())
	case "mysql":
		dialector = mysql.Open(databaseConfig.GetDSN())
	default:
		return errors.New("historical media audit requires PostgreSQL or MySQL")
	}
	// Do not initialize application services, migrations, workers or writable
	// storage. Silent SQL logging also keeps report stdout valid JSON.
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("connect historical media audit database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)
	shared := storage.LoadConfigFromEnv()
	storefrontURL := envDefault("STOREFRONT_BASE_URL", shared.BaseURL)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	report, err := service.NewHistoricalMediaReadOnlyInventoryService(db, shared, storage.LoadSiteLogoConfigFromEnv(shared), storefrontURL).AuditHistoricalRegistrationsAndBusinessReferences(ctx, service.HistoricalMediaReadOnlyInventoryOptions{BatchSize: *batchSize, MaxRecordsPerSource: *maxRecords, MaxIssues: *maxIssues})
	if err != nil {
		return err
	}
	databaseTarget := url.URL{Scheme: databaseConfig.Driver, Host: net.JoinHostPort(databaseConfig.Host, strconv.Itoa(databaseConfig.Port)), Path: "/" + databaseConfig.Database}
	report.DatabaseTarget = databaseTarget.String()
	if *format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	return writeHistoricalMediaReadOnlyInventoryTextReport(stdout, report)
}

func writeHistoricalMediaReadOnlyInventoryTextReport(writer io.Writer, report service.HistoricalMediaReadOnlyInventoryReport) error {
	if _, err := fmt.Fprintf(writer, "Historical media read-only inventory generated_at=%s registrations=%d references=%d issues=%d\nDatabase: %s\nSnapshot: %s; registration_index_complete=%t; requested_sources_complete=%t; issue_details_truncated=%t\nCoverage: %s\nPhysical objects and unreferenced objects were not evaluated. Findings are for review, not deletion authorization.\n", report.GeneratedAt.Format(time.RFC3339), report.RegistrationsScanned, report.ReferencesScanned, report.TotalIssues, report.DatabaseTarget, report.SnapshotMode, report.RegistrationIndexComplete, report.RequestedSourcesComplete, report.IssuesTruncated, report.Coverage); err != nil {
		return err
	}
	configurationKeys := make([]string, 0, len(report.IdentityConfiguration))
	for key := range report.IdentityConfiguration {
		configurationKeys = append(configurationKeys, key)
	}
	sort.Strings(configurationKeys)
	for _, key := range configurationKeys {
		if _, err := fmt.Fprintf(writer, "identity_configuration.%s=%s\n", key, report.IdentityConfiguration[key]); err != nil {
			return err
		}
	}
	codes := make([]string, 0, len(report.IssueCounts))
	for code := range report.IssueCounts {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	for _, code := range codes {
		if _, err := fmt.Fprintf(writer, "%s=%d\n", code, report.IssueCounts[code]); err != nil {
			return err
		}
	}
	for _, source := range report.Sources {
		if _, err := fmt.Fprintf(writer, "source=%s.%s rows=%d last_id=%d complete=%t unavailable=%t remaining_rows=%d\n", source.Table, source.Field, source.RowsScanned, source.LastRecordID, source.Complete, source.Unavailable, source.RemainingRows); err != nil {
			return err
		}
	}
	for _, issue := range report.Issues {
		if _, err := fmt.Fprintf(writer, "%s table=%s id=%d field=%s asset_id=%d related_id=%d origin=%s reference_sha256=%s retained_deleted_record=%t\n", issue.Code, issue.Table, issue.RecordID, issue.Field, issue.AssetID, issue.RelatedRecordID, issue.Origin, issue.ReferenceSHA256, issue.RetainedDeletedRecord); err != nil {
			return err
		}
	}
	return nil
}
