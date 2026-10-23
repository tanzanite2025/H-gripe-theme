package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func verifyCustomerServiceIndependentPairedSwitchRehearsalAndRollbackWithCompleteMigrationCandidate(t *testing.T, manager *Manager, task *Task, record *customerServicePersistentCandidateLifecycle) {
	t.Helper()
	ctx := context.Background()
	journal, err := manager.prepareCustomerServiceIndependentPairedSwitchRehearsalJournal(ctx, task.ID, record.SHA256)
	require.NoError(t, err)
	require.NoError(t, manager.executeCustomerServiceIndependentPairedSwitchRehearsal(ctx, task.ID, journal.SHA256))
	require.NoError(t, manager.rollbackCustomerServiceIndependentPairedSwitch(ctx, task.ID))
	journal, err = manager.loadCustomerServiceIndependentPairedSwitchJournal(task.ID)
	require.NoError(t, err)
	require.Equal(t, "rolled_back_isolated", journal.State)
	require.False(t, journal.ExecutionAvailable)
	currentTask, err := manager.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, task, currentTask)
	if evidenceDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); evidenceDirectory != "" {
		require.NoError(t, os.MkdirAll(evidenceDirectory, 0700))
		require.NoError(t, writeDurableJSON(filepath.Join(evidenceDirectory, "customer-service-complete-migration-independent-paired-switch-rolled-back-journal.json"), journal))
	}
	t.Logf("complete migration independent paired rehearsal: %d retained tables, %d sequences, %d bound media renames; switched under isolation and rolled back without releasing fence", record.Validation.VerifiedRetainedTables, record.Validation.VerifiedSequenceStates, len(journal.Moves))
}

func verifyCustomerServicePairedSwitchRestoredOriginalAndPreservedStage(t *testing.T, fixture *customerServiceDisposableCandidateTestFixture, manager *Manager, task *Task, record *customerServicePersistentCandidateLifecycle) {
	t.Helper()
	ctx := context.Background()
	_, fingerprint, _, err := captureDatabaseInventory(ctx, fixture.Target)
	require.NoError(t, err)
	require.Equal(t, record.Binding.TargetDatabaseFingerprint, fingerprint)
	_, sourceFingerprint, _, err := captureDatabaseInventory(ctx, fixture.Source)
	require.NoError(t, err)
	require.Equal(t, task.StageFingerprint, sourceFingerprint)
	mediaFingerprint, err := fixture.Configuration.mediaFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, record.Binding.TargetMediaFingerprint, mediaFingerprint)
	currentTask, err := manager.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, task, currentTask)
	fenced, err := maintenancegate.IsApplicationFenced(fixture.Configuration.MaintenanceDirectory)
	require.NoError(t, err)
	require.True(t, fenced)
}

