package localbackup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
)

// Preparation only: no database/media rename, ownership transfer or permission
// to resume a switched database. RecoveryJournal remains a separate contract.
type ProductCatalogStoppedHandoffPreparationJournal struct {
	Version                   int       `json:"version"`
	WindowID                  string    `json:"window_id"`
	TaskID                    string    `json:"task_id"`
	PackageSHA256             string    `json:"package_sha256"`
	ConfigurationSHA256       string    `json:"configuration_sha256"`
	ApplicationBoundarySHA256 string    `json:"application_boundary_sha256"`
	FenceSHA256               string    `json:"fence_sha256,omitempty"`
	Operator                  string    `json:"operator"`
	Phase                     string    `json:"phase"`
	StartedAt                 time.Time `json:"started_at"`
	PreparedAt                time.Time `json:"prepared_at,omitempty"`
	PreviewBindingSHA256      string    `json:"preview_binding_sha256,omitempty"`
	DrainProgressSHA256       string    `json:"drain_progress_sha256,omitempty"`
	CancellationVerifiedAt    time.Time `json:"cancellation_verified_at,omitempty"`
	SHA256                    string    `json:"sha256"`
}

type ProductCatalogStoppedHandoffPreparationRequest struct {
	Confirmation                                    string `json:"confirmation"`
	AllListedWritersStoppedAndUseThisSingleHostGate bool   `json:"all_listed_writers_stopped_and_use_this_single_host_gate"`
}

type ProductCatalogStoppedHandoffCancellationRequest struct {
	Confirmation string `json:"confirmation"`
}

func taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task *Task) bool {
	if taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
		return !taskHasDurablyReleasedOriginalAfterProductCatalogStoppedSwitchRollback(task)
	}
	if task == nil || task.ProductCatalogStoppedHandoffPreparation == nil {
		return false
	}
	journal := task.ProductCatalogStoppedHandoffPreparation
	if journal.Phase != "cancelled" {
		return true
	}
	copy := *journal
	copy.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(copy)
	// A damaged terminal marker must not turn off release/TTL protection. A
	// valid historical cancellation need not match a later reverified preview.
	return err != nil || digest != journal.SHA256 || journal.Version != 1 || journal.TaskID != task.ID || journal.CancellationVerifiedAt.IsZero()
}

func productCatalogStoppedHandoffImmutablePackageSHA256(task *Task) (string, error) {
	return customerServiceRecoveryPlanCanonicalJSONSHA256(struct {
		TaskID           string
		Manifest         *Manifest
		CiphertextSHA256 string
		ExpiresAt        time.Time
	}{task.ID, task.Manifest, task.CiphertextSHA256, task.ExpiresAt})
}

func (manager *Manager) saveProductCatalogStoppedHandoffPreparationJournal(task *Task) error {
	journal := task.ProductCatalogStoppedHandoffPreparation
	journal.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(journal)
	if err != nil {
		return err
	}
	journal.SHA256 = digest
	return manager.saveTask(task)
}

func (manager *Manager) validateProductCatalogStoppedHandoffPreparationJournal(task *Task) error {
	journal := task.ProductCatalogStoppedHandoffPreparation
	if journal == nil {
		return errors.New("stopped product handoff preparation journal is missing")
	}
	copy := *journal
	copy.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(copy)
	packageDigest, packageErr := productCatalogStoppedHandoffImmutablePackageSHA256(task)
	configurationDigest, configurationErr := customerServiceRecoveryPlanCanonicalJSONSHA256(manager.Configuration)
	if err != nil || packageErr != nil || configurationErr != nil || journal.SHA256 != digest || journal.Version != 1 || !safeTaskID.MatchString(journal.WindowID) || journal.TaskID != task.ID || journal.PackageSHA256 != packageDigest || journal.ConfigurationSHA256 != configurationDigest || strings.TrimSpace(journal.Operator) == "" || journal.StartedAt.IsZero() || task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != productCatalogBackupScope || task.Journal != nil || task.WorkCopyExpired {
		return errors.New("stopped product handoff task, package, configuration or journal changed; retain isolation")
	}
	switch journal.Phase {
	case "waiting_for_stop", "preparing":
	case "ready", "cancelling", "cancelled":
		previewDigest, previewErr := customerServiceRecoveryPlanCanonicalJSONSHA256(task.ProductCatalogPreviewBinding)
		if previewErr != nil || task.ProductCatalogPreviewBinding == nil || journal.PreviewBindingSHA256 != previewDigest || task.ProductViewStoppedDrain == nil || task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" || journal.DrainProgressSHA256 != task.ProductViewStoppedDrain.SHA256 || journal.PreparedAt.IsZero() || journal.FenceSHA256 == "" {
			return errors.New("stopped product handoff readiness binding changed; revalidate before proceeding")
		}
		if journal.Phase != "ready" && journal.CancellationVerifiedAt.IsZero() {
			return errors.New("product handoff cancellation has no verified release intent")
		}
	default:
		return errors.New("stopped product handoff preparation phase is invalid")
	}
	return manager.validateProductViewStoppedDrainProgress(task)
}

