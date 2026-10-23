package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/lib/pq"
)

// This is an operator-requested isolated switch, independent of internal
// rehearsals. The candidate retains its isolated owner; no service-resume API
// accepts this journal or its read-only post-switch inspection as approval.
type ProductCatalogContinuouslyStoppedSwitchJournal struct {
	Version                   int                            `json:"version"`
	TaskID                    string                         `json:"task_id"`
	WindowID                  string                         `json:"window_id"`
	PreparationSHA256         string                         `json:"preparation_sha256"`
	PackageSHA256             string                         `json:"package_sha256"`
	PreviewBindingSHA256      string                         `json:"preview_binding_sha256"`
	Operator                  string                         `json:"operator"`
	StartedAt                 time.Time                      `json:"started_at"`
	Phase                     string                         `json:"phase"`
	Moves                     []moduleRecoveryMediaEntryMove `json:"moves"`
	RollbackVerifiedAt        time.Time                      `json:"rollback_verified_at,omitempty"`
	OriginalServiceReleasedAt time.Time                      `json:"original_service_released_at,omitempty"`
	SHA256                    string                         `json:"sha256"`
}

type ProductCatalogStoppedSwitchConfirmationRequest struct {
	Confirmation string `json:"confirmation"`
}

type ProductCatalogStoppedPostSwitchInspection struct {
	Version                   int                                           `json:"version"`
	Scope                     string                                        `json:"scope"`
	Observation               string                                        `json:"observation"`
	TaskID                    string                                        `json:"task_id"`
	WindowID                  string                                        `json:"window_id"`
	PackageID                 string                                        `json:"package_id"`
	PackageSHA256             string                                        `json:"package_sha256"`
	SwitchJournalSHA256       string                                        `json:"switch_journal_sha256"`
	PreviewBindingSHA256      string                                        `json:"preview_binding_sha256"`
	ConfigurationSHA256       string                                        `json:"configuration_sha256"`
	ApplicationBoundarySHA256 string                                        `json:"application_boundary_sha256"`
	FenceSHA256               string                                        `json:"fence_sha256"`
	RedisOwnershipSHA256      string                                        `json:"redis_ownership_sha256"`
	DatabaseName              string                                        `json:"database_name"`
	DatabaseOID               int64                                         `json:"database_oid"`
	DatabaseFingerprint       string                                        `json:"database_fingerprint"`
	CompleteMediaSHA256       string                                        `json:"complete_media_sha256"`
	ApplicationRemainsStopped bool                                          `json:"application_remains_stopped"`
	CandidateRemainsIsolated  bool                                          `json:"candidate_remains_isolated"`
	OnlineOverwriteAvailable  bool                                          `json:"online_overwrite_available"`
	ExternalActionsExecuted   bool                                          `json:"external_actions_executed"`
	ProjectionRequirements    []ExternalProjectionReconciliationRequirement `json:"projection_requirements"`
	ContractSHA256            string                                        `json:"contract_sha256"`
	CapturedAt                time.Time                                     `json:"captured_at"`
}

func taskHasProductCatalogContinuouslyStoppedSwitchJournal(task *Task) bool {
	return task != nil && task.Journal != nil && task.Journal.ProductCatalogStoppedSwitch != nil
}

func productCatalogStoppedSwitchCompleteJournalSHA256(task *Task) (string, error) {
	if !taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
		return "", errors.New("stopped product switch journal is missing")
	}
	outer := *task.Journal
	inner := *outer.ProductCatalogStoppedSwitch
	inner.SHA256 = ""
	outer.ProductCatalogStoppedSwitch = &inner
	return customerServiceRecoveryPlanCanonicalJSONSHA256(outer)
}

