package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

type ModuleRecoveryDependencyBlocker struct {
	Code    string `json:"code"`
	Table   string `json:"table,omitempty"`
	Message string `json:"message"`
}

type ModuleRecoveryDependencyTableComparison struct {
	Table         string `json:"table"`
	Role          string `json:"role"`
	SourceRecords *int64 `json:"source_records"`
	TargetRecords *int64 `json:"target_records"`
	Status        string `json:"status"`
}

type ModuleRecoveryDependencyReport struct {
	RegistryVersion                     int                                       `json:"registry_version"`
	RegistryMigrationVersion            int64                                     `json:"registry_migration_version"`
	TaskID                              string                                    `json:"task_id"`
	PackageID                           string                                    `json:"package_id"`
	Module                              string                                    `json:"module"`
	ModuleLabel                         string                                    `json:"module_label"`
	InspectedAt                         time.Time                                 `json:"inspected_at"`
	PackageCapturedAt                   time.Time                                 `json:"package_captured_at"`
	RestoreAvailable                    bool                                      `json:"restore_available"`
	TargetDatabaseAbsent                bool                                      `json:"target_database_absent"`
	TargetMatchesWholeStorefrontPreview bool                                      `json:"target_matches_whole_storefront_preview"`
	SourceDatabaseFingerprint           string                                    `json:"source_database_fingerprint"`
	TargetDatabaseFingerprint           string                                    `json:"target_database_fingerprint"`
	Tables                              []ModuleRecoveryDependencyTableComparison `json:"tables"`
	References                          []ModuleRecoveryDependencyReference       `json:"references"`
	Blockers                            []ModuleRecoveryDependencyBlocker         `json:"blockers"`
	RequiredReviews                     []string                                  `json:"required_reviews"`
	ConversationRows                    *ModuleRecoveryRowInspection              `json:"conversation_rows,omitempty"`
	OrderRows                           *ModuleRecoveryRowInspection              `json:"order_rows,omitempty"`
}

type moduleRecoveryDatabaseObservation struct {
	Tables           map[string]int64
	Columns          map[string]map[string]bool
	Shapes           map[string]string
	PrimaryKeys      map[string][]string
	References       []ModuleRecoveryDependencyReference
	Fingerprint      string
	Migration        int64
	Indexes          map[string][]moduleRecoveryIndexDefinition
	WriteSideEffects map[string][]moduleRecoveryDatabaseWriteSideEffectDefinition
}

// InspectVerifiedPackageModuleDependencies observes two independent read-only
// snapshots. It never fences writers, changes database connection policy,
// replaces rows or files, or promotes a task into another restore state.
// A report is advisory and cannot be submitted to the whole-storefront apply
// endpoint as a partial-recovery plan.
func (manager *Manager) InspectVerifiedPackageModuleDependencies(ctx context.Context, id, module string) (*ModuleRecoveryDependencyReport, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	return manager.inspectVerifiedPackageModuleDependenciesWhileOperationLocked(ctx, id, module)
}

func (manager *Manager) inspectVerifiedPackageModuleDependenciesWhileOperationLocked(ctx context.Context, id, module string) (*ModuleRecoveryDependencyReport, error) {
	return manager.inspectVerifiedPackageModuleDependenciesWithCustomerServiceSelectionObserverWhileOperationLocked(ctx, id, module, nil)
}