func (manager *Manager) verifyProductCatalogStoppedHandoffFenceAndRuntimeBoundary(task *Task) error {
	if err := manager.validateProductCatalogStoppedHandoffPreparationJournal(task); err != nil {
		return err
	}
	return manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task)
}

// Callers first validate either the preparation or the distinct switch journal.
func (manager *Manager) verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task *Task) error {
	journal := task.ProductCatalogStoppedHandoffPreparation
	data, err := readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename), 4096)
	if err != nil || maintenancegate.FingerprintExternalProjectionConfiguration(string(data)) != journal.FenceSHA256 {
		return errors.New("stopped product handoff fence changed or disappeared; retain isolation for inspection")
	}
	data, err = os.ReadFile(filepath.Join(manager.Configuration.MaintenanceDirectory, "application-boundary.json"))
	if err != nil || maintenancegate.FingerprintExternalProjectionConfiguration(string(data)) != journal.ApplicationBoundarySHA256 {
		return errors.New("registered runtime boundary changed during stopped product handoff")
	}
	return manager.Configuration.validateRuntimeBoundary()
}

// Callers hold operationMutex, including Close. The fence deliberately survives
// executor exit: a new executor cannot inherit this process's readiness/lease.
func (manager *Manager) closeProductCatalogStoppedHandoffLease() error {
	lease := manager.productCatalogStoppedHandoffLease
	manager.productCatalogStoppedHandoffLease = nil
	manager.productCatalogStoppedHandoffLeaseTaskID = ""
	if lease == nil {
		return nil
	}
	return lease.Close()
}

func (manager *Manager) acquireOrRetainProductCatalogStoppedHandoffLease(ctx context.Context, task *Task) error {
	journal := task.ProductCatalogStoppedHandoffPreparation
	if journal.FenceSHA256 != "" {
		if err := manager.verifyProductCatalogStoppedHandoffFenceAndRuntimeBoundary(task); err != nil {
			return err
		}
	}
	if manager.productCatalogStoppedHandoffLease != nil {
		if manager.productCatalogStoppedHandoffLeaseTaskID != task.ID {
			return errors.New("another product handoff holds the stopped application lease")
		}
		return nil
	}
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	manager.productCatalogStoppedHandoffLease, manager.productCatalogStoppedHandoffLeaseTaskID = lease, task.ID
	data, err := readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename), 4096)
	if err != nil {
		return err
	}
	journal.FenceSHA256 = maintenancegate.FingerprintExternalProjectionConfiguration(string(data))
	if err := manager.saveProductCatalogStoppedHandoffPreparationJournal(task); err != nil {
		return err
	}
	return manager.verifyProductCatalogStoppedHandoffFenceAndRuntimeBoundary(task)
}

