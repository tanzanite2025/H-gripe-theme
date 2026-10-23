package localbackup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func verifyProductCatalogStoppedRevalidationPairedSwitchAndEveryInterruptedRenameRollback(t *testing.T, managerPointer **Manager, packageData []byte) {
	t.Helper()
	ctx := context.Background()
	cfg := (*managerPointer).Configuration
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	driftScenarios := []struct {
		name   string
		change func(*Task) func()
	}{
		{"latest_inventory", func(task *Task) func() {
			db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec(`UPDATE product_variants SET stock=stock+1 WHERE sku='product-backup-sku'`)
			require.NoError(t, err)
			return func() {}
		}},
		{"latest_order_detail", func(task *Task) func() {
			db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec(`UPDATE order_items SET product_name='Updated retained order description' WHERE product_name='Historical order description'`)
			require.NoError(t, err)
			return func() {}
		}},
		{"latest_customer_message", func(task *Task) func() {
			db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec(`UPDATE ticket_messages SET content='Latest message after product preview' WHERE ticket_id=(SELECT id FROM tickets WHERE ticket_number='product-backup-chat')`)
			require.NoError(t, err)
			return func() {}
		}},
		{"new_unselected_attachment", func(task *Task) func() {
			require.NoError(t, os.WriteFile(filepath.Join(cfg.StorageRoots["uploads_private"], "after-product-preview.txt"), []byte("new customer attachment"), 0600))
			return func() {}
		}},
		{"new_empty_directory", func(task *Task) func() {
			require.NoError(t, os.Mkdir(filepath.Join(cfg.StorageRoots["uploads_private"], "after-product-preview-empty"), 0700))
			return func() {}
		}},
		{"target_schema", func(task *Task) func() {
			db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			_, err = db.Exec(`CREATE INDEX synthetic_target_product_name_index ON products(name)`)
			require.NoError(t, err)
			db.Close()
			return func() {
				db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				defer db.Close()
				_, err = db.Exec(`DROP INDEX synthetic_target_product_name_index`)
				require.NoError(t, err)
			}
		}},
		{"candidate_data", func(task *Task) func() {
			db, err := cfg.connectDatabase(ctx, task.StageDatabase)
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec(`UPDATE products SET name='Tampered candidate' WHERE locale='en'`)
			require.NoError(t, err)
			return func() {}
		}},
		{"candidate_sequence", func(task *Task) func() {
			db, err := cfg.connectDatabase(ctx, task.StageDatabase)
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec(`SELECT nextval(pg_get_serial_sequence('products','id'))`)
			require.NoError(t, err)
			return func() {}
		}},
		{"candidate_schema", func(task *Task) func() {
			db, err := cfg.connectDatabase(ctx, task.StageDatabase)
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec(`CREATE INDEX synthetic_candidate_product_name_index ON products(name)`)
			require.NoError(t, err)
			return func() {}
		}},
		{"candidate_file", func(task *Task) func() {
			roots := (*managerPointer).productCatalogCandidateMediaConfiguration(task).StorageRoots
			require.NoError(t, os.WriteFile(filepath.Join(roots["uploads_private"], "retained-chat.txt"), []byte("tampered candidate attachment"), 0600))
			return func() {}
		}},
		{"bundled_static_image", func(task *Task) func() {
			path := filepath.Join(cfg.ProductStaticAssetsDirectory, "company/products/bundled-product-photo.png")
			original, err := os.ReadFile(path)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(path, []byte("different deployed static image"), 0600))
			return func() { require.NoError(t, os.WriteFile(path, original, 0600)) }
		}},
	}
	executedDriftScenarios := 0
	for _, scenario := range driftScenarios {
		t.Run("product_stopped_revalidation_rejects_"+scenario.name, func(t *testing.T) {
			executedDriftScenarios++
			manager := *managerPointer
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			restore := scenario.change(task)
			defer restore()
			_, inspectionError := manager.InspectProductCatalogCandidateHandoff(ctx, task.ID)
			require.Error(t, inspectionError, "read-only handoff inspection must also reject this actual preview drift")
			err := manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, nil)
			require.Error(t, err)
			expected := "product target changed after preview"
			if strings.HasPrefix(scenario.name, "candidate_") {
				expected = "product candidate database identity, data or schema changed"
			}
			if scenario.name == "candidate_file" {
				expected = "product candidate complete media changed"
			}
			if scenario.name == "bundled_static_image" {
				expected = "bundled product image is absent or differs"
			}
			require.ErrorContains(t, err, expected)
			task, err = manager.GetTask(task.ID)
			require.NoError(t, err)
			require.Nil(t, task.Journal, "stale candidates must refuse before a switch journal or rename")
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced, "a rejected preview has not changed live data")
		})
	}
	checkpoints := []string{}
	t.Run("product_complete_paired_switch_and_verified_rollback", func(t *testing.T) {
		manager := *managerPointer
		task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
		require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
		require.NoError(t, manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, func(phase string) error {
			checkpoints = append(checkpoints, phase)
			if phase == "paired_switch_verified" {
				current, err := manager.GetTask(task.ID)
				require.NoError(t, err)
				require.True(t, current.Journal.DatabaseSwitched)
				require.False(t, current.ProductCatalogRecovery.OnlineOverwriteAvailable)
				require.False(t, taskHasSwitchedWholeApplicationRecoveryJournal(current))
				var oid int64
				var allowed bool
				require.NoError(t, control.QueryRow(`SELECT oid::bigint,datallowconn FROM pg_database WHERE datname=$1`, cfg.DatabaseName).Scan(&oid, &allowed))
				require.Equal(t, current.Journal.NewDatabaseOID, oid)
				require.False(t, allowed)
				require.NoError(t, manager.validateProductCatalogPairedPhysicalLayout(ctx, current))
				_, err = manager.ApplyVerifiedRecovery(task.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.Manifest.PackageID)
				require.Error(t, err)
				require.Error(t, manager.CompleteReconciledRecovery(task.ID, "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID, ReconciliationEvidence{}))
			}
			return nil
		}))
		task, err = manager.GetTask(task.ID)
		require.NoError(t, err)
		require.Equal(t, "rolled_back", task.State)
		require.True(t, task.Journal.ProductCatalog.RollbackVerified)
		requireProductCatalogOriginalDatabaseAndCompleteMediaRestored(t, manager, task)
	})
	require.NotEmpty(t, checkpoints)
	executedInterruptions := 0
	for _, phase := range checkpoints {
		t.Run("product_restart_rolls_back_"+phase, func(t *testing.T) {
			executedInterruptions++
			manager := *managerPointer
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			interruption := errors.New("synthetic executor interruption at " + phase)
			err := manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, func(actual string) error {
				if actual == phase {
					return interruption
				}
				return nil
			})
			require.ErrorIs(t, err, interruption)
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.True(t, fenced)
			require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
			require.NoError(t, manager.ExpireWorkCopies(ctx))
			require.NoError(t, manager.Close())
			restarted, err := NewManager(cfg)
			require.NoError(t, err)
			*managerPointer = restarted
			task, err = restarted.GetTask(task.ID)
			require.NoError(t, err)
			require.Equal(t, "rolled_back", task.State, task.Error)
			requireProductCatalogOriginalDatabaseAndCompleteMediaRestored(t, restarted, task)
		})
	}
	t.Logf("product stopped revalidation rejected %d stale/tampered candidates; complete paired switch and %d durable interruption/restart boundaries restored the original database and all media including empty directories", executedDriftScenarios, executedInterruptions)
	verifyProductCatalogInterruptedRollbackRejectsChangedDatabaseMediaJournalAndFenceIdentities(t, managerPointer, packageData)
	// Rollback artifacts belong to this task. Later legitimate live changes must
	// not turn their expiry into deletion of new online data or a permanent leak.
	t.Run("product_verified_rollback_artifact_cleanup_preserves_new_live_media", func(t *testing.T) {
		manager := *managerPointer
		tasks, err := manager.ListTasks()
		require.NoError(t, err)
		var selected *Task
		for _, task := range tasks {
			if task.Journal != nil && task.Journal.ProductCatalog != nil && task.State == "rolled_back" {
				selected = task
				break
			}
		}
		require.NotNil(t, selected)
		path := filepath.Join(cfg.StorageRoots["uploads_private"], "legitimate-new-file-after-rollback.txt")
		require.NoError(t, os.WriteFile(path, []byte("must survive cleanup"), 0600))
		require.NoError(t, manager.cleanupProductCatalogVerifiedPairedMediaVolumes(ctx, selected))
		body, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, "must survive cleanup", string(body))
		for _, rootName := range mediaRoots {
			_, err := os.Stat(manager.productCatalogPairedMediaVolumeDirectory(selected, rootName))
			require.True(t, os.IsNotExist(err))
		}
	})
}

