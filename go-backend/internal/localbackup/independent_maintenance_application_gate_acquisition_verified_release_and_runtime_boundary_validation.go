package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/gofrs/flock"

	"commerce-platform/internal/pkg/maintenancegate"
)

// Own maintenance isolation and verified release; domain journals continue to decide whether release is safe.
func (manager *Manager) acquireMaintenance(ctx context.Context, task *Task) (*flock.Flock, error) {
	if manager.productCatalogStoppedHandoffLease != nil {
		return nil, errors.New("a stopped product handoff already holds the application lease; use its preparation or cancellation operation")
	}
	if err := manager.requireNoUnresolvedProductCatalogPairedRehearsalExceptTask(task.ID); err != nil {
		return nil, err
	}
	filename := filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename)
	if data, err := os.ReadFile(filename); err == nil { // #nosec G304 -- The path is a fixed record under the configured private maintenance root or validated task directory.
		var owner struct {
			TaskID string `json:"task_id"`
		}
		if json.Unmarshal(data, &owner) != nil || owner.TaskID != task.ID {
			return nil, errors.New("another maintenance window is active")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	} else {
		if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) && task.ProductCatalogStoppedHandoffPreparation.FenceSHA256 != "" {
			return nil, errors.New("stopped product handoff fence is missing; retain preparation for inspection")
		}
		if task.Journal != nil {
			return nil, errors.New("paired recovery fence is missing; verify the environment before any further recovery action")
		}
		if task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" && task.ProductViewStoppedDrain.Phase != "waiting_for_stop" {
			return nil, errors.New("stopped product view drain fence is missing; retain progress for inspection")
		}
		if err := writeDurableJSON(filename, struct {
			TaskID    string    `json:"task_id"`
			StartedAt time.Time `json:"started_at"`
		}{task.ID, time.Now().UTC()}); err != nil {
			return nil, err
		}
	}
	wait, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	return maintenancegate.AcquireExclusiveRecoveryLease(wait, manager.Configuration.MaintenanceDirectory)
}

func (manager *Manager) releaseMaintenance(task *Task) error {
	return manager.releaseMaintenanceWithVerifiedProductHandoffCancellation(task, false)
}

func (manager *Manager) releaseMaintenanceWithVerifiedProductHandoffCancellation(task *Task, cancellationVerified bool) error {
	if taskHasAnyTransactionalBusinessModuleRestoreProgress(task) {
		return errors.New("module restore requires its dedicated commit and attachment recovery before service release")
	}
	if taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
		return errors.New("stopped product switch requires its dedicated verified paired rollback before original service release")
	}
	if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
		if !cancellationVerified || task.ProductCatalogStoppedHandoffPreparation.Phase != "canceling" {
			return errors.New("stopped product handoff preparation requires its verified cancellation; ordinary release is forbidden")
		}
		if err := manager.validateProductCatalogStoppedHandoffPreparationJournal(task); err != nil {
			return err
		}
	}
	if err := manager.validateProductViewStoppedDrainProgress(task); err != nil {
		return err
	}
	if task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
		return errors.New("product view drain or candidate rebuild is unresolved; maintenance isolation must remain")
	}
	if taskHasProductCatalogPairedRecoveryJournal(task) {
		if err := manager.validateProductCatalogPairedRehearsalBinding(task); err != nil {
			return err
		}
		if !task.Journal.ProductCatalog.RollbackVerified || (task.State != "releasing_rollback" && task.State != "rolled_back") {
			return errors.New("product paired rehearsal must verify rollback before releasing maintenance")
		}
	}
	if err := manager.requireCustomerServiceIndependentPairedSwitchRolledBack(task.ID); err != nil {
		return err
	}
	filename := filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename)
	data, err := os.ReadFile(filename) // #nosec G304 -- The path is a fixed record under the configured private maintenance root or validated task directory.
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var owner struct {
		TaskID string `json:"task_id"`
	}
	if json.Unmarshal(data, &owner) != nil || owner.TaskID != task.ID {
		return errors.New("maintenance fence owner mismatch")
	}
	if err := os.Remove(filename); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(filename))
}

// Ordinary release is not a universal recovery operation. Each transactional,
// stopped-product or older candidate lifecycle must first resolve its own journal.
func (manager *Manager) ReleaseUnmodifiedMaintenance(id string) error {
	if !manager.operationMutex.TryLock() {
		return errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return err
	}
	if taskHasAnyTransactionalBusinessModuleRestoreProgress(task) {
		return errors.New("module restore journal exists; use its dedicated recovery")
	}
	if task.Journal != nil {
		return errors.New("live switch journal exists; rollback or complete reconciliation is required")
	}
	if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
		return errors.New("cancel the stopped product handoff preparation through its dedicated verification")
	}
	if task.State == "running" || task.State == "queued" {
		return errors.New("task is still active")
	}
	if err := manager.requireCustomerServicePersistentCandidateDiscarded(id); err != nil {
		return err
	}
	if err := manager.validateProductViewStoppedDrainProgress(task); err != nil {
		return err
	}
	if task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
		return errors.New("resume the unresolved product view drain before releasing maintenance")
	}
	if err := manager.cleanupUnswitchedPreparedMedia(task); err != nil {
		return err
	}
	return manager.releaseMaintenance(task)
}

func (cfg Configuration) validateRuntimeBoundary() error {
	data, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json"))
	if err != nil {
		return errors.New("application maintenance protocol has not been registered; upgrade and start the gated application first")
	}
	var boundary maintenancegate.ApplicationBoundary
	if json.Unmarshal(data, &boundary) != nil || boundary.ValidateRegisteredRecoveryProjectionProtocol() != nil || boundary.Database != cfg.DatabaseName || boundary.ApplicationVersion != cfg.ApplicationVersion || !reflect.DeepEqual(boundary.StorageRoots, cfg.StorageRoots) {
		return errors.New("configured database, application version or media roots do not match the gated application")
	}
	for name, secret := range cfg.RecoverySecrets {
		digest := sha256.Sum256([]byte(secret))
		if boundary.RecoverySecretSHA256[name] != hex.EncodeToString(digest[:]) {
			return errors.New("application recovery keys differ from independently configured recovery materials")
		}
	}
	if len(boundary.RecoverySecretSHA256) != len(cfg.RecoverySecrets) {
		return errors.New("independent recovery materials omit additional application encryption keys")
	}
	if strings.TrimSpace(cfg.ApplicationVersion) == "dev" {
		return errors.New("explicit application build identity is required for recoverable packages")
	}
	return nil
}
