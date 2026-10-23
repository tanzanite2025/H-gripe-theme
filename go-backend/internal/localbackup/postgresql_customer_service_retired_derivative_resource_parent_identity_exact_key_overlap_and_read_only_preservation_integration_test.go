package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/stretchr/testify/require"
)

const customerServiceRetiredDerivativeCleanupSyntheticRegistration = `
ALTER TABLE media_asset_derivatives ADD COLUMN deleted_at timestamptz;
INSERT INTO media_assets(id,storage_key,url,visibility) VALUES(9001,'unrelated-parent.png','/uploads/unrelated-parent.png','public');
INSERT INTO media_asset_derivatives(id,media_asset_id,storage_key,url,size,deleted_at) VALUES(501,9001,'media-derivatives/9001/retired.png','/uploads/media-derivatives/9001/retired.png',8,'2026-10-01');
`

func TestPostgreSQLCustomerServiceMediaCleanupUsesResourceSpecificIdentitiesAndPreservesOriginalReceipts(t *testing.T) {
	for _, scenario := range []struct {
		name, resourceType, resourceID, aggregateID, key, bothSQL, status, code string
		blocked, incomplete                                                     bool
	}{
		{name: "retired_registration_and_parent_ids_differ"},
		{name: "retired_registration_id_can_equal_an_unrelated_retained_asset_id", resourceID: "2", bothSQL: `UPDATE media_asset_derivatives SET id=2 WHERE id=501;`},
		{name: "absent_registration_is_already_cleaned", bothSQL: `DELETE FROM media_asset_derivatives WHERE id=501;`},
		{name: "url_only_retired_registration", bothSQL: `UPDATE media_asset_derivatives SET storage_key='' WHERE id=501;`},
		{name: "ignored_receipt_stays_suppressed", status: "ignored"},
		{name: "resource_id_is_not_the_parent_aggregate", aggregateID: "501", code: "retired_derivative_cleanup_registration_identity_conflict", blocked: true},
		{name: "active_generation_is_not_retired", bothSQL: `UPDATE media_asset_derivatives SET deleted_at=NULL WHERE id=501;`, code: "retired_derivative_cleanup_registration_identity_conflict", blocked: true},
		{name: "retired_registration_key_differs", key: "media-derivatives/9001/other.png", code: "retired_derivative_cleanup_registered_object_identity_conflict", blocked: true, incomplete: true},
		{name: "retired_registration_url_and_key_disagree", bothSQL: `UPDATE media_asset_derivatives SET url='/uploads/media-derivatives/9001/other.png' WHERE id=501;`, code: "retired_derivative_cleanup_registered_object_identity_conflict", blocked: true, incomplete: true},
		{name: "retired_registration_foreign_url_unresolved", bothSQL: `UPDATE media_asset_derivatives SET url='https://foreign.invalid/retired.png' WHERE id=501;`, code: "retired_derivative_cleanup_registered_object_identity_conflict", blocked: true, incomplete: true},
		{name: "retired_registration_parent_missing", bothSQL: `ALTER TABLE media_asset_derivatives DROP CONSTRAINT media_asset_derivatives_media_asset_id_fkey; DELETE FROM media_assets WHERE id=9001;`, code: "retired_derivative_cleanup_registered_parent_missing", blocked: true, incomplete: true},
		{name: "retired_registration_state_column_missing", bothSQL: `ALTER TABLE media_asset_derivatives DROP COLUMN deleted_at;`, code: "retired_derivative_cleanup_registration_schema_missing", blocked: true, incomplete: true},
		{name: "retired_registration_resource_id_is_not_canonical", resourceID: "0501", code: "durable_cleanup_resource_identity_conflict", blocked: true},
		{name: "retired_registration_aggregate_id_is_not_canonical", aggregateID: "09001", code: "durable_cleanup_resource_identity_conflict", blocked: true},
		{name: "retired_registration_key_list_missing", key: "omit", code: "durable_derivative_cleanup_exact_key_required", blocked: true, incomplete: true},
		{name: "retired_registration_requires_one_key", key: "multiple", code: "durable_derivative_cleanup_exact_key_required", blocked: true, incomplete: true},
		{name: "retired_registration_exact_key_overlaps_retained_file", key: "legacy-attachment.txt", bothSQL: `UPDATE media_asset_derivatives SET storage_key='legacy-attachment.txt',url='/uploads/legacy-attachment.txt' WHERE id=501;`, code: "durable_cleanup_overlaps_retained_recovery_object", blocked: true},
		{name: "unregistered_derivative_parent_does_not_own_other_keys", resourceType: "unregistered_media_derivative", resourceID: "1", aggregateID: "1", key: "customer-service/attachments/derivatives/1/unregistered.png"},
		{name: "unregistered_derivative_still_requires_its_parent_identity", resourceType: "unregistered_media_derivative", resourceID: "1", aggregateID: "2", key: "customer-service/attachments/derivatives/1/unregistered.png", code: "durable_cleanup_resource_identity_conflict", blocked: true},
		{name: "unregistered_derivative_exact_key_still_overlaps", resourceType: "unregistered_media_derivative", resourceID: "1", aggregateID: "1", key: "customer-service/attachments/derivatives/synthetic.txt", code: "durable_cleanup_overlaps_retained_recovery_object", blocked: true},
		{name: "unregistered_derivative_cannot_have_an_empty_key_list", resourceType: "unregistered_media_derivative", resourceID: "1", aggregateID: "1", key: "omit", code: "durable_derivative_cleanup_exact_key_required", blocked: true, incomplete: true},
		{name: "whole_asset_cleanup_still_overlaps_by_parent_id", resourceType: "media_asset", resourceID: "1", aggregateID: "1", key: "omit", code: "durable_cleanup_overlaps_retained_recovery_object", blocked: true},
		{name: "noncanonical_whole_asset_id_cannot_hide_retained_overlap", resourceType: "media_asset", resourceID: "01", aggregateID: "1", key: "omit", code: "durable_cleanup_resource_identity_conflict", blocked: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			ctx := context.Background()
			resourceType, resourceID, aggregateID, key, status := scenario.resourceType, scenario.resourceID, scenario.aggregateID, scenario.key, scenario.status
			if resourceType == "" {
				resourceType = "retired_media_derivative"
			}
			if resourceID == "" {
				resourceID = "501"
			}
			if aggregateID == "" {
				aggregateID = "9001"
			}
			if key == "" {
				key = "media-derivatives/9001/retired.png"
			}
			if status == "" {
				status = "processed"
			}
			payload := map[string]any{"resource_type": resourceType, "resource_id": resourceID}
			if key != "omit" {
				keys := []string{key}
				if key == "multiple" {
					keys = []string{"media-derivatives/9001/retired.png", "media-derivatives/9001/other.png"}
				}
				payload["object_keys"] = keys
			}
			encodedPayload, err := json.Marshal(payload)
			require.NoError(t, err)
			for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, err := database.Exec(customerServiceRetiredDerivativeCleanupSyntheticRegistration + scenario.bothSQL)
				require.NoError(t, err)
				_, err = database.Exec(`INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(91,'private-retired-derivative-receipt','storage.object_cleanup_requested','media_asset',$1,$2::jsonb,$3)`, aggregateID, string(encodedPayload), status)
				require.NoError(t, err)
			}
			var before []string
			for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, fingerprint, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				before = append(before, fingerprint)
			}
			mediaBefore, err := fixture.Configuration.mediaFingerprint(ctx)
			require.NoError(t, err)
			source, target, plan, review, request := fixture.inspectAndBindSavedCandidateRehearsalSelections(t, nil)
			selection, err := fixture.Configuration.inspectCustomerServiceEffectiveSelectedRows(ctx, source, target, plan, review, request, fixture.Manifest, fixture.VerifiedDirectory)
			require.NoError(t, err)
			require.Equal(t, !scenario.blocked, selection.SelectionDependenciesSatisfied, "%v", selection.DependencyChecks.IssueCounts)
			require.Equal(t, !scenario.incomplete, selection.DependencyChecks.Complete)
			if scenario.code != "" {
				require.Positive(t, selection.DependencyChecks.IssueCounts[scenario.code], "%v", selection.DependencyChecks.IssueCounts)
				if resourceType == "media_asset" {
					require.Positive(t, plan.BlockingIssueCounts[scenario.code], "whole-asset ID hazards must also appear in the original inspection")
				}
			} else {
				require.Zero(t, selection.DependencyChecks.BlockingIssues, "%v", selection.DependencyChecks.IssueCounts)
				require.Zero(t, plan.BlockingIssueCounts["durable_cleanup_resource_identity_conflict"])
				require.Zero(t, plan.BlockingIssueCounts["durable_cleanup_overlaps_retained_recovery_object"])
			}
			dispositions := selection.DependencyChecks.DurableEvents.CleanupDispositions
			require.Equal(t, 1, dispositions.TargetReceipts)
			if scenario.blocked {
				require.Equal(t, 1, dispositions.ActionCounts["block_restore_requires_cleanup_reconciliation"])
			} else {
				require.NoError(t, validateCustomerServiceRecoveryPreservedCleanupDispositionEntries(dispositions.entries))
			}
			for index, database := range []*sql.DB{fixture.Source, fixture.Target} {
				_, fingerprint, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, before[index], fingerprint)
			}
			mediaAfter, err := fixture.Configuration.mediaFingerprint(ctx)
			require.NoError(t, err)
			require.Equal(t, mediaBefore, mediaAfter)
			encoded, err := json.Marshal(selection)
			require.NoError(t, err)
			for _, privateValue := range []string{"private-retired-derivative-receipt", "retired.png", "unrelated-parent.png"} {
				require.NotContains(t, string(encoded), privateValue)
			}
			if directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); directory != "" {
				require.NoError(t, os.MkdirAll(directory, 0700))
				require.NoError(t, writeDurableJSON(filepath.Join(directory, scenario.name+"-derivative-cleanup-selection.json"), selection))
			}
		})
	}
}