// The observer runs inside the same two read-only snapshots as the bound plan.
func (manager *Manager) inspectVerifiedPackageModuleDependenciesWithCustomerServiceSelectionObserverWhileOperationLocked(ctx context.Context, id, module string, observe func(context.Context, *ModuleRecoveryDependencyReport, *moduleRecoveryRowDependencyGraph, *moduleRecoveryRowDependencyGraph) error) (*ModuleRecoveryDependencyReport, error) {
	contract, supported := registeredOrderAndCustomerServiceRecoveryDependencyContracts()[module]
	if !supported {
		return nil, errors.New("only orders and customer_service dependency reports are supported")
	}
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	cfg := manager.Configuration
	if task.Kind != "upload" || task.State != "verified" || task.Journal != nil || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) || task.Manifest == nil || task.StageFingerprint == "" || task.StageDatabase != "local_backup_stage_"+id[:16] {
		return nil, errors.New("a verified whole-storefront upload with its original isolated database is required")
	}
	manifest := task.Manifest
	if manifest.FormatVersion != PackageVersion || manifest.Scope != "whole_application" || manifest.SiteID != cfg.SiteID || manifest.Environment != cfg.Environment || manifest.ApplicationVersion != cfg.ApplicationVersion || manifest.MigrationVersion != cfg.ExpectedMigrationVersion || len(manifest.Gaps) != 0 {
		return nil, errors.New("package identity, compatibility or completeness no longer permits inspection")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	source, err := cfg.connectDatabase(ctx, task.StageDatabase)
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(source.Close)
	sourceTransaction, err := source.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(sourceTransaction.Rollback)
	sourceObservation, err := inspectModuleRecoveryDatabaseObservation(ctx, sourceTransaction)
	if err != nil {
		return nil, err
	}
	if sourceObservation.Fingerprint != task.StageFingerprint || sourceObservation.Fingerprint != manifest.DatabaseFingerprint || sourceObservation.Migration != manifest.MigrationVersion {
		return nil, errors.New("isolated database changed after package validation; dependency report refused")
	}
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	var targetExists bool
	if err := control.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, cfg.DatabaseName).Scan(&targetExists); err != nil {
		return nil, err
	}
	var targetObservation *moduleRecoveryDatabaseObservation
	var targetTransaction *sql.Tx
	if targetExists {
		target, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
		if err != nil {
			return nil, err
		}
		defer func(closeResource func() error) { _ = closeResource() }(target.Close)
		targetTransaction, err = target.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return nil, err
		}
		defer func(closeResource func() error) { _ = closeResource() }(targetTransaction.Rollback)
		targetObservation, err = inspectModuleRecoveryDatabaseObservation(ctx, targetTransaction)
		if err != nil {
			return nil, err
		}
	}
	report := buildModuleRecoveryDependencyReport(contract, task, sourceObservation, targetObservation)
	// Reuse the existing whole-storefront preview identity exactly. No report
	// writes a replacement preview or erases its protection against later writes.
	mediaFingerprint, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return nil, err
	}
	databaseFingerprint := "absent_database"
	if targetObservation != nil {
		databaseFingerprint = targetObservation.Fingerprint
	}
	report.TargetDatabaseFingerprint = databaseFingerprint
	previewDigest := sha256.Sum256([]byte(databaseFingerprint + ":" + mediaFingerprint))
	report.TargetMatchesWholeStorefrontPreview = hex.EncodeToString(previewDigest[:]) == task.PreviewTargetFingerprint
	if !report.TargetMatchesWholeStorefrontPreview {
		report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "target_changed_since_whole_storefront_preview", Message: "目标数据库、序列或附件自整体预览后已变化；本报告不会更新原覆盖预览。"})
	}
	for index := range report.Tables {
		comparison := &report.Tables[index]
		if comparison.Role == "preserved" || comparison.SourceRecords == nil || comparison.TargetRecords == nil {
			continue
		}
		name := comparison.Table
		sourceIdentity, err := fingerprintModuleRecoveryTableContent(ctx, sourceTransaction, name, sourceObservation.Shapes[name])
		if err != nil {
			return nil, err
		}
		targetIdentity, err := fingerprintModuleRecoveryTableContent(ctx, targetTransaction, name, targetObservation.Shapes[name])
		if err != nil {
			return nil, err
		}
		if comparison.Status == "schema_conflict" {
			continue
		}
		if sourceIdentity != targetIdentity {
			comparison.Status = "content_differs"
			if comparison.Role == "shared_dependency" {
				report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "shared_dependency_content_differs", Table: name, Message: "共享依赖的内容与包内不同，尚未证明所引用行兼容；不能整体覆盖共享表。"})
			}
		} else {
			comparison.Status = "table_contents_match"
		}
	}
	var conversationSourceGraph, conversationTargetGraph *moduleRecoveryRowDependencyGraph
	if module == "customer_service" {
		directory, _ := manager.taskDirectory(id)
		report.ConversationRows, err = cfg.inspectCustomerServiceConversationRows(ctx, sourceTransaction, targetTransaction, sourceObservation, targetObservation, manifest, filepath.Join(directory, "verified"), defaultModuleRecoveryRowInspectionLimits(), func(source, target *moduleRecoveryRowDependencyGraph) {
			conversationSourceGraph, conversationTargetGraph = source, target
		})
		if err != nil {
			return nil, err
		}
		if !report.ConversationRows.Complete {
			report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "conversation_row_inspection_incomplete", Message: "会话行检查未覆盖全部对象，请查看缺表、资源上限或未支持引用的明细。"})
		}
		if report.ConversationRows.BlockingIssues > 0 {
			report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "conversation_row_conflicts", Message: "会话行或附件存在缺失、内容冲突或归属待确认对象，请查看逐对象明细。"})
		}
	}
	if module == "orders" {
		directory, _ := manager.taskDirectory(id)
		report.OrderRows, err = cfg.inspectOrderTransactionRows(ctx, sourceTransaction, targetTransaction, sourceObservation, targetObservation, manifest, filepath.Join(directory, "verified"), defaultModuleRecoveryRowInspectionLimits())
		if err != nil {
			return nil, err
		}
		if !report.OrderRows.Complete {
			report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "order_row_inspection_incomplete", Message: "订单逐行检查不完整，请查看必要结构、对象引用或资源上限。"})
		}
		if report.OrderRows.BlockingIssues > 0 {
			report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "order_row_conflicts", Message: "订单、支付退款、共享依赖或证据文件存在冲突，请查看逐对象明细。"})
		}
	}
	sortModuleRecoveryDependencyReport(report)
	if report.ConversationRows != nil && report.ConversationRows.ConversationRecoveryPlan != nil {
		if err := cfg.bindCustomerServiceRecoveryPlanToObservedSnapshots(report.ConversationRows.ConversationRecoveryPlan, task, report, mediaFingerprint, sourceObservation, targetObservation); err != nil {
			return nil, err
		}
		if observe != nil {
			if err := observe(ctx, report, conversationSourceGraph, conversationTargetGraph); err != nil {
				return nil, err
			}
		}
	}
	if err := sourceTransaction.Commit(); err != nil {
		return nil, err
	}
	if targetTransaction != nil {
		if err := targetTransaction.Commit(); err != nil {
			return nil, err
		}
	}
	directory, _ := manager.taskDirectory(id)
	if report.ConversationRows != nil && report.ConversationRows.ConversationRecoveryPlan != nil {
		plan := report.ConversationRows.ConversationRecoveryPlan
		if err := writeDurableJSON(filepath.Join(directory, "customer-service-conversation-recovery-plan-draft-v1.json"), plan); err != nil {
			return nil, errors.New("customer service recovery plan draft could not be persisted")
		}
	}
	if err := writeDurableJSON(filepath.Join(directory, "module-dependency-report-"+module+".json"), report); err != nil {
		return nil, errors.New("module dependency report could not be persisted")
	}
	return report, nil
}