func TestPostgreSQLCustomerServiceIndependentPairedSwitchRehearsalRetainsCurrentEventsAndStageAndRequiresRollbackBeforeDisposal(t *testing.T) {
	fixture, manager, task, request := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
	ctx := context.Background()
	record, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, task.ID, request)
	require.NoError(t, err)
	journal, err := manager.prepareCustomerServiceIndependentPairedSwitchRehearsalJournal(ctx, task.ID, record.SHA256)
	require.NoError(t, err)
	require.False(t, journal.ExecutionAvailable)
	require.Equal(t, "prepared_isolated_rehearsal", journal.State)
	require.Error(t, manager.discardCustomerServicePersistentCandidate(ctx, task.ID))
	require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
	require.NoError(t, manager.executeCustomerServiceIndependentPairedSwitchRehearsal(ctx, task.ID, journal.SHA256))
	journal, err = manager.loadCustomerServiceIndependentPairedSwitchJournal(task.ID)
	require.NoError(t, err)
	require.Equal(t, "switched_isolated_rollback_required", journal.State)
	var targetOID int64
	require.NoError(t, readDatabaseOID(ctx, fixture.Control, fixture.Configuration.DatabaseName, &targetOID))
	require.Equal(t, record.Identity.DatabaseOID, targetOID)
	for _, name := range []string{journal.TargetDatabase, journal.RollbackDatabase} {
		var allowed bool
		require.NoError(t, fixture.Control.QueryRowContext(context.Background(), `SELECT datallowconn FROM pg_database WHERE datname=$1`, name).Scan(&allowed))
		require.False(t, allowed)
	}
	_, err = fixture.Configuration.connectDatabase(ctx, journal.TargetDatabase)
	require.Error(t, err)
	// A deliberate privileged test connection proves the changed message and
	// the preserved current event state. It is resealed before any recovery.
	_, err = fixture.Control.ExecContext(context.Background(), `ALTER DATABASE `+pq.QuoteIdentifier(journal.TargetDatabase)+` ALLOW_CONNECTIONS true`)
	require.NoError(t, err)
	candidate, err := fixture.Configuration.connectDatabase(ctx, journal.TargetDatabase)
	require.NoError(t, err)
	var body, status string
	var attempts, messages int
	require.NoError(t, candidate.QueryRowContext(context.Background(), `SELECT content FROM ticket_messages WHERE id=1`).Scan(&body))
	require.Equal(t, "private retained synthetic message", body)
	require.NoError(t, candidate.QueryRowContext(context.Background(), `SELECT count(*) FROM ticket_messages WHERE id=100`).Scan(&messages))
	require.Equal(t, 1, messages)
	require.NoError(t, candidate.QueryRowContext(context.Background(), `SELECT status,attempts FROM outbox_events WHERE id=1`).Scan(&status, &attempts))
	require.Equal(t, "ignored", status)
	require.Equal(t, 7, attempts)
	require.NoError(t, candidate.Close())
	require.NoError(t, sealDatabase(ctx, fixture.Control, journal.TargetDatabase))
	for _, rootName := range mediaRoots {
		content, err := os.ReadFile(filepath.Join(fixture.Configuration.StorageRoots[rootName], "unselected-current-only.txt"))
		require.NoError(t, err)
		require.Equal(t, "synthetic retained "+rootName, string(content))
	}
	require.Error(t, manager.discardCustomerServicePersistentCandidate(ctx, task.ID))
	require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
	_, err = manager.ApplyVerifiedRecovery(task.ID, "RESTORE "+fixture.Configuration.SiteID+"/"+fixture.Configuration.Environment+"/"+task.Manifest.PackageID)
	require.Error(t, err)
	if evidenceDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_VALIDATION_DIRECTORY"); evidenceDirectory != "" {
		require.NoError(t, os.MkdirAll(evidenceDirectory, 0700))
		require.NoError(t, writeDurableJSON(filepath.Join(evidenceDirectory, "customer-service-independent-paired-switch-switched-isolated-journal.json"), journal))
	}
	require.NoError(t, manager.rollbackCustomerServiceIndependentPairedSwitch(ctx, task.ID))
	require.NoError(t, manager.rollbackCustomerServiceIndependentPairedSwitch(ctx, task.ID))
	verifyCustomerServicePairedSwitchRestoredOriginalAndPreservedStage(t, fixture, manager, task, record)
	require.NoError(t, manager.discardCustomerServicePersistentCandidate(ctx, task.ID))
	require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
}

// Create real database/media rename boundaries, deliberately omitting the next
// acknowledgement where named. Production execution is exercised separately.
func createCustomerServiceIndependentPairedSwitchInterruptionCheckpoint(t *testing.T, manager *Manager, control *sql.DB, record *customerServicePersistentCandidateLifecycle, journal *customerServiceIndependentPairedSwitchJournal, checkpoint string) {
	t.Helper()
	ctx := context.Background()
	if checkpoint == "journal_only" {
		return
	}
	journal.State, journal.DatabasePhase = "switching_isolated_rehearsal", "retain_original_database_intent"
	require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
	for _, volume := range record.MediaUnion.Volumes {
		directory, _ := customerServiceCandidateMediaVolumeDirectory(record, volume)
		require.NoError(t, os.Mkdir(filepath.Join(directory, "previous"), 0700))
		require.NoError(t, syncDirectory(directory))
	}
	require.NoError(t, sealDatabase(ctx, control, journal.TargetDatabase))
	if checkpoint == "old_database_sealed_before_rename" {
		return
	}
	_, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.TargetDatabase)+` RENAME TO `+pq.QuoteIdentifier(journal.RollbackDatabase))
	require.NoError(t, err)
	if checkpoint == "old_database_rename_without_ack" {
		return
	}
	journal.DatabasePhase = "original_database_retained"
	for index := range journal.Moves {
		move := &journal.Moves[index]
		move.Started = true
		require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
		if checkpoint == "first_current_move_intent_before_rename" {
			return
		}
		require.NoError(t, manager.performCustomerServicePairedSwitchBoundMediaRename(ctx, record, *move, false))
		if checkpoint == "current_move_rename_without_ack" || (checkpoint == "candidate_move_rename_without_ack" && move.Direction == "install_candidate") {
			return
		}
		move.Completed = true
		require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
	}
	journal.DatabasePhase = "install_candidate_database_intent"
	require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
	_, err = control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.CandidateDatabase)+` RENAME TO `+pq.QuoteIdentifier(journal.TargetDatabase))
	require.NoError(t, err)
	if checkpoint == "candidate_database_rename_without_ack" {
		return
	}
	journal.State, journal.DatabasePhase = "rolling_back_isolated_rehearsal", "return_candidate_database_intent"
	require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
	_, err = control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.TargetDatabase)+` RENAME TO `+pq.QuoteIdentifier(journal.CandidateDatabase))
	require.NoError(t, err)
	for index := len(journal.Moves) - 1; index >= 0; index-- {
		move := &journal.Moves[index]
		move.RollbackStarted = true
		require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
		require.NoError(t, manager.performCustomerServicePairedSwitchBoundMediaRename(ctx, record, *move, true))
		if checkpoint == "reverse_move_rename_without_ack" {
			return
		}
		move.RolledBack = true
		require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
	}
	journal.DatabasePhase = "restore_original_database_intent"
	require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
	_, err = control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.RollbackDatabase)+` RENAME TO `+pq.QuoteIdentifier(journal.TargetDatabase))
	require.NoError(t, err)
}

