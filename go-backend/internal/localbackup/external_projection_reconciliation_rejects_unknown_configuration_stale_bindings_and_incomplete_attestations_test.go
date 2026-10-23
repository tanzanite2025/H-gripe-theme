package localbackup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/stretchr/testify/require"
)

func syntheticExternalProjectionBoundaryForRecoveryReleaseTests() maintenancegate.ExternalProjectionBoundary {
	return maintenancegate.ExternalProjectionBoundary{
		Version: 1, CustomerServiceSearchMode: "external_delete_only", SearchDestinationSHA256: strings.Repeat("a", 64),
		RedisDestinationSHA256: strings.Repeat("b", 64), CustomerServiceRealtimeMode: "redis_stream",
		RealtimeStreamSHA256: strings.Repeat("c", 64), RealtimeDedupeNamespaceSHA256: strings.Repeat("d", 64),
		MediaDeliveryConfigurationSHA256: strings.Repeat("e", 64),
	}
}

func syntheticExternalProjectionAttestationsForReconciliationReport(report *ReconciliationReport) *ExternalProjectionReconciliationEvidence {
	contract := report.ExternalProjectionReconciliation
	evidence := &ExternalProjectionReconciliationEvidence{ContractSHA256: contract.ContractSHA256}
	for _, requirement := range contract.Requirements {
		evidence.Attestations = append(evidence.Attestations, ExternalProjectionReconciliationAttestation{
			Projection: requirement.Projection, Action: requirement.Action,
			EvidenceReference: "synthetic-isolated-provider-report-no-external-connection",
			EvidenceSHA256:    strings.Repeat("f", 64),
		})
	}
	return evidence
}

func TestExternalProjectionReconciliationRequiresKnownBusinessConfigurationAndAllFourAttestations(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager := &Manager{Configuration: cfg}
	task := &Task{ID: newTaskID(), Kind: "upload", Manifest: &Manifest{PackageID: newTaskID(), Scope: "whole_application"}, Journal: &RecoveryJournal{NewDatabaseOID: 123, DatabaseSwitched: true}}
	report := &ReconciliationReport{DatabaseFingerprint: strings.Repeat("1", 64), MediaFingerprint: strings.Repeat("2", 64)}
	legacy := manager.buildTaskBoundExternalProjectionReconciliationContract(task, report)
	require.Contains(t, legacy.Blockers, "external_projection_configuration_unknown_reregister_with_current_business_build")
	require.Error(t, validateTaskBoundExternalProjectionReconciliationEvidence(legacy, nil))
	boundary := syntheticExternalProjectionBoundaryForRecoveryReleaseTests()
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets, boundary))
	report.ExternalProjectionReconciliation = manager.buildTaskBoundExternalProjectionReconciliationContract(task, report)
	contract := report.ExternalProjectionReconciliation
	require.Empty(t, contract.Blockers)
	require.Equal(t, "external_search_rebuilt_and_verified_against_restored_sql", contract.Requirements[0].Action)
	require.Equal(t, "scoped_stream_and_dedupe_invalidated_with_http_baseline_verified", contract.Requirements[2].Action)
	evidence := syntheticExternalProjectionAttestationsForReconciliationReport(report)
	require.NoError(t, validateTaskBoundExternalProjectionReconciliationEvidence(contract, evidence))
	for _, scenario := range []struct {
		name   string
		change func(*ExternalProjectionReconciliationEvidence)
	}{
		{"wrong_contract", func(e *ExternalProjectionReconciliationEvidence) { e.ContractSHA256 = strings.Repeat("a", 64) }},
		{"missing_projection", func(e *ExternalProjectionReconciliationEvidence) { e.Attestations = e.Attestations[:3] }},
		{"duplicate_projection", func(e *ExternalProjectionReconciliationEvidence) { e.Attestations[1] = e.Attestations[0] }},
		{"unknown_projection", func(e *ExternalProjectionReconciliationEvidence) { e.Attestations[0].Projection = "all_done" }},
		{"unsupported_action", func(e *ExternalProjectionReconciliationEvidence) { e.Attestations[0].Action = "delete_all" }},
		{"shared_redis_flush", func(e *ExternalProjectionReconciliationEvidence) { e.Attestations[1].Action = "flushdb" }},
		{"stream_only", func(e *ExternalProjectionReconciliationEvidence) { e.Attestations[2].Action = "stream_cleared" }},
		{"empty_reference", func(e *ExternalProjectionReconciliationEvidence) { e.Attestations[0].EvidenceReference = "    " }},
		{"oversized_reference", func(e *ExternalProjectionReconciliationEvidence) {
			e.Attestations[0].EvidenceReference = strings.Repeat("x", 2001)
		}},
		{"missing_record_digest", func(e *ExternalProjectionReconciliationEvidence) { e.Attestations[0].EvidenceSHA256 = "" }},
		{"nonhex_record_digest", func(e *ExternalProjectionReconciliationEvidence) {
			e.Attestations[0].EvidenceSHA256 = strings.Repeat("z", 64)
		}},
		{"zero_record_digest", func(e *ExternalProjectionReconciliationEvidence) {
			e.Attestations[0].EvidenceSHA256 = strings.Repeat("0", 64)
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			encoded, err := json.Marshal(evidence)
			require.NoError(t, err)
			var changed ExternalProjectionReconciliationEvidence
			require.NoError(t, json.Unmarshal(encoded, &changed))
			scenario.change(&changed)
			require.Error(t, validateTaskBoundExternalProjectionReconciliationEvidence(contract, &changed))
		})
	}
	t.Run("sql_only_and_process_local_are_explicit", func(t *testing.T) {
		boundary.CustomerServiceSearchMode = "sql_only"
		boundary.SearchDestinationSHA256 = ""
		boundary.CustomerServiceRealtimeMode = "process_local"
		boundary.RealtimeStreamSHA256 = ""
		boundary.RealtimeDedupeNamespaceSHA256 = ""
		require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets, boundary))
		updated := manager.buildTaskBoundExternalProjectionReconciliationContract(task, report)
		require.Equal(t, "sql_search_verified_without_external_index", updated.Requirements[0].Action)
		require.Equal(t, "process_local_realtime_restart_and_http_baseline_verified", updated.Requirements[2].Action)
		require.Error(t, validateTaskBoundExternalProjectionReconciliationEvidence(updated, evidence))
	})
}

