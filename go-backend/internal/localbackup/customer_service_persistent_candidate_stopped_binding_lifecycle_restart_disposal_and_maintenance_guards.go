package localbackup

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
	"github.com/lib/pq"
)

const customerServicePersistentCandidateLifecycleFilename = "customer-service-persistent-candidate-lifecycle-v1.json"

// This is an internal preparation/disposal ledger, never a switch journal.
// StageDatabase, StageFingerprint and the whole-storefront preview stay intact.
type customerServicePersistentCandidateLifecycle struct {
	Version                       int                                                  `json:"version"`
	State                         string                                               `json:"state"`
	ExecutionAvailable            bool                                                 `json:"execution_available"`
	TaskID                        string                                               `json:"task_id"`
	AttemptID                     string                                               `json:"attempt_id"`
	CreatedAt                     time.Time                                            `json:"created_at"`
	UpdatedAt                     time.Time                                            `json:"updated_at"`
	Request                       CustomerServiceEffectiveSelectionInspectionRequest   `json:"saved_selection_request"`
	Binding                       *CustomerServiceRecoveryPlanSnapshotBinding          `json:"stopped_snapshot_binding,omitempty"`
	Identity                      customerServiceDisposableCandidateDatabaseIdentity   `json:"database_identity"`
	Validation                    *customerServiceDisposableCandidateRehearsalResult   `json:"validation,omitempty"`
	MediaUnion                    *customerServiceCandidateCompleteMediaUnionLifecycle `json:"complete_media_union,omitempty"`
	PeakCapacity                  *customerServiceCandidateCombinedPeakCapacityBudget  `json:"combined_peak_capacity,omitempty"`
	CleanupDispositionAuditSHA256 string                                               `json:"cleanup_disposition_audit_sha256,omitempty"`
	ErrorCode                     string                                               `json:"error_code,omitempty"`
	SHA256                        string                                               `json:"sha256"`
}

func (manager *Manager) saveCustomerServicePersistentCandidateLifecycle(record *customerServicePersistentCandidateLifecycle) error {
	directory, err := manager.taskDirectory(record.TaskID)
	if err != nil {
		return err
	}
	record.UpdatedAt = time.Now().UTC()
	record.SHA256 = ""
	record.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(record)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	if len(encoded) > 4<<20 {
		return errors.New("candidate lifecycle exceeds its size limit")
	}
	if err := manager.Configuration.checkWorkspaceCapacity(int64(len(encoded)) * 2); err != nil {
		return err
	}
	return writeDurableFile(filepath.Join(directory, customerServicePersistentCandidateLifecycleFilename), encoded)
}

