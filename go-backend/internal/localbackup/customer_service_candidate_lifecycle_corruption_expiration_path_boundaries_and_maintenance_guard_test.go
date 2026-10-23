package localbackup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/stretchr/testify/require"
)

func createUnallocatedCustomerServiceCandidateLifecycleForMaintenanceGuardTest(t *testing.T) (*Manager, *Task, *customerServicePersistentCandidateLifecycle) {
	t.Helper()
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { manager.Close() })
	task, err := manager.createTask("upload", "synthetic-operator")
	require.NoError(t, err)
	task.State = "verified"
	task.Manifest = &Manifest{PackageID: newTaskID()}
	task.ExpiresAt = time.Now().Add(-time.Hour)
	require.NoError(t, manager.saveTask(task))
	attemptID := newTaskID()
	record := &customerServicePersistentCandidateLifecycle{Version: 1, State: "waiting_for_application_stop", TaskID: task.ID, AttemptID: attemptID, CreatedAt: time.Now().UTC(), Request: CustomerServiceEffectiveSelectionInspectionRequest{PlanSHA256: strings.Repeat("a", 64)}, Identity: customerServiceDisposableCandidateDatabaseIdentity{Version: 1, State: "allocation_not_attempted", DatabaseName: "local_backup_candidate_" + attemptID[:16], PlanSHA256: strings.Repeat("a", 64)}}
	require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
	return manager, task, record
}

