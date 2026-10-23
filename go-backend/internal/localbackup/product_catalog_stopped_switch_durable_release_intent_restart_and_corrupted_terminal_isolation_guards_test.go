package localbackup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/service"
)

func createSyntheticStoppedProductSwitchWithVerifiedOriginalReleaseIntent(t *testing.T) (*Manager, *Task) {
	t.Helper()
	manager, task, _ := createSyntheticProductCatalogCandidateHandoffTaskAndRuntimeBoundary(t)
	directory, err := manager.taskDirectory(task.ID)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(directory, 0700))
	configurationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(manager.Configuration)
	require.NoError(t, err)
	task.ProductCatalogPreviewBinding.ConfigurationSHA256 = configurationDigest
	previewDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(task.ProductCatalogPreviewBinding)
	require.NoError(t, err)
	task.ProductViewStoppedDrain = &ProductViewStoppedDrainAndCandidateRebuildProgress{Version: 1, TaskID: task.ID, PackageID: task.Manifest.PackageID, ConfigurationSHA256: configurationDigest, Operator: "synthetic", TargetDatabaseOID: 1, Phase: "candidate_rebuilt", RebuiltPreviewSHA256: previewDigest, After: &service.ProductViewCountRecoveryBoundaryInspection{RedisBuffersEmpty: true}}
	require.NoError(t, manager.saveProductViewStoppedDrainProgress(task))
	packageDigest, err := productCatalogStoppedHandoffImmutablePackageSHA256(task)
	require.NoError(t, err)
	task.ProductCatalogStoppedHandoffPreparation = &ProductCatalogStoppedHandoffPreparationJournal{Version: 1, WindowID: newTaskID(), TaskID: task.ID, PackageSHA256: packageDigest, ConfigurationSHA256: configurationDigest, FenceSHA256: "synthetic-fence", Operator: "synthetic", Phase: "ready", StartedAt: time.Now().Add(-time.Hour), PreparedAt: time.Now().Add(-time.Minute), PreviewBindingSHA256: previewDigest, DrainProgressSHA256: task.ProductViewStoppedDrain.SHA256}
	require.NoError(t, manager.saveProductCatalogStoppedHandoffPreparationJournal(task))
	moves, err := buildModuleRecoveryPairedMediaMoves(task.ProductCatalogPreviewBinding.TargetMedia, task.ProductCatalogPreviewBinding.CandidateMedia)
	require.NoError(t, err)
	task.Journal = &RecoveryJournal{OldDatabaseOID: 1, NewDatabaseOID: 2, RollbackDatabase: "local_backup_product_old_" + task.ID[:16], ProductCatalogStoppedSwitch: &ProductCatalogContinuouslyStoppedSwitchJournal{Version: 1, TaskID: task.ID, WindowID: task.ProductCatalogStoppedHandoffPreparation.WindowID, PreparationSHA256: task.ProductCatalogStoppedHandoffPreparation.SHA256, PackageSHA256: packageDigest, PreviewBindingSHA256: previewDigest, Operator: "synthetic", StartedAt: time.Now(), Phase: "releasing_original", Moves: moves, RollbackVerifiedAt: time.Now()}}
	task.State = "running"
	require.NoError(t, manager.saveProductCatalogStoppedSwitchProgress(task))
	require.NoError(t, manager.validateProductCatalogStoppedSwitchJournalBinding(task))
	return manager, task
}

func TestStoppedProductSwitchRestartCompletesOnlyItsDurableVerifiedOriginalReleaseWithoutDatabaseAccess(t *testing.T) {
	previous, task := createSyntheticStoppedProductSwitchWithVerifiedOriginalReleaseIntent(t)
	manager, err := NewManager(previous.Configuration)
	require.NoError(t, err)
	defer manager.Close()
	recovered, err := manager.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, "rolled_back", recovered.State)
	require.True(t, taskHasDurablyReleasedOriginalAfterProductCatalogStoppedSwitchRollback(recovered))
	require.False(t, taskHasUnresolvedProductCatalogStoppedHandoffPreparation(recovered))
	require.Nil(t, manager.productCatalogStoppedHandoffLease)
	_, err = os.Stat(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename))
	require.True(t, os.IsNotExist(err))
}

func TestStoppedProductSwitchCorruptedTerminalAndScopeBindingsCannotDisableIsolationGuards(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		mutate func(*Task)
	}{
		{"outer_oid", func(task *Task) { task.Journal.OldDatabaseOID++ }},
		{"preparation_window", func(task *Task) { task.ProductCatalogStoppedHandoffPreparation.WindowID = newTaskID() }},
		{"package_expiration", func(task *Task) { task.ExpiresAt = task.ExpiresAt.Add(time.Hour) }},
		{"preview_database_owner", func(task *Task) { task.ProductCatalogPreviewBinding.Candidate.Owner = "forged" }},
		{"drain_receipts", func(task *Task) { task.ProductViewStoppedDrain.FlushCalls++ }},
		{"mixed_rehearsal_scope", func(task *Task) { task.Journal.ProductCatalog = &productCatalogPairedRehearsalJournal{} }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			manager, task := createSyntheticStoppedProductSwitchWithVerifiedOriginalReleaseIntent(t)
			require.NoError(t, manager.acknowledgeStoppedProductCatalogOriginalRelease(task))
			require.True(t, taskHasDurablyReleasedOriginalAfterProductCatalogStoppedSwitchRollback(task))
			scenario.mutate(task)
			require.Error(t, manager.validateProductCatalogStoppedSwitchJournalBinding(task))
			require.False(t, taskHasDurablyReleasedOriginalAfterProductCatalogStoppedSwitchRollback(task))
			require.True(t, taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task))
			require.Error(t, manager.releaseMaintenance(task))
		})
	}
}