func taskHasDurablyReleasedOriginalAfterProductCatalogStoppedSwitchRollback(task *Task) bool {
	if !taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
		return false
	}
	journal := task.Journal.ProductCatalogStoppedSwitch
	digest, err := productCatalogStoppedSwitchCompleteJournalSHA256(task)
	if err != nil || journal.SHA256 != digest || journal.Version != 1 || journal.TaskID != task.ID || journal.Phase != "original_released" || journal.RollbackVerifiedAt.IsZero() || journal.OriginalServiceReleasedAt.IsZero() || task.ProductCatalogStoppedHandoffPreparation == nil || task.ProductCatalogPreviewBinding == nil || task.ProductViewStoppedDrain == nil {
		return false
	}
	preparation := *task.ProductCatalogStoppedHandoffPreparation
	preparation.SHA256 = ""
	preparationDigest, preparationErr := customerServiceRecoveryPlanCanonicalJSONSHA256(preparation)
	packageDigest, packageErr := productCatalogStoppedHandoffImmutablePackageSHA256(task)
	previewDigest, previewErr := customerServiceRecoveryPlanCanonicalJSONSHA256(task.ProductCatalogPreviewBinding)
	drain := *task.ProductViewStoppedDrain
	drain.SHA256 = ""
	drainDigest, drainErr := customerServiceRecoveryPlanCanonicalJSONSHA256(drain)
	return preparationErr == nil && packageErr == nil && previewErr == nil && drainErr == nil && preparationDigest == task.ProductCatalogStoppedHandoffPreparation.SHA256 && preparationDigest == journal.PreparationSHA256 && preparation.WindowID == journal.WindowID && preparation.TaskID == task.ID && packageDigest == journal.PackageSHA256 && previewDigest == journal.PreviewBindingSHA256 && drainDigest == preparation.DrainProgressSHA256
}

func (manager *Manager) saveProductCatalogStoppedSwitchProgress(task *Task) error {
	digest, err := productCatalogStoppedSwitchCompleteJournalSHA256(task)
	if err != nil {
		return err
	}
	task.Journal.ProductCatalogStoppedSwitch.SHA256 = digest
	return manager.saveTask(task)
}