func TestExternalProjectionReconciliationInvalidatesEveryRecoveryIdentityAndConfigurationChange(t *testing.T) {
	for _, scenario := range []string{"task", "package", "site", "environment", "database_oid", "database_content", "media", "redis_destination", "search_destination", "realtime_stream", "realtime_dedupe", "media_delivery", "missing_boundary", "invalid_boundary", "legacy_boundary"} {
		t.Run(scenario, func(t *testing.T) {
			cfg := createIndependentRecoveryTestConfiguration(t)
			boundary := syntheticExternalProjectionBoundaryForRecoveryReleaseTests()
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets, boundary))
			manager := &Manager{Configuration: cfg}
			task := &Task{ID: newTaskID(), Kind: "upload", Manifest: &Manifest{PackageID: newTaskID(), Scope: "whole_application"}, Journal: &RecoveryJournal{NewDatabaseOID: 123, DatabaseSwitched: true}}
			report := &ReconciliationReport{DatabaseFingerprint: strings.Repeat("1", 64), MediaFingerprint: strings.Repeat("2", 64)}
			report.ExternalProjectionReconciliation = manager.buildTaskBoundExternalProjectionReconciliationContract(task, report)
			evidence := syntheticExternalProjectionAttestationsForReconciliationReport(report)
			require.Equal(t, evidence.ContractSHA256, manager.buildTaskBoundExternalProjectionReconciliationContract(task, report).ContractSHA256)
			switch scenario {
			case "task":
				task.ID = newTaskID()
			case "package":
				task.Manifest.PackageID = newTaskID()
			case "site":
				manager.Configuration.SiteID = "other-site"
			case "environment":
				manager.Configuration.Environment = "other-environment"
			case "database_oid":
				task.Journal.NewDatabaseOID++
			case "database_content":
				report.DatabaseFingerprint = strings.Repeat("3", 64)
			case "media":
				report.MediaFingerprint = strings.Repeat("4", 64)
			case "redis_destination":
				boundary.RedisDestinationSHA256 = strings.Repeat("5", 64)
			case "search_destination":
				boundary.SearchDestinationSHA256 = strings.Repeat("5", 64)
			case "realtime_stream":
				boundary.RealtimeStreamSHA256 = strings.Repeat("5", 64)
			case "realtime_dedupe":
				boundary.RealtimeDedupeNamespaceSHA256 = strings.Repeat("5", 64)
			case "media_delivery":
				boundary.MediaDeliveryConfigurationSHA256 = strings.Repeat("5", 64)
			}
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets, boundary))
			boundaryPath := filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json")
			switch scenario {
			case "missing_boundary":
				require.NoError(t, os.Remove(boundaryPath))
			case "invalid_boundary":
				require.NoError(t, os.WriteFile(boundaryPath, []byte(`{"protocol":1,"external_projections":{"version":99}}`), 0600))
			case "legacy_boundary":
				require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
			}
			updated := manager.buildTaskBoundExternalProjectionReconciliationContract(task, report)
			require.Error(t, validateTaskBoundExternalProjectionReconciliationEvidence(updated, evidence))
		})
	}
}

func TestPartialAndUnswitchedCandidatesCannotUseWholeApplicationReconciliationEndpoints(t *testing.T) {
	for _, scenario := range []string{"partial_scope", "backup_task", "unswitched", "missing_journal"} {
		t.Run(scenario, func(t *testing.T) {
			cfg := createIndependentRecoveryTestConfiguration(t)
			manager, err := NewManager(cfg)
			require.NoError(t, err)
			defer manager.Close()
			task, err := manager.createTask("upload", "synthetic-operator")
			require.NoError(t, err)
			task.State = "awaiting_reconciliation"
			task.Manifest = &Manifest{Scope: "whole_application"}
			task.Journal = &RecoveryJournal{NewDatabaseOID: 123, DatabaseSwitched: true}
			switch scenario {
			case "partial_scope":
				task.Manifest.Scope = "customer_service"
			case "backup_task":
				task.Kind = "backup"
			case "unswitched":
				task.Journal.DatabaseSwitched = false
			case "missing_journal":
				task.Journal = nil
			}
			require.NoError(t, manager.saveTask(task))
			_, err = manager.InspectReconciliation(task.ID)
			require.ErrorContains(t, err, "paired recovery")
			require.ErrorContains(t, manager.OpenReconciliationConnections(task.ID, "RECONCILE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID), "confirmation")
			require.ErrorContains(t, manager.CompleteReconciledRecovery(task.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID, ReconciliationEvidence{}), "paired recovery")
		})
	}
}