func verifyProductCatalogInterruptedRollbackRejectsChangedDatabaseMediaJournalAndFenceIdentities(t *testing.T, managerPointer **Manager, packageData []byte) {
	t.Helper()
	ctx := context.Background()
	cfg := (*managerPointer).Configuration
	control, err := cfg.connectDatabase(ctx, "postgres")
	require.NoError(t, err)
	defer control.Close()
	for _, scenario := range []string{"previous_media_content", "volume_identity", "journal_digest", "missing_fence", "candidate_owner", "rollback_database_oid"} {
		t.Run("product_restart_retains_fence_for_changed_"+scenario, func(t *testing.T) {
			manager := *managerPointer
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, packageData)
			require.Equal(t, "product_module_isolated_verified", task.State, task.Error)
			stop := errors.New("synthetic interrupted paired switch")
			require.ErrorIs(t, manager.rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx, task.ID, func(phase string) error {
				if phase == "paired_switch_verified" {
					return stop
				}
				return nil
			}), stop)
			task, err = manager.GetTask(task.ID)
			require.NoError(t, err)
			var restore func()
			switch scenario {
			case "previous_media_content", "volume_identity", "missing_fence":
				path := filepath.Join(manager.productCatalogPairedMediaVolumeDirectory(task, "uploads_private"), "previous", "retained-chat.txt")
				if scenario == "volume_identity" {
					path = filepath.Join(manager.productCatalogPairedMediaVolumeDirectory(task, "uploads_private"), "product-paired-volume-binding.json")
				}
				if scenario == "missing_fence" {
					path = filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename)
				}
				original, err := os.ReadFile(path)
				require.NoError(t, err)
				if scenario == "missing_fence" {
					require.NoError(t, os.Remove(path))
				} else {
					require.NoError(t, os.WriteFile(path, []byte("changed synthetic identity"), 0600))
				}
				restore = func() { require.NoError(t, os.WriteFile(path, original, 0600)) }
			case "journal_digest":
				original := task.Journal.ProductCatalog.SHA256
				task.Journal.ProductCatalog.SHA256 = strings.Repeat("0", 64)
				require.NoError(t, manager.saveTask(task))
				restore = func() {
					current, err := (*managerPointer).GetTask(task.ID)
					require.NoError(t, err)
					current.Journal.ProductCatalog.SHA256 = original
					require.NoError(t, (*managerPointer).saveTask(current))
				}
			case "candidate_owner":
				_, err := control.Exec(`ALTER DATABASE ` + pq.QuoteIdentifier(cfg.DatabaseName) + ` OWNER TO ` + pq.QuoteIdentifier(cfg.ApplicationDatabaseRole))
				require.NoError(t, err)
				restore = func() {
					_, err := control.Exec(`ALTER DATABASE ` + pq.QuoteIdentifier(cfg.DatabaseName) + ` OWNER TO local_backup_isolated_owner`)
					require.NoError(t, err)
				}
			case "rollback_database_oid":
				holdingName := "local_backup_fixture_original_" + task.ID[:16]
				_, err := control.Exec(`ALTER DATABASE ` + pq.QuoteIdentifier(task.Journal.RollbackDatabase) + ` RENAME TO ` + pq.QuoteIdentifier(holdingName))
				require.NoError(t, err)
				_, err = control.Exec(`CREATE DATABASE ` + pq.QuoteIdentifier(task.Journal.RollbackDatabase))
				require.NoError(t, err)
				restore = func() {
					_, err := control.Exec(`DROP DATABASE ` + pq.QuoteIdentifier(task.Journal.RollbackDatabase))
					require.NoError(t, err)
					_, err = control.Exec(`ALTER DATABASE ` + pq.QuoteIdentifier(holdingName) + ` RENAME TO ` + pq.QuoteIdentifier(task.Journal.RollbackDatabase))
					require.NoError(t, err)
				}
			}
			require.NoError(t, manager.Close())
			restarted, err := NewManager(cfg)
			require.NoError(t, err)
			*managerPointer = restarted
			current, err := restarted.GetTask(task.ID)
			require.NoError(t, err)
			require.Equal(t, "interrupted", current.State)
			require.Equal(t, "product_paired_rehearsal_retained_for_inspection", current.Phase)
			require.False(t, current.Journal.ProductCatalog.RollbackVerified)
			require.Error(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
			var oid int64
			require.NoError(t, readDatabaseOID(ctx, control, cfg.DatabaseName, &oid))
			require.Equal(t, task.Journal.NewDatabaseOID, oid, "failed rollback must not replace an unverified database")
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			if scenario == "missing_fence" {
				require.False(t, fenced, "restart must not manufacture a missing fence")
				unrelatedTask, err := restarted.createTask("backup", "synthetic-unrelated-operator")
				require.NoError(t, err)
				_, err = restarted.acquireMaintenance(ctx, unrelatedTask)
				require.ErrorContains(t, err, "unresolved product paired journal")
			} else {
				require.True(t, fenced)
			}
			require.NoError(t, restarted.ExpireWorkCopies(ctx))
			restore()
			require.NoError(t, restarted.Close())
			restarted, err = NewManager(cfg)
			require.NoError(t, err)
			*managerPointer = restarted
			current, err = restarted.GetTask(task.ID)
			require.NoError(t, err)
			require.Equal(t, "rolled_back", current.State, current.Error)
			requireProductCatalogOriginalDatabaseAndCompleteMediaRestored(t, restarted, current)
		})
	}
}

func requireProductCatalogOriginalDatabaseAndCompleteMediaRestored(t *testing.T, manager *Manager, task *Task) {
	t.Helper()
	ctx := context.Background()
	binding := task.ProductCatalogPreviewBinding
	actual, err := manager.Configuration.captureProductCatalogRecoveryDatabaseBinding(ctx, manager.Configuration.DatabaseName)
	require.NoError(t, err)
	require.Equal(t, binding.Target, *actual)
	media, err := manager.Configuration.inventoryModuleRecoveryCompleteMedia(ctx)
	require.NoError(t, err)
	require.Equal(t, binding.TargetMedia, media)
	fingerprint, _, err := manager.Configuration.targetFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, task.PreviewTargetFingerprint, fingerprint)
	fenced, err := maintenancegate.IsApplicationFenced(manager.Configuration.MaintenanceDirectory)
	require.NoError(t, err)
	require.False(t, fenced)
	for _, move := range task.Journal.ProductCatalog.Moves {
		if move.Started {
			require.True(t, move.RolledBack, fmt.Sprint(move))
		}
		require.False(t, strings.Contains(move.EntryName, "/"))
	}
}
