package localbackup

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMediaModuleWritableTablesMatchStandaloneExportWithoutAuthorizingOtherModules(t *testing.T) {
	require.ElementsMatch(t, mediaModuleOwnedTableNames(), sortedBusinessModuleScopeTableNamesForArchitectureBoundaryChecks(selectedBusinessModuleTableSelectors([]string{"media"})))
	require.Len(t, mediaModuleOwnedTableNames(), 11)
	for _, modules := range [][]string{{"media"}, {"orders"}, {"customers"}, {"customer_service"}, {"configuration"}, {"media", "orders"}, {"configuration", "media"}, {"media", "products"}} {
		for _, table := range mediaModuleOwnedTableNames() {
			require.Equal(t, len(modules) == 1 && modules[0] == "media", mediaModuleHasDedicatedMissingRowContract(selectedBusinessModuleCompleteSnapshot{Modules: modules}, table))
		}
	}
	for _, table := range []string{"users", "products", "orders", "tickets", "ticket_messages", "settings", "outbox_events", "audit_logs"} {
		require.False(t, mediaModuleHasDedicatedMissingRowContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"media"}}, table))
	}
}

func TestMediaModuleMissingLifecycleRejectsDeletionUnfinishedUploadsAndRunnableJobs(t *testing.T) {
	for _, scenario := range []struct {
		table, row string
		allowed    bool
	}{
		{"media_assets", `{"status":"active"}`, true},
		{"media_assets", `{"status":"inactive"}`, false},
		{"media_assets", `{"status":"active","deleted_at":"2026-10-01"}`, false},
		{"media_assets", `{"status":"active","customer_service_derivative_write_token":"lease"}`, false},
		{"media_assets", `{"status":"active","customer_service_derivative_write_expires_at":"2026-01-01"}`, false},
		{"media_upload_sessions", `{"status":"bound"}`, true},
		{"media_upload_sessions", `{"status":"ready"}`, false},
		{"media_upload_sessions", `{"status":"uploading"}`, false},
		{"media_upload_sessions", `{"status":"failed"}`, false},
		{"business_draft_media_upload_receipts", `{"status":"bound"}`, true},
		{"business_draft_media_upload_receipts", `{"status":"ready"}`, false},
		{"media_derivative_rebuild_jobs", `{"status":"succeeded","finished_at":"2026-10-01"}`, true},
		{"media_derivative_rebuild_jobs", `{"status":"pending"}`, false},
		{"media_derivative_rebuild_jobs", `{"status":"running"}`, false},
		{"media_derivative_rebuild_jobs", `{"status":"succeeded"}`, false},
		{"media_derivative_rebuild_jobs", `{"status":"succeeded","finished_at":"2026-10-01","locked_by":"original-worker"}`, false},
		{"media_asset_derivatives", `{"deleted_at":"2026-10-01"}`, false},
	} {
		t.Run(scenario.table+scenario.row, func(t *testing.T) {
			require.Equal(t, scenario.allowed, inspectMediaModuleMissingRowLifecycle(scenario.table, json.RawMessage(scenario.row)) == "")
		})
	}
}

func TestMediaModuleFormalRestoreRequiresCurrentPolicyDigestAndExclusiveJournal(t *testing.T) {
	for _, scenario := range []string{"eligible", "ineligible", "old_policy", "mutated_digest", "missing_binding", "mixed_scope", "expired", "other_progress", "already_started", "work_expired"} {
		t.Run(scenario, func(t *testing.T) {
			report := &SelectedBusinessModuleIsolatedRestorePreview{Modules: []string{"media"}, Binding: &BusinessModuleRestoreSnapshotBinding{}, TargetUnchanged: true, IsolatedRestoreVerified: true, DisposableStageRemoved: true, MediaRestoreEligibility: &MediaModuleRestoreEligibility{Policy: mediaModuleFormalRestorePolicy, Eligible: true, Blockers: []string{}}}
			report.SHA256, _ = businessModuleRecoveryCanonicalJSONSHA256(report)
			task := &Task{Kind: "upload", Manifest: &Manifest{Scope: selectedBusinessModulesBackupScope, SelectedModules: []string{"media"}}, State: "module_isolated_restore_verified", ExpiresAt: time.Now().Add(time.Hour), SelectedModuleRestorePreview: report}
			switch scenario {
			case "ineligible":
				report.MediaRestoreEligibility.Eligible = false
			case "old_policy":
				report.MediaRestoreEligibility.Policy = "old"
			case "mutated_digest":
				report.RestoredRows = 10
			case "missing_binding":
				report.Binding = nil
			case "mixed_scope":
				task.Manifest.SelectedModules = append(task.Manifest.SelectedModules, "orders")
			case "expired":
				task.ExpiresAt = time.Now().Add(-time.Hour)
			case "other_progress":
				task.OrderModuleRestore = &BusinessModuleTransactionalRestoreProgress{}
			case "already_started":
				task.MediaModuleRestore = &BusinessModuleTransactionalRestoreProgress{}
			case "work_expired":
				task.WorkCopyExpired = true
			}
			err := validateMediaModuleFormalRestorePreview(task, report.SHA256)
			if scenario == "eligible" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
	task := &Task{MediaModuleRestore: &BusinessModuleTransactionalRestoreProgress{}, CustomerProfileModuleRestore: &BusinessModuleTransactionalRestoreProgress{}}
	require.Equal(t, 2, countTransactionalBusinessModuleRestoreProgressRecords(task))
	require.Empty(t, transactionalBusinessModuleRestoreContractForTask(task).ModuleName)
	require.Error(t, validateTransactionalBusinessModuleRestoreProgress(task))
	encoded, err := json.Marshal(&SelectedBusinessModuleIsolatedRestorePreview{})
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "media_restore_eligibility")
	encoded, err = json.Marshal(&Task{})
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "media_module_restore")
}
