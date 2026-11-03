package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
)

type CustomerProfileModuleFormalRestoreRequest struct {
	Confirmation          string `json:"confirmation"`
	ExpectedPreviewSHA256 string `json:"expected_preview_sha256"`
}

func validateCustomerProfileModuleFormalRestorePreview(task *Task, expected string) error {
	if task == nil || task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != selectedBusinessModulesBackupScope || !reflect.DeepEqual(task.Manifest.SelectedModules, []string{"customers"}) || task.State != "module_isolated_restore_verified" || taskHasAnyTransactionalBusinessModuleRestoreProgress(task) || task.Journal != nil || task.StageDatabase != "" || task.SelectedModuleDisposableStage != nil || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("请使用未过期的单独客户包并完成本版隔离演练")
	}
	report := task.SelectedModuleRestorePreview
	if report == nil || report.Binding == nil || report.SHA256 == "" || report.SHA256 != expected || !report.TargetUnchanged || !report.DisposableStageRemoved || !report.IsolatedRestoreVerified || report.OnlineOverwriteAvailable || !reflect.DeepEqual(report.Modules, []string{"customers"}) || report.CustomerProfileRestoreEligibility == nil || !report.CustomerProfileRestoreEligibility.Eligible || report.CustomerProfileRestoreEligibility.Policy != customerProfileModuleFormalRestorePolicy {
		return errors.New("客户恢复资格或绑定的预览缺失")
	}
	copyReport := *report
	copyReport.SHA256 = ""
	digest, err := businessModuleRecoveryCanonicalJSONSHA256(copyReport)
	if err != nil || digest != expected {
		return errors.New("客户恢复预览摘要已变化")
	}
	return nil
}