func inspectModuleRecoveryDatabaseObservation(ctx context.Context, transaction *sql.Tx) (*moduleRecoveryDatabaseObservation, error) {
	tables, fingerprint, migration, err := captureDatabaseTransactionInventory(ctx, transaction)
	if err != nil {
		return nil, err
	}
	observation, err := inspectModuleRecoveryDatabaseSchemaWithoutRowFingerprints(ctx, transaction)
	if err != nil {
		return nil, err
	}
	observation.Tables, observation.Fingerprint, observation.Migration = tables, fingerprint, migration
	return observation, nil
}

// Reference capture needs catalog definitions, not a second content scan of
// every unrelated table. This observation has no row counts or data fingerprint
// and must never replace the full observation used to bind a recovery candidate.
func inspectModuleRecoveryDatabaseSchemaWithoutRowFingerprints(ctx context.Context, transaction *sql.Tx) (*moduleRecoveryDatabaseObservation, error) {
	observation := &moduleRecoveryDatabaseObservation{Columns: map[string]map[string]bool{}, Shapes: map[string]string{}, PrimaryKeys: map[string][]string{}}
	rows, err := transaction.QueryContext(ctx, `SELECT relation.relname,attribute.attname,jsonb_build_array(attribute.attnum,format_type(attribute.atttypid,attribute.atttypmod),attribute.attnotnull,attribute.attidentity,attribute.attgenerated,pg_get_expr(default_value.adbin,default_value.adrelid))::text FROM pg_class relation JOIN pg_attribute attribute ON attribute.attrelid=relation.oid LEFT JOIN pg_attrdef default_value ON default_value.adrelid=relation.oid AND default_value.adnum=attribute.attnum WHERE relation.relnamespace='public'::regnamespace AND relation.relkind IN ('r','p') AND attribute.attnum>0 AND NOT attribute.attisdropped ORDER BY relation.relname,attribute.attnum`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var table, column, shape string
		if err := rows.Scan(&table, &column, &shape); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if observation.Columns[table] == nil {
			observation.Columns[table] = map[string]bool{}
		}
		observation.Columns[table][column] = true
		observation.Shapes[table] += fmt.Sprintf("%q:%s\n", column, shape)
	}
	readErr := rows.Err()
	_ = rows.Close()
	if readErr != nil {
		return nil, readErr
	}
	primaryKeyRows, err := transaction.QueryContext(ctx, `SELECT relation.relname,attribute.attname FROM pg_constraint constraint_record JOIN pg_class relation ON relation.oid=constraint_record.conrelid CROSS JOIN LATERAL unnest(constraint_record.conkey) WITH ORDINALITY column_number(number,position) JOIN pg_attribute attribute ON attribute.attrelid=relation.oid AND attribute.attnum=column_number.number WHERE constraint_record.contype='p' AND relation.relnamespace='public'::regnamespace ORDER BY relation.relname,column_number.position`)
	if err != nil {
		return nil, err
	}
	for primaryKeyRows.Next() {
		var table, column string
		if err := primaryKeyRows.Scan(&table, &column); err != nil {
			_ = primaryKeyRows.Close()
			return nil, err
		}
		observation.PrimaryKeys[table] = append(observation.PrimaryKeys[table], column)
	}
	readErr = primaryKeyRows.Err()
	_ = primaryKeyRows.Close()
	if readErr != nil {
		return nil, readErr
	}
	rows, err = transaction.QueryContext(ctx, `SELECT relation.relname,pg_get_constraintdef(constraint_record.oid,true),constraint_record.convalidated FROM pg_constraint constraint_record JOIN pg_class relation ON relation.oid=constraint_record.conrelid WHERE relation.relnamespace='public'::regnamespace ORDER BY relation.relname,pg_get_constraintdef(constraint_record.oid,true),constraint_record.convalidated`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var table, definition string
		var validated bool
		if err := rows.Scan(&table, &definition, &validated); err != nil {
			_ = rows.Close()
			return nil, err
		}
		observation.Shapes[table] += fmt.Sprintf("constraint:%q:%t\n", definition, validated)
	}
	readErr = rows.Err()
	_ = rows.Close()
	if readErr != nil {
		return nil, readErr
	}
	rows, err = transaction.QueryContext(ctx, `SELECT child.relname,parent.relname,child_namespace.nspname,parent_namespace.nspname,array_to_json(ARRAY(SELECT attribute.attname FROM unnest(constraint_record.conkey) WITH ORDINALITY column_number(number,position) JOIN pg_attribute attribute ON attribute.attrelid=child.oid AND attribute.attnum=column_number.number ORDER BY column_number.position))::text,array_to_json(ARRAY(SELECT attribute.attname FROM unnest(constraint_record.confkey) WITH ORDINALITY column_number(number,position) JOIN pg_attribute attribute ON attribute.attrelid=parent.oid AND attribute.attnum=column_number.number ORDER BY column_number.position))::text FROM pg_constraint constraint_record JOIN pg_class child ON child.oid=constraint_record.conrelid JOIN pg_class parent ON parent.oid=constraint_record.confrelid JOIN pg_namespace child_namespace ON child_namespace.oid=child.relnamespace JOIN pg_namespace parent_namespace ON parent_namespace.oid=parent.relnamespace WHERE constraint_record.contype='f' AND (child_namespace.nspname='public' OR parent_namespace.nspname='public') ORDER BY child.relname,parent.relname,constraint_record.conname`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var reference ModuleRecoveryDependencyReference
		var childNamespace, parentNamespace, childColumns, parentColumns string
		if err := rows.Scan(&reference.SourceTable, &reference.TargetTable, &childNamespace, &parentNamespace, &childColumns, &parentColumns); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if childNamespace != "public" || parentNamespace != "public" {
			_ = rows.Close()
			return nil, errors.New("module dependency inspection refuses cross-schema references")
		}
		if err := json.Unmarshal([]byte(childColumns), &reference.SourceColumns); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if err := json.Unmarshal([]byte(parentColumns), &reference.TargetColumns); err != nil {
			_ = rows.Close()
			return nil, err
		}
		reference.Kind = "foreign_key"
		reference.Description = "实际数据库外键"
		observation.References = append(observation.References, reference)
	}
	readErr = rows.Err()
	_ = rows.Close()
	if readErr != nil {
		return nil, readErr
	}
	observation.Indexes, err = inspectModuleRecoveryDatabaseIndexDefinitions(ctx, transaction)
	if err != nil {
		return nil, err
	}
	observation.WriteSideEffects, err = inspectModuleRecoveryDatabaseWriteSideEffectDefinitions(ctx, transaction)
	if err != nil {
		return nil, err
	}
	return observation, nil
}