func TestPostgreSQLCustomerServiceIndependentPairedSwitchRestartRollsBackEachDurableRenameBoundaryBeforeCandidateDisposal(t *testing.T) {
	for _, checkpoint := range []string{"journal_only", "old_database_sealed_before_rename", "old_database_rename_without_ack", "first_current_move_intent_before_rename", "current_move_rename_without_ack", "candidate_move_rename_without_ack", "candidate_database_rename_without_ack", "reverse_move_rename_without_ack", "original_database_restored_without_ack", "fully_switched"} {
		t.Run(checkpoint, func(t *testing.T) {
			fixture, manager, task, request := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
			record, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(context.Background(), task.ID, request)
			require.NoError(t, err)
			journal, err := manager.prepareCustomerServiceIndependentPairedSwitchRehearsalJournal(context.Background(), task.ID, record.SHA256)
			require.NoError(t, err)
			if checkpoint == "fully_switched" {
				require.NoError(t, manager.executeCustomerServiceIndependentPairedSwitchRehearsal(context.Background(), task.ID, journal.SHA256))
			} else {
				createCustomerServiceIndependentPairedSwitchInterruptionCheckpoint(t, manager, fixture.Control, record, journal, checkpoint)
			}
			require.NoError(t, manager.Close())
			restarted, err := NewManager(fixture.Configuration)
			require.NoError(t, err)
			defer restarted.Close()
			journal, err = restarted.loadCustomerServiceIndependentPairedSwitchJournal(task.ID)
			require.NoError(t, err)
			require.Equal(t, "rolled_back_isolated", journal.State)
			loaded, err := restarted.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			require.Equal(t, "discarded_not_switchable", loaded.State)
			var candidateCount int
			require.NoError(t, fixture.Control.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_database WHERE oid=$1`, record.Identity.DatabaseOID).Scan(&candidateCount))
			require.Zero(t, candidateCount)
			for _, volume := range record.MediaUnion.Volumes {
				directory, _ := customerServiceCandidateMediaVolumeDirectory(record, volume)
				_, err := os.Stat(directory)
				require.True(t, os.IsNotExist(err))
			}
			verifyCustomerServicePairedSwitchRestoredOriginalAndPreservedStage(t, fixture, restarted, task, record)
			require.NoError(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
		})
	}
}

func TestPostgreSQLCustomerServiceIndependentPairedSwitchIdentityFailureRetainsCandidatesAndRejectsForgedRollbackCompletion(t *testing.T) {
	for _, scenario := range []string{"media_marker_changed", "maintenance_fence_missing", "extra_current_media", "candidate_content_changed", "candidate_connections_enabled", "rollback_name_occupied", "original_owner_changed", "forged_rolled_back_flag"} {
		t.Run(scenario, func(t *testing.T) {
			fixture, manager, task, request := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
			ctx := context.Background()
			record, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, task.ID, request)
			require.NoError(t, err)
			journal, err := manager.prepareCustomerServiceIndependentPairedSwitchRehearsalJournal(ctx, task.ID, record.SHA256)
			require.NoError(t, err)
			repairKnownTestChange := func() {}
			switch scenario {
			case "media_marker_changed", "maintenance_fence_missing":
				filename := filepath.Join(fixture.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename)
				if scenario == "media_marker_changed" {
					directory, _ := customerServiceCandidateMediaVolumeDirectory(record, record.MediaUnion.Volumes[0])
					filename = filepath.Join(directory, customerServiceCandidateMediaVolumeIdentityFilename)
				}
				original, err := os.ReadFile(filename)
				require.NoError(t, err)
				if scenario == "media_marker_changed" {
					require.NoError(t, os.WriteFile(filename, []byte(`{"version":999}`), 0600))
				} else {
					require.NoError(t, os.Remove(filename))
				}
				repairKnownTestChange = func() { require.NoError(t, writeDurableFile(filename, original)) }
			case "extra_current_media":
				filename := filepath.Join(fixture.Configuration.StorageRoots["uploads"], "synthetic-unexpected-file.txt")
				require.NoError(t, os.WriteFile(filename, []byte("unexpected"), 0600))
				repairKnownTestChange = func() { require.NoError(t, os.Remove(filename)) }
			case "candidate_content_changed":
				changeCandidateContent := func(body string) {
					_, err := fixture.Control.ExecContext(context.Background(), `ALTER DATABASE `+pq.QuoteIdentifier(journal.CandidateDatabase)+` ALLOW_CONNECTIONS true`)
					require.NoError(t, err)
					database, err := fixture.Configuration.connectDatabase(ctx, journal.CandidateDatabase)
					require.NoError(t, err)
					_, err = database.ExecContext(context.Background(), `UPDATE ticket_messages SET content=$1 WHERE id=1`, body)
					require.NoError(t, err)
					require.NoError(t, database.Close())
					require.NoError(t, sealDatabase(ctx, fixture.Control, journal.CandidateDatabase))
				}
				changeCandidateContent("synthetic changed candidate")
				repairKnownTestChange = func() { changeCandidateContent("private retained synthetic message") }
			case "candidate_connections_enabled":
				_, err := fixture.Control.ExecContext(context.Background(), `ALTER DATABASE `+pq.QuoteIdentifier(journal.CandidateDatabase)+` ALLOW_CONNECTIONS true`)
				require.NoError(t, err)
				repairKnownTestChange = func() { require.NoError(t, sealDatabase(ctx, fixture.Control, journal.CandidateDatabase)) }
			case "rollback_name_occupied":
				_, err := fixture.Control.ExecContext(context.Background(), `CREATE DATABASE `+pq.QuoteIdentifier(journal.RollbackDatabase))
				require.NoError(t, err)
				repairKnownTestChange = func() {
					_, err := fixture.Control.ExecContext(context.Background(), `DROP DATABASE `+pq.QuoteIdentifier(journal.RollbackDatabase))
					require.NoError(t, err)
				}
			case "original_owner_changed":
				_, err := fixture.Control.ExecContext(context.Background(), `ALTER DATABASE `+pq.QuoteIdentifier(journal.TargetDatabase)+` OWNER TO `+pq.QuoteIdentifier(fixture.Configuration.ApplicationDatabaseRole))
				require.NoError(t, err)
				repairKnownTestChange = func() {
					_, err := fixture.Control.ExecContext(context.Background(), `ALTER DATABASE `+pq.QuoteIdentifier(journal.TargetDatabase)+` OWNER TO `+pq.QuoteIdentifier(journal.OriginalDatabaseOwner))
					require.NoError(t, err)
				}
			case "forged_rolled_back_flag":
				require.NoError(t, manager.executeCustomerServiceIndependentPairedSwitchRehearsal(ctx, task.ID, journal.SHA256))
				journal, err = manager.loadCustomerServiceIndependentPairedSwitchJournal(task.ID)
				require.NoError(t, err)
				original, err := json.Marshal(journal)
				require.NoError(t, err)
				journal.State, journal.DatabasePhase = "rolled_back_isolated", "original_database_and_media_verified"
				for index := range journal.Moves {
					journal.Moves[index].RollbackStarted, journal.Moves[index].RolledBack = true, true
				}
				require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
				repairKnownTestChange = func() {
					require.NoError(t, json.Unmarshal(original, journal))
					require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
				}
			}
			if scenario != "forged_rolled_back_flag" {
				require.Error(t, manager.executeCustomerServiceIndependentPairedSwitchRehearsal(ctx, task.ID, journal.SHA256))
			}
			require.NoError(t, manager.Close())
			restarted, err := NewManager(fixture.Configuration)
			require.NoError(t, err)
			defer restarted.Close()
			require.Error(t, restarted.discardCustomerServicePersistentCandidate(ctx, task.ID))
			require.Error(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
			var candidateCount int
			require.NoError(t, fixture.Control.QueryRowContext(context.Background(), `SELECT count(*) FROM pg_database WHERE oid=$1`, record.Identity.DatabaseOID).Scan(&candidateCount))
			require.Equal(t, 1, candidateCount)
			for _, volume := range record.MediaUnion.Volumes {
				directory, _ := customerServiceCandidateMediaVolumeDirectory(record, volume)
				_, err := os.Stat(directory)
				require.NoError(t, err)
			}
			if scenario == "maintenance_fence_missing" {
				_, err := os.Stat(filepath.Join(fixture.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename))
				require.True(t, os.IsNotExist(err), "recovery must not invent a replacement fence")
			}
			repairKnownTestChange()
			require.NoError(t, restarted.rollbackCustomerServiceIndependentPairedSwitch(ctx, task.ID))
			verifyCustomerServicePairedSwitchRestoredOriginalAndPreservedStage(t, fixture, restarted, task, record)
			require.NoError(t, restarted.discardCustomerServicePersistentCandidate(ctx, task.ID))
			require.NoError(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
		})
	}
}