func (manager *Manager) validateProductCatalogStoppedSwitchJournalBinding(task *Task) error {
	if !taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) || !safeTaskID.MatchString(task.ID) || task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != productCatalogBackupScope || task.Journal.ProductCatalog != nil || len(task.Journal.Moves) != 0 || task.ProductCatalogPreviewBinding == nil || task.ProductCatalogStoppedHandoffPreparation == nil {
		return errors.New("independent stopped product switch binding is missing or mixed with another recovery scope")
	}
	journal, preparation, binding := task.Journal.ProductCatalogStoppedSwitch, task.ProductCatalogStoppedHandoffPreparation, task.ProductCatalogPreviewBinding
	digest, err := productCatalogStoppedSwitchCompleteJournalSHA256(task)
	if err != nil || digest != journal.SHA256 || journal.Version != 1 || journal.TaskID != task.ID || !safeTaskID.MatchString(journal.WindowID) || journal.WindowID != preparation.WindowID || strings.TrimSpace(journal.Operator) == "" || journal.StartedAt.IsZero() {
		return errors.New("stopped product switch journal identity or digest changed")
	}
	copy := *preparation
	copy.SHA256 = ""
	preparationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(copy)
	if err != nil || preparationDigest != preparation.SHA256 || preparation.SHA256 != journal.PreparationSHA256 || preparation.Version != 1 || preparation.TaskID != task.ID || preparation.Phase != "ready" || preparation.PreparedAt.IsZero() || preparation.FenceSHA256 == "" {
		return errors.New("stopped product switch lost its exact ready preparation and maintenance window")
	}
	packageDigest, packageErr := productCatalogStoppedHandoffImmutablePackageSHA256(task)
	configurationDigest, configurationErr := customerServiceRecoveryPlanCanonicalJSONSHA256(manager.Configuration)
	previewDigest, previewErr := customerServiceRecoveryPlanCanonicalJSONSHA256(binding)
	if packageErr != nil || configurationErr != nil || previewErr != nil || packageDigest != journal.PackageSHA256 || packageDigest != preparation.PackageSHA256 || configurationDigest != preparation.ConfigurationSHA256 || configurationDigest != binding.ConfigurationSHA256 || previewDigest != journal.PreviewBindingSHA256 || previewDigest != preparation.PreviewBindingSHA256 || binding.Version != 1 || binding.TargetMedia == nil || binding.CandidateMedia == nil || binding.Candidate.Owner != "local_backup_isolated_owner" || task.StageDatabase != "local_backup_stage_"+task.ID[:16] || task.StageFingerprint != binding.Candidate.Fingerprint || task.Journal.RollbackDatabase != "local_backup_product_old_"+task.ID[:16] || task.Journal.OldDatabaseOID != binding.Target.OID || task.Journal.NewDatabaseOID != binding.Candidate.OID || binding.Target.OID <= 0 || binding.Candidate.OID <= 0 || binding.Target.OID == binding.Candidate.OID {
		return errors.New("stopped product switch package, configuration, preview or database identities changed")
	}
	if err := manager.validateProductViewStoppedDrainProgress(task); err != nil {
		return err
	}
	if task.ProductViewStoppedDrain == nil || task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" || task.ProductViewStoppedDrain.SHA256 != preparation.DrainProgressSHA256 {
		return errors.New("stopped product switch drain and receipt observation changed")
	}
	expected, err := buildModuleRecoveryPairedMediaMoves(binding.TargetMedia, binding.CandidateMedia)
	if err != nil || len(expected) != len(journal.Moves) {
		return errors.New("stopped product switch complete media move contract changed")
	}
	for index, move := range journal.Moves {
		if (move.Completed && !move.Started) || (move.RollbackStarted && !move.Started) || (move.RolledBack && !move.RollbackStarted) {
			return errors.New("stopped product switch media acknowledgement order is invalid")
		}
		move.Started, move.Completed, move.RollbackStarted, move.RolledBack = false, false, false, false
		if move != expected[index] {
			return errors.New("stopped product switch media identity changed")
		}
	}
	switch journal.Phase {
	case "switching", "switched_still_isolated", "rolling_back":
	case "releasing_original", "original_released":
		if journal.RollbackVerifiedAt.IsZero() {
			return errors.New("stopped product rollback has no durable verified release intent")
		}
		for _, move := range journal.Moves {
			if move.Started && !move.RolledBack {
				return errors.New("stopped product rollback still has pending reverse moves")
			}
		}
		if journal.Phase == "original_released" && journal.OriginalServiceReleasedAt.IsZero() {
			return errors.New("stopped product rollback release acknowledgement is missing")
		}
	default:
		return errors.New("stopped product switch phase is invalid")
	}
	return nil
}

func (manager *Manager) validateProductCatalogStoppedSwitchPhysicalMediaLayout(ctx context.Context, task *Task) error {
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err != nil {
		return err
	}
	layout, err := manager.inspectProductCatalogPairedPhysicalMediaLayout(ctx, task, true)
	if err != nil {
		return err
	}
	return validateModuleRecoveryPairedMediaLayout(task.Journal.ProductCatalogStoppedSwitch.Moves, layout)
}

func (manager *Manager) renameStoppedProductSwitchMediaEntryWithVerifiedJournalAndLayout(ctx context.Context, task *Task, move moduleRecoveryMediaEntryMove, reverse bool) error {
	if err := manager.validateProductCatalogStoppedSwitchPhysicalMediaLayout(ctx, task); err != nil {
		return err
	}
	return manager.renameProductCatalogPairedMediaEntryAfterLayoutValidation(ctx, task, move, reverse)
}

