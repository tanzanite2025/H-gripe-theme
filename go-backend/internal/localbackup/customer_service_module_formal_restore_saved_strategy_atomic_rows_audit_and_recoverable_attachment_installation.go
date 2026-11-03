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

const customerServiceModuleFormalRestorePolicy = "restore_selected_conversation_rows_preserve_current_shared_facts_v1"

type CustomerServiceModuleFormalRestoreRequest struct {
	Confirmation            string `json:"confirmation"`
	ExpectedChoiceSHA256    string `json:"expected_choice_sha256"`
	ExpectedRehearsalSHA256 string `json:"expected_rehearsal_sha256"`
}

func validateCustomerServiceModuleFormalRestoreAuthorization(task *Task, choiceSHA256, rehearsalSHA256 string) error {
	if task == nil || task.State != "module_chat_candidate_rehearsal_verified" || taskHasAnyTransactionalBusinessModuleRestoreProgress(task) {
		return errors.New("请先保存聊天恢复策略并完成本版完整行演练")
	}
	choice, rehearsal := task.SelectedCustomerServiceRestoreChoice, task.SelectedCustomerServiceCandidateRehearsal
	if choice == nil || choice.Version != 1 || choice.State != "recorded_not_executable" || choice.ExecutionAvailable || choice.TaskID != task.ID || strings.TrimSpace(choice.Operator) == "" || choice.SHA256 == "" || choice.SHA256 != choiceSHA256 {
		return errors.New("所选聊天策略缺失或已变化")
	}
	copyChoice := *choice
	copyChoice.SHA256 = ""
	actualChoiceSHA256, err := businessModuleRecoveryCanonicalJSONSHA256(copyChoice)
	if err != nil || actualChoiceSHA256 != choiceSHA256 {
		return errors.New("所选聊天策略摘要不一致")
	}
	if err := validateSelectedCustomerServiceModuleTaskAndPreview(task, choice.PreviewSHA256); err != nil {
		return err
	}
	if !reflect.DeepEqual(choice.Binding, *task.SelectedModuleRestorePreview.Binding) {
		return errors.New("聊天策略与原始预览不对应")
	}
	if rehearsal == nil || rehearsal.Binding == nil || rehearsal.SHA256 == "" || rehearsal.SHA256 != rehearsalSHA256 || rehearsal.SavedChoiceSHA256 != choiceSHA256 || rehearsal.CustomerServiceFormalRestorePolicy != customerServiceModuleFormalRestorePolicy || !reflect.DeepEqual(rehearsal.Modules, []string{"customer_service"}) || !reflect.DeepEqual(*rehearsal.Binding, choice.Binding) || !rehearsal.TargetUnchanged || !rehearsal.DisposableStageRemoved || !rehearsal.IsolatedRestoreVerified || rehearsal.OnlineOverwriteAvailable || rehearsal.CandidateDatabaseFingerprint == "" {
		return errors.New("请使用当前所选策略重新完成本版聊天完整行演练")
	}
	copyRehearsal := *rehearsal
	copyRehearsal.SHA256 = ""
	actualRehearsalSHA256, err := businessModuleRecoveryCanonicalJSONSHA256(copyRehearsal)
	if err != nil || actualRehearsalSHA256 != rehearsalSHA256 {
		return errors.New("聊天演练报告摘要不一致")
	}
	return validateSelectedCustomerServiceModuleChosenStrategyForFullRowRehearsal(choice, rehearsal)
}

