package localbackup

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type customerServiceCandidateForceKilledExecutorTestConfiguration struct {
	Configuration        Configuration `json:"configuration"`
	TaskID               string        `json:"task_id"`
	AllocationCheckpoint string        `json:"allocation_checkpoint"`
	ReadyFilename        string        `json:"ready_filename"`
}

func createPartiallyAllocatedCustomerServiceCandidateMediaVolumeForForceKillCheckpointTest(t *testing.T, manager *Manager, record *customerServicePersistentCandidateLifecycle, checkpoint string) {
	t.Helper()
	record.MediaUnion = &customerServiceCandidateCompleteMediaUnionLifecycle{Version: 1, State: "allocation_intent", InventorySHA256: strings.Repeat("a", 64)}
	for _, rootName := range mediaRoots {
		record.MediaUnion.Volumes = append(record.MediaUnion.Volumes, customerServiceCandidateMediaVolumeState{RootName: rootName, StorageRoot: manager.Configuration.StorageRoots[rootName], State: "unallocated"})
	}
	volume := &record.MediaUnion.Volumes[0]
	volume.State = "allocation_intent"
	require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
	directory, err := customerServiceCandidateMediaVolumeDirectory(record, *volume)
	require.NoError(t, err)
	require.NoError(t, os.Mkdir(directory, 0700))
	require.NoError(t, syncDirectory(volume.StorageRoot))
	if checkpoint == "media_identity_recorded_with_partial_tree" {
		require.NoError(t, writeDurableJSON(filepath.Join(directory, customerServiceCandidateMediaVolumeIdentityFilename), customerServiceCandidateExpectedMediaVolumeIdentity(record, *volume)))
		volume.State = "identity_recorded"
		require.NoError(t, manager.saveCustomerServicePersistentCandidateLifecycle(record))
		tree := filepath.Join(directory, "tree")
		require.NoError(t, os.Mkdir(tree, 0700))
		require.NoError(t, writeDurableFile(filepath.Join(tree, "synthetic-partial-copy.txt"), []byte("synthetic partial media bytes")))
		require.NoError(t, syncDirectory(directory))
	}
}