func (manager *Manager) StartProductCatalogIsolatedSwitchWithinContinuouslyStoppedWindow(id, actor string, request ProductCatalogStoppedSwitchConfirmationRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	cfg := manager.Configuration
	if request.Confirmation != "SWITCH STOPPED PRODUCT CATALOG "+cfg.SiteID+"/"+cfg.Environment+"/"+id || strings.TrimSpace(actor) == "" {
		return nil, errors.New("explicit stopped product catalog switch confirmation is required")
	}
	if task.State != "product_module_stopped_handoff_prepared" || task.Journal != nil || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return nil, errors.New("a currently prepared, unexpired continuously stopped product window is required")
	}
	if err := manager.validateProductCatalogStoppedHandoffPreparationJournal(task); err != nil || task.ProductCatalogStoppedHandoffPreparation.Phase != "ready" {
		return nil, errors.New("the exact ready product preparation must be revalidated before isolated switching")
	}
	err = manager.startTaskOperation(task, "switching_product_with_application_continuously_stopped", func(ctx context.Context, task *Task) error {
		return manager.switchProductCatalogWithinContinuouslyHeldStoppedWindow(ctx, task, actor, nil)
	})
	if err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

// checkpoint is a test-only fault-injection parameter, never an HTTP input.
// Every failure retains the fence, held lease and durable rollback materials.
func (manager *Manager) switchProductCatalogWithinContinuouslyHeldStoppedWindow(ctx context.Context, task *Task, actor string, checkpoint func(string) error) error {
	if task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("product package expired before the switch began")
	}
	if err := manager.verifyProductCatalogStoppedHandoffFenceAndRuntimeBoundary(task); err != nil {
		return err
	}
	if err := manager.verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx, task); err != nil {
		return err
	}
	if err := manager.validateProductCatalogCandidateStillMatchesIsolatedPreview(ctx, task); err != nil {
		return err
	}
	if err := manager.verifyPreparedRecoveryFiles(task); err != nil {
		return err
	}
	cfg, binding := manager.Configuration, task.ProductCatalogPreviewBinding
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer control.Close()
	var unsafeApplicationRole bool
	if err := control.QueryRowContext(ctx, `SELECT rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls OR pg_has_role(oid,'local_backup_isolated_owner','MEMBER') FROM pg_roles WHERE rolname=$1`, cfg.ApplicationDatabaseRole).Scan(&unsafeApplicationRole); err != nil {
		return err
	}
	if unsafeApplicationRole {
		return errors.New("stopped product switching requires an unprivileged application role independent of the isolated owner")
	}
	rollbackName := "local_backup_product_old_" + task.ID[:16]
	var occupied int64
	if err := readDatabaseOID(ctx, control, rollbackName, &occupied); err != nil {
		return err
	}
	if occupied != 0 {
		return errors.New("stopped product rollback database name is occupied")
	}
	preparation := task.ProductCatalogStoppedHandoffPreparation
	moves, err := buildModuleRecoveryPairedMediaMoves(binding.TargetMedia, binding.CandidateMedia)
	if err != nil {
		return err
	}
	if err := manager.prepareProductCatalogPairedCompleteMediaVolumes(ctx, task, preparation.PreviewBindingSHA256); err != nil {
		return err
	}
	task.Journal = &RecoveryJournal{OldDatabaseOID: binding.Target.OID, NewDatabaseOID: binding.Candidate.OID, RollbackDatabase: rollbackName, Moves: []RecoveryFileMove{}, ProductCatalogStoppedSwitch: &ProductCatalogContinuouslyStoppedSwitchJournal{Version: 1, TaskID: task.ID, WindowID: preparation.WindowID, PreparationSHA256: preparation.SHA256, PackageSHA256: preparation.PackageSHA256, PreviewBindingSHA256: preparation.PreviewBindingSHA256, Operator: actor, StartedAt: time.Now().UTC(), Phase: "switching", Moves: moves}}
	task.Phase = "product_stopped_switch_journal_persisted"
	if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("journal_persisted"); err != nil {
			return err
		}
	}
	// Recheck after copying complete media, before any privilege/seal/rename.
	if task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("product package expired while preparing the paired switch; rollback is required")
	}
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return err
	}
	if err := manager.verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx, task); err != nil {
		return err
	}
	if err := manager.validateProductCatalogCandidateStillMatchesIsolatedPreview(ctx, task); err != nil {
		return err
	}
	if err := manager.revokeApplicationAccessToBoundStoppedProductCandidate(ctx, control, task); err != nil {
		return err
	}
	for _, name := range []string{cfg.DatabaseName, task.StageDatabase} {
		if err := sealDatabase(ctx, control, name); err != nil {
			return err
		}
	}
	if _, _, err := manager.inspectProductCatalogPairedDatabaseTopology(ctx, control, task); err != nil {
		return err
	}
	if err := manager.validateProductCatalogStoppedSwitchPhysicalMediaLayout(ctx, task); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("databases_sealed"); err != nil {
			return err
		}
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` RENAME TO `+pq.QuoteIdentifier(rollbackName)); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("original_database_renamed"); err != nil {
			return err
		}
	}
	for index := range task.Journal.ProductCatalogStoppedSwitch.Moves {
		move := &task.Journal.ProductCatalogStoppedSwitch.Moves[index]
		move.Started = true
		if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
			return err
		}
		if err := manager.renameStoppedProductSwitchMediaEntryWithVerifiedJournalAndLayout(ctx, task, *move, false); err != nil {
			return err
		}
		if checkpoint != nil {
			if err := checkpoint(fmt.Sprintf("media_%d_renamed_before_acknowledgement", index)); err != nil {
				return err
			}
		}
		move.Completed = true
		if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
			return err
		}
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(task.StageDatabase)+` RENAME TO `+pq.QuoteIdentifier(cfg.DatabaseName)); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("candidate_database_renamed_before_acknowledgement"); err != nil {
			return err
		}
	}
	task.Journal.DatabaseSwitched = true
	task.Journal.ProductCatalogStoppedSwitch.Phase = "switched_still_isolated"
	task.State, task.Phase = "product_module_switched_still_isolated", "product_database_and_complete_media_switched_application_stays_stopped"
	if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
		return err
	}
	_, err = manager.inspectStoppedProductCatalogPostSwitchWithHeldLease(ctx, task)
	return err
}

