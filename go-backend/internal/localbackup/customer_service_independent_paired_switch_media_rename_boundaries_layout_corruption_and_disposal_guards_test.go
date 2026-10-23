package localbackup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createCustomerServicePairedSwitchMediaAndJournalWithoutDatabaseTestFixture(t *testing.T) (*Manager, *Task, *customerServicePersistentCandidateLifecycle, *customerServiceIndependentPairedSwitchJournal) {
	t.Helper()
	manager, task, record, selection, manifest, verifiedDirectory := createCustomerServiceCompleteMediaUnionPlanTestFixture(t)
	inventory, _, err := manager.Configuration.planCustomerServiceCompleteCurrentMediaUnion(context.Background(), selection, manifest, verifiedDirectory)
	require.NoError(t, err)
	attachCustomerServicePlannedMediaUnionLifecycleForConstructionTest(t, manager, record, inventory)
	require.NoError(t, manager.constructCustomerServiceCompleteMediaUnionOnOwnedVolumes(context.Background(), record, inventory, verifiedDirectory))
	baseline, err := manager.Configuration.inventoryCustomerServiceCompleteCurrentMedia(context.Background())
	require.NoError(t, err)
	moves, err := buildCustomerServicePairedSwitchExpectedMediaMoves(baseline, inventory)
	require.NoError(t, err)
	journal := &customerServiceIndependentPairedSwitchJournal{Version: 1, State: "prepared_isolated_rehearsal", DatabasePhase: "not_started", TaskID: task.ID, AttemptID: record.AttemptID, CreatedAt: time.Now().UTC(), TargetDatabase: manager.Configuration.DatabaseName, CandidateDatabase: record.Identity.DatabaseName, RollbackDatabase: "local_backup_cs_old_" + record.AttemptID[:16], OriginalDatabaseOID: 101, CandidateDatabaseOID: 102, OriginalDatabaseOwner: "synthetic_owner", ConfigurationSHA256: strings.Repeat("a", 64), OriginalTaskSHA256: strings.Repeat("a", 64), CandidateLifecycleSHA256: record.SHA256, InventorySHA256: inventory.SHA256, PlanSHA256: inventory.PlanSHA256, SelectionSHA256: inventory.SelectionSHA256, BaselineDatabaseFingerprint: strings.Repeat("b", 64), CandidateDatabaseFingerprint: strings.Repeat("c", 64), BaselineDirectories: baseline.Directories, Moves: moves}
	_, err = reconstructCustomerServicePairedSwitchBaselineAndValidateMoves(journal, inventory)
	require.NoError(t, err)
	for _, volume := range record.MediaUnion.Volumes {
		directory, err := customerServiceCandidateMediaVolumeDirectory(record, volume)
		require.NoError(t, err)
		require.NoError(t, os.Mkdir(filepath.Join(directory, "previous"), 0700))
	}
	return manager, task, record, journal
}

func TestCustomerServiceIndependentPairedSwitchMediaLayoutRecognizesEveryUnacknowledgedForwardAndReverseRename(t *testing.T) {
	manager, _, record, journal := createCustomerServicePairedSwitchMediaAndJournalWithoutDatabaseTestFixture(t)
	ctx := context.Background()
	verifyLayout := func() {
		t.Helper()
		layout, err := manager.inspectCustomerServicePairedSwitchCompletePhysicalMediaLayout(ctx, record)
		require.NoError(t, err)
		require.NoError(t, validateCustomerServicePairedSwitchPhysicalLayoutAgainstDurableProgress(journal, layout))
	}
	verifyLayout()
	for index := range journal.Moves {
		journal.Moves[index].Started = true
		verifyLayout()
		require.NoError(t, manager.performCustomerServicePairedSwitchBoundMediaRename(ctx, record, journal.Moves[index], false))
		verifyLayout()
		journal.Moves[index].Completed = true
		verifyLayout()
	}
	for index := len(journal.Moves) - 1; index >= 0; index-- {
		journal.Moves[index].RollbackStarted = true
		verifyLayout()
		require.NoError(t, manager.performCustomerServicePairedSwitchBoundMediaRename(ctx, record, journal.Moves[index], true))
		verifyLayout()
		journal.Moves[index].RolledBack = true
		verifyLayout()
	}
	t.Logf("verified before/after acknowledgement boundaries for %d forward and reverse media renames", len(journal.Moves))
}

