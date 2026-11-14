package localbackup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
)

// Serialize background operations and persist their original running/error transitions.
func (manager *Manager) startTaskOperation(task *Task, phase string, operation func(context.Context, *Task) error) error {
	if !manager.operationMutex.TryLock() {
		return errors.New("another backup or restore operation is running")
	}
	if err := manager.requireNoUnresolvedProductCatalogPairedRehearsalExceptTask(task.ID); err != nil {
		manager.operationMutex.Unlock()
		return err
	}
	if err := manager.requireCustomerServicePersistentCandidateDiscarded(task.ID); err != nil {
		manager.operationMutex.Unlock()
		return err
	}
	latest, err := manager.GetTask(task.ID)
	if err != nil || !latest.UpdatedAt.Equal(task.UpdatedAt) || latest.State != task.State {
		manager.operationMutex.Unlock()
		return errors.New("task changed before operation; refresh and retry")
	}
	task.State = "running"
	task.Phase = phase
	task.Error = ""
	if err := manager.saveTask(task); err != nil {
		manager.operationMutex.Unlock()
		return err
	}
	go func() {
		defer manager.operationMutex.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
		defer cancel()
		if err := operation(ctx, task); err != nil {
			task.State = "failed"
			task.Error = err.Error()
			if saveErr := manager.saveTask(task); saveErr != nil {
				_ = writeBackupTaskPersistenceFailureDiagnosticWithoutPrivatePaths(os.Stderr, task, saveErr)
			}
		}
	}()
	return nil
}

// Preserve the OS error code needed to diagnose durable writes without logging
// workspace paths, task errors, credentials or filesystem error messages. A
// release may already have removed the fence, so do not claim it is retained.
func writeBackupTaskPersistenceFailureDiagnosticWithoutPrivatePaths(destination io.Writer, task *Task, persistenceError error) error {
	filesystemOperation := "unavailable"
	var pathError *os.PathError
	var linkError *os.LinkError
	if errors.As(persistenceError, &pathError) {
		filesystemOperation = pathError.Op
	} else if errors.As(persistenceError, &linkError) {
		filesystemOperation = linkError.Op
	}
	switch filesystemOperation {
	case "open", "write", "sync", "close", "rename":
	default:
		filesystemOperation = "unavailable"
	}
	var osErrorCode syscall.Errno
	_ = errors.As(persistenceError, &osErrorCode)
	_, err := fmt.Fprintf(destination, "backup task result could not be persisted; task=%s phase=%s persistence_error_type=%T filesystem_operation=%s os_error_code=%d; verify maintenance isolation before continuing\n", task.ID, task.Phase, persistenceError, filesystemOperation, uint64(osErrorCode))
	return err
}