func (manager *Manager) InspectStoppedProductCatalogPostSwitch(ctx context.Context, id string) (*ProductCatalogStoppedPostSwitchInspection, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("another maintenance operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.State != "product_module_switched_still_isolated" {
		return nil, errors.New("a switched product candidate with this executor's continuous lease is required")
	}
	return manager.inspectStoppedProductCatalogPostSwitchWithHeldLease(ctx, task)
}

func (manager *Manager) inspectStoppedProductCatalogPostSwitchWithHeldLease(ctx context.Context, task *Task) (*ProductCatalogStoppedPostSwitchInspection, error) {
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err != nil {
		return nil, err
	}
	journal := task.Journal.ProductCatalogStoppedSwitch
	if journal.Phase != "switched_still_isolated" || !task.Journal.DatabaseSwitched {
		return nil, errors.New("only the fully switched stopped product result can be inspected")
	}
	for _, move := range journal.Moves {
		if !move.Completed || move.RollbackStarted || move.RolledBack {
			return nil, errors.New("product post-switch inspection has unfinished or reversed media moves")
		}
	}
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return nil, err
	}
	if err := manager.verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx, task); err != nil {
		return nil, err
	}
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return nil, err
	}
	defer control.Close()
	var connectedDatabasesAllowed int
	if err := control.QueryRowContext(ctx, `SELECT count(*) FROM pg_database WHERE oid=ANY($1) AND datallowconn`, pq.Array([]int64{task.Journal.OldDatabaseOID, task.Journal.NewDatabaseOID})).Scan(&connectedDatabasesAllowed); err != nil || connectedDatabasesAllowed != 0 {
		return nil, errors.New("post-switch product databases must both remain sealed")
	}
	original, candidate, err := manager.inspectProductCatalogPairedDatabaseTopology(ctx, control, task)
	if err != nil {
		return nil, err
	}
	if original != task.Journal.RollbackDatabase || candidate != manager.Configuration.DatabaseName {
		return nil, errors.New("post-switch product database positions changed")
	}
	if err := manager.validateProductCatalogStoppedSwitchPhysicalMediaLayout(ctx, task); err != nil {
		return nil, err
	}
	if err := manager.validateProductCatalogBundledStaticImagesAlreadyMatchTarget(ctx, task); err != nil {
		return nil, err
	}
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return nil, err
	}
	if err := manager.verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx, task); err != nil {
		return nil, err
	}
	mediaDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(task.ProductCatalogPreviewBinding.CandidateMedia)
	if err != nil {
		return nil, err
	}
	preparation, binding := task.ProductCatalogStoppedHandoffPreparation, task.ProductCatalogPreviewBinding
	report := &ProductCatalogStoppedPostSwitchInspection{Version: 1, Scope: productCatalogBackupScope, Observation: "database_and_complete_media_verified_with_application_stopped_external_actions_pending", TaskID: task.ID, WindowID: journal.WindowID, PackageID: task.Manifest.PackageID, PackageSHA256: journal.PackageSHA256, SwitchJournalSHA256: journal.SHA256, PreviewBindingSHA256: journal.PreviewBindingSHA256, ConfigurationSHA256: binding.ConfigurationSHA256, ApplicationBoundarySHA256: preparation.ApplicationBoundarySHA256, FenceSHA256: preparation.FenceSHA256, RedisOwnershipSHA256: task.ProductViewStoppedDrain.OwnershipSHA256, DatabaseName: candidate, DatabaseOID: binding.Candidate.OID, DatabaseFingerprint: binding.Candidate.Fingerprint, CompleteMediaSHA256: mediaDigest, ApplicationRemainsStopped: true, CandidateRemainsIsolated: true, ProjectionRequirements: productCatalogPartialProjectionReconciliationRequirements("verify_owned_redis_buffers_remain_empty_and_preserve_committed_sql_receipts", "已在本次停写窗口核对 Redis 归属和空缓冲，并核对完整 SQL 与收据；继续停写，最终接收前仍须重新核对，禁止清理浏览数或重放批次。")}
	report.ContractSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(report)
	if err != nil {
		return nil, err
	}
	report.CapturedAt = time.Now().UTC()
	return report, nil
}

