package localbackup

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type customerServicePairedSwitchForceKilledExecutorConfiguration struct {
	Configuration Configuration                                      `json:"configuration"`
	TaskID        string                                             `json:"task_id"`
	Request       CustomerServiceEffectiveSelectionInspectionRequest `json:"request"`
	Checkpoint    string                                             `json:"checkpoint"`
	ReadyFilename string                                             `json:"ready_filename"`
}

func TestPostgreSQLCustomerServiceIndependentPairedSwitchForceKilledExecutorRollsBackBeforeDisposalWithoutReleasingFence(t *testing.T) {
	if configurationFilename := os.Getenv("LOCAL_BACKUP_CUSTOMER_SERVICE_PAIRED_SWITCH_FORCE_KILL_CONFIGURATION"); configurationFilename != "" {
		data, err := readCustomerServicePairedRecoveryBoundedRegularFile(configurationFilename, 4<<20)
		require.NoError(t, err)
		var configuration customerServicePairedSwitchForceKilledExecutorConfiguration
		require.NoError(t, decodeStrictJSON(data, &configuration))
		manager, err := NewManager(configuration.Configuration)
		require.NoError(t, err)
		defer manager.Close()
		ctx := context.Background()
		record, err := manager.prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx, configuration.TaskID, configuration.Request)
		require.NoError(t, err)
		journal, err := manager.prepareCustomerServiceIndependentPairedSwitchRehearsalJournal(ctx, configuration.TaskID, record.SHA256)
		require.NoError(t, err)
		task, err := manager.GetTask(configuration.TaskID)
		require.NoError(t, err)
		lease, err := manager.acquireCustomerServicePairedSwitchExistingMaintenanceLease(ctx, task)
		require.NoError(t, err)
		defer lease.Close()
		control, err := configuration.Configuration.connectDatabase(ctx, "postgres")
		require.NoError(t, err)
		defer control.Close()
		createCustomerServiceIndependentPairedSwitchInterruptionCheckpoint(t, manager, control, record, journal, configuration.Checkpoint)
		require.NoError(t, writeDurableFile(configuration.ReadyFilename, []byte("paired rename checkpoint is durable")))
		for {
			time.Sleep(time.Second)
		}
	}
	for _, checkpoint := range []string{"candidate_move_rename_without_ack", "candidate_database_rename_without_ack", "reverse_move_rename_without_ack"} {
		t.Run(checkpoint, func(t *testing.T) {
			fixture, manager, task, request := createCustomerServicePersistentCandidateLifecycleTestFixture(t)
			require.NoError(t, manager.Close())
			directory, _ := manager.taskDirectory(task.ID)
			configurationFilename := filepath.Join(directory, "synthetic-paired-switch-force-killed-executor-configuration.json")
			readyFilename := filepath.Join(directory, "synthetic-paired-switch-force-killed-executor-ready.txt")
			require.NoError(t, writeDurableJSON(configurationFilename, customerServicePairedSwitchForceKilledExecutorConfiguration{fixture.Configuration, task.ID, request, checkpoint, readyFilename}))
			executable, err := os.Executable()
			require.NoError(t, err)
			command := exec.CommandContext(context.Background(), executable, "-test.run=^TestPostgreSQLCustomerServiceIndependentPairedSwitchForceKilledExecutorRollsBackBeforeDisposalWithoutReleasingFence$", "-test.timeout=3m")
			command.Env = append(os.Environ(), "LOCAL_BACKUP_CUSTOMER_SERVICE_PAIRED_SWITCH_FORCE_KILL_CONFIGURATION="+configurationFilename)
			var output bytes.Buffer
			command.Stdout, command.Stderr = &output, &output
			require.NoError(t, command.Start())
			finished := make(chan error, 1)
			go func() { finished <- command.Wait() }()
			reaped := false
			t.Cleanup(func() {
				if !reaped {
					_ = command.Process.Kill()
					<-finished
				}
			})
			deadline := time.Now().Add(45 * time.Second)
			for {
				if _, err := os.Stat(readyFilename); err == nil {
					break
				}
				select {
				case err := <-finished:
					reaped = true
					t.Fatalf("paired executor exited before checkpoint: %v: %s", err, output.String())
				default:
				}
				require.True(t, time.Now().Before(deadline), "paired executor failed to reach its durable checkpoint")
				time.Sleep(20 * time.Millisecond)
			}
			require.NoError(t, command.Process.Kill())
			require.Error(t, <-finished)
			reaped = true
			restarted, err := NewManager(fixture.Configuration)
			require.NoError(t, err)
			defer restarted.Close()
			journal, err := restarted.loadCustomerServiceIndependentPairedSwitchJournal(task.ID)
			require.NoError(t, err)
			require.Equal(t, "rolled_back_isolated", journal.State)
			record, err := restarted.loadCustomerServicePersistentCandidateLifecycle(task.ID)
			require.NoError(t, err)
			require.Equal(t, "discarded_not_switchable", record.State)
			verifyCustomerServicePairedSwitchRestoredOriginalAndPreservedStage(t, fixture, restarted, task, record)
			require.NoError(t, restarted.ReleaseUnmodifiedMaintenance(task.ID))
			t.Logf("force-killed executor at %s: original database/media restored before candidate disposal; original stage and fence retained", checkpoint)
		})
	}
}
