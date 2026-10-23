package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func createCustomerServicePersistentCandidateLifecycleTestFixture(t *testing.T) (*customerServiceDisposableCandidateTestFixture, *Manager, *Task, CustomerServiceEffectiveSelectionInspectionRequest) {
	t.Helper()
	fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
	ctx := context.Background()
	cfg := fixture.Configuration
	require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
	_, err := fixture.Source.ExecContext(ctx, `INSERT INTO ticket_messages(id,ticket_id,content,attachments,attachment_reference_format,is_internal) VALUES(100,1,'private package message','[]','legacy_urls',true);`)
	require.NoError(t, err)
	_, err = fixture.Target.ExecContext(ctx, `UPDATE ticket_messages SET content='private current body' WHERE id=1; UPDATE outbox_events SET status='ignored',attempts=7 WHERE id=1`)
	require.NoError(t, err)
	addUnselectedCurrentMediaFilesAndEmptyDirectoriesForPersistentCandidateTest(t, cfg)
	manager, task, request := registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t, fixture)
	return fixture, manager, task, request
}

func addUnselectedCurrentMediaFilesAndEmptyDirectoriesForPersistentCandidateTest(t *testing.T, cfg Configuration) {
	t.Helper()
	for _, rootName := range mediaRoots {
		require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots[rootName], "unselected-current-only.txt"), []byte("synthetic retained "+rootName), 0600))
		require.NoError(t, os.MkdirAll(filepath.Join(cfg.StorageRoots[rootName], "unselected-empty-directory"), 0700))
	}
	require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads"], ".ordinary-current-file"), []byte("synthetic ordinary dot file"), 0600))
}