// Supported separate customer profile entry: bounded inserts in the current
// database with the shared atomic receipt/attachment recovery lifecycle. This is
// not account registration, arbitrary identity overwrite or a candidate switch.
func (manager *Manager) StartCustomerProfileModuleFormalRestore(id, actor string, request CustomerProfileModuleFormalRestoreRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if request.Confirmation != "RESTORE CUSTOMER PROFILES "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id || strings.TrimSpace(actor) == "" {
		return nil, errors.New("请明确确认本次客户资料恢复和维护窗口")
	}
	if err := validateCustomerProfileModuleFormalRestorePreview(task, request.ExpectedPreviewSHA256); err != nil {
		return nil, err
	}
	if err := manager.startTaskOperation(task, "restoring_missing_customer_profiles", func(ctx context.Context, task *Task) error {
		return manager.restoreCustomerProfilesWithAtomicReceiptAndRecoverableFiles(ctx, task, actor, nil)
	}); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

// checkpoint injects failures only in disposable integration tests.
func (manager *Manager) restoreCustomerProfilesWithAtomicReceiptAndRecoverableFiles(ctx context.Context, task *Task, actor string, checkpoint func(string) error) (resultErr error) {
	cfg := manager.Configuration
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return err
	}
	if task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("客户上传包在恢复前已过期")
	}
	snapshot, extracted, cleanup, err := manager.readAuthenticatedTransactionalBusinessModuleRestorePackage(task, "customers")
	if err != nil {
		return err
	}
	defer cleanup()
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	defer func() {
		if task.CustomerProfileModuleRestore == nil {
			resultErr = errors.Join(resultErr, manager.releaseMaintenance(task))
		}
	}()
	live, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(live.Close)
	if err := assertNoUnfencedDatabaseSessions(ctx, live, cfg.DatabaseName); err != nil {
		return err
	}
	media, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return err
	}
	binding, err := manager.captureBusinessModuleRestoreSnapshotBinding(ctx, live, task, media)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(binding, task.SelectedModuleRestorePreview.Binding) {
		return errors.New("当前客户恢复目标已变化，请重新演练")
	}
	fence, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename))
	if err != nil {
		return err
	}
	boundary, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json"))
	if err != nil {
		return err
	}
	progress := &BusinessModuleTransactionalRestoreProgress{Version: 1, TaskID: task.ID, Operator: actor, Phase: "installing_files", PreviewSHA256: task.SelectedModuleRestorePreview.SHA256, PackageSHA256: task.CiphertextSHA256, ConfigurationSHA256: binding.ConfigurationSHA256, RuntimeBoundarySHA256: transactionalBusinessModuleRestoreBytesSHA256(boundary), FenceSHA256: transactionalBusinessModuleRestoreBytesSHA256(fence), DatabaseOID: binding.TargetDatabaseOID, BeforeDatabaseFingerprint: binding.TargetDatabaseFingerprint, BeforeMediaFingerprint: media, Files: []BusinessModuleTransactionalRestoreFile{}, CreatedDirectories: []string{}, StartedAt: time.Now().UTC()}
	for _, entry := range task.Manifest.Files {
		if entry.Path == selectedBusinessModulesRowsFilename {
			continue
		}
		rootName, key, err := transactionalBusinessModuleRestoreMediaPath(entry.Path)
		if err != nil {
			return err
		}
		filename := filepath.Join(cfg.StorageRoots[rootName], filepath.FromSlash(key))
		if _, err := os.Lstat(filename); errors.Is(err, os.ErrNotExist) {
			progress.Files = append(progress.Files, BusinessModuleTransactionalRestoreFile{Identity: entry})
		} else if err != nil {
			return err
		} else {
			identity, err := fileIdentity(filename, entry.Path)
			if err != nil || identity != entry {
				return errors.New("客户包引用文件的当前内容冲突")
			}
		}
	}
	task.CustomerProfileModuleRestore = progress
	if err := manager.saveTransactionalBusinessModuleRestoreProgress(task); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("restore_intent_persisted"); err != nil {
			return err
		}
	}
	if err := manager.verifyTransactionalBusinessModuleRestoreFenceAndRuntime(task); err != nil {
		return err
	}
	tx, err := live.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	if _, err := tx.ExecContext(ctx, `SET LOCAL timezone='UTC'; SET LOCAL search_path=pg_catalog,public`); err != nil {
		return err
	}
	if err := requireRegisteredStorefrontTablesForProductCatalogModule(ctx, tx); err != nil {
		return err
	}
	observation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
	if err != nil {
		return err
	}
	if observation.Fingerprint != progress.BeforeDatabaseFingerprint {
		return errors.New("客户目标在事务写入前发生变化")
	}
	report := &SelectedBusinessModuleIsolatedRestorePreview{Modules: snapshot.Modules, Tables: []SelectedBusinessModuleTableRestorePreview{}}
	missing, blocked, err := inspectSelectedBusinessModuleRowsForRestore(ctx, tx, snapshot, observation, report)
	if err != nil {
		return err
	}
	eligibility, err := inspectCustomerProfileModuleRestoreEligibility(ctx, tx, snapshot, missing)
	if err != nil {
		return err
	}
	if blocked || !eligibility.Eligible {
		return errors.New("客户身份、积分闭环或共享依赖存在冲突，请重新检查")
	}
	for table, rows := range missing {
		if len(rows) > 0 {
			if err := validateCustomerProfileModuleInsertOnlyWriteContract(ctx, tx, snapshot, table, observation); err != nil {
				return err
			}
		}
	}
	if err := manager.installTransactionalBusinessModuleRestoreFiles(ctx, task, extracted, checkpoint); err != nil {
		return err
	}
	progress.RestoredRows, err = restoreCustomerProfileModuleMissingRowsWithNormalConstraintsAndVerifiedPreservation(ctx, tx, snapshot, observation, missing, true)
	if err != nil {
		return err
	}
	afterObservation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
	if err != nil {
		return err
	}
	if progress.RestoredRows != task.SelectedModuleRestorePreview.RestoredRows {
		return errors.New("客户实际恢复行数与演练不一致")
	}
	if err := validateSelectedBusinessModuleIsolatedInsertWriteContract(ctx, tx, "audit_logs", afterObservation); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `LOCK TABLE public.audit_logs IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(id),0)+1 FROM public.audit_logs`).Scan(&progress.AuditID); err != nil {
		return err
	}
	receipt, err := json.Marshal(struct {
		Version       int    `json:"version"`
		TaskID        string `json:"task_id"`
		PackageSHA256 string `json:"package_sha256"`
		PreviewSHA256 string `json:"preview_sha256"`
		RestoredRows  int    `json:"restored_rows"`
	}{1, task.ID, progress.PackageSHA256, progress.PreviewSHA256, progress.RestoredRows})
	if err != nil {
		return err
	}
	progress.AuditReceipt = string(receipt)
	sequences, err := readBusinessModuleRestoreSequenceStates(ctx, tx)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO public.audit_logs(id,username,action,resource,method,path,status,new_value,created_at) VALUES($1,$2,'restore','customer_profile_module_formal_restore','POST',$3,'success',$4,$5)`, progress.AuditID, actor, "/api/tasks/"+task.ID+"/customer-profile-module-formal-restore", progress.AuditReceipt, progress.StartedAt); err != nil {
		return err
	}
	if err := advanceBusinessModuleRestoreOwnedSequencesTransactionallyAndVerifyPreservation(ctx, tx, sequences, map[string][]json.RawMessage{"audit_logs": {json.RawMessage(`{}`)}}); err != nil {
		return err
	}
	_, progress.AfterDatabaseFingerprint, _, err = captureDatabaseTransactionInventory(ctx, tx)
	if err != nil {
		return err
	}
	progress.AfterMediaFingerprint, err = cfg.mediaFingerprint(ctx)
	if err != nil {
		return err
	}
	progress.Phase = "committing_rows"
	task.Phase = "customer_profile_rows_and_audit_ready_for_atomic_commit"
	if err := manager.saveTransactionalBusinessModuleRestoreProgress(task); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("rows_and_sequences_before_commit"); err != nil {
			return err
		}
	}
	if err := manager.verifyTransactionalBusinessModuleRestoreFenceAndRuntime(task); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("rows_committed_before_acknowledgement"); err != nil {
			return err
		}
	}
	progress.Phase = "rows_committed"
	if err := manager.saveTransactionalBusinessModuleRestoreProgress(task); err != nil {
		return err
	}
	return manager.completeOrCancelInterruptedTransactionalBusinessModuleRestoreWithHeldLease(ctx, task, live, checkpoint)
}

func (manager *Manager) StartInterruptedCustomerProfileModuleRestoreRecovery(id, actor, confirmation string) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.CustomerProfileModuleRestore == nil || countTransactionalBusinessModuleRestoreProgressRecords(task) != 1 || strings.TrimSpace(actor) == "" || confirmation != "RECOVER CUSTOMER PROFILES "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id {
		return nil, errors.New("请明确确认本次客户资料中断恢复核验")
	}
	if err := validateTransactionalBusinessModuleRestoreProgress(task); err != nil {
		return nil, err
	}
	if !taskHasUnresolvedTransactionalBusinessModuleRestore(task) || task.State == "running" || task.State == "queued" {
		return nil, errors.New("只有中断的客户恢复任务可继续核验")
	}
	if err := manager.startTaskOperation(task, "recovering_interrupted_customer_profile_restore", func(ctx context.Context, task *Task) error {
		fenced, err := maintenancegate.IsApplicationFenced(manager.Configuration.MaintenanceDirectory)
		if err != nil {
			return err
		}
		if !fenced && task.CustomerProfileModuleRestore.Phase == "releasing" {
			return manager.acknowledgeTransactionalBusinessModuleRestoreServiceRelease(task)
		}
		if err := manager.verifyTransactionalBusinessModuleRestoreFenceAndRuntime(task); err != nil {
			return err
		}
		lease, err := manager.acquireMaintenance(ctx, task)
		if err != nil {
			return err
		}
		defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
		live, err := manager.Configuration.connectDatabase(ctx, manager.Configuration.DatabaseName)
		if err != nil {
			return err
		}
		defer func(closeResource func() error) { _ = closeResource() }(live.Close)
		return manager.completeOrCancelInterruptedTransactionalBusinessModuleRestoreWithHeldLease(ctx, task, live, nil)
	}); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}