func TestPostgreSQLCustomerServiceRetiredDerivativeParentConflictRejectsPersistentCandidateBeforeAllocation(t *testing.T) {
	fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
	ctx := context.Background()
	for _, database := range []*sql.DB{fixture.Source, fixture.Target} {
		_, err := database.Exec(customerServiceRetiredDerivativeCleanupSyntheticRegistration + `INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(91,'private-retired-derivative-receipt','storage.object_cleanup_requested','media_asset','501','{"resource_type":"retired_media_derivative","resource_id":"501","object_keys":["media-derivatives/9001/retired.png"]}','processed');`)
		require.NoError(t, err)
	}
	cfg := fixture.Configuration
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	manager, task, request := registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t, fixture)
	before, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	_, err = manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, task.ID, request)
	require.Error(t, err)
	record, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	require.NoError(t, err)
	require.Zero(t, record.Identity.DatabaseOID)
	require.Nil(t, record.MediaUnion)
	var candidates int
	require.NoError(t, fixture.Control.QueryRow(`SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&candidates))
	require.Zero(t, candidates)
	after, err := cfg.mediaFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, before, after)
	require.NoError(t, manager.discardCustomerServicePersistentCandidate(ctx, task.ID))
	require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
}

func verifyCustomerServiceRetiredDerivativeCleanupIdentitiesAgainstActualMigrationSchema(t *testing.T, cfg Configuration, database *sql.DB, manifest *Manifest, verifiedDirectory string) {
	t.Helper()
	for _, scenario := range []struct {
		name, aggregateID, statement, code string
	}{
		{name: "distinct_registration_and_parent_ids", aggregateID: "900101"},
		{name: "forged_parent_id_equal_to_registration", aggregateID: "900102", code: "retired_derivative_cleanup_registration_identity_conflict"},
		{name: "active_registration_cannot_be_retired", aggregateID: "900101", statement: `UPDATE media_asset_derivatives SET deleted_at=NULL WHERE id=900102;`, code: "retired_derivative_cleanup_registration_identity_conflict"},
		{name: "already_removed_registration", aggregateID: "900101", statement: `DELETE FROM media_asset_derivatives WHERE id=900102;`},
		{name: "historical_url_only_registration", aggregateID: "900101", statement: `UPDATE media_asset_derivatives SET storage_key='' WHERE id=900102;`},
		{name: "mismatched_registered_url", aggregateID: "900101", statement: `UPDATE media_asset_derivatives SET url='/uploads/media-derivatives/900101/other.png' WHERE id=900102;`, code: "retired_derivative_cleanup_registered_object_identity_conflict"},
	} {
		t.Run("complete_migration_retired_derivative_"+scenario.name, func(t *testing.T) {
			ctx := context.Background()
			transaction, err := database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
			require.NoError(t, err)
			defer transaction.Rollback()
			_, err = transaction.Exec(`INSERT INTO media_assets(id,filename,storage_key,url,visibility) VALUES(900101,'private-retired-derivative-parent.png','private-retired-derivative-parent.png','/uploads/private-retired-derivative-parent.png','public');
INSERT INTO media_asset_derivatives(id,media_asset_id,preset,storage_key,url,size,deleted_at) VALUES(900102,900101,'private-retired-generation','media-derivatives/900101/private-retired.png','/uploads/media-derivatives/900101/private-retired.png',8,'2026-10-01');` + scenario.statement)
			require.NoError(t, err)
			_, err = transaction.Exec(`INSERT INTO outbox_events(id,event_key,event_type,aggregate_type,aggregate_id,payload,status) VALUES(900103,'private-complete-migration-retired-derivative-receipt','storage.object_cleanup_requested','media_asset',$1,'{"resource_type":"retired_media_derivative","resource_id":"900102","object_keys":["media-derivatives/900101/private-retired.png"]}','processed')`, scenario.aggregateID)
			require.NoError(t, err)
			observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
			require.NoError(t, err)
			var currentGraph *moduleRecoveryRowDependencyGraph
			report, err := cfg.inspectCustomerServiceConversationRows(ctx, transaction, transaction, observation, observation, manifest, verifiedDirectory, defaultModuleRecoveryRowInspectionLimits(), func(source, target *moduleRecoveryRowDependencyGraph) {
				currentGraph = target
			})
			require.NoError(t, err)
			// Run the same durable audit used by a mixed selection, using current
			// rows and the actual migration's relationship projection.
			currentGraph.SelectedCustomerServiceRows = map[string]map[string]*moduleRecoveryDependencyRow{}
			require.NoError(t, cfg.inspectCustomerServiceDurableEventDependencies(ctx, []*moduleRecoveryRowDependencyGraph{currentGraph}, report))
			if scenario.code == "" {
				require.True(t, report.Complete, "%v", report.IssueCounts)
				require.Zero(t, report.BlockingIssues, "%v", report.IssueCounts)
			} else {
				require.Positive(t, report.IssueCounts[scenario.code], "%v", report.IssueCounts)
			}
			_, after, _, err := captureDatabaseTransactionInventory(ctx, transaction)
			require.NoError(t, err)
			require.Equal(t, observation.Fingerprint, after)
			encoded, err := json.Marshal(report)
			require.NoError(t, err)
			for _, privateValue := range []string{"private-retired", "private-complete-migration-retired-derivative-receipt"} {
				require.NotContains(t, string(encoded), privateValue)
			}
			if directory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); directory != "" {
				require.NoError(t, os.MkdirAll(directory, 0700))
				require.NoError(t, writeDurableJSON(filepath.Join(directory, "complete-migration-"+scenario.name+"-derivative-cleanup-report.json"), report))
			}
		})
	}
}