// Supported separate customer-service module entry: restore the selected four
// conversation tables in the current database transaction. The older persistent
// candidate and independent paired-switch journals are not execution inputs.
func (manager *Manager) StartCustomerServiceModuleFormalRestore(id, actor string, request CustomerServiceModuleFormalRestoreRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(actor) == "" || request.Confirmation != "RESTORE CHAT MODULE "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id+"/"+request.ExpectedChoiceSHA256 {
		return nil, errors.New("请明确确认本次聊天策略恢复和商城维护窗口")
	}
	if err := validateCustomerServiceModuleFormalRestoreAuthorization(task, request.ExpectedChoiceSHA256, request.ExpectedRehearsalSHA256); err != nil {
		return nil, err
	}
	if err := manager.startTaskOperation(task, "restoring_customer_service_module_with_saved_strategy", func(ctx context.Context, task *Task) error {
		return manager.restoreCustomerServiceModuleWithSavedStrategyAtomicReceiptAndRecoverableFiles(ctx, task, actor, nil)
	}); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

func (cfg Configuration) inspectCustomerServiceModuleFormalRestoreStrategyWithoutSharedWrites(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, task *Task, extracted string) error {
	report := &SelectedBusinessModuleIsolatedRestorePreview{Modules: snapshot.Modules, Tables: []SelectedBusinessModuleTableRestorePreview{}}
	missing, _, err := inspectSelectedBusinessModuleRowsForRestore(ctx, tx, snapshot, observation, report)
	if err != nil {
		return err
	}
	// The disposable metadata adapter is never allowed to insert shared data on
	// a formal restore. Reject first, before invoking the shared domain checks.
	for table, rows := range missing {
		if len(rows) > 0 && !customerServiceRecoveryPlanProposedRowReplacementAllowed(table) {
			return errors.New("聊天正式恢复不能补造共享媒体、上传收据或事件事实")
		}
	}
	report.CustomerServiceOverwritePreview, err = cfg.inspectSelectedCustomerServiceModuleBoundedOverwriteStrategies(ctx, tx, snapshot, observation, missing, task.Manifest, extracted, report)
	if err != nil {
		return err
	}
	return validateSelectedCustomerServiceModuleChosenStrategyForFullRowRehearsal(task.SelectedCustomerServiceRestoreChoice, report)
}

// Checkpoints exercise interruption boundaries only in disposable tests. The
// importer writes complete original rows and never invokes business services.
func (manager *Manager) restoreCustomerServiceModuleWithSavedStrategyAtomicReceiptAndRecoverableFiles(ctx context.Context, task *Task, actor string, checkpoint func(string) error) (resultErr error) {
	cfg := manager.Configuration
	choice, rehearsal := task.SelectedCustomerServiceRestoreChoice, task.SelectedCustomerServiceCandidateRehearsal
	if choice == nil || rehearsal == nil || taskHasAnyTransactionalBusinessModuleRestoreProgress(task) || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("聊天恢复授权缺失、已开始或工作副本已到期")
	}
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return err
	}
	snapshot, extracted, cleanup, err := manager.readAuthenticatedTransactionalBusinessModuleRestorePackage(task, "customer_service")
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
		if task.CustomerServiceModuleRestore == nil {
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
	mediaBefore, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return err
	}
	binding, err := manager.captureBusinessModuleRestoreSnapshotBinding(ctx, live, task, mediaBefore)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(*binding, choice.Binding) || rehearsal.Binding == nil || !reflect.DeepEqual(*binding, *rehearsal.Binding) {
		return errors.New("聊天恢复的数据库、媒体或运行边界已变化，请重新预览和演练")
	}
	fence, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, maintenancegate.FenceFilename))
	if err != nil {
		return err
	}
	boundary, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json"))
	if err != nil {
		return err
	}
	progress := &BusinessModuleTransactionalRestoreProgress{Version: 1, TaskID: task.ID, Operator: actor, Phase: "installing_files", PreviewSHA256: rehearsal.SHA256, ChoiceSHA256: choice.SHA256, PackageSHA256: task.CiphertextSHA256, ConfigurationSHA256: binding.ConfigurationSHA256, RuntimeBoundarySHA256: transactionalBusinessModuleRestoreBytesSHA256(boundary), FenceSHA256: transactionalBusinessModuleRestoreBytesSHA256(fence), DatabaseOID: binding.TargetDatabaseOID, BeforeDatabaseFingerprint: binding.TargetDatabaseFingerprint, BeforeMediaFingerprint: mediaBefore, Files: []BusinessModuleTransactionalRestoreFile{}, CreatedDirectories: []string{}, StartedAt: time.Now().UTC()}
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
		} else if identity, err := fileIdentity(filename, entry.Path); err != nil || identity != entry {
			return errors.New("聊天附件路径已有不同内容，不能覆盖共享文件")
		}
	}
	task.CustomerServiceModuleRestore = progress
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
		return errors.New("聊天目标在事务写入前发生变化")
	}
	if err := cfg.inspectCustomerServiceModuleFormalRestoreStrategyWithoutSharedWrites(ctx, tx, snapshot, observation, task, extracted); err != nil {
		return err
	}
	if err := manager.installTransactionalBusinessModuleRestoreFiles(ctx, task, extracted, checkpoint); err != nil {
		return err
	}
	progress.RestoredRows, progress.ReplacedRows, err = restoreSelectedCustomerServiceModuleChosenCompleteRowsWithExplicitTransactionPolicy(ctx, tx, snapshot, observation, choice.Policy, true)
	if err != nil {
		return err
	}
	if progress.RestoredRows != rehearsal.RestoredRows || progress.ReplacedRows != rehearsal.ReplacedRows {
		return errors.New("聊天正式写入行数与所选策略演练不一致")
	}
	afterObservation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
	if err != nil {
		return err
	}
	if err := cfg.inspectCustomerServiceModuleFormalRestoreStrategyWithoutSharedWrites(ctx, tx, snapshot, afterObservation, task, extracted); err != nil {
		return err
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
		Version         int    `json:"version"`
		TaskID          string `json:"task_id"`
		PackageSHA256   string `json:"package_sha256"`
		ChoiceSHA256    string `json:"choice_sha256"`
		RehearsalSHA256 string `json:"rehearsal_sha256"`
		Policy          string `json:"policy"`
		RestoredRows    int    `json:"restored_rows"`
		ReplacedRows    int    `json:"replaced_rows"`
	}{1, task.ID, progress.PackageSHA256, progress.ChoiceSHA256, progress.PreviewSHA256, choice.Policy, progress.RestoredRows, progress.ReplacedRows})
	if err != nil {
		return err
	}
	progress.AuditReceipt = string(receipt)
	sequences, err := readBusinessModuleRestoreSequenceStates(ctx, tx)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO public.audit_logs(id,username,action,resource,method,path,status,new_value,created_at) VALUES($1,$2,'restore','customer_service_module_formal_restore','POST',$3,'success',$4,$5)`, progress.AuditID, actor, "/api/tasks/"+task.ID+"/customer-service-module-formal-restore", progress.AuditReceipt, progress.StartedAt); err != nil {
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
	progress.Phase, task.Phase = "committing_rows", "customer_service_rows_and_audit_ready_for_atomic_commit"
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

func (manager *Manager) StartInterruptedCustomerServiceModuleRestoreRecovery(id, actor, confirmation string) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.CustomerServiceModuleRestore == nil || task.OrderModuleRestore != nil || task.CustomerProfileModuleRestore != nil || task.MediaModuleRestore != nil || task.ConfigurationModuleSiteContentRestore != nil || strings.TrimSpace(actor) == "" || confirmation != "RECOVER CHAT RESTORE "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id {
		return nil, errors.New("请明确确认本次聊天中断恢复核验")
	}
	if err := validateTransactionalBusinessModuleRestoreProgress(task); err != nil {
		return nil, err
	}
	if !taskHasUnresolvedTransactionalBusinessModuleRestore(task) || task.State == "running" || task.State == "queued" {
		return nil, errors.New("只有中断且尚未核验的聊天恢复任务可继续处理")
	}
	if err := manager.startTaskOperation(task, "recovering_interrupted_customer_service_module_restore", func(ctx context.Context, task *Task) error {
		fenced, err := maintenancegate.IsApplicationFenced(manager.Configuration.MaintenanceDirectory)
		if err != nil {
			return err
		}
		if !fenced && task.CustomerServiceModuleRestore.Phase == "releasing" {
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
