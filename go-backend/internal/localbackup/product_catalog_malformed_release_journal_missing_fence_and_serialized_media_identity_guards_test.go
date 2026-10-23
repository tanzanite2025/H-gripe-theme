package localbackup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/stretchr/testify/require"
)

func TestProductCatalogMalformedReleaseIntentCannotBecomeCompletedOrAcquireAnotherMaintenanceWindow(t *testing.T) {
	for _, journal := range []*RecoveryJournal{{}, {ProductCatalog: &productCatalogPairedRehearsalJournal{Version: 1, RollbackVerified: true}}} {
		t.Run(func() string {
			if journal.ProductCatalog == nil {
				return "missing_product_extension"
			}
			return "missing_preview_binding"
		}(), func(t *testing.T) {
			cfg := createIndependentRecoveryTestConfiguration(t)
			manager, err := NewManager(cfg)
			require.NoError(t, err)
			task, err := manager.createTask("upload", "synthetic-operator")
			require.NoError(t, err)
			task.State = "releasing_rollback"
			task.Manifest = &Manifest{Scope: productCatalogBackupScope}
			task.Journal = journal
			task.ExpiresAt = time.Now().Add(-time.Hour)
			require.NoError(t, manager.saveTask(task))
			require.NoError(t, manager.Close())
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			defer manager.Close()
			task, err = manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Equal(t, "interrupted", task.State)
			require.Error(t, manager.releaseMaintenance(task))
			require.Error(t, manager.rollbackPairedRecovery(context.Background(), task))
			require.NoError(t, manager.ExpireWorkCopies(context.Background()))
			retained, err := manager.GetTask(task.ID)
			require.NoError(t, err)
			require.False(t, retained.WorkCopyExpired, "an expired unresolved journal must retain its recovery work copies")
			unrelated, err := manager.createTask("backup", "synthetic-unrelated-operator")
			require.NoError(t, err)
			_, err = manager.acquireMaintenance(context.Background(), unrelated)
			require.ErrorContains(t, err, "unresolved product paired journal")
			_, err = os.Stat(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename))
			require.True(t, os.IsNotExist(err), "a missing fence must not be silently replaced")
		})
	}
}

func TestProductCatalogCompleteMediaIdentityRemainsEqualAfterDurableTaskSerialization(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads_private"], "retained-message-file.txt"), []byte("retained private file"), 0600))
	require.NoError(t, os.Mkdir(filepath.Join(cfg.StorageRoots["uploads_private"], "retained-empty-directory"), 0700))
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	inventory, err := cfg.inventoryModuleRecoveryCompleteMedia(context.Background())
	require.NoError(t, err)
	task, err := manager.createTask("upload", "synthetic-operator")
	require.NoError(t, err)
	task.ProductCatalogPreviewBinding = &productCatalogRecoveryPreviewBinding{Version: 1, TargetMedia: inventory, CandidateMedia: inventory}
	require.NoError(t, manager.saveTask(task))
	reloaded, err := manager.GetTask(task.ID)
	require.NoError(t, err)
	current, err := cfg.inventoryModuleRecoveryCompleteMedia(context.Background())
	require.NoError(t, err)
	require.Equal(t, reloaded.ProductCatalogPreviewBinding.TargetMedia, current)
	require.Len(t, current.Files, 1)
	require.Len(t, current.Directories, 1)
}