func verifyPreparedPersistentCandidateCompleteMediaAndCombinedCapacity(t *testing.T, manager *Manager, record *customerServicePersistentCandidateLifecycle) {
	t.Helper()
	require.NotNil(t, record.MediaUnion)
	require.Equal(t, "prepared_not_switchable", record.MediaUnion.State)
	require.Len(t, record.MediaUnion.Volumes, len(mediaRoots))
	require.NotNil(t, record.PeakCapacity)
	require.Positive(t, record.PeakCapacity.CurrentDatabaseBytes)
	require.Positive(t, record.PeakCapacity.OriginalStageDatabaseBytes)
	require.Equal(t, record.MediaUnion.TotalBytes, record.PeakCapacity.CandidateMediaBytes)
	expectedComponents := map[string][]string{}
	for name, directory := range map[string]string{
		"workspace_dump_inventory_and_reserve": manager.Configuration.Workspace,
		"database_clone_wal_and_reserve":       manager.Configuration.PostgreSQLCapacityDirectory,
		"media_uploads":                        manager.Configuration.StorageRoots["uploads"],
		"media_uploads_private":                manager.Configuration.StorageRoots["uploads_private"],
		"media_site_logo_uploads":              manager.Configuration.StorageRoots["site_logo_uploads"],
	} {
		identity, err := customerServiceCandidateCapacityFilesystemIdentity(directory)
		require.NoError(t, err)
		expectedComponents[identity] = append(expectedComponents[identity], name)
	}
	require.Len(t, record.PeakCapacity.Requirements, len(expectedComponents))
	for _, requirement := range record.PeakCapacity.Requirements {
		require.ElementsMatch(t, expectedComponents[requirement.FilesystemIdentity], requirement.Components)
		require.GreaterOrEqual(t, requirement.AdditionalBytes, int64(len(requirement.Components))*(64<<20))
		require.GreaterOrEqual(t, requirement.AvailableBytes, uint64(requirement.AdditionalBytes))
	}
	workspace, err := manager.customerServicePersistentCandidateConstructionWorkspace(record)
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(workspace.Directory, customerServiceCandidateMediaInventoryFilename))
	require.NoError(t, err)
	var inventory customerServiceCandidateCompleteMediaUnionInventory
	require.NoError(t, decodeStrictJSON(data, &inventory))
	require.Equal(t, record.Validation.SelectionSHA256, inventory.SelectionSHA256)
	require.Equal(t, record.MediaUnion.InventorySHA256, inventory.SHA256)
	require.NoError(t, manager.verifyCustomerServiceCompleteCandidateMediaUnion(context.Background(), record, &inventory))
	for _, volume := range record.MediaUnion.Volumes {
		require.Equal(t, "tree_prepared", volume.State)
		directory, err := customerServiceCandidateMediaVolumeDirectory(record, volume)
		require.NoError(t, err)
		content, err := os.ReadFile(filepath.Join(directory, "tree", "unselected-current-only.txt"))
		require.NoError(t, err)
		require.Equal(t, "synthetic retained "+volume.RootName, string(content))
		info, err := os.Stat(filepath.Join(directory, "tree", "unselected-empty-directory"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	}
	t.Logf("complete media union: %d current files, %d selected additions, %d bytes; five capacity components combined into %d filesystems", record.MediaUnion.CurrentFiles, record.MediaUnion.AddedPackageFiles, record.MediaUnion.TotalBytes, len(record.PeakCapacity.Requirements))
}

func registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t *testing.T, fixture *customerServiceDisposableCandidateTestFixture) (*Manager, *Task, CustomerServiceEffectiveSelectionInspectionRequest) {
	t.Helper()
	ctx := context.Background()
	cfg := fixture.Configuration
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { manager.Close() })
	directory, _ := manager.taskDirectory(fixture.TaskID)
	require.NoError(t, os.Mkdir(directory, 0700))
	for _, identity := range fixture.Manifest.Files {
		filename := filepath.Join(directory, "verified", filepath.FromSlash(identity.Path))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		source, err := os.Open(filepath.Join(fixture.VerifiedDirectory, filepath.FromSlash(identity.Path)))
		require.NoError(t, err)
		destination, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		require.NoError(t, err)
		_, err = io.Copy(destination, source)
		require.NoError(t, err)
		require.NoError(t, source.Close())
		require.NoError(t, destination.Close())
	}
	fixture.VerifiedDirectory = filepath.Join(directory, "verified")
	_, fingerprint, _, err := captureDatabaseInventory(ctx, fixture.Source)
	require.NoError(t, err)
	fixture.Manifest.DatabaseFingerprint = fingerprint
	targetFingerprint, tables, err := cfg.targetFingerprint(ctx)
	require.NoError(t, err)
	task := &Task{ID: fixture.TaskID, Kind: "upload", Actor: "synthetic-operator", State: "verified", Phase: "original_whole_storefront_preview", CreatedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(time.Hour), Manifest: fixture.Manifest, StageDatabase: fixture.SourceName, StageFingerprint: fingerprint, PreviewTargetFingerprint: targetFingerprint, PreviewTargetTables: tables}
	require.NoError(t, manager.saveTask(task))
	report, err := manager.InspectVerifiedPackageModuleDependencies(ctx, task.ID, "customer_service")
	require.NoError(t, err)
	plan := report.ConversationRows.ConversationRecoveryPlan
	reviewRequest := CustomerServiceRecoveryConflictReviewRequest{PlanSHA256: plan.SHA256, Confirmation: "REVIEW " + cfg.SiteID + "/" + cfg.Environment + "/" + fixture.Manifest.PackageID}
	for _, row := range plan.RowActions {
		if customerServiceRecoveryRowRequiresContentDecision(row) {
			reviewRequest.Decisions = append(reviewRequest.Decisions, CustomerServiceRecoveryContentDecision{Table: row.Table, IdentitySHA256: row.IdentitySHA256, Choice: "use_package_row", Reason: "synthetic persistent candidate verification"})
		}
	}
	request := CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: plan.SHA256}
	if len(reviewRequest.Decisions) != 0 {
		review, err := manager.SaveCustomerServiceRecoveryConflictReview(ctx, task.ID, "synthetic-operator", reviewRequest)
		require.NoError(t, err)
		request.ReviewSHA256 = review.SHA256
	}
	return manager, task, request
}

