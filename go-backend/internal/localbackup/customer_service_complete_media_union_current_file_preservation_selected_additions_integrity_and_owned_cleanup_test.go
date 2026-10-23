package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/stretchr/testify/require"
)

func createCustomerServiceCompleteMediaUnionPlanTestFixture(t *testing.T) (*Manager, *Task, *customerServicePersistentCandidateLifecycle, *CustomerServiceEffectiveSelectionInspection, *Manifest, string) {
	t.Helper()
	manager, task, record := createUnallocatedCustomerServiceCandidateLifecycleForMaintenanceGuardTest(t)
	for _, rootName := range mediaRoots {
		require.NoError(t, os.WriteFile(filepath.Join(manager.Configuration.StorageRoots[rootName], "unselected-target-only.txt"), []byte("synthetic current "+rootName), 0600))
		require.NoError(t, os.MkdirAll(filepath.Join(manager.Configuration.StorageRoots[rootName], "empty-current-directory"), 0700))
	}
	require.NoError(t, os.WriteFile(filepath.Join(manager.Configuration.StorageRoots["uploads"], ".current-dot-file"), []byte("synthetic ordinary dot file"), 0600))
	verifiedDirectory := filepath.Join(t.TempDir(), "verified")
	manifest := &Manifest{}
	for _, path := range []string{"media/uploads_private/selected/nested/required.txt", "media/uploads/unselected-package-only.txt"} {
		filename := filepath.Join(verifiedDirectory, filepath.FromSlash(path))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, []byte("synthetic package bytes"), 0600))
		identity, err := fileIdentity(filename, path)
		require.NoError(t, err)
		manifest.Files = append(manifest.Files, identity)
	}
	baseline, err := manager.Configuration.mediaFingerprint(context.Background())
	require.NoError(t, err)
	expected := manifest.Files[0]
	pathDigest := sha256.Sum256([]byte(expected.Path))
	selection := &CustomerServiceEffectiveSelectionInspection{PlanSHA256: record.Request.PlanSHA256, SHA256: strings.Repeat("b", 64), SelectionDependenciesSatisfied: true, Binding: CustomerServiceRecoveryPlanSnapshotBinding{TargetMediaFingerprint: baseline}, RequiredFiles: []CustomerServiceRecoveryPlanFileAction{{PathSHA256: hex.EncodeToString(pathDigest[:]), Action: "selected_add_package_file", SourceSHA256: expected.SHA256, SourceSize: expected.Size}}}
	return manager, task, record, selection, manifest, verifiedDirectory
}

func attachCustomerServicePlannedMediaUnionLifecycleForConstructionTest(t *testing.T, manager *Manager, record *customerServicePersistentCandidateLifecycle, inventory *customerServiceCandidateCompleteMediaUnionInventory) {
	t.Helper()
	record.State = "constructing_not_switchable"
	record.MediaUnion = &customerServiceCandidateCompleteMediaUnionLifecycle{Version: 1, State: "allocation_intent", InventorySHA256: inventory.SHA256, BaselineMediaFingerprint: inventory.BaselineMediaFingerprint, CandidateMediaFingerprint: inventory.CandidateMediaFingerprint, CurrentFiles: len(inventory.Files) - inventory.AddedPackageFiles, AddedPackageFiles: inventory.AddedPackageFiles, TotalBytes: inventory.TotalBytes}
	for _, rootName := range mediaRoots {
		record.MediaUnion.Volumes = append(record.MediaUnion.Volumes, customerServiceCandidateMediaVolumeState{RootName: rootName, StorageRoot: manager.Configuration.StorageRoots[rootName], State: "unallocated"})
	}
	require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
}

