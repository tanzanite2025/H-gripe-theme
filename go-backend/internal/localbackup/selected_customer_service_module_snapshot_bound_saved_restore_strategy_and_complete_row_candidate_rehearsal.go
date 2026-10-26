package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

// No full customer rows, connection strings or credentials enter these records.
// Saving a choice authorizes only a disposable rehearsal, never an online switch.
type SelectedCustomerServiceModuleSnapshotBinding struct {
	Version                   int       `json:"version"`
	TaskID                    string    `json:"task_id"`
	PackageID                 string    `json:"package_id"`
	SiteID                    string    `json:"site_id"`
	Environment               string    `json:"environment"`
	ApplicationVersion        string    `json:"application_version"`
	MigrationVersion          int64     `json:"migration_version"`
	ManifestSHA256            string    `json:"manifest_sha256"`
	CiphertextSHA256          string    `json:"ciphertext_sha256"`
	TargetDatabaseOID         int64     `json:"target_database_oid"`
	TargetDatabaseFingerprint string    `json:"target_database_fingerprint"`
	TargetSchemaSHA256        string    `json:"target_schema_sha256"`
	TargetMediaFingerprint    string    `json:"target_media_fingerprint"`
	ConfigurationSHA256       string    `json:"configuration_sha256"`
	WorkCopyExpiresAt         time.Time `json:"work_copy_expires_at"`
}

type SelectedCustomerServiceModuleSavedRestoreChoice struct {
	Version            int                                          `json:"version"`
	State              string                                       `json:"state"`
	TaskID             string                                       `json:"task_id"`
	Operator           string                                       `json:"operator"`
	Policy             string                                       `json:"policy"`
	PreviewSHA256      string                                       `json:"preview_sha256"`
	Binding            SelectedCustomerServiceModuleSnapshotBinding `json:"binding"`
	RecordedAt         time.Time                                    `json:"recorded_at"`
	SHA256             string                                       `json:"sha256"`
	ExecutionAvailable bool                                         `json:"execution_available"`
}

func (manager *Manager) captureSelectedCustomerServiceModuleSnapshotBinding(ctx context.Context, live *sql.DB, task *Task, mediaFingerprint string) (*SelectedCustomerServiceModuleSnapshotBinding, error) {
	cfg := manager.Configuration
	binding := &SelectedCustomerServiceModuleSnapshotBinding{Version: 1, TaskID: task.ID, PackageID: task.Manifest.PackageID, SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, MigrationVersion: cfg.ExpectedMigrationVersion, CiphertextSHA256: task.CiphertextSHA256, TargetMediaFingerprint: mediaFingerprint, WorkCopyExpiresAt: task.ExpiresAt}
	var err error
	binding.ManifestSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(task.Manifest)
	if err != nil {
		return nil, err
	}
	binding.ConfigurationSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(cfg)
	if err != nil {
		return nil, err
	}
	if err := live.QueryRowContext(ctx, `SELECT oid::bigint FROM pg_database WHERE datname=current_database()`).Scan(&binding.TargetDatabaseOID); err != nil {
		return nil, err
	}
	tx, err := live.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	observation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
	if err != nil {
		return nil, err
	}
	if observation.Migration != cfg.ExpectedMigrationVersion {
		return nil, errors.New("目标数据库迁移版本已变化，请重新预览")
	}
	binding.TargetDatabaseFingerprint = observation.Fingerprint
	binding.TargetSchemaSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(struct{ Shapes, Indexes, WriteSideEffects any }{observation.Shapes, observation.Indexes, observation.WriteSideEffects})
	return binding, err
}

