package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/service"
)

func TestHistoricalMediaReadOnlyInventoryCommandValidatesArgumentsBeforeDatabaseConnection(t *testing.T) {
	for _, args := range [][]string{
		{"-batch-size", "0"}, {"-batch-size", "1001"}, {"-max-records-per-source", "0"},
		{"-max-issues", "0"}, {"-timeout", "0s"}, {"-format", "csv"}, {"unexpected-positional-argument"},
	} {
		var stdout, stderr bytes.Buffer
		err := run(append([]string{"audit-historical-media"}, args...), &stdout, &stderr)
		require.Error(t, err)
		require.Empty(t, stdout.String())
		require.NotContains(t, err.Error(), "connect")
	}
}

type historicalMediaInventoryReportWriterThatRejectsOutput struct{}

func (historicalMediaInventoryReportWriterThatRejectsOutput) Write([]byte) (int, error) {
	return 0, errors.New("historical inventory report destination unavailable")
}

func TestHistoricalMediaReadOnlyInventoryTextReportShowsCoverageAndPropagatesOutputFailure(t *testing.T) {
	report := service.HistoricalMediaReadOnlyInventoryReport{
		RegistrationIndexComplete: false, IssuesTruncated: true,
		Sources:     []service.HistoricalMediaReadOnlyInventorySourceCoverage{{Table: "posts", Field: "content", RowsScanned: 2, Complete: false, LastRecordID: 5}},
		IssueCounts: map[string]int{"business_managed_url_not_registered": 3},
		Issues:      []service.HistoricalMediaReadOnlyInventoryIssue{{Code: "business_managed_url_not_registered", Table: "posts", RecordID: 5, Field: "content.img.src", ReferenceSHA256: "address-fingerprint"}},
	}
	var output bytes.Buffer
	require.NoError(t, writeHistoricalMediaReadOnlyInventoryTextReport(&output, report))
	for _, required := range []string{"registration_index_complete=false", "requested_sources_complete=false", "issue_details_truncated=true", "Physical objects and unreferenced objects were not evaluated", "source=posts.content rows=2 last_id=5 complete=false", "reference_sha256=address-fingerprint"} {
		require.Contains(t, output.String(), required)
	}
	require.Error(t, writeHistoricalMediaReadOnlyInventoryTextReport(historicalMediaInventoryReportWriterThatRejectsOutput{}, report))
}

func TestHistoricalMediaReadOnlyInventoryCommandRejectsMissingExplicitStorageOriginBeforeLoadingConfig(t *testing.T) {
	t.Setenv("STORAGE_BASE_URL", "")
	var stdout, stderr bytes.Buffer
	err := runHistoricalMediaReadOnlyInventoryCommand([]string{"-config", "missing-config-that-must-not-be-opened.yaml"}, &stdout, &stderr)
	require.EqualError(t, err, "STORAGE_BASE_URL is required when auditing with -config")
	require.Empty(t, stdout.String())
}
