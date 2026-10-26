package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Reports contain counts and registered table names only, never customer data,
// attachment keys, credentials or full rows. This is a disposable rehearsal,
// not a candidate accepted by any switch or whole-application apply endpoint.
type SelectedBusinessModuleIsolatedRestorePreview struct {
	SHA256                          string                                                `json:"sha256,omitempty"`
	Binding                         *SelectedCustomerServiceModuleSnapshotBinding         `json:"binding,omitempty"`
	SavedChoiceSHA256               string                                                `json:"saved_choice_sha256,omitempty"`
	ReplacedRows                    int                                                   `json:"replaced_rows"`
	CandidateDatabaseFingerprint    string                                                `json:"candidate_database_fingerprint,omitempty"`
	Modules                         []string                                              `json:"modules"`
	Tables                          []SelectedBusinessModuleTableRestorePreview           `json:"tables"`
	DependencyConflicts             int                                                   `json:"dependency_conflicts"`
	DependencyIssues                []SelectedBusinessModuleExternalDependencyIssue       `json:"dependency_issues,omitempty"`
	DependencyReviews               int                                                   `json:"dependency_reviews"`
	FileConflicts                   int                                                   `json:"file_conflicts"`
	VerifiedFiles                   int                                                   `json:"verified_files"`
	RestoredRows                    int                                                   `json:"restored_rows"`
	PreservedTables                 int                                                   `json:"preserved_tables"`
	TargetUnchanged                 bool                                                  `json:"target_unchanged"`
	IsolatedRestoreVerified         bool                                                  `json:"isolated_restore_verified"`
	DisposableStageRemoved          bool                                                  `json:"disposable_stage_removed"`
	OnlineOverwriteAvailable        bool                                                  `json:"online_overwrite_available"`
	CustomerServiceOverwritePreview *SelectedCustomerServiceModuleBoundedOverwritePreview `json:"customer_service_overwrite_preview,omitempty"`
	Limitations                     []string                                              `json:"limitations"`
}

type SelectedBusinessModuleTableRestorePreview struct {
	Table                       string `json:"table"`
	IdenticalRows               int    `json:"identical_rows"`
	MissingRows                 int    `json:"missing_rows"`
	ConflictingRows             int    `json:"conflicting_rows"`
	ProtectedNewRows            int    `json:"protected_new_rows"`
	PreservedCurrentRuntimeRows int    `json:"preserved_current_runtime_rows"`
	UnsupportedReason           string `json:"unsupported_reason,omitempty"`
}

func (manager *Manager) StartSelectedBusinessModuleIsolatedRestorePreview(id, confirmation string) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != selectedBusinessModulesBackupScope || task.Journal != nil || task.StageDatabase != "" || task.SelectedModuleDisposableStage != nil || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return nil, errors.New("请选择未到期且已通过完整性校验的业务模块上传包")
	}
	if task.State != "module_package_integrity_verified" && task.State != "module_isolated_restore_verified" && task.State != "module_restore_conflicts_found" && task.State != "module_chat_candidate_rehearsal_verified" && ((task.State != "failed" && task.State != "interrupted") || task.SelectedModuleRestorePreview == nil) {
		return nil, errors.New("请先重新校验上传包，再验证模块恢复")
	}
	if confirmation != "VERIFY MODULE RESTORE "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id {
		return nil, errors.New("请确认验证期间的商城维护窗口")
	}
	if err := manager.Configuration.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	task.SelectedModuleRestorePreview = nil
	task.SelectedCustomerServiceRestoreChoice = nil
	task.SelectedCustomerServiceCandidateRehearsal = nil
	if err := manager.startTaskOperation(task, "preparing_selected_module_isolated_restore", manager.previewSelectedBusinessModuleRestoreInDisposableTargetClone); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

// Product composition needs the existing inventory/identity-specific merger.
// Execution facts need event reconciliation before any insertion is allowed.
// Conservatively refuse those writes; identical protected records are readable.
func selectedBusinessModuleTableRequiresDedicatedMissingRowRecovery(table string) bool {
	for _, name := range productCatalogOwnedTableNames {
		if table == name {
			return true
		}
	}
	return strings.HasPrefix(table, "referral") || strings.HasPrefix(table, "user_referral") ||
		strings.HasPrefix(table, "loyalty") || table == "user_loyalty" || table == "check_ins" ||
		strings.Contains(table, "outbox") || strings.Contains(table, "delivery") ||
		strings.Contains(table, "jobs") || strings.Contains(table, "receipt") ||
		table == "media_upload_sessions" || table == "payment_risk_alert_states"
}

