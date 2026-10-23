package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"

	"commerce-platform/internal/pkg/maintenancegate"
)

type Task struct {
	ID                                           string                                                 `json:"id"`
	Kind                                         string                                                 `json:"kind"`
	Actor                                        string                                                 `json:"actor"`
	State                                        string                                                 `json:"state"`
	Phase                                        string                                                 `json:"phase"`
	Error                                        string                                                 `json:"error,omitempty"`
	CreatedAt                                    time.Time                                              `json:"created_at"`
	UpdatedAt                                    time.Time                                              `json:"updated_at"`
	ExpiresAt                                    time.Time                                              `json:"expires_at"`
	Size                                         int64                                                  `json:"size"`
	UploadedBytes                                int64                                                  `json:"uploaded_bytes"`
	CiphertextSHA256                             string                                                 `json:"ciphertext_sha256,omitempty"`
	DownloadRequests                             int                                                    `json:"download_requests"`
	WorkCopyExpired                              bool                                                   `json:"work_copy_expired"`
	Manifest                                     *Manifest                                              `json:"manifest,omitempty"`
	PreviewTargetFingerprint                     string                                                 `json:"preview_target_fingerprint,omitempty"`
	PreviewTargetTables                          map[string]int64                                       `json:"preview_target_tables,omitempty"`
	StageDatabase                                string                                                 `json:"stage_database,omitempty"`
	StageFingerprint                             string                                                 `json:"stage_fingerprint,omitempty"`
	Journal                                      *RecoveryJournal                                       `json:"recovery_journal,omitempty"`
	Reconciliation                               *ReconciliationEvidence                                `json:"reconciliation,omitempty"`
	ExternalProjectionReconciliationContract     *ExternalProjectionReconciliationContract              `json:"external_projection_reconciliation_contract,omitempty"`
	ReconciliationConnectionsEnabled             bool                                                   `json:"reconciliation_connections_enabled"`
	ProductCatalogRecovery                       *ProductCatalogIsolatedRecoveryResult                  `json:"product_catalog_recovery,omitempty"`
	ProductCatalogPreviewBinding                 *productCatalogRecoveryPreviewBinding                  `json:"product_catalog_preview_binding,omitempty"`
	ProductViewStoppedDrain                      *ProductViewStoppedDrainAndCandidateRebuildProgress    `json:"product_view_stopped_drain,omitempty"`
	ProductCatalogStoppedHandoffPreparation      *ProductCatalogStoppedHandoffPreparationJournal        `json:"product_catalog_stopped_handoff_preparation,omitempty"`
	ProductCatalogStoppedDetailCacheInvalidation *ProductCatalogStoppedDetailCacheInvalidationProgress  `json:"product_catalog_stopped_detail_cache_invalidation,omitempty"`
	ProductCatalogStoppedProjectionEvidence      *ProductCatalogStoppedProjectionOperatorEvidenceRecord `json:"product_catalog_stopped_projection_evidence,omitempty"`
}

type Manager struct {
	Configuration                           Configuration
	stateMutex                              sync.Mutex
	operationMutex                          sync.Mutex
	processLease                            *flock.Flock
	gateProcessLease                        *flock.Flock
	productCatalogStoppedHandoffLease       *flock.Flock
	productCatalogStoppedHandoffLeaseTaskID string
}