func TestPostgreSQLCustomerServiceCandidateExecutorForceKillRetainsDurableAllocationIdentityAndFence(t *testing.T) {
	if configurationFile := os.Getenv("LOCAL_BACKUP_CANDIDATE_FORCE_KILL_TEST_CONFIGURATION"); configurationFile != "" {
		// This child really owns the executor's filesystem leases and is killed by
		// its parent without running defers or database cleanup callbacks.
		data, err := os.ReadFile(configurationFile)
		require.NoError(t, err)
		var configuration customerServiceCandidateForceKilledExecutorTestConfiguration
		require.NoError(t, json.Unmarshal(data, &configuration))
		manager, err := NewManager(configuration.Configuration)
		require.NoError(t, err)
		task, err := manager.GetTask(configuration.TaskID)
		require.NoError(t, err)
		lease, err := manager.acquireMaintenance(context.Background(), task)
		require.NoError(t, err)
		defer lease.Close()
		defer manager.Close()
		control, err := configuration.Configuration.connectDatabase(context.Background(), "postgres")
		require.NoError(t, err)
		defer control.Close()
		identityState := configuration.AllocationCheckpoint
		if strings.HasPrefix(identityState, "media_") {
			identityState = "allocation_not_attempted"
		}
		record := createRecordedCustomerServiceCandidateDatabaseForLifecycleRecoveryTest(t, manager, task, control, identityState)
		if strings.HasPrefix(configuration.AllocationCheckpoint, "media_") {
			createPartiallyAllocatedCustomerServiceCandidateMediaVolumeForForceKillCheckpointTest(t, manager, record, configuration.AllocationCheckpoint)
		}
		require.NoError(t, writeDurableFile(configuration.ReadyFilename, []byte("durable allocation checkpoint reached")))
		for {
			time.Sleep(time.Second)
		}
	}
	for _, checkpoint := range []string{"allocated_not_switchable", "allocation_intent", "media_directory_allocated_without_identity", "media_identity_recorded_with_partial_tree"} {
		t.Run(checkpoint, func(t *testing.T) {
			fixture, manager, task, _ := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
			require.NoError(t, manager.Close())
			directory, _ := manager.taskDirectory(task.ID)
			configurationFilename := filepath.Join(directory, "synthetic-candidate-force-killed-executor-configuration.json")
			readyFilename := filepath.Join(directory, "synthetic-candidate-force-killed-executor-ready.txt")
			require.NoError(t, writeDurableJSON(configurationFilename, customerServiceCandidateForceKilledExecutorTestConfiguration{Configuration: fixture.Configuration, TaskID: task.ID, AllocationCheckpoint: checkpoint, ReadyFilename: readyFilename}))
			executable, err := os.Executable()
			require.NoError(t, err)
			command := exec.Command(executable, "-test.run=^TestPostgreSQLCustomerServiceCandidateExecutorForceKillRetainsDurableAllocationIdentityAndFence$", "-test.timeout=1m")
			command.Env = append(os.Environ(), "LOCAL_BACKUP_CANDIDATE_FORCE_KILL_TEST_CONFIGURATION="+configurationFilename)
			var output bytes.Buffer
			command.Stdout, command.Stderr = &output, &output
			require.NoError(t, command.Start())
			finished := make(chan error, 1)
			go func() { finished <- command.Wait() }()
			childReaped := false
			t.Cleanup(func() {
				if !childReaped {
					_ = command.Process.Kill()
					<-finished
				}
			})
			deadline := time.Now().Add(20 * time.Second)
			for {
				if _, err := os.Stat(readyFilename); err == nil {
					break
				}
				select {
				case err := <-finished:
					childReaped = true
					t.Fatalf("candidate child exited before checkpoint: %v: %s", err, output.String())
				default:
				}
				require.True(t, time.Now().Before(deadline), "candidate child did not reach durable checkpoint")
				time.Sleep(20 * time.Millisecond)
			}
			require.NoError(t, command.Process.Kill())
			require.Error(t, <-finished)
			childReaped = true
			restarted, err := NewManager(fixture.Configuration)
			require.NoError(t, err)
			defer restarted.Close()
			record, err := restarted.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			t.Cleanup(func() {
				_, _ = fixture.Control.Exec(`DROP DATABASE IF EXISTS ` + pq.QuoteIdentifier(record.Identity.DatabaseName) + ` WITH (FORCE)`)
			})
			if strings.HasPrefix(checkpoint, "media_") {
				directory, err := customerServiceCandidateMediaVolumeDirectory(record, record.MediaUnion.Volumes[0])
				require.NoError(t, err)
				_, statErr := os.Stat(directory)
				if checkpoint == "media_directory_allocated_without_identity" {
					require.Equal(t, "inspection_required", record.State)
					require.NoError(t, statErr)
					require.Error(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
				} else {
					require.Equal(t, "discarded_not_switchable", record.State)
					require.Equal(t, "discarded_not_switchable", record.MediaUnion.State)
					require.True(t, os.IsNotExist(statErr))
				}
			} else if checkpoint == "allocated_not_switchable" {
				require.Equal(t, "discarded_not_switchable", record.State)
				var count int
				require.NoError(t, fixture.Control.QueryRow(`SELECT count(*) FROM pg_database WHERE oid=$1`, record.Identity.DatabaseOID).Scan(&count))
				require.Zero(t, count)
			} else {
				require.Equal(t, "inspection_required", record.State)
				require.Zero(t, record.Identity.DatabaseOID)
				var count int
				require.NoError(t, fixture.Control.QueryRow(`SELECT count(*) FROM pg_database WHERE datname=$1`, record.Identity.DatabaseName).Scan(&count))
				require.Equal(t, 1, count)
				require.Error(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
			}
			fenced, err := maintenancegate.IsApplicationFenced(fixture.Configuration.MaintenanceDirectory)
			require.NoError(t, err)
			require.True(t, fenced)
			t.Logf("executor force-killed at %s: restart state=%s, maintenance fence retained", checkpoint, record.State)
		})
	}
}