func (manager *Manager) loadCustomerServicePersistentCandidateLifecycle(id string) (*customerServicePersistentCandidateLifecycle, error) {
	directory, err := manager.taskDirectory(id)
	if err != nil {
		return nil, err
	}
	if info, err := os.Lstat(directory); err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
		return nil, errors.New("candidate task directory is not a private directory")
	} else if err != nil {
		return nil, err
	}
	filename := filepath.Join(directory, customerServicePersistentCandidateLifecycleFilename)
	info, err := os.Lstat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 4<<20 {
		return nil, errors.New("candidate lifecycle must be a bounded regular file")
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, (4<<20)+1))
	if err != nil || len(data) > 4<<20 {
		return nil, errors.New("candidate lifecycle exceeds its size limit")
	}
	var record customerServicePersistentCandidateLifecycle
	if decodeStrictJSON(data, &record) != nil {
		return nil, errors.New("candidate lifecycle is invalid; inspection is required")
	}
	copyOfRecord := record
	copyOfRecord.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(&copyOfRecord)
	planDigest, planErr := hex.DecodeString(record.Request.PlanSHA256)
	if err != nil || digest != record.SHA256 || record.Version != 1 || record.ExecutionAvailable || record.TaskID != id || !safeTaskID.MatchString(record.AttemptID) || planErr != nil || len(planDigest) != 32 || record.Identity.Version != 1 || record.Identity.DatabaseName != "local_backup_candidate_"+record.AttemptID[:16] || record.Identity.DatabaseName == manager.Configuration.DatabaseName || record.Identity.PlanSHA256 != record.Request.PlanSHA256 {
		return nil, errors.New("candidate lifecycle digest, task or database identity changed; inspection is required")
	}
	switch record.State {
	case "waiting_for_application_stop", "revalidating_stopped_binding", "constructing_not_switchable", "sealed_not_switchable", "discard_required", "inspection_required", "discarded_not_switchable":
	default:
		return nil, errors.New("candidate lifecycle state is unsupported")
	}
	if record.State == "sealed_not_switchable" && (record.Identity.State != "allocated_not_switchable" || record.Binding == nil || record.Binding.TaskID != id || record.Validation == nil || record.Validation.ExecutionAvailable || record.Validation.CandidateDeleted || record.Validation.PlanSHA256 != record.Request.PlanSHA256 || record.Validation.State != "validated_not_switchable") {
		return nil, errors.New("sealed candidate validation or binding is invalid")
	}
	if record.State == "discarded_not_switchable" && (record.Identity.State == "allocation_intent" || (record.Validation != nil && !record.Validation.CandidateDeleted)) {
		return nil, errors.New("discarded candidate cleanup evidence is invalid")
	}
	if err := manager.validateCustomerServiceCandidateMediaLifecycle(&record); err != nil {
		return nil, err
	}
	if (record.Identity.State == "allocation_not_attempted" || record.Identity.State == "allocation_intent") && record.Identity.DatabaseOID == 0 {
		return &record, nil
	}
	if record.Identity.State != "allocated_not_switchable" || record.Identity.DatabaseOID <= 0 {
		return nil, errors.New("candidate allocation state or OID is invalid")
	}
	return &record, nil
}

func (manager *Manager) requireCustomerServicePersistentCandidateDiscarded(id string) error {
	if err := manager.requireCustomerServiceIndependentPairedSwitchRolledBack(id); err != nil {
		return err
	}
	record, err := manager.loadCustomerServicePersistentCandidateLifecycle(id)
	if err != nil {
		return err
	}
	if record != nil && record.State != "discarded_not_switchable" {
		return errors.New("customer-service candidate must be safely discarded before releasing maintenance, expiring work copies or whole-storefront recovery")
	}
	return nil
}

func (manager *Manager) customerServicePersistentCandidateConstructionWorkspace(record *customerServicePersistentCandidateLifecycle) (*customerServiceRecordedCandidateConstructionWorkspace, error) {
	directory, err := manager.taskDirectory(record.TaskID)
	if err != nil || !safeTaskID.MatchString(record.AttemptID) {
		return nil, errors.New("candidate workspace task or attempt identity is invalid")
	}
	return &customerServiceRecordedCandidateConstructionWorkspace{
		Directory: filepath.Join(directory, "customer-service-persistent-candidate-"+record.AttemptID), Identity: record.Identity,
		PersistIdentity: func(identity customerServiceDisposableCandidateDatabaseIdentity) error {
			record.Identity = identity
			return manager.saveCustomerServicePersistentCandidateLifecycle(record)
		},
	}, nil
}