func TestCustomerServiceIndependentPairedSwitchMediaLayoutRejectsUnexpectedContentAndImpossibleDurableProgress(t *testing.T) {
	for _, scenario := range []string{"changed_current_bytes", "extra_current_file", "extra_previous_file", "changed_candidate_empty_directory", "changed_volume_marker", "forward_progress_gap", "reverse_progress_gap", "unbound_move_path", "changed_baseline_directory"} {
		t.Run(scenario, func(t *testing.T) {
			manager, _, record, journal := createCustomerServicePairedSwitchMediaAndJournalWithoutDatabaseTestFixture(t)
			volume := record.MediaUnion.Volumes[0]
			directory, err := customerServiceCandidateMediaVolumeDirectory(record, volume)
			require.NoError(t, err)
			switch scenario {
			case "changed_current_bytes":
				require.NoError(t, os.WriteFile(filepath.Join(volume.StorageRoot, "unselected-target-only.txt"), []byte("changed"), 0600))
			case "extra_current_file":
				require.NoError(t, os.WriteFile(filepath.Join(volume.StorageRoot, "extra.txt"), []byte("extra"), 0600))
			case "extra_previous_file":
				require.NoError(t, os.WriteFile(filepath.Join(directory, "previous", "extra.txt"), []byte("extra"), 0600))
			case "changed_candidate_empty_directory":
				require.NoError(t, os.Remove(filepath.Join(directory, "tree", "empty-current-directory")))
			case "changed_volume_marker":
				require.NoError(t, os.WriteFile(filepath.Join(directory, customerServiceCandidateMediaVolumeIdentityFilename), []byte(`{"version":999}`), 0600))
			case "forward_progress_gap":
				journal.Moves[1].Started = true
			case "reverse_progress_gap":
				for index := range journal.Moves {
					journal.Moves[index].Started, journal.Moves[index].Completed = true, true
					require.NoError(t, manager.performCustomerServicePairedSwitchBoundMediaRename(context.Background(), record, journal.Moves[index], false))
				}
				journal.Moves[0].RollbackStarted = true
			case "unbound_move_path":
				journal.Moves[0].EntryName = "unregistered-name"
				inventory, _, err := manager.Configuration.planCustomerServiceCompleteCurrentMediaUnion(context.Background(), &CustomerServiceEffectiveSelectionInspection{SelectionDependenciesSatisfied: true, Binding: CustomerServiceRecoveryPlanSnapshotBinding{TargetMediaFingerprint: record.MediaUnion.BaselineMediaFingerprint}}, &Manifest{}, "")
				require.NoError(t, err)
				_, err = reconstructCustomerServicePairedSwitchBaselineAndValidateMoves(journal, inventory)
				require.Error(t, err)
				return
			case "changed_baseline_directory":
				journal.BaselineDirectories = append(journal.BaselineDirectories, customerServiceCandidateMediaDirectory{volume.RootName, "unknown-directory"})
				inventory := &customerServiceCandidateCompleteMediaUnionInventory{BaselineTreeSHA256: strings.Repeat("a", 64)}
				_, err := reconstructCustomerServicePairedSwitchBaselineAndValidateMoves(journal, inventory)
				require.Error(t, err)
				return
			}
			layout, err := manager.inspectCustomerServicePairedSwitchCompletePhysicalMediaLayout(context.Background(), record)
			if scenario == "changed_volume_marker" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Error(t, validateCustomerServicePairedSwitchPhysicalLayoutAgainstDurableProgress(journal, layout))
		})
	}
}

func TestCustomerServiceIndependentPairedSwitchActiveOrMalformedJournalBlocksCandidateDisposalAndMaintenanceRelease(t *testing.T) {
	for _, scenario := range []string{"active", "invalid_json", "digest_changed", "execution_enabled", "path_escape", "invalid_database_phase", "same_database_oid", "invalid_progress", "record_symlink"} {
		t.Run(scenario, func(t *testing.T) {
			manager, task, record, journal := createCustomerServicePairedSwitchMediaAndJournalWithoutDatabaseTestFixture(t)
			directory, _ := manager.taskDirectory(task.ID)
			filename := filepath.Join(directory, customerServiceIndependentPairedSwitchJournalFilename)
			switch scenario {
			case "execution_enabled":
				journal.ExecutionAvailable = true
			case "path_escape":
				journal.Moves[0].EntryName = "../outside"
			case "invalid_database_phase":
				journal.DatabasePhase = "unknown"
			case "same_database_oid":
				journal.CandidateDatabaseOID = journal.OriginalDatabaseOID
			case "invalid_progress":
				journal.Moves[0].Completed = true
			}
			require.NoError(t, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
			switch scenario {
			case "invalid_json":
				require.NoError(t, os.WriteFile(filename, []byte("{"), 0600))
			case "digest_changed":
				journal.SHA256 = strings.Repeat("0", 64)
				require.NoError(t, writeDurableJSON(filename, journal))
			case "record_symlink":
				require.NoError(t, os.Remove(filename))
				if err := os.Symlink(filepath.Join(directory, customerServicePersistentCandidateLifecycleFilename), filename); err != nil {
					t.Skip("symlink creation unavailable")
				}
			}
			require.Error(t, manager.requireCustomerServicePersistentCandidateDiscarded(task.ID))
			require.Error(t, manager.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
			require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
			for _, volume := range record.MediaUnion.Volumes {
				candidateDirectory, err := customerServiceCandidateMediaVolumeDirectory(record, volume)
				require.NoError(t, err)
				_, err = os.Stat(candidateDirectory)
				require.NoError(t, err)
			}
		})
	}
}
