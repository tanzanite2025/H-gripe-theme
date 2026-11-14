package localbackup

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func TestBackupTaskPersistenceFailureDiagnosticIdentifiesFilesystemOperationAndOSCodeWithoutPrivateValues(t *testing.T) {
	for _, scenario := range []struct {
		name      string
		failure   error
		operation string
		code      string
	}{
		{"open_denied", &os.PathError{Op: "open", Path: "private-workspace/credential.json.next", Err: syscall.Errno(5)}, "open", "5"},
		{"rename_sharing_violation", &os.LinkError{Op: "rename", Old: "private-workspace/task.json.next", New: "private-workspace/task.json", Err: syscall.Errno(32)}, "rename", "32"},
		{"joined_sync_failure", errors.Join(errors.New("private-credential-value"), &os.PathError{Op: "sync", Path: "private-workspace/task.json.next", Err: syscall.Errno(112)}), "sync", "112"},
		{"unknown_error", errors.New("private-credential-value"), "unavailable", "0"},
		{"unrecognized_operation", &os.PathError{Op: "private-credential-value", Path: "private-workspace/task.json.next", Err: syscall.Errno(5)}, "unavailable", "5"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			task := &Task{ID: "0123456789abcdef0123456789abcdef", Phase: "configuration_site_content_releasing", Error: "private-credential-value"}
			var diagnostic bytes.Buffer
			writeBackupTaskPersistenceFailureDiagnosticWithoutPrivatePaths(&diagnostic, task, scenario.failure)
			require.Contains(t, diagnostic.String(), "task="+task.ID)
			require.Contains(t, diagnostic.String(), "phase="+task.Phase)
			require.Contains(t, diagnostic.String(), "filesystem_operation="+scenario.operation)
			require.Contains(t, diagnostic.String(), "os_error_code="+scenario.code+";")
			require.NotContains(t, diagnostic.String(), "private-workspace")
			require.NotContains(t, diagnostic.String(), "private-credential-value")
			require.NotContains(t, diagnostic.String(), "maintenance fence is retained")
		})
	}
}

func TestBackupTaskResultPersistenceFailurePreservesLastDurableRecordAndMaintenanceFence(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	task, err := manager.createTask("upload", "synthetic-operator")
	require.NoError(t, err)
	fencePath := filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename)
	require.NoError(t, os.WriteFile(fencePath, []byte(`{"task_id":"`+task.ID+`"}`), 0600))
	taskDirectory, err := manager.taskDirectory(task.ID)
	require.NoError(t, err)
	taskPath := filepath.Join(taskDirectory, "task.json")
	obstruction := taskPath + ".next"
	diagnostic, err := os.CreateTemp(t.TempDir(), "task-persistence-diagnostic-*.log")
	require.NoError(t, err)
	defer diagnostic.Close()
	originalStderr := os.Stderr
	os.Stderr = diagnostic
	defer func() { os.Stderr = originalStderr }()
	lastDurableRecord := make(chan []byte, 1)
	require.NoError(t, manager.startTaskOperation(task, "synthetic_persistence_failure", func(_ context.Context, _ *Task) error {
		data, readErr := os.ReadFile(taskPath)
		lastDurableRecord <- data
		if readErr != nil {
			return readErr
		}
		if err := os.Mkdir(obstruction, 0700); err != nil {
			return err
		}
		return errors.New("private-credential-value")
	}))
	require.Eventually(t, func() bool {
		if !manager.operationMutex.TryLock() {
			return false
		}
		manager.operationMutex.Unlock()
		return true
	}, 5*time.Second, 10*time.Millisecond)
	os.Stderr = originalStderr
	before := <-lastDurableRecord
	after, err := os.ReadFile(taskPath)
	require.NoError(t, err)
	require.Equal(t, before, after)
	persisted, err := manager.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, "running", persisted.State)
	require.Empty(t, persisted.Error)
	fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
	require.NoError(t, err)
	require.True(t, fenced)
	logBytes, err := os.ReadFile(diagnostic.Name())
	require.NoError(t, err)
	require.Contains(t, string(logBytes), "backup task result could not be persisted")
	require.Contains(t, string(logBytes), "filesystem_operation=open")
	require.NotContains(t, string(logBytes), cfg.Workspace)
	require.NotContains(t, string(logBytes), "private-credential-value")
	// Remove only the injected obstruction. Restart must mark the unacknowledged
	// task interrupted and preserve its fence, rather than claim completion.
	require.NoError(t, os.Remove(obstruction))
	require.NoError(t, manager.Close())
	restarted, err := NewManager(cfg)
	require.NoError(t, err)
	defer restarted.Close()
	persisted, err = restarted.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, "interrupted", persisted.State)
	fenced, err = maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
	require.NoError(t, err)
	require.True(t, fenced)
}
