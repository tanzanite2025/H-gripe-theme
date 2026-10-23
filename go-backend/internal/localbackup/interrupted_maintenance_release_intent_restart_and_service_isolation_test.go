package localbackup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func TestRestartCompletesPersistedReleaseOnlyAfterMaintenanceFenceWasReleased(t *testing.T) {
	for _, state := range []string{"resuming", "releasing_rollback"} {
		for _, fenced := range []bool{false, true} {
			name := state + "/released"
			if fenced {
				name = state + "/still_fenced"
			}
			t.Run(name, func(t *testing.T) {
				cfg := createIndependentRecoveryTestConfiguration(t)
				manager, err := NewManager(cfg)
				require.NoError(t, err)
				task, err := manager.createTask("upload", "synthetic-operator")
				require.NoError(t, err)
				task.State = state
				require.NoError(t, manager.saveTask(task))
				if fenced {
					require.NoError(t, os.WriteFile(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename), []byte(`{"task_id":"`+task.ID+`"}`), 0600))
				}
				require.NoError(t, manager.Close())
				manager, err = NewManager(cfg)
				require.NoError(t, err)
				defer manager.Close()
				persisted, err := manager.GetTask(task.ID)
				require.NoError(t, err)
				expected := "completed"
				if state == "releasing_rollback" {
					expected = "rolled_back"
				}
				if fenced {
					expected = "awaiting_reconciliation"
					if state == "releasing_rollback" {
						expected = "interrupted"
					}
				}
				require.Equal(t, expected, persisted.State)
				actualFence, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
				require.NoError(t, err)
				require.Equal(t, fenced, actualFence)
			})
		}
	}
}