func (manager *Manager) StartVerifiedPairedRollbackOfStoppedProductCatalogSwitch(id, actor string, request ProductCatalogStoppedSwitchConfirmationRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if request.Confirmation != "ROLLBACK STOPPED PRODUCT CATALOG "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id || strings.TrimSpace(actor) == "" {
		return nil, errors.New("explicit stopped product catalog paired rollback confirmation is required")
	}
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err != nil {
		return nil, err
	}
	if task.State == "running" || task.State == "queued" || task.State == "completed" || task.State == "rolled_back" {
		return nil, errors.New("stopped product switching is active or its original service has already been restored")
	}
	err = manager.startTaskOperation(task, "rolling_back_stopped_product_database_and_complete_media", func(ctx context.Context, task *Task) error {
		return manager.rollbackStoppedProductCatalogSwitchWithVerifiedOriginalRelease(ctx, task, nil)
	})
	if err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

func (manager *Manager) rollbackStoppedProductCatalogSwitchWithVerifiedOriginalRelease(ctx context.Context, task *Task, checkpoint func(string) error) error {
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err != nil {
		return err
	}
	journal, cfg := task.Journal.ProductCatalogStoppedSwitch, manager.Configuration
	fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
	if err != nil {
		return err
	}
	// A persisted release intent may have already removed the fence. Do not
	// reseal, inspect or roll back a database after service was allowed to resume.
	if !fenced && (journal.Phase == "releasing_original" || journal.Phase == "original_released") {
		return manager.acknowledgeStoppedProductCatalogOriginalRelease(task)
	}
	if journal.Phase == "original_released" {
		return errors.New("a released product rollback has a new fence; retain it for inspection")
	}
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return err
	}
	if manager.productCatalogStoppedHandoffLease == nil {
		wait, cancel := context.WithTimeout(ctx, 60*time.Second)
		lease, err := manager.acquireCustomerServicePairedSwitchExistingMaintenanceLease(wait, task)
		cancel()
		if err != nil {
			return err
		}
		manager.productCatalogStoppedHandoffLease, manager.productCatalogStoppedHandoffLeaseTaskID = lease, task.ID
	}
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return err
	}
	if err := manager.verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx, task); err != nil {
		return err
	}
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer control.Close()
	// Journal persistence can precede candidate ACL isolation. Reapply only
	// during the initial switch phase and only to the exact isolated OID/owner.
	if journal.Phase == "switching" {
		if err := manager.revokeApplicationAccessToBoundStoppedProductCandidate(ctx, control, task); err != nil {
			return err
		}
	}
	original, candidate, err := manager.inspectProductCatalogPairedDatabaseTopology(ctx, control, task)
	if err != nil {
		return err
	}
	if err := manager.validateProductCatalogStoppedSwitchPhysicalMediaLayout(ctx, task); err != nil {
		return err
	}
	journal.Phase = "rolling_back"
	if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
		return err
	}
	for _, name := range []string{original, candidate} {
		if err := sealDatabase(ctx, control, name); err != nil {
			return err
		}
	}
	if candidate == cfg.DatabaseName {
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(candidate)+` RENAME TO `+pq.QuoteIdentifier(task.StageDatabase)); err != nil {
			return err
		}
	}
	for index := len(journal.Moves) - 1; index >= 0; index-- {
		move := &journal.Moves[index]
		if !move.Started || move.RolledBack {
			continue
		}
		move.RollbackStarted = true
		if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
			return err
		}
		if err := manager.renameStoppedProductSwitchMediaEntryWithVerifiedJournalAndLayout(ctx, task, *move, true); err != nil {
			return err
		}
		if checkpoint != nil {
			if err := checkpoint(fmt.Sprintf("media_%d_reversed_before_acknowledgement", index)); err != nil {
				return err
			}
		}
		move.RolledBack = true
		if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
			return err
		}
	}
	if original == task.Journal.RollbackDatabase {
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(original)+` RENAME TO `+pq.QuoteIdentifier(cfg.DatabaseName)); err != nil {
			return err
		}
	}
	original, candidate, err = manager.inspectProductCatalogPairedDatabaseTopology(ctx, control, task)
	if err != nil || original != cfg.DatabaseName || candidate != task.StageDatabase {
		return errors.Join(err, errors.New("stopped product original database positions could not be verified"))
	}
	if err := manager.validateProductCatalogStoppedSwitchPhysicalMediaLayout(ctx, task); err != nil {
		return err
	}
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return err
	}
	if err := manager.verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx, task); err != nil {
		return err
	}
	journal.Phase, journal.RollbackVerifiedAt = "releasing_original", time.Now().UTC()
	task.Phase = "product_original_database_and_complete_media_verified_before_release"
	if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
		return err
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` ALLOW_CONNECTIONS true`); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("original_verified_before_fence_removal"); err != nil {
			return err
		}
	}
	// Ordinary release and preparation cancellation reject this journal. Only
	// this already verified paired rollback may remove its exact original fence.
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename)); err != nil {
		return err
	}
	if err := syncDirectory(cfg.MaintenanceDirectory); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("fence_removed_before_release_acknowledgement"); err != nil {
			return err
		}
	}
	return manager.acknowledgeStoppedProductCatalogOriginalRelease(task)
}

