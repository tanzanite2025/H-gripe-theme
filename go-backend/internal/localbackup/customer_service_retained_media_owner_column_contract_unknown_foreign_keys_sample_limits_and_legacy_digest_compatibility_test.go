package localbackup

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func createCustomerServiceRetainedMediaOwnerColumnObservationForTest() *moduleRecoveryDatabaseObservation {
	observation := createCustomerServiceCandidateColumnContractObservationForTest()
	observation.Columns["customer_service_agent_profiles"] = map[string]bool{"id": true, "avatar": true, "deleted_at": true}
	observation.Columns["ticket_auto_replies"] = map[string]bool{"id": true, "message_type": true, "metadata": true, "attachments": true, "is_active": true}
	return observation
}

func TestCustomerServiceRetainedMediaOwnerColumnContractsRejectUnknownColumnsEvenWithActualForeignKeys(t *testing.T) {
	for _, table := range []string{"customer_service_agent_profiles", "ticket_auto_replies"} {
		t.Run(table, func(t *testing.T) {
			observation := createCustomerServiceRetainedMediaOwnerColumnObservationForTest()
			coverage, err := inspectCustomerServiceRetainedMediaOwnerColumnReferenceContractCoverage(observation, "target")
			require.NoError(t, err)
			require.True(t, coverage.Complete)
			require.Equal(t, 8, coverage.RegisteredColumns)
			observation.Columns[table]["custom_media_asset_id"] = true
			observation.Columns["media_assets"] = map[string]bool{"id": true}
			observation.References = []ModuleRecoveryDependencyReference{{Kind: "foreign_key", SourceTable: table, SourceColumns: []string{"custom_media_asset_id"}, TargetTable: "media_assets", TargetColumns: []string{"id"}}}
			coverage, err = inspectCustomerServiceRetainedMediaOwnerColumnReferenceContractCoverage(observation, "target")
			require.NoError(t, err)
			require.False(t, coverage.Complete)
			require.Equal(t, 1, coverage.UnregisteredColumns)
			require.Equal(t, 1, coverage.UnregisteredColumnCountsByTable[table])
			_, _, err = readCustomerServiceCandidateSelectedFullRows(context.Background(), &moduleRecoveryRowDependencyGraph{Observation: observation}, nil)
			require.ErrorContains(t, err, "retained avatar profile or auto reply")
			require.ErrorContains(t, assertCustomerServiceCandidateTablesHaveSupportedWriteContracts(context.Background(), nil, observation), "retained avatar profile or auto reply")
		})
	}
}

func TestCustomerServiceRetainedMediaOwnerColumnContractsCountAllUnknownFieldsAndKeepLegacyReportDigests(t *testing.T) {
	observation := createCustomerServiceRetainedMediaOwnerColumnObservationForTest()
	for index := range 208 {
		observation.Columns["ticket_auto_replies"][fmt.Sprintf("custom_reference_%03d", index)] = true
	}
	report := &ModuleRecoveryRowInspection{Complete: true, IssueCounts: map[string]int{}, IssueSampleLimit: 2, ConversationReferenceChecks: &CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection{}}
	graph := newModuleRecoveryRowDependencyGraph(nil, observation, "target", report, defaultModuleRecoveryRowInspectionLimits())
	require.NoError(t, graph.inspectCustomerServiceRetainedMediaOwnerColumnReferenceSemantics())
	require.False(t, report.Complete)
	require.Equal(t, 208, report.BlockingIssueCounts["retained_media_owner_column_reference_contract_unregistered"])
	require.Len(t, report.Issues, 2)
	require.True(t, report.IssuesTruncated)
	coverage := report.ConversationReferenceChecks.RetainedMediaOwnerColumnContracts[0]
	require.Len(t, coverage.UnregisteredColumnSamples, 200)
	require.True(t, coverage.SamplesTruncated)
	delete(observation.Columns["ticket_auto_replies"], "custom_reference_207")
	changed, err := inspectCustomerServiceRetainedMediaOwnerColumnReferenceContractCoverage(observation, "target")
	require.NoError(t, err)
	require.Equal(t, coverage.UnregisteredColumnSamples, changed.UnregisteredColumnSamples)
	require.NotEqual(t, coverage.ColumnContractsSHA256, changed.ColumnContractsSHA256)
	delete(observation.Columns, "customer_service_agent_profiles")
	missing, err := inspectCustomerServiceRetainedMediaOwnerColumnReferenceContractCoverage(observation, "target")
	require.NoError(t, err)
	require.Equal(t, []string{"customer_service_agent_profiles"}, missing.MissingRetainedOwnerTables)
	require.False(t, missing.Complete)
	// Version 5 reports already contained candidate column contracts. Adding the
	// retained-owner field must not change their canonical serialized identity.
	legacy := json.RawMessage(`{"version":1,"complete":true,"natural_key_comparisons":0,"snapshot_reference_checks":0,"shared_business_rows_preserved":0,"column_contracts":[{"version":1,"side":"target","complete":true,"observed_columns":0,"registered_columns":0,"foreign_key_columns":0,"unregistered_columns":0,"unregistered_column_counts_by_table":{},"missing_candidate_tables":[],"column_contracts_sha256":"","unregistered_column_samples":[],"samples_truncated":false,"sample_limit":200}]}`)
	var legacyReport CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection
	require.NoError(t, json.Unmarshal(legacy, &legacyReport))
	legacyDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(legacy)
	require.NoError(t, err)
	currentDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(legacyReport)
	require.NoError(t, err)
	require.Equal(t, legacyDigest, currentDigest)
}

func TestCustomerServiceAvatarRelationshipProjectionUsesTheActualPersistedColumnWithoutProfileContactValues(t *testing.T) {
	observation := createCustomerServiceRetainedMediaOwnerColumnObservationForTest()
	observation.PrimaryKeys["customer_service_agent_profiles"] = []string{"id"}
	for _, column := range []string{"name", "email", "whats_app", "avatar_url", "custom_metadata"} {
		observation.Columns["customer_service_agent_profiles"][column] = true
	}
	require.Equal(t, []string{"avatar", "id"}, moduleRecoveryDependencyProjectionColumns(observation, "customer_service_agent_profiles"))
	coverage, err := inspectCustomerServiceRetainedMediaOwnerColumnReferenceContractCoverage(observation, "target")
	require.NoError(t, err)
	require.False(t, coverage.Complete)
	require.Equal(t, 2, coverage.UnregisteredColumns)
}