func buildModuleRecoveryDependencyReport(contract moduleRecoveryDependencyContract, task *Task, source, target *moduleRecoveryDatabaseObservation) *ModuleRecoveryDependencyReport {
	report := &ModuleRecoveryDependencyReport{
		RegistryVersion: moduleDependencyRegistryVersion, RegistryMigrationVersion: moduleDependencyRegistryMigrationVersion,
		TaskID: task.ID, PackageID: task.Manifest.PackageID, Module: contract.Name, ModuleLabel: contract.Label,
		InspectedAt: time.Now().UTC(), PackageCapturedAt: task.Manifest.CapturedAt, SourceDatabaseFingerprint: source.Fingerprint,
		TargetDatabaseAbsent: target == nil, Tables: []ModuleRecoveryDependencyTableComparison{}, References: []ModuleRecoveryDependencyReference{},
		RequiredReviews: append([]string{}, contract.ReviewMessages...),
		Blockers: []ModuleRecoveryDependencyBlocker{
			{Code: "module_recovery_not_enabled", Message: "当前仅提供模块依赖检查，单模块生成和覆盖恢复尚未开放。"},
			{Code: "row_and_file_ownership_review_required", Message: "表级内容相同不能证明局部恢复安全；仍需实现共享行、非外键引用、附件归属和成对回退契约。"},
		},
	}
	roles := map[string]string{}
	for _, name := range contract.OwnedTables {
		roles[name] = "module_owned"
	}
	for _, name := range contract.SharedTables {
		roles[name] = "shared_dependency"
	}
	observations := []*moduleRecoveryDatabaseObservation{source}
	if target == nil {
		report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "target_database_absent", Message: "目标业务库不存在，须先使用整体商城恢复建立完整基线。"})
	} else {
		observations = append(observations, target)
	}
	registeredTables := registeredStorefrontModuleRecoveryTableNames()
	allTables := map[string]bool{}
	references := append([]ModuleRecoveryDependencyReference{}, registeredNonForeignKeyModuleRecoveryReferences()...)
	for _, observation := range observations {
		if !moduleDependencyRegistrySupportsAuditedMigrationVersion(observation.Migration) {
			report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "module_registry_migration_incompatible", Message: "数据库迁移版本不在本版模块依赖登记范围内。"})
		}
		for table := range observation.Tables {
			allTables[table] = true
			if !registeredTables[table] {
				report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "unregistered_table", Table: table, Message: "存在未登记的表，须核对其模块归属与 JSON/文本引用；空表也不能静默忽略。"})
			}
		}
		references = append(references, observation.References...)
	}
	// Follow outgoing dependencies transitively. Incoming references are listed
	// as protected boundaries below; they do not authorize replacing that table.
	for changed := true; changed; {
		changed = false
		for _, reference := range references {
			if roles[reference.SourceTable] != "" && allTables[reference.SourceTable] && allTables[reference.TargetTable] && roles[reference.TargetTable] == "" {
				roles[reference.TargetTable] = "shared_dependency"
				changed = true
			}
		}
	}
	for table, role := range roles {
		if _, exists := source.Tables[table]; !exists {
			report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "source_dependency_table_missing", Table: table, Message: "包内缺少本模块登记的必要表。"})
		}
		if role != "" {
			allTables[table] = true
		}
	}
	for _, reference := range references {
		if roles[reference.SourceTable] == "" && roles[reference.TargetTable] == "" || !allTables[reference.SourceTable] {
			continue
		}
		report.References = append(report.References, reference)
		for _, observation := range observations {
			if _, exists := observation.Tables[reference.SourceTable]; !exists {
				continue
			}
			if _, exists := observation.Tables[reference.TargetTable]; !exists {
				report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "referenced_table_missing", Table: reference.TargetTable, Message: "外键或已登记业务引用的目标表缺失。"})
			}
			for _, field := range reference.SourceColumns {
				if !observation.Columns[reference.SourceTable][field] {
					report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "registered_reference_column_missing", Table: reference.SourceTable, Message: "已登记引用字段不存在：" + field})
				}
			}
			for _, field := range reference.TargetColumns {
				if !observation.Columns[reference.TargetTable][field] {
					report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "registered_reference_column_missing", Table: reference.TargetTable, Message: "已登记引用字段不存在：" + field})
				}
			}
		}
		if roles[reference.SourceTable] == "" && roles[reference.TargetTable] == "module_owned" {
			report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "preserved_table_references_module", Table: reference.SourceTable, Message: "保留表引用模块内数据，须逐行验证恢复后仍有效。"})
		}
	}
	for name := range allTables {
		role := roles[name]
		if role == "" {
			role = "preserved"
		}
		comparison := ModuleRecoveryDependencyTableComparison{Table: name, Role: role, Status: "not_compared"}
		if count, exists := source.Tables[name]; exists {
			comparison.SourceRecords = &count
		}
		if target != nil {
			if count, exists := target.Tables[name]; exists {
				comparison.TargetRecords = &count
			}
		}
		if role != "preserved" {
			if comparison.SourceRecords == nil {
				comparison.Status = "source_missing"
			} else if comparison.TargetRecords == nil {
				comparison.Status = "target_missing"
				report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "target_dependency_table_missing", Table: name, Message: "目标缺少本模块需要的表，尚无已验证的局部补全契约。"})
			} else if source.Shapes[name] != target.Shapes[name] {
				comparison.Status = "schema_conflict"
				report.Blockers = append(report.Blockers, ModuleRecoveryDependencyBlocker{Code: "dependency_schema_differs", Table: name, Message: "包内与目标的字段或约束结构不同。"})
			}
		}
		report.Tables = append(report.Tables, comparison)
	}
	return report
}