func validateSelectedCustomerServiceModuleTaskAndPreview(task *Task, previewSHA256 string) error {
	if task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != selectedBusinessModulesBackupScope || !reflect.DeepEqual(task.Manifest.SelectedModules, []string{"customer_service"}) || task.Journal != nil || task.StageDatabase != "" || task.SelectedModuleDisposableStage != nil || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("请选择未到期、已预览的单独聊天模块上传包")
	}
	if task.State != "module_isolated_restore_verified" && task.State != "module_restore_conflicts_found" && task.State != "module_chat_candidate_rehearsal_verified" {
		return errors.New("请先重新预览聊天模块恢复")
	}
	report := task.SelectedModuleRestorePreview
	if report == nil || report.Binding == nil || report.SHA256 == "" || report.SHA256 != previewSHA256 || !report.TargetUnchanged || !report.DisposableStageRemoved || report.OnlineOverwriteAvailable || !reflect.DeepEqual(report.Modules, []string{"customer_service"}) {
		return errors.New("聊天恢复预览缺失或已变化，请重新预览")
	}
	copyReport := *report
	copyReport.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(&copyReport)
	if err != nil || digest != previewSHA256 {
		return errors.New("聊天恢复预览摘要不一致，请重新预览")
	}
	if report.Binding.Version != 1 || report.Binding.TaskID != task.ID || report.Binding.PackageID != task.Manifest.PackageID || !report.Binding.WorkCopyExpiresAt.Equal(task.ExpiresAt) {
		return errors.New("聊天恢复预览绑定不一致，请重新预览")
	}
	return nil
}

func validateSelectedCustomerServiceModuleChosenStrategyForFullRowRehearsal(choice *SelectedCustomerServiceModuleSavedRestoreChoice, report *SelectedBusinessModuleIsolatedRestorePreview) error {
	if choice == nil || report == nil || report.CustomerServiceOverwritePreview == nil || !report.CustomerServiceOverwritePreview.PrerequisitesSatisfied || report.DependencyConflicts != 0 || report.FileConflicts != 0 {
		return errors.New("聊天恢复的共享依赖、文件或领域检查未通过")
	}
	var strategy *SelectedCustomerServiceModuleOverwriteStrategyInspection
	switch choice.Policy {
	case "use_package_conversation_rows":
		strategy = report.CustomerServiceOverwritePreview.PackageRows
	case "preserve_current_conflicting_conversation_rows":
		strategy = report.CustomerServiceOverwritePreview.PreserveCurrentRows
	default:
		return errors.New("请选择已登记的聊天恢复策略")
	}
	if strategy == nil || strategy.Policy != choice.Policy || !strategy.Complete || !strategy.DomainChecksSatisfied || strategy.BlockingIssues != 0 {
		return errors.New("所选聊天恢复策略存在阻断或尚未完成检查")
	}
	for _, table := range report.Tables {
		if table.UnsupportedReason != "" || table.ProtectedNewRows > 0 || !customerServiceRecoveryPlanProposedRowReplacementAllowed(table.Table) && (table.MissingRows > 0 || table.ConflictingRows > 0) {
			return errors.New("完整行候选暂不支持补入或覆盖共享元数据、执行事实")
		}
	}
	return nil
}