func inspectSelectedBusinessModuleRowsForRestore(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, report *SelectedBusinessModuleIsolatedRestorePreview) (map[string][]json.RawMessage, bool, error) {
	missing := map[string][]json.RawMessage{}
	blocked := false
	names := make([]string, 0, len(snapshot.Tables))
	for name := range snapshot.Tables {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		contents := snapshot.Tables[name]
		preview := SelectedBusinessModuleTableRestorePreview{Table: name}
		keys := observation.PrimaryKeys[name]
		if len(keys) == 0 && len(contents.Rows) > 0 {
			preview.UnsupportedReason = "该表缺少可验证的主键，尚不支持恢复"
			blocked = true
		} else {
			seen := map[string]bool{}
			for _, row := range contents.Rows {
				var fields map[string]json.RawMessage
				if err := json.Unmarshal(row, &fields); err != nil || len(fields) != len(observation.Columns[name]) {
					return nil, false, fmt.Errorf("模块记录字段与目标结构不一致：%s", name)
				}
				for column := range observation.Columns[name] {
					if _, ok := fields[column]; !ok {
						return nil, false, fmt.Errorf("模块记录缺少持久化字段：%s", name)
					}
				}
				identity := map[string]json.RawMessage{}
				for _, key := range keys {
					if string(fields[key]) == "null" || len(fields[key]) == 0 {
						return nil, false, errors.New("模块记录主键为空")
					}
					identity[key] = fields[key]
				}
				encoded, _ := json.Marshal(identity)
				if seen[string(encoded)] {
					return nil, false, fmt.Errorf("模块记录主键重复：%s", name)
				}
				seen[string(encoded)] = true
				var same bool
				query := `SELECT to_jsonb(record)=$1::jsonb FROM public.` + pq.QuoteIdentifier(name) + ` record WHERE ` + customerServiceCandidateTypedPrimaryKeyPredicate(name, keys)
				err := tx.QueryRowContext(ctx, query, string(row)).Scan(&same)
				switch {
				case err == sql.ErrNoRows:
					preview.MissingRows++
					if selectedBusinessModuleTableRequiresDedicatedMissingRowRecovery(name) || len(snapshot.Modules) > 1 && customerServiceRecoveryPlanProposedRowReplacementAllowed(name) {
						preview.ProtectedNewRows++
						blocked = true
					} else {
						missing[name] = append(missing[name], row)
					}
				case err != nil:
					return nil, false, fmt.Errorf("无法检查模块记录身份：%s", name)
				case !same:
					preserve, err := canPreserveAdvancedSelectedOrderPaymentRiskAlertState(ctx, tx, snapshot, name, row)
					if err != nil {
						return nil, false, err
					}
					if preserve {
						preview.PreservedCurrentRuntimeRows++
					} else {
						preview.ConflictingRows++
						blocked = true
					}
				default:
					preview.IdenticalRows++
				}
			}
		}
		report.Tables = append(report.Tables, preview)
	}
	actual, err := captureSelectedBusinessModuleExternalReferenceFingerprintsWithObservedSchema(ctx, tx, snapshot, observation)
	if err != nil {
		return nil, false, err
	}
	classifySelectedBusinessModuleExternalDependencyIssues(snapshot, observation, actual, report)
	return missing, blocked || report.DependencyConflicts > 0, nil
}