func (manager *Manager) StartProductCatalogContinuouslyStoppedHandoffPreparation(id, actor string, request ProductCatalogStoppedHandoffPreparationRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	cfg := manager.Configuration
	if request.Confirmation != "PREPARE STOPPED PRODUCT HANDOFF "+cfg.SiteID+"/"+cfg.Environment+"/"+id || !request.AllListedWritersStoppedAndUseThisSingleHostGate || strings.TrimSpace(actor) == "" {
		return nil, errors.New("explicit stopped product handoff confirmation and complete single-host writer attestation are required")
	}
	if cfg.ProductViewRecoveryRedis == nil || cfg.ProductViewRecoveryRedis.validateExplicitProductViewRecoveryRedisConnection() != nil || len(cfg.ProductViewRecoveryWriterInventory) == 0 {
		return nil, errors.New("configure Redis ownership and the complete single-host writer inventory first")
	}
	for _, writer := range cfg.ProductViewRecoveryWriterInventory {
		if strings.TrimSpace(writer) == "" {
			return nil, errors.New("writer inventory has an empty entry")
		}
	}
	if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
		if err := manager.validateProductCatalogStoppedHandoffPreparationJournal(task); err != nil {
			return nil, err
		}
		if task.State != "failed" && task.State != "interrupted" && task.State != "product_module_stopped_handoff_prepared" {
			return nil, errors.New("stopped handoff preparation is already active")
		}
		if task.ProductCatalogStoppedHandoffPreparation.Phase == "cancelling" {
			return nil, errors.New("finish verified handoff cancellation instead of preparing again")
		}
	} else {
		if err := validateFreshProductCatalogCandidateTaskForHandoffInspection(task); err != nil {
			return nil, err
		}
		if fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory); err != nil || fenced {
			return nil, errors.New("resolve the existing maintenance window before a new stopped handoff preparation")
		}
	}
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	err = manager.startTaskOperation(task, "preparing_product_handoff_with_continuous_isolation", func(ctx context.Context, task *Task) error {
		if !taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
			packageDigest, err := productCatalogStoppedHandoffImmutablePackageSHA256(task)
			if err != nil {
				return err
			}
			configurationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(cfg)
			if err != nil {
				return err
			}
			boundary, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json"))
			if err != nil {
				return err
			}
			task.ProductCatalogStoppedHandoffPreparation = &ProductCatalogStoppedHandoffPreparationJournal{Version: 1, WindowID: newTaskID(), TaskID: task.ID, PackageSHA256: packageDigest, ConfigurationSHA256: configurationDigest, ApplicationBoundarySHA256: maintenancegate.FingerprintExternalProjectionConfiguration(string(boundary)), Operator: actor, Phase: "waiting_for_stop", StartedAt: time.Now().UTC()}
			if err := manager.saveProductCatalogStoppedHandoffPreparationJournal(task); err != nil {
				return err
			}
		}
		if err := manager.acquireOrRetainProductCatalogStoppedHandoffLease(ctx, task); err != nil {
			return err
		}
		journal := task.ProductCatalogStoppedHandoffPreparation
		// Persist invalidation before touching Redis, SQL, or the old candidate.
		journal.Phase, journal.Operator = "preparing", actor
		journal.PreparedAt, journal.CancellationVerifiedAt = time.Time{}, time.Time{}
		journal.PreviewBindingSHA256, journal.DrainProgressSHA256 = "", ""
		if err := manager.saveProductCatalogStoppedHandoffPreparationJournal(task); err != nil {
			return err
		}
		if err := manager.drainProductViewsAndRebuildCandidateWithHeldMaintenanceLease(ctx, task, actor); err != nil {
			return err
		}
		if err := manager.verifyProductCatalogStoppedHandoffFenceAndRuntimeBoundary(task); err != nil {
			return err
		}
		if err := manager.validateProductCatalogCandidateStillMatchesIsolatedPreview(ctx, task); err != nil {
			return err
		}
		if err := manager.verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx, task); err != nil {
			return err
		}
		journal.PreviewBindingSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(task.ProductCatalogPreviewBinding)
		if err != nil {
			return err
		}
		journal.DrainProgressSHA256 = task.ProductViewStoppedDrain.SHA256
		journal.Phase, journal.PreparedAt = "ready", time.Now().UTC()
		task.State, task.Phase = "product_module_stopped_handoff_prepared", "product_handoff_prepared_application_remains_stopped"
		return manager.saveProductCatalogStoppedHandoffPreparationJournal(task)
	})
	if err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