func NewManager(cfg Configuration) (*Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	lease := flock.New(filepath.Join(cfg.Workspace, "maintenance-executor.lock"))
	locked, err := lease.TryLock()
	if err != nil || !locked {
		return nil, errors.New("another independent backup executor owns this workspace")
	}
	manager := &Manager{Configuration: cfg, processLease: lease}
	gateLease := flock.New(filepath.Join(cfg.MaintenanceDirectory, "independent-maintenance-executor.lock"))
	gateLocked, gateErr := gateLease.TryLock()
	if gateErr != nil || !gateLocked {
		_ = lease.Close()
		_ = gateLease.Close()
		return nil, errors.New("another backup executor owns the application maintenance gate")
	}
	manager.gateProcessLease = gateLease
	tasks, err := manager.ListTasks()
	if err != nil {
		_ = manager.Close()
		return nil, err
	}
	for _, task := range tasks {
		if taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
			if err := manager.retainStoppedProductCatalogSwitchForExplicitRollbackAfterRestart(task); err != nil {
				_ = manager.Close()
				return nil, err
			}
			continue
		}
		if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
			task.State, task.Phase = "interrupted", "product_stopped_handoff_requires_explicit_revalidation"
			task.Error = "stopped product handoff preparation retains isolation; explicitly revalidate or cancel, never reuse readiness after restart"
			if err := manager.saveTask(task); err != nil {
				_ = manager.Close()
				return nil, err
			}
			continue
		}
		if task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
			task.State, task.Phase = "interrupted", "product_view_drain_retained_for_explicit_retry"
			task.ProductCatalogRecovery = nil
			task.Error = "stopped product view drain is unresolved; retain maintenance and explicitly retry after inspecting ownership and writers"
			if err := manager.saveTask(task); err != nil {
				_ = manager.Close()
				return nil, err
			}
		}
		if task.State == "resuming" || task.State == "releasing_rollback" {
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			if err != nil {
				_ = manager.Close()
				return nil, err
			}
			if !fenced {
				if taskHasProductCatalogPairedRecoveryJournal(task) {
					if err := manager.validateProductCatalogPairedRehearsalBinding(task); err != nil || !task.Journal.ProductCatalog.RollbackVerified {
						task.State, task.Phase, task.Error = "interrupted", "product_paired_rehearsal_retained_for_inspection", "product release intent has no valid verified rollback; inspection required"
						if err := manager.saveTask(task); err != nil {
							_ = manager.Close()
							return nil, err
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
				_ = manager.Close()
				return nil, err
			}
		}
		if task.State == "running" || task.State == "queued" {
			task.State = "interrupted"
			task.Error = "executor stopped; inspect persisted phase and recovery journal before continuing"
			if err := manager.saveTask(task); err != nil {
				_ = manager.Close()
				return nil, err
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
	return manager, nil
}

func (manager *Manager) Close() error {
	manager.operationMutex.Lock()
	defer manager.operationMutex.Unlock()
	return errors.Join(manager.closeProductCatalogStoppedHandoffLease(), manager.processLease.Close(), manager.gateProcessLease.Close())
}
func (manager *Manager) taskDirectory(id string) (string, error) {
	if !safeTaskID.MatchString(id) {
		return "", errors.New("invalid task identity")
	}
	return filepath.Join(manager.Configuration.Workspace, id), nil
}
func (manager *Manager) taskPackagePath(id string) (string, error) {
	directory, err := manager.taskDirectory(id)
	return filepath.Join(directory, "application-backup.age"), err
}

func (manager *Manager) GetTask(id string) (*Task, error) {
	manager.stateMutex.Lock()
	defer manager.stateMutex.Unlock()
	directory, err := manager.taskDirectory(id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(directory, "task.json")) // #nosec G304 -- The path is a fixed record under the configured private maintenance root or validated task directory.
	if err != nil {
		return nil, err
	}
	var task Task
	if err := decodeStrictJSON(data, &task); err != nil {
		return nil, err
	}
	if task.ID != id {
		return nil, errors.New("task identity mismatch")
	}
	return &task, nil
}

func (manager *Manager) ListTasks() ([]*Task, error) {
	entries, err := os.ReadDir(manager.Configuration.Workspace)
	if err != nil {
		return nil, err
	}
	tasks := make([]*Task, 0)
	for _, entry := range entries {
		if !entry.IsDir() || !safeTaskID.MatchString(entry.Name()) {
			continue
		}
		task, err := manager.GetTask(entry.Name())
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].CreatedAt.After(tasks[j].CreatedAt) })
	return tasks, nil
}

func (manager *Manager) saveTask(task *Task) error {
	manager.stateMutex.Lock()
	defer manager.stateMutex.Unlock()
	directory, err := manager.taskDirectory(task.ID)
	if err != nil {
		return err
	}
	task.UpdatedAt = time.Now().UTC()
	return writeDurableJSON(filepath.Join(directory, "task.json"), task)
}

func (manager *Manager) createTask(kind, actor string) (*Task, error) {
	task := &Task{ID: newTaskID(), Kind: kind, Actor: actor, State: "queued", Phase: "created", CreatedAt: time.Now().UTC()}
	task.ExpiresAt = task.CreatedAt.Add(time.Duration(manager.Configuration.WorkCopyLifetimeHours) * time.Hour)
	directory, _ := manager.taskDirectory(task.ID)
	if err := os.Mkdir(directory, 0700); err != nil {
		return nil, err
	}
	return task, manager.saveTask(task)
}

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

func (manager *Manager) CreateBackup(actor, confirmation string) (*Task, error) {
	cfg := manager.Configuration
	if confirmation != "BACKUP "+cfg.SiteID+"/"+cfg.Environment {
		return nil, errors.New("explicit confirmation of the storefront maintenance window is required")
	}
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	preflight, cancelPreflight := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelPreflight()
	required, err := cfg.estimateBackupWorkspaceBytes(preflight)
	if err != nil {
		return nil, err
	}
	if err := cfg.checkWorkspaceCapacity(required); err != nil {
		return nil, err
	}
	task, err := manager.createTask("backup", actor)
	if err != nil {
		return nil, err
	}
	if err := manager.startTaskOperation(task, "waiting_for_application_stop", manager.generateBackup); err != nil {
		task.State = "failed"
		task.Error = err.Error()
		_ = manager.saveTask(task)
		return nil, err
	}
	return manager.GetTask(task.ID)
}

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

func (manager *Manager) generateBackup(ctx context.Context, task *Task) (resultErr error) {
	cfg := manager.Configuration
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	// No live data is changed by generation. Successful or failed capture can
	// release maintenance; restore and switch failures deliberately cannot.
	defer func() { resultErr = errors.Join(resultErr, manager.releaseMaintenance(task)) }()
	directory, _ := manager.taskDirectory(task.ID)
	snapshot := filepath.Join(directory, "snapshot")
	if err := os.Mkdir(snapshot, 0700); err != nil {
		return err
	}
	defer func(cleanupPath string) { _ = os.RemoveAll(cleanupPath) }(snapshot)
	manifest := Manifest{FormatVersion: PackageVersion, PackageID: task.ID, SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, CapturedAt: time.Now().UTC(), Scope: "whole_application", Files: []FileIdentity{}, Gaps: []string{}}
	task.Phase = "capturing_consistent_database_snapshot"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if err := cfg.captureSnapshotAndDump(ctx, snapshot, &manifest); err != nil {
		return err
	}
	if manifest.MigrationVersion != cfg.ExpectedMigrationVersion {
		return errors.New("application migration does not match configured backup contract")
	}
	task.Phase = "copying_and_verifying_three_media_roots"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if err := cfg.copyMediaSnapshot(ctx, snapshot, &manifest); err != nil {
		return err
	}
	if err := cfg.copyWholeApplicationCompanyStaticSnapshot(ctx, snapshot, &manifest); err != nil {
		return err
	}
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	err = cfg.findMediaReferenceGaps(ctx, db, &manifest)
	if err == nil {
		err = findBusinessRelationshipGaps(ctx, db, &manifest)
	}
	if err == nil {
		_, fingerprint, _, inventoryErr := captureDatabaseInventory(ctx, db)
		err = inventoryErr
		if err == nil && fingerprint != manifest.DatabaseFingerprint {
			err = errors.New("database changed despite the maintenance fence; backup rejected")
		}
	}
	_ = db.Close()
	if err != nil {
		return err
	}
	if err := writeDurableJSON(filepath.Join(snapshot, "recovery-secrets.json"), cfg.RecoverySecrets); err != nil {
		return err
	}
	for _, name := range []string{"database.dump", "recovery-secrets.json"} {
		entry, err := fileIdentity(filepath.Join(snapshot, name), name)
		if err != nil {
			return err
		}
		manifest.Files = append(manifest.Files, entry)
	}
	cfg.sortedManifestFiles(&manifest)
	liveMediaFingerprint, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return err
	}
	if liveMediaFingerprint != fingerprintManifestMedia(manifest) {
		return errors.New("media changed despite maintenance isolation; package rejected")
	}
	if wholeApplicationPackageIncludesCompanyStatic(&manifest) {
		fingerprint, err := cfg.companyStaticReleaseFileFingerprint(ctx)
		if err != nil || fingerprint != fingerprintCompanyStaticManifestFiles(manifest.Files) {
			return errors.New("company static files changed during backup; package rejected")
		}
	}
	// The copied files are now independent of live roots. Encryption can run
	// after capture, but retaining the fence also protects against rogue writers.
	task.Phase = "encrypting_package"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	if err != nil {
		return err
	}
	destination, _ := manager.taskPackagePath(task.ID)
	if err := WriteEncryptedPackage(destination, snapshot, manifest, passphrase, cfg.MaxPackageBytes); err != nil {
		return err
	}
	identity, err := fileIdentity(destination, "")
	if err != nil {
		return err
	}
	task.Manifest = &manifest
	task.Size = identity.Size
	task.CiphertextSHA256 = identity.SHA256
	task.State = "ready_to_download"
	task.Phase = "package_generated"
	if len(manifest.Gaps) > 0 {
		task.State = "generated_with_gaps"
		task.Phase = "package_generated_restore_blocked"
	}
	return manager.saveTask(task)
}

func (manager *Manager) VerifyUploadedPackage(id string) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
		return nil, errors.New("resume or cancel the stopped product handoff preparation before package verification")
	}
	if task.Kind != "upload" || task.UploadedBytes != task.Size || (task.State != "uploaded" && task.State != "failed" && task.State != "interrupted") {
		return nil, errors.New("package upload is incomplete or not eligible for validation")
	}
	if task.Journal != nil {
		return nil, errors.New("a recovery journal already exists; use rollback or reconciliation")
	}
	if task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
		return nil, errors.New("resume the stopped product view drain and candidate rebuild first")
	}
	if err := manager.startTaskOperation(task, "authenticating_and_extracting", manager.verifyAndPreview); err != nil {
		return nil, err
	}
	return manager.GetTask(task.ID)
}

