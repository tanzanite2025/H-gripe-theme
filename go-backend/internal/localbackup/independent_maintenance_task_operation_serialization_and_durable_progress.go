package localbackup

import (
	"context"
	"errors"
	"fmt"
	"os"
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
				fmt.Fprintln(os.Stderr, "backup task result could not be persisted; maintenance fence is retained")
			}
		}
	}()
	return nil
}