func TestCustomerServiceCompleteMediaUnionPreservesAllCurrentFilesEmptyDirectoriesAndOnlyAddsSelectedPackageFiles(t *testing.T) {
	manager, task, record, selection, manifest, verifiedDirectory := createCustomerServiceCompleteMediaUnionPlanTestFixture(t)
	ctx := context.Background()
	inventory, encoded, err := manager.Configuration.planCustomerServiceCompleteCurrentMediaUnion(ctx, selection, manifest, verifiedDirectory)
	require.NoError(t, err)
	require.Equal(t, 5, len(inventory.Files))
	require.Equal(t, 1, inventory.AddedPackageFiles)
	require.NotEqual(t, inventory.BaselineMediaFingerprint, inventory.CandidateMediaFingerprint)
	require.NotEmpty(t, encoded)
	attachCustomerServicePlannedMediaUnionLifecycleForConstructionTest(t, manager, record, inventory)
	require.NoError(t, manager.constructCustomerServiceCompleteMediaUnionOnOwnedVolumes(ctx, record, inventory, verifiedDirectory))
	require.NoError(t, manager.verifyCustomerServiceCompleteCandidateMediaUnion(ctx, record, inventory))
	for _, volume := range record.MediaUnion.Volumes {
		directory, err := customerServiceCandidateMediaVolumeDirectory(record, volume)
		require.NoError(t, err)
		content, err := os.ReadFile(filepath.Join(directory, "tree", "unselected-target-only.txt"))
		require.NoError(t, err)
		require.Equal(t, "synthetic current "+volume.RootName, string(content))
		info, err := os.Stat(filepath.Join(directory, "tree", "empty-current-directory"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
		_, err = os.Stat(filepath.Join(directory, "tree", "unselected-package-only.txt"))
		require.True(t, os.IsNotExist(err))
		if volume.RootName == "uploads_private" {
			content, err := os.ReadFile(filepath.Join(directory, "tree", "selected", "nested", "required.txt"))
			require.NoError(t, err)
			require.Equal(t, "synthetic package bytes", string(content))
			if runtime.GOOS != "windows" {
				info, err := os.Stat(filepath.Join(directory, "tree", "selected", "nested", "required.txt"))
				require.NoError(t, err)
				require.Equal(t, os.FileMode(0600), info.Mode().Perm())
			}
		}
	}
	require.NoError(t, manager.discardCustomerServicePersistentCandidate(ctx, task.ID))
	loaded, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	require.NoError(t, err)
	require.Equal(t, "discarded_not_switchable", loaded.MediaUnion.State)
	actual, err := manager.Configuration.mediaFingerprint(ctx)
	require.NoError(t, err)
	require.Equal(t, selection.Binding.TargetMediaFingerprint, actual)
}

func TestCustomerServiceCompleteMediaUnionRejectsUnboundConflictingUnsafeOrUnboundedPlans(t *testing.T) {
	for _, scenario := range []string{"incomplete_selection", "stale_current_media", "source_bytes_changed", "source_digest_changed", "required_identity_missing", "required_identity_duplicate", "blocked_action", "preserve_missing_file", "addition_overwrites_current", "file_directory_collision", "case_alias_collision", "entry_limit", "byte_limit", "source_symlink", "nested_reserved_path", "current_symlink", "metadata_limit"} {
		t.Run(scenario, func(t *testing.T) {
			manager, _, _, selection, manifest, verifiedDirectory := createCustomerServiceCompleteMediaUnionPlanTestFixture(t)
			switch scenario {
			case "incomplete_selection":
				selection.SelectionDependenciesSatisfied = false
			case "stale_current_media":
				selection.Binding.TargetMediaFingerprint = strings.Repeat("0", 64)
			case "source_bytes_changed":
				require.NoError(t, os.WriteFile(filepath.Join(verifiedDirectory, filepath.FromSlash(manifest.Files[0].Path)), []byte("synthetic changed"), 0600))
			case "source_digest_changed":
				selection.RequiredFiles[0].SourceSHA256 = strings.Repeat("0", 64)
			case "required_identity_missing":
				selection.RequiredFiles[0].PathSHA256 = strings.Repeat("0", 64)
			case "required_identity_duplicate":
				selection.RequiredFiles = append(selection.RequiredFiles, selection.RequiredFiles[0])
			case "blocked_action":
				selection.RequiredFiles[0].Action = "blocked_effective_file_registration_conflict"
			case "preserve_missing_file":
				selection.RequiredFiles[0].Action = "preserve_current_required_file"
			case "addition_overwrites_current":
				require.NoError(t, os.MkdirAll(filepath.Join(manager.Configuration.StorageRoots["uploads_private"], "selected", "nested"), 0700))
				require.NoError(t, os.WriteFile(filepath.Join(manager.Configuration.StorageRoots["uploads_private"], "selected", "nested", "required.txt"), []byte("synthetic current bytes"), 0600))
			case "file_directory_collision":
				require.NoError(t, os.MkdirAll(filepath.Join(manager.Configuration.StorageRoots["uploads_private"], "selected", "nested", "required.txt"), 0700))
			case "case_alias_collision":
				require.NoError(t, os.MkdirAll(filepath.Join(manager.Configuration.StorageRoots["uploads_private"], "Selected"), 0700))
			case "entry_limit":
				manager.Configuration.MaxFiles = 1
			case "byte_limit":
				manager.Configuration.MaxExpandedBytes = 1
			case "source_symlink":
				filename := filepath.Join(verifiedDirectory, filepath.FromSlash(manifest.Files[0].Path))
				require.NoError(t, os.Remove(filename))
				if err := os.Symlink(filepath.Join(verifiedDirectory, filepath.FromSlash(manifest.Files[1].Path)), filename); err != nil {
					t.Skip("symlink creation unavailable")
				}
			case "nested_reserved_path":
				require.NoError(t, os.MkdirAll(filepath.Join(manager.Configuration.StorageRoots["uploads"], "nested", ".local-backup-hidden"), 0700))
			case "current_symlink":
				if err := os.Symlink(verifiedDirectory, filepath.Join(manager.Configuration.StorageRoots["uploads"], "unsafe-link")); err != nil {
					t.Skip("symlink creation unavailable")
				}
			case "metadata_limit":
				inventory := &customerServiceCandidateCompleteMediaUnionInventory{EncodedBytes: customerServiceCandidateMediaMaximumInventoryBytes}
				require.Error(t, accountCustomerServiceCandidateMediaInventoryMetadata(inventory, "media/uploads/next"))
				return
			}
			// Bind the deliberately changed current tree for collision tests so the
			// specific tree/path guard is reached rather than just freshness.
			if scenario == "addition_overwrites_current" || scenario == "file_directory_collision" || scenario == "case_alias_collision" {
				fingerprint, err := manager.Configuration.mediaFingerprint(context.Background())
				require.NoError(t, err)
				selection.Binding.TargetMediaFingerprint = fingerprint
			}
			_, _, err := manager.Configuration.planCustomerServiceCompleteCurrentMediaUnion(context.Background(), selection, manifest, verifiedDirectory)
			require.Error(t, err)
		})
	}
}

func TestCustomerServiceCompleteMediaUnionActualVerificationDetectsMissingChangedExtraAndReservedFilesOrDirectories(t *testing.T) {
	for _, scenario := range []string{"changed_file", "missing_file", "extra_file", "reserved_extra_file", "removed_empty_directory", "current_file_changed", "current_empty_directory_added"} {
		t.Run(scenario, func(t *testing.T) {
			manager, _, record, selection, manifest, verifiedDirectory := createCustomerServiceCompleteMediaUnionPlanTestFixture(t)
			inventory, _, err := manager.Configuration.planCustomerServiceCompleteCurrentMediaUnion(context.Background(), selection, manifest, verifiedDirectory)
			require.NoError(t, err)
			attachCustomerServicePlannedMediaUnionLifecycleForConstructionTest(t, manager, record, inventory)
			require.NoError(t, manager.constructCustomerServiceCompleteMediaUnionOnOwnedVolumes(context.Background(), record, inventory, verifiedDirectory))
			directory, err := customerServiceCandidateMediaVolumeDirectory(record, record.MediaUnion.Volumes[0])
			require.NoError(t, err)
			tree := filepath.Join(directory, "tree")
			switch scenario {
			case "changed_file":
				require.NoError(t, os.WriteFile(filepath.Join(tree, "unselected-target-only.txt"), []byte("synthetic corrupted"), 0600))
			case "missing_file":
				require.NoError(t, os.Remove(filepath.Join(tree, "unselected-target-only.txt")))
			case "extra_file":
				require.NoError(t, os.WriteFile(filepath.Join(tree, "unexpected.txt"), []byte("synthetic extra"), 0600))
			case "reserved_extra_file":
				require.NoError(t, os.WriteFile(filepath.Join(tree, ".local-backup-injected"), []byte("synthetic extra"), 0600))
			case "removed_empty_directory":
				require.NoError(t, os.Remove(filepath.Join(tree, "empty-current-directory")))
			case "current_file_changed":
				require.NoError(t, os.WriteFile(filepath.Join(manager.Configuration.StorageRoots["uploads"], "unselected-target-only.txt"), []byte("synthetic changed live"), 0600))
			case "current_empty_directory_added":
				require.NoError(t, os.Mkdir(filepath.Join(manager.Configuration.StorageRoots["uploads"], "unexpected-empty-directory"), 0700))
			}
			require.Error(t, manager.verifyCustomerServiceCompleteCandidateMediaUnion(context.Background(), record, inventory))
		})
	}
}

func TestCustomerServiceCompleteMediaUnionCleanupKeepsUnownedDirectoriesAndResumesRecordedPartialDisposal(t *testing.T) {
	for _, scenario := range []string{"missing_marker", "changed_marker", "symlink_marker", "partial_cleanup"} {
		t.Run(scenario, func(t *testing.T) {
			manager, task, record, selection, manifest, verifiedDirectory := createCustomerServiceCompleteMediaUnionPlanTestFixture(t)
			inventory, _, err := manager.Configuration.planCustomerServiceCompleteCurrentMediaUnion(context.Background(), selection, manifest, verifiedDirectory)
			require.NoError(t, err)
			attachCustomerServicePlannedMediaUnionLifecycleForConstructionTest(t, manager, record, inventory)
			require.NoError(t, manager.constructCustomerServiceCompleteMediaUnionOnOwnedVolumes(context.Background(), record, inventory, verifiedDirectory))
			directory, err := customerServiceCandidateMediaVolumeDirectory(record, record.MediaUnion.Volumes[0])
			require.NoError(t, err)
			marker := filepath.Join(directory, customerServiceCandidateMediaVolumeIdentityFilename)
			switch scenario {
			case "missing_marker":
				require.NoError(t, os.Remove(marker))
			case "changed_marker":
				require.NoError(t, os.WriteFile(marker, []byte(`{"version":999}`), 0600))
			case "symlink_marker":
				require.NoError(t, os.Remove(marker))
				if err := os.Symlink(filepath.Join(directory, "tree", "unselected-target-only.txt"), marker); err != nil {
					t.Skip("symlink creation unavailable")
				}
			case "partial_cleanup":
				record.MediaUnion.State = "discard_required"
				record.MediaUnion.Volumes[0].State = "removed"
				record.State = "discard_required"
				require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
				require.NoError(t, os.RemoveAll(directory))
			}
			require.NoError(t, os.WriteFile(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename), []byte(`{"task_id":"`+task.ID+`"}`), 0600))
			require.NoError(t, manager.Close())
			restarted, err := NewManager(manager.Configuration)
			require.NoError(t, err)
			defer restarted.Close()
			loaded, err := restarted.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			if scenario == "partial_cleanup" {
				require.Equal(t, "discarded_not_switchable", loaded.State)
				require.Equal(t, "discarded_not_switchable", loaded.MediaUnion.State)
			} else {
				require.Equal(t, "inspection_required", loaded.State)
				require.Error(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
				_, err = os.Stat(directory)
				require.NoError(t, err)
			}
			fenced, err := maintenancegate.IsApplicationFenced(manager.Configuration.MaintenanceDirectory)
			require.NoError(t, err)
			require.True(t, fenced)
		})
	}
}

func TestCustomerServiceCandidateCombinedCapacityRejectsAggregateShortageAndArithmeticOverflow(t *testing.T) {
	require.Error(t, validateCustomerServiceCandidateCombinedFilesystemCapacity([]customerServiceCandidateFilesystemCapacityRequirement{{Components: []string{"workspace", "database", "media"}, AdditionalBytes: 300, AvailableBytes: 250}}))
	require.NoError(t, validateCustomerServiceCandidateCombinedFilesystemCapacity([]customerServiceCandidateFilesystemCapacityRequirement{{AdditionalBytes: 250, AvailableBytes: 250}}))
	_, err := addCustomerServiceCandidateCapacityBytesWithoutOverflow(9223372036854775800, 100)
	require.Error(t, err)
	_, err = addCustomerServiceCandidateCapacityBytesWithoutOverflow(-1, 1)
	require.Error(t, err)
}