func (manager *Manager) acknowledgeStoppedProductCatalogOriginalRelease(task *Task) error {
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err != nil {
		return err
	}
	journal := task.Journal.ProductCatalogStoppedSwitch
	if journal.Phase != "releasing_original" && journal.Phase != "original_released" {
		return errors.New("stopped product original release has no durable verified intent")
	}
	if manager.productCatalogStoppedHandoffLease != nil && manager.productCatalogStoppedHandoffLeaseTaskID != task.ID {
		return errors.New("another product window holds the runtime lease")
	}
	if err := manager.closeProductCatalogStoppedHandoffLease(); err != nil {
		return err
	}
	journal.Phase = "original_released"
	if journal.OriginalServiceReleasedAt.IsZero() {
		journal.OriginalServiceReleasedAt = time.Now().UTC()
	}
	task.State, task.Phase, task.Error = "rolled_back", "product_stopped_switch_original_service_restored_after_verified_paired_rollback", ""
	return manager.saveProductCatalogStoppedSwitchProgress(task)
}

func (manager *Manager) revokeApplicationAccessToBoundStoppedProductCandidate(ctx context.Context, control *sql.DB, task *Task) error {
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err != nil {
		return err
	}
	var name, owner string
	if err := control.QueryRowContext(ctx, `SELECT datname,pg_get_userbyid(datdba) FROM pg_database WHERE oid=$1`, task.Journal.NewDatabaseOID).Scan(&name, &owner); err != nil {
		return err
	}
	if (name != task.StageDatabase && name != manager.Configuration.DatabaseName) || owner != task.ProductCatalogPreviewBinding.Candidate.Owner || owner != "local_backup_isolated_owner" {
		return errors.New("cannot isolate stopped product candidate with a changed database identity or owner")
	}
	if err := assertNoUnfencedDatabaseSessions(ctx, control, name); err != nil {
		return err
	}
	_, err := control.ExecContext(ctx, `REVOKE ALL ON DATABASE `+pq.QuoteIdentifier(name)+` FROM PUBLIC,`+pq.QuoteIdentifier(manager.Configuration.ApplicationDatabaseRole))
	return err
}