func (manager *Manager) verifyAndPreview(ctx context.Context, task *Task) error {
	cfg := manager.Configuration
	if err := cfg.checkWorkspaceCapacity(64 << 20); err != nil {
		return err
	}
	directory, _ := manager.taskDirectory(task.ID)
	extracted := filepath.Join(directory, "verified")
	if err := os.RemoveAll(extracted); err != nil {
		return err
	}
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	if err != nil {
		return err
	}
	filename, _ := manager.taskPackagePath(task.ID)
	identity, err := fileIdentity(filename, "")
	if err != nil || identity.Size != task.Size {
		return errors.New("uploaded ciphertext length changed")
	}
	task.CiphertextSHA256 = identity.SHA256
	manifest, err := extractAndVerifyEncryptedPackageWithAllowedScopes(filename, extracted, passphrase, cfg.MaxExpandedBytes, cfg.MaxFiles, true, cfg.checkWorkspaceCapacity)
	if err != nil {
		return err
	}
	if manifest.SiteID != cfg.SiteID || manifest.Environment != cfg.Environment || manifest.ApplicationVersion != cfg.ApplicationVersion || manifest.MigrationVersion != cfg.ExpectedMigrationVersion {
		return errors.New("package site, environment, application or migration version is incompatible")
	}
	task.Manifest = &manifest
	if len(manifest.Gaps) > 0 {
		return errors.New("package contains unresolved attachment or media gaps; overwrite is blocked")
	}
	if manifest.Scope == productCatalogBackupScope {
		return manager.restoreProductCatalogIntoIsolatedTargetClone(ctx, task, extracted)
	}
	if err := cfg.validateWholeApplicationCompanyStaticPackage(manifest, extracted); err != nil {
		return err
	}
	materials, err := readRecoveryMaterials(extracted)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(materials, cfg.RecoverySecrets) {
		return errors.New("recovery signing/encryption keys do not match the target; provision the original keys separately")
	}
	task.Phase = "restoring_isolated_database"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if task.StageDatabase != "" {
		if err := cfg.dropOwnedStageDatabase(ctx, task.StageDatabase); err != nil {
			return err
		}
	}
	task.StageDatabase = "local_backup_stage_" + task.ID[:16]
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if err := cfg.restoreIsolatedDatabase(ctx, extracted, task.StageDatabase, manifest); err != nil {
		return err
	}
	staged, err := cfg.connectDatabase(ctx, task.StageDatabase)
	if err != nil {
		return err
	}
	_, task.StageFingerprint, _, err = captureDatabaseInventory(ctx, staged)
	_ = staged.Close()
	if err != nil {
		return err
	}
	task.PreviewTargetFingerprint, task.PreviewTargetTables, err = cfg.targetFingerprint(ctx)
	if err != nil {
		return err
	}
	task.State = "verified"
	task.Phase = "isolated_restore_verified_overwrite_preview_ready"
	return manager.saveTask(task)
}

func (cfg Configuration) targetFingerprint(ctx context.Context) (string, map[string]int64, error) {
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return "", nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	var exists bool
	if err := control.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, cfg.DatabaseName).Scan(&exists); err != nil {
		return "", nil, err
	}
	fingerprint := "absent_database"
	tables := map[string]int64{}
	if exists {
		db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
		if err != nil {
			return "", nil, err
		}
		tables, fingerprint, _, err = captureDatabaseInventory(ctx, db)
		_ = db.Close()
		if err != nil {
			return "", nil, err
		}
	}
	media, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return "", nil, err
	}
	identity := fingerprint + ":" + media
	if cfg.ProductStaticAssetsDirectory != "" {
		staticFingerprint, err := cfg.companyStaticReleaseFileFingerprint(ctx)
		if err != nil {
			return "", nil, err
		}
		identity += ":company-static:" + staticFingerprint
	}
	digest := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(digest[:]), tables, nil
}

func (manager *Manager) ReleaseUnmodifiedMaintenance(id string) error {
	if !manager.operationMutex.TryLock() {
		return errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return err
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