func TestCustomerServiceCandidateLifecyclePreventsExpirationAndReleaseUntilUnallocatedAttemptIsDurablyDiscarded(t *testing.T) {
	manager, task, record := createUnallocatedCustomerServiceCandidateLifecycleForMaintenanceGuardTest(t)
	directory, _ := manager.taskDirectory(task.ID)
	filename := filepath.Join(directory, "application-backup.age")
	require.NoError(t, os.WriteFile(filename, []byte("synthetic upload must remain available"), 0600))
	require.ErrorContains(t, manager.ExpireWorkCopies(context.Background()), "safely discarded")
	_, err := os.Stat(filename)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename), []byte(`{"task_id":"`+task.ID+`"}`), 0600))
	require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
	require.NoError(t, manager.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
	require.NoError(t, manager.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
	loaded, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	require.NoError(t, err)
	require.Equal(t, "discarded_not_switchable", loaded.State)
	require.Equal(t, record.AttemptID, loaded.AttemptID)
	require.NoError(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
	require.NoError(t, manager.ExpireWorkCopies(context.Background()))
	_, err = os.Stat(filename)
	require.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(directory, customerServicePersistentCandidateLifecycleFilename))
	require.NoError(t, err, "audit lifecycle remains after TTL")
}

func TestCustomerServiceCandidateLifecycleMalformedOrInconsistentRecordCannotUnlockOrExpireWorkCopies(t *testing.T) {
	for _, scenario := range []string{"invalid_json", "changed_digest", "unknown_state", "other_task", "other_database", "switchable_flag", "missing_oid", "impossible_discard", "missing_validation", "oversized_record"} {
		t.Run(scenario, func(t *testing.T) {
			manager, task, record := createUnallocatedCustomerServiceCandidateLifecycleForMaintenanceGuardTest(t)
			directory, _ := manager.taskDirectory(task.ID)
			filename := filepath.Join(directory, customerServicePersistentCandidateLifecycleFilename)
			switch scenario {
			case "invalid_json":
				require.NoError(t, os.WriteFile(filename, []byte("{"), 0600))
			case "changed_digest":
				record.SHA256 = strings.Repeat("0", 64)
				require.NoError(t, writeDurableJSON(filename, record))
			case "unknown_state":
				record.State = "ready_to_switch"
				require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
			case "other_task":
				record.TaskID = newTaskID()
				require.NoError(t, writeDurableJSON(filename, record))
			case "other_database":
				record.Identity.DatabaseName = manager.Configuration.DatabaseName
				require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
			case "switchable_flag":
				record.ExecutionAvailable = true
				require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
			case "missing_oid":
				record.Identity.State = "allocated_not_switchable"
				require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
			case "impossible_discard":
				record.State = "discarded_not_switchable"
				record.Identity.State = "allocation_intent"
				require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
			case "missing_validation":
				record.State = "sealed_not_switchable"
				require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
			case "oversized_record":
				require.NoError(t, os.WriteFile(filename, []byte(strings.Repeat(" ", (4<<20)+1)), 0600))
			}
			_, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.Error(t, err)
			require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
			require.Error(t, manager.ExpireWorkCopies(context.Background()))
			_, err = manager.ApplyVerifiedRecovery(task.ID, "RESTORE "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+task.Manifest.PackageID)
			require.Error(t, err)
		})
	}
}

func TestCustomerServiceCandidateDisposalCannotAdvanceWhenItsDurableDiscardIntentCannotBeWritten(t *testing.T) {
	manager, task, record := createUnallocatedCustomerServiceCandidateLifecycleForMaintenanceGuardTest(t)
	workspace, err := manager.customerServicePersistentCandidateConstructionWorkspace(record)
	require.NoError(t, err)
	require.NoError(t, os.Mkdir(workspace.Directory, 0700))
	directory, _ := manager.taskDirectory(task.ID)
	blockedTemporaryFile := filepath.Join(directory, customerServicePersistentCandidateLifecycleFilename+".next")
	require.NoError(t, os.Mkdir(blockedTemporaryFile, 0700))
	require.Error(t, manager.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
	loaded, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	require.NoError(t, err)
	require.Equal(t, record.State, loaded.State)
	_, err = os.Stat(workspace.Directory)
	require.NoError(t, err)
	require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
	require.NoError(t, os.Remove(blockedTemporaryFile))
	require.NoError(t, manager.discardCustomerServicePersistentCandidate(context.Background(), task.ID))
	loaded, err = manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	require.NoError(t, err)
	require.Equal(t, "discarded_not_switchable", loaded.State)
}

func TestCustomerServiceCandidateWorkspaceCleanupRejectsOutsidePathsAndSymbolicLinkAncestors(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	outside := filepath.Join(t.TempDir(), "outside-candidate-directory")
	require.NoError(t, os.Mkdir(outside, 0700))
	sentinel := filepath.Join(outside, "must-survive.txt")
	require.NoError(t, os.WriteFile(sentinel, []byte("synthetic outside data"), 0600))
	identity := customerServiceDisposableCandidateDatabaseIdentity{Version: 1, State: "allocation_not_attempted"}
	for _, directory := range []string{cfg.Workspace, filepath.Dir(cfg.Workspace), outside} {
		err := cfg.discardCustomerServiceRecordedCandidateConstructionWorkspace(context.Background(), &customerServiceRecordedCandidateConstructionWorkspace{Directory: directory, Identity: identity})
		require.ErrorContains(t, err, "outside")
	}
	link := filepath.Join(cfg.Workspace, "synthetic-symbolic-link")
	if err := os.Symlink(filepath.Dir(outside), link); err != nil {
		t.Log("symbolic link creation unavailable on this platform")
		return
	}
	err := cfg.discardCustomerServiceRecordedCandidateConstructionWorkspace(context.Background(), &customerServiceRecordedCandidateConstructionWorkspace{Directory: filepath.Join(link, filepath.Base(outside)), Identity: identity})
	require.ErrorContains(t, err, "symbolic link")
	_, err = os.Stat(sentinel)
	require.NoError(t, err)
}

func TestCustomerServiceDisposableCandidateStartupOnlyDiscardsExplicitUnallocatedIdentityAndRetainsUnknownFiles(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	for _, scenario := range []string{"unallocated", "missing_identity", "malformed_identity", "allocation_intent_without_oid"} {
		directory := filepath.Join(cfg.Workspace, "customer-service-candidate-rehearsal-"+scenario)
		require.NoError(t, os.Mkdir(directory, 0700))
		if scenario == "missing_identity" {
			continue
		}
		identity := customerServiceDisposableCandidateDatabaseIdentity{Version: 1, State: "allocation_not_attempted", DatabaseName: "local_backup_candidate_" + newTaskID()[:16]}
		if scenario == "allocation_intent_without_oid" {
			identity.State = "allocation_intent"
		}
		filename := filepath.Join(directory, "customer-service-disposable-candidate-database-identity.json")
		if scenario == "malformed_identity" {
			require.NoError(t, os.WriteFile(filename, []byte("{"), 0600))
		} else {
			require.NoError(t, writeDurableJSON(filename, identity))
		}
	}
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	_, err = os.Stat(filepath.Join(cfg.Workspace, "customer-service-candidate-rehearsal-unallocated"))
	require.True(t, os.IsNotExist(err))
	for _, scenario := range []string{"missing_identity", "malformed_identity", "allocation_intent_without_oid"} {
		_, err = os.Stat(filepath.Join(cfg.Workspace, "customer-service-candidate-rehearsal-"+scenario))
		require.NoError(t, err)
	}
}