func (manager *Manager) retainStoppedProductCatalogSwitchForExplicitRollbackAfterRestart(task *Task) error {
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err == nil {
		fenced, fenceErr := maintenancegate.IsApplicationFenced(manager.Configuration.MaintenanceDirectory)
		phase := task.Journal.ProductCatalogStoppedSwitch.Phase
		if fenceErr == nil && !fenced && (phase == "releasing_original" || phase == "original_released") {
			return manager.acknowledgeStoppedProductCatalogOriginalRelease(task)
		}
		if taskHasDurablyReleasedOriginalAfterProductCatalogStoppedSwitchRollback(task) {
			return nil
		}
	}
	task.State, task.Phase = "interrupted", "product_stopped_switch_requires_explicit_verified_rollback"
	task.Error = "stopped product switch retains its fence and materials; historical readiness or inspection cannot resume service; explicitly verify paired rollback"
	return manager.saveTask(task)
}

func (manager *Manager) cleanupVerifiedRolledBackStoppedProductSwitchMediaVolumes(ctx context.Context, task *Task) error {
	if !taskHasDurablyReleasedOriginalAfterProductCatalogStoppedSwitchRollback(task) || task.State != "rolled_back" {
		return errors.New("stopped product media cleanup requires durably verified original service release")
	}
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err != nil {
		return err
	}
	actual, err := manager.inspectProductCatalogPairedPhysicalMediaLayout(ctx, task, false)
	if err != nil {
		return err
	}
	expected := map[string]string{}
	for _, move := range task.Journal.ProductCatalogStoppedSwitch.Moves {
		if move.Direction == "install_candidate" {
			source, _ := customerServicePairedSwitchMediaMoveLogicalPaths(move)
			expected[source] = move.EntrySHA256
		}
	}
	actualDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(actual)
	expectedDigest, expectedErr := customerServiceRecoveryPlanCanonicalJSONSHA256(expected)
	if err != nil || expectedErr != nil || actualDigest != expectedDigest {
		return errors.New("stopped product cleanup media changed; retain materials for inspection")
	}
	for _, rootName := range mediaRoots {
		if err := os.RemoveAll(manager.productCatalogPairedMediaVolumeDirectory(task, rootName)); err != nil {
			return err
		}
		if err := syncDirectory(manager.Configuration.StorageRoots[rootName]); err != nil {
			return err
		}
	}
	return nil
}
