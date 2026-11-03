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

type ConfigurationModuleSiteContentFormalRestoreRequest struct {
	Confirmation          string `json:"confirmation"`
	ExpectedPreviewSHA256 string `json:"expected_preview_sha256"`
}

func validateConfigurationModuleSiteContentFormalRestorePreview(task *Task, expected string) error {
	if task == nil || task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != selectedBusinessModulesBackupScope || !reflect.DeepEqual(task.Manifest.SelectedModules, []string{"configuration"}) || task.State != "module_isolated_restore_verified" || taskHasAnyTransactionalBusinessModuleRestoreProgress(task) || task.Journal != nil || task.StageDatabase != "" || task.SelectedModuleDisposableStage != nil || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("请使用未过期的单独商城配置包并完成本版隔离演练")
	}
	report := task.SelectedModuleRestorePreview
	if report == nil || report.Binding == nil || report.SHA256 == "" || report.SHA256 != expected || !report.TargetUnchanged || !report.DisposableStageRemoved || !report.IsolatedRestoreVerified || report.OnlineOverwriteAvailable || !reflect.DeepEqual(report.Modules, []string{"configuration"}) || report.ConfigurationSiteContentRestoreEligibility == nil || !report.ConfigurationSiteContentRestoreEligibility.Eligible || report.ConfigurationSiteContentRestoreEligibility.Policy != configurationModuleSiteContentRestorePolicy {
		return errors.New("配置内容恢复资格或绑定的预览缺失")
	}
	copyReport := *report
	copyReport.SHA256 = ""
	digest, err := businessModuleRecoveryCanonicalJSONSHA256(copyReport)
	if err != nil || digest != expected {
		return errors.New("配置内容恢复预览摘要已变化")
	}
	return nil
}

// Restore missing site content only; other configuration and business facts remain current.
func (manager *Manager) StartConfigurationModuleSiteContentFormalRestore(id, actor string, request ConfigurationModuleSiteContentFormalRestoreRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if request.Confirmation != "RESTORE CONFIGURATION CONTENT "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id || strings.TrimSpace(actor) == "" {
		return nil, errors.New("请明确确认本次网站内容恢复和维护窗口")
	}
	if err := validateConfigurationModuleSiteContentFormalRestorePreview(task, request.ExpectedPreviewSHA256); err != nil {
		return nil, err
	}
	if err := manager.startTaskOperation(task, "restoring_missing_configuration_site_contents", func(ctx context.Context, task *Task) error {
		return manager.restoreConfigurationModuleMissingSiteContentWithAtomicReceipt(ctx, task, actor, nil)
	}); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

// checkpoint injects failures only in disposable integration tests.
func (manager *Manager) restoreConfigurationModuleMissingSiteContentWithAtomicReceipt(ctx context.Context, task *Task, actor string, checkpoint func(string) error) (resultErr error) {
	cfg := manager.Configuration
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return err
	}
	if task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("配置上传包在恢复前已过期")
	}
	snapshot, extracted, cleanup, err := manager.readAuthenticatedTransactionalBusinessModuleRestorePackage(task, "configuration")
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
		if task.ConfigurationModuleSiteContentRestore == nil {
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
		return errors.New("当前配置内容恢复目标已变化，请重新演练")
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
	if err := cfg.verifyConfigurationModuleReferencedFilesAlreadyMatchPackage(task.Manifest); err != nil {
		return err
	}

	task.ConfigurationModuleSiteContentRestore = progress
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
		return errors.New("配置目标在事务写入前发生变化")
	}
	report := &SelectedBusinessModuleIsolatedRestorePreview{Modules: snapshot.Modules, Tables: []SelectedBusinessModuleTableRestorePreview{}}
	missing, blocked, err := inspectSelectedBusinessModuleRowsForRestore(ctx, tx, snapshot, observation, report)
	if err != nil {
		return err
	}
	eligibility, err := cfg.inspectConfigurationModuleMissingSiteContentRestoreEligibility(ctx, tx, snapshot, missing, task.Manifest, report)
	if err != nil {
		return err
	}
	if blocked || !eligibility.Eligible {
		return errors.New("配置范围、共享媒体或内容存在冲突，请重新检查")
	}
	for table, rows := range missing {
		if len(rows) > 0 {
			if err := validateConfigurationModuleMissingSiteContentInsertContract(ctx, tx, snapshot, table, observation); err != nil {
				return err
			}
		}
	}
	if err := manager.installTransactionalBusinessModuleRestoreFiles(ctx, task, extracted, checkpoint); err != nil {
		return err
	}
	progress.RestoredRows, err = restoreConfigurationModuleMissingBusinessConfigurationRowsWithNormalConstraintsAndVerifiedPreservation(ctx, tx, snapshot, observation, missing)
	if err != nil {
		return err
	}
	afterObservation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
	if err != nil {
		return err
	}
	if progress.RestoredRows != task.SelectedModuleRestorePreview.RestoredRows {
		return errors.New("配置实际恢复行数与演练不一致")
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
	if _, err := tx.ExecContext(ctx, `INSERT INTO public.audit_logs(id,username,action,resource,method,path,status,new_value,created_at) VALUES($1,$2,'restore','configuration_module_site_content_restore','POST',$3,'success',$4,$5)`, progress.AuditID, actor, "/api/tasks/"+task.ID+"/configuration-site-content-formal-restore", progress.AuditReceipt, progress.StartedAt); err != nil {
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
	task.Phase = "configuration_site_content_rows_and_audit_ready_for_atomic_commit"
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

func (manager *Manager) StartInterruptedConfigurationModuleSiteContentRestoreRecovery(id, actor, confirmation string) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.ConfigurationModuleSiteContentRestore == nil || countTransactionalBusinessModuleRestoreProgressRecords(task) != 1 || strings.TrimSpace(actor) == "" || confirmation != "RECOVER CONFIGURATION CONTENT "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id {
		return nil, errors.New("请明确确认本次网站内容中断恢复核验")
	}
	if err := validateTransactionalBusinessModuleRestoreProgress(task); err != nil {
		return nil, err
	}
	if !taskHasUnresolvedTransactionalBusinessModuleRestore(task) || task.State == "running" || task.State == "queued" {
		return nil, errors.New("只有中断的配置内容恢复任务可继续核验")
	}
	if err := manager.startTaskOperation(task, "recovering_interrupted_configuration_site_content_restore", func(ctx context.Context, task *Task) error {
		fenced, err := maintenancegate.IsApplicationFenced(manager.Configuration.MaintenanceDirectory)
		if err != nil {
			return err
		}
		if !fenced && task.ConfigurationModuleSiteContentRestore.Phase == "releasing" {
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
