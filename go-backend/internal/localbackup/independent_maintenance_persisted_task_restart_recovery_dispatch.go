package localbackup

import (
	"context"
	"fmt"
	"os"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
)

// Retain the original restart branch order and cleanup sequence. Process-lease cleanup belongs to the constructor on failure.
// See backup-recovery-lifecycle-supported-entrypoints-internal-rehearsals-and-persisted-task-compatibility.zh-cn.md.
// Transactions recover receipts; stopped product switches retain explicit rollback;
// internal paired rehearsals roll back before older candidates can be discarded.
// Unexported/test-started construction methods still have live recovery consumers.
func (manager *Manager) recoverPersistedMaintenanceTasksAfterExecutorRestart() error {
	cfg := manager.Configuration
	tasks, err := manager.ListTasks()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if taskHasAnyTransactionalBusinessModuleRestoreProgress(task) {
			if err := manager.retainInterruptedTransactionalBusinessModuleRestoreAfterRestart(task); err != nil {
				return err
			}
			continue
		}
		if task.SelectedModuleDisposableStage != nil {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cleanupErr := manager.cleanupSelectedBusinessModuleDisposableStageByPersistedIdentity(cleanupCtx, task)
			cancel()
			if cleanupErr == nil {
				cleanupErr = manager.cleanupSelectedBusinessModuleDisposableWorkspaceDirectories(task)
			}
			if cleanupErr != nil {
				task.State, task.Phase, task.Error = "interrupted", "selected_module_disposable_cleanup_requires_inspection", cleanupErr.Error()
				if err := manager.saveTask(task); err != nil {
					return err
				}
			}
		}
		if taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
			if err := manager.retainStoppedProductCatalogSwitchForExplicitRollbackAfterRestart(task); err != nil {
				return err
			}
			continue
		}
		if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
			task.State, task.Phase = "interrupted", "product_stopped_handoff_requires_explicit_revalidation"
			task.Error = "stopped product handoff preparation retains isolation; explicitly revalidate or cancel, never reuse readiness after restart"
			if err := manager.saveTask(task); err != nil {
				return err
			}
			continue
		}
		if task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
			task.State, task.Phase = "interrupted", "product_view_drain_retained_for_explicit_retry"
			task.ProductCatalogRecovery = nil
			task.Error = "stopped product view drain is unresolved; retain maintenance and explicitly retry after inspecting ownership and writers"
			if err := manager.saveTask(task); err != nil {
				return err
			}
		}
		if task.State == "resuming" || task.State == "releasing_rollback" {
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			if err != nil {
				return err
			}
			if !fenced {
				if taskHasProductCatalogPairedRecoveryJournal(task) {
					if err := manager.validateProductCatalogPairedRehearsalBinding(task); err != nil || !task.Journal.ProductCatalog.RollbackVerified {
						task.State, task.Phase, task.Error = "interrupted", "product_paired_rehearsal_retained_for_inspection", "product release intent has no valid verified rollback; inspection required"
						if err := manager.saveTask(task); err != nil {
							return err
						}
						continue
					}
				}
				if task.State == "resuming" {
					task.State = "completed"
				} else {
					task.State = "rolled_back"
				}
				task.Phase = "durable_release_intent_completed_after_executor_restart"
			} else if task.State == "resuming" {
				task.State = "awaiting_reconciliation"
			} else {
				task.State = "interrupted"
			}
			if err := manager.saveTask(task); err != nil {
				return err
			}
		}
		if task.State == "running" || task.State == "queued" {
			task.State = "interrupted"
			task.Error = "executor stopped; inspect persisted phase and recovery journal before continuing"
			if err := manager.saveTask(task); err != nil {
				return err
			}
		}
	}
	manager.recoverInterruptedProductCatalogPairedRehearsals(tasks)
	manager.recoverInterruptedCustomerServiceIndependentPairedSwitches(tasks)
	manager.discardInterruptedCustomerServicePersistentCandidates(tasks)
	cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), 30*time.Second)
	cleanupErr := manager.discardInterruptedCustomerServiceDisposableRehearsals(cleanupCtx)
	cancelCleanup()
	if cleanupErr != nil {
		fmt.Fprintln(os.Stderr, "interrupted disposable customer-service candidate identity is retained for inspection")
	}
	return nil
}