func fingerprintModuleRecoveryTableContent(ctx context.Context, transaction *sql.Tx, table, shape string) (string, error) {
	var count int64
	var first, second string
	query := `SELECT count(*),coalesce(sum(('x'||substr(md5(to_jsonb(record)::text),1,15))::bit(60)::bigint)::text,'0'),coalesce(sum(('x'||substr(md5(to_jsonb(record)::text),17,15))::bit(60)::bigint)::text,'0') FROM public.` + pq.QuoteIdentifier(table) + ` record`
	if err := transaction.QueryRowContext(ctx, query).Scan(&count, &first, &second); err != nil {
		return "", err
	}
	hash := sha256.New()
	_, _ = fmt.Fprintf(hash, "%s:%d:%s:%s", shape, count, first, second)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func sortModuleRecoveryDependencyReport(report *ModuleRecoveryDependencyReport) {
	sort.Slice(report.Tables, func(i, j int) bool {
		if report.Tables[i].Role != report.Tables[j].Role {
			return report.Tables[i].Role < report.Tables[j].Role
		}
		return report.Tables[i].Table < report.Tables[j].Table
	})
	uniqueBlockers := map[string]bool{}
	blockers := report.Blockers[:0]
	for _, blocker := range report.Blockers {
		key := blocker.Code + ":" + blocker.Table + ":" + blocker.Message
		if !uniqueBlockers[key] {
			blockers = append(blockers, blocker)
			uniqueBlockers[key] = true
		}
	}
	report.Blockers = blockers
	sort.Slice(report.Blockers, func(i, j int) bool {
		return report.Blockers[i].Code+report.Blockers[i].Table+report.Blockers[i].Message < report.Blockers[j].Code+report.Blockers[j].Table+report.Blockers[j].Message
	})
	uniqueReferences := map[string]bool{}
	references := report.References[:0]
	for _, reference := range report.References {
		key := reference.Kind + ":" + reference.SourceTable + ":" + strings.Join(reference.SourceColumns, ",") + ":" + reference.TargetTable + ":" + strings.Join(reference.TargetColumns, ",")
		if !uniqueReferences[key] {
			references = append(references, reference)
			uniqueReferences[key] = true
		}
	}
	report.References = references
	sort.Slice(report.References, func(i, j int) bool {
		left, _ := json.Marshal(report.References[i])
		right, _ := json.Marshal(report.References[j])
		return string(left) < string(right)
	})
}