func verifyCustomerServicePersistentCandidateWithCompleteMigrationFixture(t *testing.T, fixture *customerServiceDisposableCandidateTestFixture) {
	t.Helper()
	t.Run("complete_migration_stopped_persistent_candidate_lifecycle", func(t *testing.T) {
		// A separate database, workspace and gate keep this lifecycle independent
		// of the outer read-only observation transaction. Paired switching must
		// terminate every connection to its own target when sealing that database.
		ctx := context.Background()
		cfg := createIndependentRecoveryTestConfiguration(t)
		cfg.DatabaseURL = fixture.Configuration.DatabaseURL
		cfg.DatabaseName = fixture.Configuration.DatabaseName
		cfg.ApplicationDatabaseRole = fixture.Configuration.ApplicationDatabaseRole
		cfg.ExpectedMigrationVersion = fixture.Configuration.ExpectedMigrationVersion
		if capacityDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_POSTGRESQL_CAPACITY_DIRECTORY"); capacityDirectory != "" {
			cfg.PostgreSQLCapacityDirectory = capacityDirectory
		}
		independentTargetName := "local_backup_stage_" + newTaskID()[:16]
		independentTargetDumpDirectory := t.TempDir()
		independentTargetManifest := *fixture.Manifest
		require.NoError(t, cfg.captureSnapshotAndDump(ctx, independentTargetDumpDirectory, &independentTargetManifest))
		t.Cleanup(func() {
			_, cleanupError := fixture.Control.ExecContext(context.Background(), `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(independentTargetName)+` WITH (FORCE)`)
			require.NoError(t, cleanupError)
		})
		require.NoError(t, cfg.restoreIsolatedDatabase(ctx, independentTargetDumpDirectory, independentTargetName, independentTargetManifest))
		cfg.DatabaseName = independentTargetName
		independentTargetDatabase, err := cfg.connectDatabase(ctx, independentTargetName)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, independentTargetDatabase.Close()) })
		for _, identity := range fixture.Manifest.Files {
			if !strings.HasPrefix(identity.Path, "media/") {
				continue
			}
			parts := strings.SplitN(strings.TrimPrefix(identity.Path, "media/"), "/", 2)
			require.Len(t, parts, 2)
			require.Contains(t, cfg.StorageRoots, parts[0])
			content, err := os.ReadFile(filepath.Join(fixture.VerifiedDirectory, filepath.FromSlash(identity.Path)))
			require.NoError(t, err)
			filename := filepath.Join(cfg.StorageRoots[parts[0]], filepath.FromSlash(parts[1]))
			require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
			require.NoError(t, os.WriteFile(filename, content, 0600))
		}
		addUnselectedCurrentMediaFilesAndEmptyDirectoriesForPersistentCandidateTest(t, cfg)
		require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
		persistentFixture := *fixture
		persistentFixture.Configuration = cfg
		persistentFixture.Target = independentTargetDatabase
		manager, task, request := registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t, &persistentFixture)
		record, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(context.Background(), task.ID, request)
		require.NoError(t, err)
		require.Equal(t, "sealed_not_switchable", record.State)
		require.Equal(t, len(registeredStorefrontModuleRecoveryTableNames()), record.Validation.VerifiedRetainedTables)
		require.Equal(t, 208, record.Validation.VerifiedSequenceStates)
		require.Equal(t, 2, record.Validation.SelectedFullRows)
		verifyPreparedPersistentCandidateCompleteMediaAndCombinedCapacity(t, manager, record)
		verifyCustomerServiceIndependentPairedSwitchRehearsalAndRollbackWithCompleteMigrationCandidate(t, manager, task, record)
		require.NoError(t, manager.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
		record, err = manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
		require.NoError(t, err)
		require.True(t, record.Validation.CandidateDeleted)
		require.Equal(t, "discarded_not_switchable", record.MediaUnion.State)
		require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
		if evidenceDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); evidenceDirectory != "" {
			require.NoError(t, os.MkdirAll(evidenceDirectory, 0700))
			require.NoError(t, writeDurableJSON(filepath.Join(evidenceDirectory, "customer-service-complete-migration-persistent-candidate-discarded-lifecycle.json"), record))
		}
		t.Logf("complete migration stopped persistent candidate: %d complete rows, %d preserved tables, %d sequences; sealed then safely discarded, original stage retained", record.Validation.SelectedFullRows, record.Validation.VerifiedRetainedTables, record.Validation.VerifiedSequenceStates)
	})
}

func TestPostgreSQLCustomerServicePersistentCandidateSealsFullRowValidationBlocksOtherFlowsAndDiscardsOnRestart(t *testing.T) {
	fixture, manager, originalTask, request := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
	ctx := context.Background()
	_, originalTargetFingerprint, _, err := captureDatabaseInventory(ctx, fixture.Target)
	require.NoError(t, err)
	originalMediaFingerprint, err := fixture.Configuration.mediaFingerprint(ctx)
	require.NoError(t, err)
	record, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, originalTask.ID, request)
	require.NoError(t, err)
	require.Equal(t, "sealed_not_switchable", record.State)
	require.False(t, record.ExecutionAvailable)
	require.False(t, record.Validation.ExecutionAvailable)
	require.False(t, record.Validation.CandidateDeleted)
	require.Equal(t, 2, record.Validation.SelectedFullRows)
	require.NotEmpty(t, record.Validation.DependencyBlockers)
	require.NotNil(t, record.Binding)
	verifyPreparedPersistentCandidateCompleteMediaAndCombinedCapacity(t, manager, record)
	preparedRecord := *record
	var oid int64
	var connectionsAllowed bool
	require.NoError(t, fixture.Control.QueryRowContext(ctx, `SELECT oid::bigint,datallowconn FROM pg_database WHERE datname=$1`, record.Identity.DatabaseName).Scan(&oid, &connectionsAllowed))
	require.Equal(t, record.Identity.DatabaseOID, oid)
	require.False(t, connectionsAllowed)
	_, err = fixture.Configuration.connectDatabase(ctx, record.Identity.DatabaseName)
	require.Error(t, err)
	fenced, err := maintenancegate.IsApplicationFenced(fixture.Configuration.MaintenanceDirectory)
	require.NoError(t, err)
	require.True(t, fenced)
	storedTask, err := manager.GetTask(originalTask.ID)
	require.NoError(t, err)
	require.Equal(t, originalTask, storedTask, "candidate must not replace the stage or whole-storefront preview/task state")
	require.ErrorContains(t, manager.ReleaseUnmodifiedMaintenance(originalTask.ID), "safely discarded")
	_, err = manager.ApplyVerifiedRecovery(originalTask.ID, "RESTORE "+fixture.Configuration.SiteID+"/"+fixture.Configuration.Environment+"/"+fixture.Manifest.PackageID)
	require.ErrorContains(t, err, "safely discarded")
	_, err = manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, originalTask.ID, request)
	require.ErrorContains(t, err, "safely discarded")
	workspace, err := manager.customerServicePersistentCandidateConstructionWorkspace(record)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(workspace.Directory, "current-target-snapshot.dump"))
	require.True(t, os.IsNotExist(err), "successful sealing removes the dump")
	require.NoError(t, manager.Close())
	restarted, err := NewManager(fixture.Configuration)
	require.NoError(t, err)
	defer restarted.Close()
	record, err = restarted.loadCustomerServicePersistentCandidateLifecycle(originalTask.ID)
	require.NoError(t, err)
	require.Equal(t, "discarded_not_switchable", record.State)
	require.True(t, record.Validation.CandidateDeleted)
	require.Equal(t, "discarded_not_switchable", record.MediaUnion.State)
	for _, volume := range preparedRecord.MediaUnion.Volumes {
		directory, err := customerServiceCandidateMediaVolumeDirectory(&preparedRecord, volume)
		require.NoError(t, err)
		_, err = os.Stat(directory)
		require.True(t, os.IsNotExist(err))
	}
	require.NoError(t, fixture.Control.QueryRowContext(ctx, `SELECT count(*) FROM pg_database WHERE oid=$1`, oid).Scan(&oid))
	require.Zero(t, oid)
	_, err = os.Stat(workspace.Directory)
	require.True(t, os.IsNotExist(err))
	fenced, err = maintenancegate.IsApplicationFenced(fixture.Configuration.MaintenanceDirectory)
	require.NoError(t, err)
	require.True(t, fenced, "restart/discard must never release maintenance")
	require.NoError(t, restarted.ReleaseUnmodifiedMaintenance(originalTask.ID))
	_, actualTargetFingerprint, _, err := captureDatabaseInventory(ctx, fixture.Target)
	require.NoError(t, err)
	require.Equal(t, originalTargetFingerprint, actualTargetFingerprint)
	actualMediaFingerprint, err := fixture.Configuration.mediaFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, originalMediaFingerprint, actualMediaFingerprint)
	storedTask, err = restarted.GetTask(originalTask.ID)
	require.NoError(t, err)
	require.Equal(t, originalTask, storedTask)
	// A new attempt must archive the safely discarded predecessor and allocate a
	// new identity; it cannot silently reuse the earlier database or validation.
	nextAttempt, err := restarted.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, originalTask.ID, request)
	require.NoError(t, err)
	require.NotEqual(t, record.AttemptID, nextAttempt.AttemptID)
	require.NotEqual(t, record.Identity.DatabaseOID, nextAttempt.Identity.DatabaseOID)
	taskDirectory, _ := restarted.taskDirectory(originalTask.ID)
	_, err = os.Stat(filepath.Join(taskDirectory, "customer-service-discarded-candidate-attempt-"+record.AttemptID+".json"))
	require.NoError(t, err)
	require.NoError(t, restarted.discardCustomerServicePersistentCandidate(ctx, originalTask.ID))
	require.NoError(t, restarted.discardCustomerServicePersistentCandidate(ctx, originalTask.ID))
	require.NoError(t, restarted.ReleaseUnmodifiedMaintenance(originalTask.ID))
	encoded, err := json.Marshal(record)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "private ")
	t.Logf("persistent candidate: %d full rows, %d retained tables, %d sequences; sealed, blocked release/whole-apply, restart discarded only candidate", record.Validation.SelectedFullRows, record.Validation.VerifiedRetainedTables, record.Validation.VerifiedSequenceStates)
}

func TestPostgreSQLCustomerServicePersistentCandidateRevalidatesSavedBindingsAfterStoppingAndRetainsFailedIntent(t *testing.T) {
	for _, scenario := range []string{"target_content_changed", "source_content_changed", "target_sequence_changed", "target_schema_changed", "target_media_changed", "review_digest_changed", "unfenced_external_session", "cancelled_prepare", "expired_work_copy"} {
		t.Run(scenario, func(t *testing.T) {
			fixture, manager, task, request := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
			ctx := context.Background()
			switch scenario {
			case "target_content_changed":
				_, err := fixture.Target.ExecContext(ctx, `UPDATE ticket_messages SET content='private concurrent changed body' WHERE id=1`)
				require.NoError(t, err)
			case "source_content_changed":
				_, err := fixture.Source.ExecContext(ctx, `UPDATE ticket_messages SET content='private changed source body' WHERE id=1`)
				require.NoError(t, err)
			case "target_sequence_changed":
				_, err := fixture.Target.ExecContext(ctx, `SELECT nextval('ticket_messages_id_seq')`)
				require.NoError(t, err)
			case "target_schema_changed":
				_, err := fixture.Target.ExecContext(ctx, `ALTER TABLE ticket_messages ADD COLUMN synthetic_concurrent_change text`)
				require.NoError(t, err)
			case "target_media_changed":
				require.NoError(t, os.WriteFile(filepath.Join(fixture.Configuration.StorageRoots["uploads"], "synthetic-concurrent-upload.txt"), []byte("synthetic change"), 0600))
			case "review_digest_changed":
				request.ReviewSHA256 = strings.Repeat("0", 64)
			case "unfenced_external_session":
				connectionURL, err := fixture.Configuration.databaseConnectionURL(fixture.Configuration.DatabaseName)
				require.NoError(t, err)
				parsed, err := url.Parse(connectionURL)
				require.NoError(t, err)
				query := parsed.Query()
				query.Set("application_name", "synthetic_unfenced_writer")
				parsed.RawQuery = query.Encode()
				database, err := sql.Open("postgres", parsed.String())
				require.NoError(t, err)
				defer database.Close()
				require.NoError(t, database.PingContext(ctx))
			case "cancelled_prepare":
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			case "expired_work_copy":
				task.ExpiresAt = time.Now().Add(-time.Hour)
				require.NoError(t, manager.saveTask(task))
			}
			_, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, task.ID, request)
			require.Error(t, err)
			var count int
			require.NoError(t, fixture.Control.QueryRow(`SELECT count(*) FROM pg_database WHERE datname LIKE 'local_backup_candidate_%'`).Scan(&count))
			require.Zero(t, count)
			if scenario == "expired_work_copy" {
				return
			}
			record, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			require.Equal(t, "discard_required", record.State)
			require.Zero(t, record.Identity.DatabaseOID)
			require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
			require.NoError(t, manager.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
			require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
			current, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Equal(t, task, current)
		})
	}
}

func TestPostgreSQLCustomerServicePersistentCandidateMediaIdentityFailureRetainsDatabaseAndMaintenanceFenceOnRestart(t *testing.T) {
	for _, scenario := range []string{"missing_media_identity", "changed_media_identity"} {
		t.Run(scenario, func(t *testing.T) {
			fixture, manager, task, request := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
			record, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(context.Background(), task.ID, request)
			require.NoError(t, err)
			directory, err := customerServiceCandidateMediaVolumeDirectory(record, record.MediaUnion.Volumes[0])
			require.NoError(t, err)
			marker := filepath.Join(directory, customerServiceCandidateMediaVolumeIdentityFilename)
			originalMarker, err := os.ReadFile(marker)
			require.NoError(t, err)
			if scenario == "missing_media_identity" {
				require.NoError(t, os.Remove(marker))
			} else {
				require.NoError(t, os.WriteFile(marker, []byte(`{"version":999}`), 0600))
			}
			require.NoError(t, manager.Close())
			restarted, err := NewManager(fixture.Configuration)
			require.NoError(t, err)
			defer restarted.Close()
			loaded, err := restarted.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			require.Equal(t, "inspection_required", loaded.State)
			require.Error(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
			var count int
			require.NoError(t, fixture.Control.QueryRow(`SELECT count(*) FROM pg_database WHERE oid=$1`, record.Identity.DatabaseOID).Scan(&count))
			require.Equal(t, 1, count, "ambiguous media must retain the candidate database as well")
			for _, volume := range loaded.MediaUnion.Volumes {
				volumeDirectory, err := customerServiceCandidateMediaVolumeDirectory(loaded, volume)
				require.NoError(t, err)
				_, err = os.Stat(volumeDirectory)
				require.NoError(t, err, "all volumes must be validated before deleting any")
			}
			fenced, err := maintenancegate.IsApplicationFenced(fixture.Configuration.MaintenanceDirectory)
			require.NoError(t, err)
			require.True(t, fenced)
			// Restore the test's known marker to exercise a deliberate retry. The
			// executor never repairs an ambiguous identity by guessing its name.
			require.NoError(t, writeDurableFile(marker, originalMarker))
			require.NoError(t, restarted.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
			require.NoError(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
		})
	}
}

func TestPostgreSQLCustomerServicePersistentCandidateRejectsPeakWorkspaceLimitAndOriginalMissingAttachmentBlockerBeforeAllocation(t *testing.T) {
	for _, scenario := range []string{"workspace_peak_capacity_exceeded", "missing_current_attachment_with_fresh_plan"} {
		t.Run(scenario, func(t *testing.T) {
			fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(fixture.Configuration.MaintenanceDirectory, fixture.Configuration.DatabaseName, fixture.Configuration.ApplicationVersion, fixture.Configuration.StorageRoots, fixture.Configuration.RecoverySecrets))
			if scenario == "missing_current_attachment_with_fresh_plan" {
				require.NoError(t, os.Remove(filepath.Join(fixture.Configuration.StorageRoots["uploads"], "legacy-attachment.txt")))
			}
			manager, task, request := registerVerifiedUploadAndSavedSelectionsForCustomerServicePersistentCandidateTest(t, fixture)
			if scenario == "workspace_peak_capacity_exceeded" {
				manager.Configuration.MaxWorkspaceBytes = 32 << 20
			}
			_, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(context.Background(), task.ID, request)
			require.Error(t, err)
			if scenario == "workspace_peak_capacity_exceeded" {
				require.ErrorContains(t, err, "workspace capacity exceeded")
			} else {
				require.ErrorContains(t, err, "block")
			}
			record, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			require.Nil(t, record.MediaUnion)
			require.Equal(t, "allocation_not_attempted", record.Identity.State)
			require.Zero(t, record.Identity.DatabaseOID)
			var count int
			require.NoError(t, fixture.Control.QueryRow(`SELECT count(*) FROM pg_database WHERE datname=$1`, record.Identity.DatabaseName).Scan(&count))
			require.Zero(t, count)
			for _, rootName := range mediaRoots {
				matches, err := filepath.Glob(filepath.Join(fixture.Configuration.StorageRoots[rootName], ".local-backup-customer-service-candidate-*"))
				require.NoError(t, err)
				require.Empty(t, matches)
			}
			require.NoError(t, manager.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
			require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
		})
	}
}

func TestPostgreSQLCustomerServiceCandidateCombinedCapacityChargesSharedFilesystemForAllFiveComponents(t *testing.T) {
	fixture := createCustomerServiceDisposableCandidateRehearsalTestFixture(t)
	ctx := context.Background()
	// Place all five capacity mappings on the workspace filesystem to exercise
	// aggregation independently of whether PGDATA is a separate Docker volume.
	cfg := fixture.Configuration
	cfg.PostgreSQLCapacityDirectory = filepath.Join(filepath.Dir(cfg.Workspace), "postgresql-capacity")
	inventory, err := cfg.inventoryCustomerServiceCompleteCurrentMedia(ctx)
	require.NoError(t, err)
	source, err := fixture.Source.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer source.Rollback()
	target, err := fixture.Target.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer target.Rollback()
	budget, err := cfg.inspectCustomerServiceCandidateCombinedPeakCapacity(ctx, source, target, inventory)
	require.NoError(t, err)
	require.Len(t, budget.Requirements, 1)
	require.Len(t, budget.Requirements[0].Components, 5)
	expectedBytes := budget.TargetDumpByteLimit + int64(inventory.EncodedBytes) + (16 << 20) + (64 << 20)
	// A bounded 32 MiB cleanup audit and its atomic replacement both coexist.
	expectedBytes += 64 << 20
	expectedBytes += budget.CurrentDatabaseBytes*3 + (256 << 20)
	expectedBytes += inventory.TotalBytes + int64(len(inventory.Files)+len(inventory.Directories))*(64<<10) + 3*(64<<20)
	require.Equal(t, expectedBytes, budget.Requirements[0].AdditionalBytes)
	budget.Requirements[0].AvailableBytes = uint64(expectedBytes - 1)
	require.ErrorContains(t, validateCustomerServiceCandidateCombinedFilesystemCapacity(budget.Requirements), "5 components")
	t.Logf("same filesystem: five additional allocations summed to %d bytes; one byte below aggregate requirement rejected", expectedBytes)
}

func createRecordedCustomerServiceCandidateDatabaseForLifecycleRecoveryTest(t *testing.T, manager *Manager, task *Task, control *sql.DB, identityState string) *customerServicePersistentCandidateLifecycle {
	t.Helper()
	attemptID := newTaskID()
	record := &customerServicePersistentCandidateLifecycle{Version: 1, State: "constructing_not_switchable", TaskID: task.ID, AttemptID: attemptID, CreatedAt: time.Now().UTC(), Request: CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: strings.Repeat("a", 64)}, Identity: customerServiceDisposableCandidateDatabaseIdentity{Version: 1, State: identityState, DatabaseName: "local_backup_candidate_" + attemptID[:16], PlanSHA256: strings.Repeat("a", 64)}}
	workspace, err := manager.customerServicePersistentCandidateConstructionWorkspace(record)
	require.NoError(t, err)
	require.NoError(t, os.Mkdir(workspace.Directory, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(workspace.Directory, "current-target-snapshot.dump"), []byte("synthetic partial dump"), 0600))
	if identityState == "allocated_not_switchable" {
		record.Identity.State = "allocation_intent"
	}
	require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
	if identityState != "allocation_not_attempted" {
		require.NoError(t, ensureCustomerServiceDisposableCandidateOwnerAndApplicationIsolation(context.Background(), control, manager.Configuration.ApplicationDatabaseRole))
		_, err := control.Exec(`CREATE DATABASE ` + pq.QuoteIdentifier(record.Identity.DatabaseName) + ` WITH OWNER ` + pq.QuoteIdentifier(customerServiceDisposableCandidateOwner) + ` ALLOW_CONNECTIONS false`)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = control.Exec(`DROP DATABASE IF EXISTS ` + pq.QuoteIdentifier(record.Identity.DatabaseName) + ` WITH (FORCE)`)
		})
		if identityState == "allocated_not_switchable" {
			require.NoError(t, control.QueryRow(`SELECT oid::bigint FROM pg_database WHERE datname=$1`, record.Identity.DatabaseName).Scan(&record.Identity.DatabaseOID))
			record.Identity.State = "allocated_not_switchable"
		}
	}
	require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
	return record
}

func TestPostgreSQLCustomerServicePersistentCandidateRestartCleanupPreservesAmbiguousOrChangedDatabaseIdentity(t *testing.T) {
	for _, scenario := range []string{"unallocated_partial_dump", "allocated_partial_restore", "allocation_response_lost", "owner_changed", "database_renamed", "same_name_new_oid", "drop_failed"} {
		t.Run(scenario, func(t *testing.T) {
			fixture, manager, task, _ := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
			state := "allocated_not_switchable"
			if scenario == "unallocated_partial_dump" {
				state = "allocation_not_attempted"
			}
			if scenario == "allocation_response_lost" {
				state = "allocation_intent"
			}
			record := createRecordedCustomerServiceCandidateDatabaseForLifecycleRecoveryTest(t, manager, task, fixture.Control, state)
			lease, err := manager.acquireMaintenance(context.Background(), task)
			require.NoError(t, err)
			require.NoError(t, lease.Close())
			switch scenario {
			case "owner_changed":
				_, err = fixture.Control.Exec(`ALTER DATABASE ` + pq.QuoteIdentifier(record.Identity.DatabaseName) + ` OWNER TO postgres`)
				require.NoError(t, err)
			case "database_renamed":
				renamed := "synthetic_renamed_" + newTaskID()[:12]
				_, err = fixture.Control.Exec(`ALTER DATABASE ` + pq.QuoteIdentifier(record.Identity.DatabaseName) + ` RENAME TO ` + pq.QuoteIdentifier(renamed))
				require.NoError(t, err)
				t.Cleanup(func() {
					_, _ = fixture.Control.Exec(`DROP DATABASE IF EXISTS ` + pq.QuoteIdentifier(renamed) + ` WITH (FORCE)`)
				})
			case "same_name_new_oid":
				_, err = fixture.Control.Exec(`DROP DATABASE ` + pq.QuoteIdentifier(record.Identity.DatabaseName))
				require.NoError(t, err)
				_, err = fixture.Control.Exec(`CREATE DATABASE ` + pq.QuoteIdentifier(record.Identity.DatabaseName) + ` WITH OWNER ` + pq.QuoteIdentifier(customerServiceDisposableCandidateOwner))
				require.NoError(t, err)
			case "drop_failed":
				_, err = fixture.Control.Exec(`ALTER DATABASE ` + pq.QuoteIdentifier(record.Identity.DatabaseName) + ` ALLOW_CONNECTIONS true`)
				require.NoError(t, err)
				database, err := fixture.Configuration.connectDatabase(context.Background(), record.Identity.DatabaseName)
				require.NoError(t, err)
				_, err = database.Exec(`CREATE TABLE synthetic_disposal_blocker(id integer); BEGIN; INSERT INTO synthetic_disposal_blocker VALUES(1); PREPARE TRANSACTION 'synthetic_candidate_cleanup_blocker'`)
				require.NoError(t, err)
				require.NoError(t, database.Close())
				t.Cleanup(func() {
					_, _ = fixture.Control.Exec(`ALTER DATABASE ` + pq.QuoteIdentifier(record.Identity.DatabaseName) + ` ALLOW_CONNECTIONS true`)
					database, err := fixture.Configuration.connectDatabase(context.Background(), record.Identity.DatabaseName)
					if err == nil {
						_, _ = database.Exec(`ROLLBACK PREPARED 'synthetic_candidate_cleanup_blocker'`)
						database.Close()
					}
				})
			}
			require.NoError(t, manager.Close())
			restarted, err := NewManager(fixture.Configuration)
			require.NoError(t, err)
			defer restarted.Close()
			loaded, err := restarted.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			blocked := scenario == "allocation_response_lost" || scenario == "owner_changed" || scenario == "database_renamed" || scenario == "drop_failed"
			if blocked {
				require.Equal(t, "inspection_required", loaded.State)
				require.Error(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
				workspace, err := restarted.customerServicePersistentCandidateConstructionWorkspace(loaded)
				require.NoError(t, err)
				_, err = os.Stat(workspace.Directory)
				require.NoError(t, err)
			} else {
				require.Equal(t, "discarded_not_switchable", loaded.State)
			}
			fenced, err := maintenancegate.IsApplicationFenced(fixture.Configuration.MaintenanceDirectory)
			require.NoError(t, err)
			require.True(t, fenced)
			if scenario == "same_name_new_oid" {
				var newOID int64
				require.NoError(t, fixture.Control.QueryRow(`SELECT oid::bigint FROM pg_database WHERE datname=$1`, record.Identity.DatabaseName).Scan(&newOID))
				require.NotEqual(t, record.Identity.DatabaseOID, newOID)
			}
		})
	}
}

func TestPostgreSQLCustomerServiceDisposableCandidateStartupCleansRecordedOIDAndRetainsLostAllocationResponse(t *testing.T) {
	for _, identityState := range []string{"allocated_not_switchable", "allocation_intent"} {
		t.Run(identityState, func(t *testing.T) {
			fixture, manager, task, _ := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
			record := createRecordedCustomerServiceCandidateDatabaseForLifecycleRecoveryTest(t, manager, task, fixture.Control, identityState)
			directory, err := os.MkdirTemp(fixture.Configuration.Workspace, "customer-service-candidate-rehearsal-")
			require.NoError(t, err)
			require.NoError(t, writeDurableJSON(filepath.Join(directory, "customer-service-disposable-candidate-database-identity.json"), record.Identity))
			// Replace the synthetic persistent record with the older disposable
			// layout, whose identity must now be discovered at executor startup.
			taskDirectory, _ := manager.taskDirectory(task.ID)
			require.NoError(t, os.Remove(filepath.Join(taskDirectory, customerServicePersistentCandidateLifecycleFilename)))
			require.NoError(t, manager.Close())
			restarted, err := NewManager(fixture.Configuration)
			require.NoError(t, err)
			defer restarted.Close()
			var count int
			require.NoError(t, fixture.Control.QueryRow(`SELECT count(*) FROM pg_database WHERE datname=$1`, record.Identity.DatabaseName).Scan(&count))
			_, err = os.Stat(directory)
			if identityState == "allocated_not_switchable" {
				require.Zero(t, count)
				require.True(t, os.IsNotExist(err))
			} else {
				require.Equal(t, 1, count)
				require.NoError(t, err)
			}
		})
	}
}