// A clone can still run database code. Only ordinary built-in column types and
// tables without application triggers/rules or user-defined write expressions
// are eligible. All fields are explicit; no business defaults are evaluated.
func validateSelectedBusinessModuleIsolatedInsertWriteContract(ctx context.Context, tx *sql.Tx, name string, observation *moduleRecoveryDatabaseObservation) error {
	if len(observation.WriteSideEffects[name]) > 0 {
		return errors.New("该表具有写入触发器或规则，需要专用恢复契约")
	}
	var unsupported bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_class WHERE oid=$1::regclass AND (relkind<>'r' OR relrowsecurity OR relforcerowsecurity)) OR EXISTS(SELECT 1 FROM pg_inherits WHERE inhrelid=$1::regclass OR inhparent=$1::regclass) OR EXISTS(SELECT 1 FROM pg_attribute a JOIN pg_type t ON t.oid=a.atttypid WHERE a.attrelid=$1::regclass AND a.attnum>0 AND NOT a.attisdropped AND (a.attidentity<>'' OR a.attgenerated<>'' OR t.typnamespace<>'pg_catalog'::regnamespace)) OR EXISTS(SELECT 1 FROM pg_depend d JOIN pg_proc p ON d.refclassid='pg_proc'::regclass AND p.oid=d.refobjid WHERE p.pronamespace<>'pg_catalog'::regnamespace AND ((d.classid='pg_constraint'::regclass AND d.objid IN (SELECT oid FROM pg_constraint WHERE conrelid=$1::regclass)) OR (d.classid='pg_class'::regclass AND d.objid IN (SELECT indexrelid FROM pg_index WHERE indrelid=$1::regclass))))`
	if err := tx.QueryRowContext(ctx, query, "public."+pq.QuoteIdentifier(name)).Scan(&unsupported); err != nil {
		return err
	}
	if unsupported {
		return errors.New("该表具有尚不支持的类型、策略或写入表达式")
	}
	return nil
}

func restoreSelectedBusinessModuleMissingRowsAndVerifyPreservation(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, missing map[string][]json.RawMessage) (int, error) {
	// Inspection and the chat strategy savepoint have completed in this same
	// transaction. With no inserts there are no changes to re-fingerprint. The
	// outer clone verification and live database/media preservation checks stay.
	hasMissingRows := false
	for _, rows := range missing {
		hasMissingRows = hasMissingRows || len(rows) > 0
	}
	if !hasMissingRows {
		return 0, nil
	}
	baselineSequences, err := readCustomerServiceCandidateSequenceStates(ctx, tx)
	if err != nil {
		return 0, err
	}
	before := map[string]CustomerServiceRecoveryPlanTablePreservation{}
	for table := range observation.Tables {
		fingerprint, err := fingerprintCustomerServiceRecoveryPlanRetainedTargetTable(ctx, tx, table, observation.Shapes[table], missing[table])
		if err != nil {
			return 0, err
		}
		before[table] = fingerprint
	}
	if _, err := tx.ExecContext(ctx, `SET LOCAL session_replication_role='replica'`); err != nil {
		return 0, err
	}
	restored := 0
	for table, rows := range missing {
		var columns []string
		for column := range observation.Columns[table] {
			columns = append(columns, pq.QuoteIdentifier(column))
		}
		sort.Strings(columns)
		list := strings.Join(columns, ",")
		for _, row := range rows {
			query := `INSERT INTO public.` + pq.QuoteIdentifier(table) + ` (` + list + `) SELECT ` + list + ` FROM jsonb_populate_record(NULL::public.` + pq.QuoteIdentifier(table) + `,$1::jsonb)` // #nosec G202 -- Table and observed column identifiers use pq.QuoteIdentifier; complete row values are bound parameters.
			if _, err := tx.ExecContext(ctx, query, string(row)); err != nil {
				return 0, fmt.Errorf("模块隔离恢复违反唯一键或业务约束，已回滚：%s", table)
			}
			restored++
		}
	}
	if err := validateProductCatalogAffectedForeignKeys(ctx, tx, productCatalogModuleSnapshot{Tables: snapshot.Tables}); err != nil {
		return 0, err
	}
	for table, contents := range snapshot.Tables {
		for _, row := range contents.Rows {
			var same bool
			query := `SELECT to_jsonb(record)=$1::jsonb FROM public.` + pq.QuoteIdentifier(table) + ` record WHERE ` + customerServiceCandidateTypedPrimaryKeyPredicate(table, observation.PrimaryKeys[table])
			if err := tx.QueryRowContext(ctx, query, string(row)).Scan(&same); err != nil {
				return 0, fmt.Errorf("无法核对隔离恢复后的完整记录：%s", table)
			}
			if !same {
				preserve, err := canPreserveAdvancedSelectedOrderPaymentRiskAlertState(ctx, tx, snapshot, table, row)
				if err != nil {
					return 0, err
				}
				if preserve {
					continue
				}
				return 0, fmt.Errorf("隔离恢复后的完整记录与备份不一致：%s", table)
			}
		}
	}
	for table, expected := range before {
		actual, err := fingerprintCustomerServiceRecoveryPlanRetainedTargetTable(ctx, tx, table, observation.Shapes[table], missing[table])
		if err != nil {
			return 0, err
		}
		if actual.ContentFingerprint != expected.ContentFingerprint || actual.RetainedRecords != expected.RetainedRecords {
			return 0, fmt.Errorf("隔离恢复改变了目标已有记录：%s", table)
		}
	}
	if _, err := advanceCustomerServiceCandidateOwnedSequencesWithoutRewinding(ctx, tx, baselineSequences, missing); err != nil {
		return 0, err
	}
	return restored, nil
}

func (manager *Manager) previewSelectedBusinessModuleRestoreInDisposableTargetClone(ctx context.Context, task *Task) (resultErr error) {
	return manager.previewSelectedBusinessModuleRestoreOrSavedChatChoiceInDisposableTargetClone(ctx, task, nil)
}

func (manager *Manager) previewSelectedBusinessModuleRestoreOrSavedChatChoiceInDisposableTargetClone(ctx context.Context, task *Task, choice *SelectedCustomerServiceModuleSavedRestoreChoice) (resultErr error) {
	var completedState, completedPhase string
	var report *SelectedBusinessModuleIsolatedRestorePreview
	defer func() {
		if resultErr == nil && completedState != "" {
			if report.Binding != nil {
				report.SHA256 = ""
				report.SHA256, resultErr = customerServiceRecoveryPlanCanonicalJSONSHA256(report)
				if resultErr != nil {
					return
				}
			}
			task.State, task.Phase = completedState, completedPhase
			resultErr = manager.saveTask(task)
		}
	}()
	cfg := manager.Configuration
	directory, _ := manager.taskDirectory(task.ID)
	// Reauthenticate the encrypted source for each rehearsal; never trust an old
	// mutable extraction directory as authority for an import.
	extractedWorkspace, err := os.MkdirTemp(directory, "selected-module-authenticated-package-")
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, os.RemoveAll(extractedWorkspace)) }()
	extracted := filepath.Join(extractedWorkspace, "verified")
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	if err != nil {
		return err
	}
	packagePath, _ := manager.taskPackagePath(task.ID)
	identity, err := fileIdentity(packagePath, "")
	if err != nil || identity.Size != task.Size || identity.SHA256 != task.CiphertextSHA256 {
		return errors.New("模块加密包已变化，请重新上传校验")
	}
	manifest, err := extractAndVerifyEncryptedPackageWithAllowedScopes(packagePath, extracted, passphrase, cfg.MaxExpandedBytes, cfg.MaxFiles, true, cfg.checkWorkspaceCapacity)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(manifest, *task.Manifest) || len(manifest.Gaps) > 0 {
		return errors.New("模块包清单已变化，请重新上传校验")
	}
	task.Phase = "checking_selected_module_target_dependencies"
	snapshot, err := manager.readValidatedSelectedBusinessModuleSnapshot(task, extracted)
	if err != nil {
		return err
	}
	report = &SelectedBusinessModuleIsolatedRestorePreview{Modules: snapshot.Modules, Tables: []SelectedBusinessModuleTableRestorePreview{}, Limitations: []string{
		"本次为一次性隔离演练；不会覆盖当前商城，验证结束删除演练库和媒体副本。",
		"订单预警保留已验证前进的当前运行状态；其他内容差异阻断，商品组合写入、事件执行事实及推荐积分缺失需要专用恢复规则。",
		"正式分模块覆盖导入及从组合包再次选择子模块尚未开放。",
	}}
	if choice == nil {
		task.SelectedModuleRestorePreview = report
	} else {
		task.SelectedCustomerServiceCandidateRehearsal = report
		report.SavedChoiceSHA256 = choice.SHA256
		report.Limitations = []string{"本次按已保存策略在一次性独立副本写入完整聊天记录，验证结束删除副本；不会覆盖当前商城。", "正式候选接收、成对切换、对账与放行尚未开放。"}
	}
	required, err := cfg.estimateBackupWorkspaceBytes(ctx)
	if err != nil {
		return err
	}
	if err := cfg.checkWorkspaceCapacity(required); err != nil {
		return err
	}
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, lease.Close(), manager.releaseMaintenance(task)) }()
	candidate, err := os.MkdirTemp(directory, "selected-module-disposable-restore-")
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, os.RemoveAll(candidate)) }()
	mediaBefore, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return err
	}
	live, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(live.Close)
	if len(snapshot.Modules) == 1 && snapshot.Modules[0] == "customer_service" {
		binding, err := manager.captureSelectedCustomerServiceModuleSnapshotBinding(ctx, live, task, mediaBefore)
		if err != nil {
			return err
		}
		report.Binding = binding
		if choice != nil && !reflect.DeepEqual(*binding, choice.Binding) {
			return errors.New("当前数据库、媒体、业务边界或备份包已变化，请重新预览并保存恢复策略")
		}
	}
	liveTx, err := live.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	if err := requireRegisteredStorefrontTablesForProductCatalogModule(ctx, liveTx); err != nil {
		_ = liveTx.Rollback()
		return err
	}
	for name, contents := range snapshot.Tables {
		schema, _, schemaErr := inspectProductCatalogTableSchema(ctx, liveTx, name)
		constraints, constraintErr := captureProductCatalogConstraintFingerprint(ctx, liveTx, name)
		if schemaErr != nil || constraintErr != nil || schema != contents.SchemaSHA256 || constraints != contents.ConstraintsSHA256 {
			_ = liveTx.Rollback()
			return fmt.Errorf("目标模块结构或约束已变化：%s", name)
		}
	}
	if err := liveTx.Commit(); err != nil {
		return err
	}
	var target Manifest
	if err := cfg.captureSnapshotAndDump(ctx, candidate, &target); err != nil {
		return err
	}
	if target.MigrationVersion != cfg.ExpectedMigrationVersion {
		return errors.New("目标迁移版本已变化")
	}
	if report.Binding != nil && report.Binding.TargetDatabaseFingerprint != target.DatabaseFingerprint {
		return errors.New("生成副本前目标数据已变化，请重新预览")
	}
	task.PreviewTargetFingerprint, task.PreviewTargetTables = target.DatabaseFingerprint, target.Tables
	// This random stage is never persisted as an applicable recovery candidate.
	stageName := "local_backup_stage_" + newTaskID()[:16]
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	var existingOID int64
	if err := readDatabaseOID(ctx, control, stageName, &existingOID); err != nil || existingOID != 0 {
		return errors.New("一次性演练库名称已被占用")
	}
	task.SelectedModuleDisposableStage = &SelectedBusinessModuleDisposableStageIdentity{DatabaseName: stageName}
	if err := manager.saveTask(task); err != nil {
		return err
	}
	// Record the created OID before writes. Cleanup refuses name/owner changes.
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cleanupErr := manager.cleanupSelectedBusinessModuleDisposableStageByPersistedIdentity(cleanupCtx, task)
		report.DisposableStageRemoved = cleanupErr == nil
		resultErr = errors.Join(resultErr, cleanupErr)
	}()
	if err := cfg.restoreIsolatedDatabaseWithCreatedIdentityBinding(ctx, candidate, stageName, target, func(oid int64) error {
		task.SelectedModuleDisposableStage.DatabaseOID = oid
		return manager.saveTask(task)
	}); err != nil {
		return err
	}
	stage, err := cfg.connectDatabase(ctx, stageName)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(stage.Close)
	tx, err := stage.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	if _, err := tx.ExecContext(ctx, `SET LOCAL timezone='UTC'; SET LOCAL search_path=pg_catalog,public`); err != nil {
		return err
	}
	var lockedTables []string
	for name := range snapshot.Tables {
		lockedTables = append(lockedTables, "public."+pq.QuoteIdentifier(name))
	}
	sort.Strings(lockedTables)
	if _, err := tx.ExecContext(ctx, `LOCK TABLE `+strings.Join(lockedTables, ",")+` IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return err
	}
	observation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
	if err != nil {
		return err
	}
	missing, blocked, err := inspectSelectedBusinessModuleRowsForRestore(ctx, tx, snapshot, observation, report)
	if err != nil {
		return err
	}
	for index := range report.Tables {
		preview := &report.Tables[index]
		if len(missing[preview.Table]) > 0 {
			if err := validateSelectedBusinessModuleIsolatedInsertWriteContract(ctx, tx, preview.Table, observation); err != nil {
				preview.UnsupportedReason = err.Error()
				blocked = true
			}
		}
	}
	// Copy current media, then merge only missing package paths. Existing bytes
	// must match. Private roots stay private, including the rehearsal copy.
	var currentMedia Manifest
	if err := cfg.copyMediaSnapshot(ctx, candidate, &currentMedia); err != nil {
		return err
	}
	for _, entry := range task.Manifest.Files {
		if entry.Path == selectedBusinessModulesRowsFilename {
			continue
		}
		destination := filepath.Join(candidate, filepath.FromSlash(entry.Path))
		if strings.HasPrefix(entry.Path, "product-static/") {
			destination, err = cfg.resolveSelectedBusinessModuleConfinedFileSource(entry.Path)
			if err != nil {
				report.FileConflicts++
				continue
			}
		}
		if _, err := os.Lstat(destination); err == nil {
			identity, err := fileIdentity(destination, entry.Path)
			if err != nil || identity != entry {
				report.FileConflicts++
				continue
			}
		} else if errors.Is(err, os.ErrNotExist) {
			if identity, err := copyVerifiedRegularFile(filepath.Join(extracted, filepath.FromSlash(entry.Path)), destination, entry.Path); err != nil || identity != entry {
				if err == nil {
					err = errors.New("隔离复制文件摘要与清单不一致")
				}
				return err
			}
		} else {
			return err
		}
		report.VerifiedFiles++
	}
	blocked = blocked || report.FileConflicts > 0
	if len(snapshot.Modules) == 1 && snapshot.Modules[0] == "customer_service" {
		report.CustomerServiceOverwritePreview, err = cfg.inspectSelectedCustomerServiceModuleBoundedOverwriteStrategies(ctx, tx, snapshot, observation, missing, task.Manifest, extracted, report)
		if err != nil {
			return err
		}
		if report.CustomerServiceOverwritePreview.PackageRows == nil || !report.CustomerServiceOverwritePreview.PackageRows.DomainChecksSatisfied {
			blocked = true
		}
	}
	if choice != nil {
		if err := validateSelectedCustomerServiceModuleChosenStrategyForFullRowRehearsal(choice, report); err != nil {
			return err
		}
		report.RestoredRows, report.ReplacedRows, err = restoreSelectedCustomerServiceModuleChosenCompleteRowsAndVerifyAllRetainedTables(ctx, tx, snapshot, observation, choice.Policy)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		_, report.CandidateDatabaseFingerprint, _, err = captureDatabaseInventory(ctx, stage)
		if err != nil {
			return err
		}
		report.PreservedTables = len(observation.Tables)
		report.IsolatedRestoreVerified = true
		blocked = false
	} else if !blocked {
		report.RestoredRows, err = restoreSelectedBusinessModuleMissingRowsAndVerifyPreservation(ctx, tx, snapshot, observation, missing)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		report.PreservedTables = len(observation.Tables)
		report.IsolatedRestoreVerified = true
	} else if err := tx.Rollback(); err != nil {
		return err
	}
	_, after, _, err := captureDatabaseInventory(ctx, live)
	if err != nil {
		return err
	}
	mediaAfter, err := cfg.mediaFingerprint(ctx)
	if err != nil || after != target.DatabaseFingerprint || mediaAfter != mediaBefore {
		return errors.New("隔离演练期间目标数据库或实际媒体发生变化，结果无效")
	}
	report.TargetUnchanged = true
	if report.Binding != nil {
		current, err := manager.captureSelectedCustomerServiceModuleSnapshotBinding(ctx, live, task, mediaAfter)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(*current, *report.Binding) {
			return errors.New("隔离演练期间目标结构或运行边界发生变化，请重新预览")
		}
	}
	completedState, completedPhase = "module_isolated_restore_verified", "selected_module_disposable_restore_and_target_preservation_verified"
	if blocked {
		completedState, completedPhase = "module_restore_conflicts_found", "selected_module_dependency_or_content_conflicts_require_resolution"
	}
	if choice != nil {
		completedState, completedPhase = "module_chat_candidate_rehearsal_verified", "selected_chat_choice_full_rows_and_retained_target_verified"
	}
	return nil
}