func (manager *Manager) SaveSelectedCustomerServiceModuleRestoreChoice(ctx context.Context, id, operator, previewSHA256, policy string) (*Task, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("another backup or restore operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if err := validateSelectedCustomerServiceModuleTaskAndPreview(task, previewSHA256); err != nil {
		return nil, err
	}
	if strings.TrimSpace(operator) == "" {
		return nil, errors.New("恢复策略必须记录独立维护操作员")
	}
	choice := &SelectedCustomerServiceModuleSavedRestoreChoice{Version: 1, State: "recorded_not_executable", TaskID: id, Operator: operator, Policy: policy, PreviewSHA256: previewSHA256, Binding: *task.SelectedModuleRestorePreview.Binding, RecordedAt: time.Now().UTC()}
	if err := validateSelectedCustomerServiceModuleChosenStrategyForFullRowRehearsal(choice, task.SelectedModuleRestorePreview); err != nil {
		return nil, err
	}
	if err := manager.requireNoUnresolvedProductCatalogPairedRehearsalExceptTask(id); err != nil {
		return nil, err
	}
	if err := manager.requireCustomerServicePersistentCandidateDiscarded(id); err != nil {
		return nil, err
	}
	if err := manager.verifySelectedCustomerServiceModuleSavedChoiceCurrentBinding(ctx, task, choice); err != nil {
		return nil, err
	}
	choice.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(choice)
	if err != nil {
		return nil, err
	}
	task.SelectedCustomerServiceRestoreChoice = choice
	task.SelectedCustomerServiceCandidateRehearsal = nil
	if task.State == "module_chat_candidate_rehearsal_verified" {
		if task.SelectedModuleRestorePreview.IsolatedRestoreVerified {
			task.State, task.Phase = "module_isolated_restore_verified", "selected_module_disposable_restore_and_target_preservation_verified"
		} else {
			task.State, task.Phase = "module_restore_conflicts_found", "selected_module_dependency_or_content_conflicts_require_resolution"
		}
	}
	if err := manager.saveTask(task); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

func (manager *Manager) verifySelectedCustomerServiceModuleSavedChoiceCurrentBinding(ctx context.Context, task *Task, choice *SelectedCustomerServiceModuleSavedRestoreChoice) error {
	cfg := manager.Configuration
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return err
	}
	packagePath, err := manager.taskPackagePath(task.ID)
	if err != nil {
		return err
	}
	identity, err := fileIdentity(packagePath, "")
	if err != nil || identity.Size != task.Size || identity.SHA256 != task.CiphertextSHA256 {
		return errors.New("模块加密包已变化，请重新上传校验")
	}
	media, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return err
	}
	live, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(live.Close)
	current, err := manager.captureSelectedCustomerServiceModuleSnapshotBinding(ctx, live, task, media)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(*current, choice.Binding) {
		return errors.New("当前数据库、媒体、业务边界或备份包已变化，请重新预览并保存恢复策略")
	}
	if !time.Now().Before(task.ExpiresAt) {
		return errors.New("聊天模块工作副本已到期，请重新上传校验")
	}
	return nil
}

func (manager *Manager) StartSelectedCustomerServiceModuleCandidateRehearsal(id, choiceSHA256, confirmation string) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	choice := task.SelectedCustomerServiceRestoreChoice
	if choice == nil || choice.Version != 1 || choice.State != "recorded_not_executable" || choice.ExecutionAvailable || choice.SHA256 == "" || choice.SHA256 != choiceSHA256 || choice.TaskID != id || choice.Operator == "" {
		return nil, errors.New("请先保存有效的聊天恢复策略")
	}
	copyChoice := *choice
	copyChoice.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(&copyChoice)
	if err != nil || digest != choiceSHA256 {
		return nil, errors.New("已保存的聊天恢复策略摘要不一致，请重新预览")
	}
	if err := validateSelectedCustomerServiceModuleTaskAndPreview(task, choice.PreviewSHA256); err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(choice.Binding, *task.SelectedModuleRestorePreview.Binding) {
		return nil, errors.New("聊天恢复策略与预览绑定不一致，请重新预览")
	}
	if err := validateSelectedCustomerServiceModuleChosenStrategyForFullRowRehearsal(choice, task.SelectedModuleRestorePreview); err != nil {
		return nil, err
	}
	cfg := manager.Configuration
	if confirmation != "VERIFY CHAT CHOICE "+cfg.SiteID+"/"+cfg.Environment+"/"+id+"/"+choiceSHA256 {
		return nil, errors.New("请确认所选聊天策略验证期间的商城维护窗口")
	}
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	task.SelectedCustomerServiceCandidateRehearsal = nil
	if err := manager.startTaskOperation(task, "preparing_selected_chat_choice_candidate_rehearsal", func(ctx context.Context, task *Task) error {
		return manager.previewSelectedBusinessModuleRestoreOrSavedChatChoiceInDisposableTargetClone(ctx, task, choice)
	}); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