func (manager *Manager) verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx context.Context, task *Task) error {
	if manager.productCatalogStoppedHandoffLease == nil || manager.productCatalogStoppedHandoffLeaseTaskID != task.ID || !manager.productCatalogStoppedHandoffLease.Locked() {
		return errors.New("stopped product handoff requires this executor's continuously held application lease")
	}
	client, ownership, err := manager.connectAndVerifyRuntimeOwnedProductViewRedis(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	if task.ProductViewStoppedDrain == nil || maintenancegate.FingerprintExternalProjectionConfiguration(ownership) != task.ProductViewStoppedDrain.OwnershipSHA256 {
		return errors.New("product view Redis ownership changed before handoff verification")
	}
	if err := requireNoOtherProductViewRedisLogicalDatabaseClients(ctx, client, manager.Configuration.ProductViewRecoveryRedis.Database); err != nil {
		return err
	}
	remaining, err := client.Exists(ctx, "product:view_counts", "product:view_counts:processing", "product:view_counts:processing_batch_id").Result()
	if err != nil || remaining != 0 {
		return errors.New("product view buffers changed; retry preparation before cancellation or handoff")
	}
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer control.Close()
	return assertNoUnfencedDatabaseSessions(ctx, control, manager.Configuration.DatabaseName)
}

func (manager *Manager) StartVerifiedCancellationOfProductCatalogStoppedHandoffPreparation(id, actor string, request ProductCatalogStoppedHandoffCancellationRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	cfg := manager.Configuration
	if request.Confirmation != "CANCEL STOPPED PRODUCT HANDOFF "+cfg.SiteID+"/"+cfg.Environment+"/"+id || strings.TrimSpace(actor) == "" {
		return nil, errors.New("explicit stopped product handoff cancellation confirmation is required")
	}
	if err := manager.validateProductCatalogStoppedHandoffPreparationJournal(task); err != nil {
		return nil, err
	}
	journal := task.ProductCatalogStoppedHandoffPreparation
	if journal.Phase != "ready" && journal.Phase != "cancelling" {
		return nil, errors.New("finish or retry the drain and candidate preparation before cancellation")
	}
	if task.State != "product_module_stopped_handoff_prepared" && task.State != "failed" && task.State != "interrupted" {
		return nil, errors.New("stopped handoff preparation is active or already cancelled")
	}
	err = manager.startTaskOperation(task, "verifying_original_before_product_handoff_cancellation", func(ctx context.Context, task *Task) error {
		journal := task.ProductCatalogStoppedHandoffPreparation
		fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
		if err != nil {
			return err
		}
		// The durable intent precedes fence removal. Only this exact interrupted
		// cancellation may finish when removal already happened before restart.
		if journal.Phase != "cancelling" || fenced {
			if err := manager.acquireOrRetainProductCatalogStoppedHandoffLease(ctx, task); err != nil {
				return err
			}
			if err := manager.validateProductCatalogCandidateStillMatchesIsolatedPreview(ctx, task); err != nil {
				return err
			}
			if err := manager.verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx, task); err != nil {
				return err
			}
			if err := manager.verifyProductCatalogStoppedHandoffFenceAndRuntimeBoundary(task); err != nil {
				return err
			}
			journal.Phase, journal.CancellationVerifiedAt = "cancelling", time.Now().UTC()
			if err := manager.saveProductCatalogStoppedHandoffPreparationJournal(task); err != nil {
				return err
			}
			if err := manager.releaseMaintenanceWithVerifiedProductHandoffCancellation(task, true); err != nil {
				return err
			}
		}
		if err := manager.closeProductCatalogStoppedHandoffLease(); err != nil {
			return err
		}
		journal.Phase = "cancelled"
		task.State, task.Phase, task.Error = "product_module_isolated_verified", "product_handoff_cancelled_original_service_released", ""
		return manager.saveProductCatalogStoppedHandoffPreparationJournal(task)
	})
	if err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}
