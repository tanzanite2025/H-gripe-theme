package localbackup

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func createCustomerServiceCandidateColumnContractObservationForTest() *moduleRecoveryDatabaseObservation {
	return &moduleRecoveryDatabaseObservation{Columns: map[string]map[string]bool{
		"tickets":                          {"id": true, "subject": true, "customer_user_id": true},
		"ticket_messages":                  {"id": true, "content": true, "metadata": true, "message_type": true},
		"customer_service_inbox_states":    {"id": true, "last_read_message_id": true},
		"ticket_message_media_attachments": {"id": true, "media_asset_id": true},
	}, PrimaryKeys: map[string][]string{}}
}

func TestCustomerServiceCandidateColumnContractsRejectUnknownColumnsBeforeFullRowReadsOrWrites(t *testing.T) {
	observation := createCustomerServiceCandidateColumnContractObservationForTest()
	coverage, err := inspectCustomerServiceCandidateColumnReferenceContractCoverage(observation, "source")
	require.NoError(t, err)
	require.True(t, coverage.Complete)
	require.Equal(t, 11, coverage.RegisteredColumns)
	for _, table := range []string{"tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments"} {
		t.Run(table, func(t *testing.T) {
			observation := createCustomerServiceCandidateColumnContractObservationForTest()
			observation.Columns[table]["custom_html_or_json_reference"] = true
			// A primary key alone is not a contract for the meaning of a column.
			observation.PrimaryKeys[table] = []string{"custom_html_or_json_reference"}
			coverage, err := inspectCustomerServiceCandidateColumnReferenceContractCoverage(observation, "source")
			require.NoError(t, err)
			require.False(t, coverage.Complete)
			require.Equal(t, 1, coverage.UnregisteredColumns)
			require.Equal(t, 1, coverage.UnregisteredColumnCountsByTable[table])
			graph := &moduleRecoveryRowDependencyGraph{Observation: observation}
			_, _, err = readCustomerServiceCandidateSelectedFullRows(context.Background(), graph, nil)
			require.ErrorContains(t, err, "unregistered reference contract")
			require.ErrorContains(t, assertCustomerServiceCandidateTablesHaveSupportedWriteContracts(context.Background(), nil, observation), "unregistered reference contract")
		})
	}
	delete(observation.Columns, "tickets")
	coverage, err = inspectCustomerServiceCandidateColumnReferenceContractCoverage(observation, "source")
	require.NoError(t, err)
	require.False(t, coverage.Complete)
	require.Equal(t, []string{"tickets"}, coverage.MissingCandidateTables)
}

func TestCustomerServiceCandidateColumnContractsRequireCompleteObservedCompositeForeignKeyGroups(t *testing.T) {
	for _, scenario := range []string{"complete_composite_foreign_key", "wrong_kind", "missing_source", "missing_target", "different_group_lengths", "empty_group"} {
		t.Run(scenario, func(t *testing.T) {
			observation := createCustomerServiceCandidateColumnContractObservationForTest()
			observation.Columns["tickets"]["custom_parent_number"] = true
			observation.Columns["tickets"]["custom_parent_locale"] = true
			observation.Columns["custom_parent"] = map[string]bool{"number": true, "locale": true}
			reference := ModuleRecoveryDependencyReference{Kind: "foreign_key", SourceTable: "tickets", SourceColumns: []string{"custom_parent_number", "custom_parent_locale"}, TargetTable: "custom_parent", TargetColumns: []string{"number", "locale"}}
			switch scenario {
			case "wrong_kind":
				reference.Kind = "unregistered_business_reference"
			case "missing_source":
				delete(observation.Columns["tickets"], "custom_parent_locale")
			case "missing_target":
				delete(observation.Columns["custom_parent"], "locale")
			case "different_group_lengths":
				reference.TargetColumns = reference.TargetColumns[:1]
			case "empty_group":
				reference.SourceColumns, reference.TargetColumns = nil, nil
			}
			observation.References = []ModuleRecoveryDependencyReference{reference}
			coverage, err := inspectCustomerServiceCandidateColumnReferenceContractCoverage(observation, "target")
			require.NoError(t, err)
			require.Equal(t, scenario == "complete_composite_foreign_key", coverage.Complete)
			if coverage.Complete {
				require.Equal(t, 2, coverage.ForeignKeyColumns)
			} else {
				require.Positive(t, coverage.UnregisteredColumns)
			}
		})
	}
}

func TestCustomerServiceCandidateColumnContractsCountAndHashAllUnknownColumnsBeyondPublicSampleLimit(t *testing.T) {
	observation := createCustomerServiceCandidateColumnContractObservationForTest()
	for index := range 208 {
		observation.Columns["ticket_messages"][fmt.Sprintf("custom_reference_%03d", index)] = true
	}
	report := &ModuleRecoveryRowInspection{Complete: true, IssueCounts: map[string]int{}, IssueSampleLimit: 3, ConversationReferenceChecks: &CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection{}}
	graph := newModuleRecoveryRowDependencyGraph(nil, observation, "source", report, defaultModuleRecoveryRowInspectionLimits())
	require.NoError(t, graph.inspectCustomerServiceCandidateColumnReferenceSemantics())
	require.False(t, report.Complete)
	require.Equal(t, 208, report.BlockingIssueCounts["conversation_column_reference_contract_unregistered"])
	require.Len(t, report.Issues, 3)
	require.True(t, report.IssuesTruncated)
	coverage := report.ConversationReferenceChecks.ColumnContracts[0]
	require.Equal(t, 208, coverage.UnregisteredColumns)
	require.Len(t, coverage.UnregisteredColumnSamples, 200)
	require.True(t, coverage.SamplesTruncated)
	again, err := inspectCustomerServiceCandidateColumnReferenceContractCoverage(observation, "source")
	require.NoError(t, err)
	require.Equal(t, coverage.ColumnContractsSHA256, again.ColumnContractsSHA256)
	delete(observation.Columns["ticket_messages"], "custom_reference_207")
	changed, err := inspectCustomerServiceCandidateColumnReferenceContractCoverage(observation, "source")
	require.NoError(t, err)
	require.NotEqual(t, coverage.ColumnContractsSHA256, changed.ColumnContractsSHA256)
	require.Equal(t, coverage.UnregisteredColumnSamples, changed.UnregisteredColumnSamples)
}

func TestCustomerServiceCandidateColumnContractOptionalReportFieldPreservesLegacyCanonicalDigests(t *testing.T) {
	legacy := json.RawMessage(`{"version":1,"complete":true,"natural_key_comparisons":3,"snapshot_reference_checks":4,"shared_business_rows_preserved":5}`)
	var report CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection
	require.NoError(t, json.Unmarshal(legacy, &report))
	encoded, err := json.Marshal(report)
	require.NoError(t, err)
	require.JSONEq(t, string(legacy), string(encoded))
	legacyDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(legacy)
	require.NoError(t, err)
	currentDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(report)
	require.NoError(t, err)
	require.Equal(t, legacyDigest, currentDigest)
}