func restoreSelectedCustomerServiceModuleChosenCompleteRowsAndVerifyAllRetainedTables(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, policy string) (inserted, replaced int, resultErr error) {
	if !reflect.DeepEqual(snapshot.Modules, []string{"customer_service"}) || (policy != "use_package_conversation_rows" && policy != "preserve_current_conflicting_conversation_rows") {
		return 0, 0, errors.New("完整行写入只接受已登记的单独聊天模块策略")
	}
	sequences, err := readCustomerServiceCandidateSequenceStates(ctx, tx)
	if err != nil {
		return 0, 0, err
	}
	var rows []customerServiceCandidateSelectedFullRow
	keys := map[string][]json.RawMessage{}
	totalBytes, totalRows, selectedCanonicalBytes := 0, 0, 0
	for _, table := range []string{"tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments"} {
		for _, raw := range snapshot.Tables[table].Rows {
			totalRows++
			totalBytes += len(raw)
			if len(raw) > customerServiceCandidateFullRowByteLimit || totalBytes > customerServiceCandidateTotalFullRowByteLimit || totalRows > 20000 {
				return 0, 0, errors.New("完整聊天记录超过候选行数或容量限制")
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				return 0, 0, err
			}
			identity := moduleRecoveryDependencyRowKey(fields, observation.PrimaryKeys[table])
			if identity == "" {
				return 0, 0, errors.New("完整聊天记录缺少主键")
			}
			var same bool
			err := tx.QueryRowContext(ctx, `SELECT to_jsonb(record)=$1::jsonb FROM public.`+pq.QuoteIdentifier(table)+` record WHERE `+customerServiceCandidateTypedPrimaryKeyPredicate(table, observation.PrimaryKeys[table]), string(raw)).Scan(&same)
			if err != nil && err != sql.ErrNoRows {
				return 0, 0, err
			}
			if err == nil && (same || policy == "preserve_current_conflicting_conversation_rows") {
				continue
			}
			var content, digest string
			if err := tx.QueryRowContext(ctx, `SELECT ($1::jsonb)::text,encode(sha256(convert_to(($1::jsonb)::text,'UTF8')),'hex')`, string(raw)).Scan(&content, &digest); err != nil {
				return 0, 0, err
			}
			selectedCanonicalBytes += len(content)
			if len(content) > customerServiceCandidateFullRowByteLimit || selectedCanonicalBytes > customerServiceCandidateTotalFullRowByteLimit {
				return 0, 0, errors.New("规范化后的完整聊天记录超过单行或总容量限制")
			}
			rows = append(rows, customerServiceCandidateSelectedFullRow{Table: table, Identity: identity, Digest: digest, Content: json.RawMessage(content)})
			keys[table] = append(keys[table], json.RawMessage(identity))
			if err == sql.ErrNoRows {
				inserted++
			} else {
				replaced++
			}
		}
	}
	order := map[string]int{"tickets": 0, "ticket_messages": 1, "customer_service_inbox_states": 2, "ticket_message_media_attachments": 3}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Table != rows[j].Table {
			return order[rows[i].Table] < order[rows[j].Table]
		}
		return rows[i].Identity < rows[j].Identity
	})
	selection := &CustomerServiceEffectiveSelectionInspection{}
	for table := range observation.Tables {
		preserved, err := fingerprintCustomerServiceRecoveryPlanRetainedTargetTable(ctx, tx, table, observation.Shapes[table], keys[table])
		if err != nil {
			return 0, 0, err
		}
		selection.TargetPreservation = append(selection.TargetPreservation, preserved)
	}
	// A normal isolated owner must satisfy ordinary constraints; never suppress
	// triggers using session_replication_role on this complete-row write path.
	if _, err := tx.ExecContext(ctx, `SET LOCAL ROLE local_backup_isolated_owner`); err != nil {
		return 0, 0, err
	}
	if err := writeCustomerServiceSelectedFullRowsAndValidateRetainedTables(ctx, tx, observation, selection, rows, keys); err != nil {
		return 0, 0, err
	}
	if _, err := advanceCustomerServiceCandidateOwnedSequencesWithoutRewinding(ctx, tx, sequences, keys); err != nil {
		return 0, 0, err
	}
	return inserted, replaced, nil
}