// Internal only: no HTTP/page entry, switch authority or automatic service release.
// Stop first, then recompute the exact saved plan/review binding inside both snapshots.
func (manager *Manager) prepareCustomerServicePersistentCandidateUnderStoppedApplication(ctx context.Context, id string, request CustomerServiceEffectiveSelectionInspectionRequest) (result *customerServicePersistentCandidateLifecycle, resultErr error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.Kind != "upload" || task.State != "verified" || task.Manifest == nil || task.Journal != nil || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) || task.StageDatabase != "local_backup_stage_"+id[:16] {
		return nil, errors.New("persistent candidate requires the original verified upload and unexpired stage")
	}
	if err := manager.requireCustomerServicePersistentCandidateDiscarded(id); err != nil {
		return nil, err
	}
	previous, err := manager.loadCustomerServicePersistentCandidateLifecycle(id)
	if err != nil {
		return nil, err
	}
	directory, _ := manager.taskDirectory(id)
	if previous != nil {
		if err := writeDurableJSON(filepath.Join(directory, "customer-service-discarded-candidate-attempt-"+previous.AttemptID+".json"), previous); err != nil {
			return nil, err
		}
	}
	planDigest, err := hex.DecodeString(request.PlanSHA256)
	if err != nil || len(planDigest) != 32 {
		return nil, errors.New("saved candidate plan digest is invalid")
	}
	if err := manager.Configuration.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	review, err := manager.loadCustomerServiceRecoveryConflictReviewWhileOperationLocked(id)
	if err != nil {
		return nil, err
	}
	attemptID := newTaskID()
	record := &customerServicePersistentCandidateLifecycle{Version: 1, State: "waiting_for_application_stop", TaskID: id, AttemptID: attemptID, CreatedAt: time.Now().UTC(), Request: request,
		Identity: customerServiceDisposableCandidateDatabaseIdentity{Version: 1, State: "allocation_not_attempted", DatabaseName: "local_backup_candidate_" + attemptID[:16], PlanSHA256: request.PlanSHA256}}
	if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
		return nil, err
	}
	// A failed/cancelled prepare remains a durable discard obligation. Even an
	// allocation with a lost response cannot disappear into an untracked temp dir.
	defer func() {
		if resultErr != nil {
			record.State, record.ErrorCode = "discard_required", "candidate_preparation_failed_requires_discard"
			resultErr = errors.Join(resultErr, manager.saveCustomerServicePersistentCandidateLifecycle(record))
		}
	}()
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return nil, err
	}
	defer lease.Close()
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return nil, err
	}
	defer control.Close()
	for _, databaseName := range []string{manager.Configuration.DatabaseName, task.StageDatabase} {
		if err := assertNoUnfencedDatabaseSessions(ctx, control, databaseName); err != nil {
			return nil, err
		}
	}
	record.State = "revalidating_stopped_binding"
	if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
		return nil, err
	}
	workspace, err := manager.customerServicePersistentCandidateConstructionWorkspace(record)
	if err != nil {
		return nil, err
	}
	if err := os.Mkdir(workspace.Directory, 0700); err != nil {
		return nil, err
	}
	if err := syncDirectory(directory); err != nil {
		return nil, err
	}
	_, err = manager.inspectVerifiedPackageModuleDependenciesWithCustomerServiceSelectionObserverWhileOperationLocked(ctx, id, "customer_service", func(ctx context.Context, report *ModuleRecoveryDependencyReport, source, target *moduleRecoveryRowDependencyGraph) error {
		plan := report.ConversationRows.ConversationRecoveryPlan
		if _, err := validateCustomerServiceSavedDecisionsAgainstLatestPlan(plan, review, request, time.Now().UTC()); err != nil {
			return err
		}
		record.State, record.Binding = "constructing_not_switchable", &plan.Binding
		if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
			return err
		}
		var inventory *customerServiceCandidateCompleteMediaUnionInventory
		workspace.PrepareCompleteMediaForSelection = func(ctx context.Context, selection *CustomerServiceEffectiveSelectionInspection) error {
			var encodedInventory []byte
			inventory, encodedInventory, err = manager.Configuration.planCustomerServiceCompleteCurrentMediaUnion(ctx, selection, task.Manifest, filepath.Join(directory, "verified"))
			if err != nil {
				return err
			}
			budget, err := manager.Configuration.inspectCustomerServiceCandidateCombinedPeakCapacity(ctx, source.Transaction, target.Transaction, inventory)
			if err != nil {
				return err
			}
			record.PeakCapacity = budget
			if err := manager.persistCustomerServiceCandidateBoundCleanupDispositionAudit(record, selection); err != nil {
				return err
			}
			record.MediaUnion = &customerServiceCandidateCompleteMediaUnionLifecycle{Version: 1, State: "allocation_intent", InventorySHA256: inventory.SHA256, BaselineMediaFingerprint: inventory.BaselineMediaFingerprint, CandidateMediaFingerprint: inventory.CandidateMediaFingerprint, CurrentFiles: len(inventory.Files) - inventory.AddedPackageFiles, AddedPackageFiles: inventory.AddedPackageFiles, TotalBytes: inventory.TotalBytes}
			for _, rootName := range mediaRoots {
				record.MediaUnion.Volumes = append(record.MediaUnion.Volumes, customerServiceCandidateMediaVolumeState{RootName: rootName, StorageRoot: manager.Configuration.StorageRoots[rootName], State: "unallocated"})
			}
			if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
				return err
			}
			if err := writeDurableFile(filepath.Join(workspace.Directory, customerServiceCandidateMediaInventoryFilename), encodedInventory); err != nil {
				return err
			}
			return manager.constructCustomerServiceCompleteMediaUnionOnOwnedVolumes(ctx, record, inventory, filepath.Join(directory, "verified"))
		}
		validation, err := manager.Configuration.constructCustomerServiceSelectedFullRowsInRecordedTargetClone(ctx, source, target, plan, review, request, task.Manifest, filepath.Join(directory, "verified"), workspace)
		if err != nil {
			return err
		}
		if validation.SelectionSHA256 != inventory.SelectionSHA256 {
			return errors.New("database and complete media candidates used different effective selections")
		}
		if err := manager.verifyCustomerServiceCompleteCandidateMediaUnion(ctx, record, inventory); err != nil {
			return err
		}
		validation.State = "validated_not_switchable"
		record.Validation = validation
		if err := manager.verifyCustomerServiceCandidateBoundCleanupDispositionAudit(record); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Recheck ownership before sealing. The candidate remains inaccessible to the
	// application role and is never renamed to the live database.
	var oid int64
	if err := control.QueryRowContext(ctx, `SELECT oid::bigint FROM pg_database WHERE datname=$1 AND pg_get_userbyid(datdba)=$2`, record.Identity.DatabaseName, customerServiceDisposableCandidateOwner).Scan(&oid); err != nil || oid != record.Identity.DatabaseOID {
		return nil, errors.New("candidate identity changed before sealing")
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(record.Identity.DatabaseName)+` ALLOW_CONNECTIONS false`); err != nil {
		return nil, err
	}
	if _, err := control.ExecContext(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datid=$1`, oid); err != nil {
		return nil, err
	}
	if err := os.Remove(filepath.Join(workspace.Directory, "current-target-snapshot.dump")); err != nil {
		return nil, err
	}
	if err := syncDirectory(workspace.Directory); err != nil {
		return nil, err
	}
	if err := manager.Configuration.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	fence, err := os.ReadFile(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename))
	var owner struct {
		TaskID string `json:"task_id"`
	}
	if err != nil || json.Unmarshal(fence, &owner) != nil || owner.TaskID != id {
		return nil, errors.New("candidate maintenance fence changed before sealing")
	}
	for _, databaseName := range []string{manager.Configuration.DatabaseName, task.StageDatabase} {
		if err := assertNoUnfencedDatabaseSessions(ctx, control, databaseName); err != nil {
			return nil, err
		}
	}
	record.State, record.ErrorCode = "sealed_not_switchable", ""
	if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
		return nil, err
	}
	return record, nil
}

// Safe for retries after any construction/disposal interruption. A missing OID
// after allocation intent is never inferred from the name, even when it matches.
func (cfg Configuration) discardCustomerServiceRecordedCandidateConstructionWorkspace(ctx context.Context, workspace *customerServiceRecordedCandidateConstructionWorkspace) error {
	root, err := filepath.Abs(cfg.Workspace)
	if err != nil {
		return err
	}
	directory, err := filepath.Abs(workspace.Directory)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(root, directory)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return errors.New("candidate cleanup directory is outside its workspace")
	}
	ancestor := root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		ancestor = filepath.Join(ancestor, component)
		info, err := os.Lstat(ancestor)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("candidate cleanup path contains a non-directory or symbolic link")
		}
	}
	info, err := os.Lstat(directory)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
		return errors.New("candidate cleanup directory is not a private directory")
	}
	if workspace.Identity.State != "allocation_not_attempted" || workspace.Identity.DatabaseOID != 0 {
		control, err := cfg.connectDatabase(ctx, "postgres")
		if err != nil {
			return err
		}
		defer control.Close()
		if err := cleanupCustomerServiceDisposableCandidateDatabaseByOID(ctx, control, workspace.Identity); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(directory); err != nil {
		return errors.New("candidate workspace cleanup failed; identity is retained")
	}
	return syncDirectory(filepath.Dir(directory))
}

func (manager *Manager) discardCustomerServicePersistentCandidateWhileOperationLocked(ctx context.Context, task *Task) error {
	if err := manager.requireCustomerServiceIndependentPairedSwitchRolledBack(task.ID); err != nil {
		return err
	}
	record, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	if err != nil || record == nil || record.State == "discarded_not_switchable" {
		return err
	}
	if err := manager.verifyCustomerServicePairedRollbackBeforeCandidateDisposal(ctx, record); err != nil {
		return err
	}
	if task.Journal != nil {
		return errors.New("candidate disposal refuses a task with a live switch journal")
	}
	record.State = "discard_required"
	if err := manager.saveCustomerServicePersistentCandidateLifecycle(record); err != nil {
		return err
	}
	workspace, err := manager.customerServicePersistentCandidateConstructionWorkspace(record)
	if err == nil {
		err = manager.cleanupCustomerServiceCandidateMediaVolumesByRecordedIdentity(record)
	}
	if err == nil {
		err = manager.Configuration.discardCustomerServiceRecordedCandidateConstructionWorkspace(ctx, workspace)
	}
	if err != nil {
		record.State, record.ErrorCode = "inspection_required", "candidate_identity_or_cleanup_requires_inspection"
		return errors.Join(err, manager.saveCustomerServicePersistentCandidateLifecycle(record))
	}
	record.State, record.ErrorCode = "discarded_not_switchable", ""
	if record.Validation != nil {
		record.Validation.CandidateDeleted = true
	}
	return manager.saveCustomerServicePersistentCandidateLifecycle(record)
}

func (manager *Manager) discardCustomerServicePersistentCandidate(ctx context.Context, id string) error {
	if !manager.operationMutex.TryLock() {
		return errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return err
	}
	return manager.discardCustomerServicePersistentCandidateWhileOperationLocked(ctx, task)
}

// Startup never resumes/promotes a candidate and never releases a fence. Failure
// retains the record; release/apply/TTL independently enforce its discard obligation.
func (manager *Manager) discardInterruptedCustomerServicePersistentCandidates(tasks []*Task) {
	for _, task := range tasks {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := manager.discardCustomerServicePersistentCandidateWhileOperationLocked(ctx, task)
		cancel()
		if err != nil {
			_, _ = os.Stderr.WriteString("customer-service candidate requires inspection; its identity and maintenance fence are retained\n")
		}
	}
}

// Earlier disposable rehearsals used a separate temp directory. Recover only
// their explicit identity; unknown/malformed/ambiguous directories remain intact.
func (manager *Manager) discardInterruptedCustomerServiceDisposableRehearsals(ctx context.Context) error {
	entries, err := os.ReadDir(manager.Configuration.Workspace)
	if err != nil {
		return err
	}
	var cleanupErrors error
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "customer-service-candidate-rehearsal-") {
			continue
		}
		directory := filepath.Join(manager.Configuration.Workspace, entry.Name())
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			cleanupErrors = errors.Join(cleanupErrors, errors.New("candidate rehearsal directory requires inspection"))
			continue
		}
		filename := filepath.Join(directory, "customer-service-disposable-candidate-database-identity.json")
		info, err := os.Lstat(filename)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 4096 {
			cleanupErrors = errors.Join(cleanupErrors, errors.New("candidate rehearsal identity requires inspection"))
			continue
		}
		data, err := os.ReadFile(filename)
		var identity customerServiceDisposableCandidateDatabaseIdentity
		if err != nil || decodeStrictJSON(data, &identity) != nil || identity.Version != 1 || !strings.HasPrefix(identity.DatabaseName, "local_backup_candidate_") || !safeDatabaseName.MatchString(identity.DatabaseName) || identity.DatabaseName == manager.Configuration.DatabaseName {
			cleanupErrors = errors.Join(cleanupErrors, errors.New("candidate rehearsal identity is invalid"))
			continue
		}
		workspace := &customerServiceRecordedCandidateConstructionWorkspace{Directory: directory, Identity: identity}
		cleanupErrors = errors.Join(cleanupErrors, manager.Configuration.discardCustomerServiceRecordedCandidateConstructionWorkspace(ctx, workspace))
	}
	return cleanupErrors
}
