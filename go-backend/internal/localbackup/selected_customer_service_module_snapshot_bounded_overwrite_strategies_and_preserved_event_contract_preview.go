package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/lib/pq"
)

// These strategies are diagnostics of an authenticated module snapshot. They
// never manufacture a whole-application source database, saved content review,
// executable plan, or accepted candidate. Only the four conversation-owned
// tables may enter the existing SELECT-only effective-row overlay.
type SelectedCustomerServiceModuleBoundedOverwritePreview struct {
	State                  string                                                    `json:"state"`
	PrerequisitesSatisfied bool                                                      `json:"prerequisites_satisfied"`
	ExecutionAvailable     bool                                                      `json:"execution_available"`
	PackageRows            *SelectedCustomerServiceModuleOverwriteStrategyInspection `json:"package_rows,omitempty"`
	PreserveCurrentRows    *SelectedCustomerServiceModuleOverwriteStrategyInspection `json:"preserve_current_rows,omitempty"`
}

type SelectedCustomerServiceModuleOverwriteStrategyInspection struct {
	Policy                   string                   `json:"policy"`
	Complete                 bool                     `json:"complete"`
	DomainChecksSatisfied    bool                     `json:"domain_checks_satisfied"`
	PackageRowsSelected      int                      `json:"package_rows_selected"`
	ConflictingRowsPreserved int                      `json:"conflicting_rows_preserved"`
	TargetOnlyRowsPreserved  int                      `json:"target_only_rows_preserved"`
	BlockingIssues           int                      `json:"blocking_issues"`
	IssueCounts              map[string]int           `json:"issue_counts"`
	Issues                   []ModuleRecoveryRowIssue `json:"issues"`
	IssuesTruncated          bool                     `json:"issues_truncated"`
	PreservedEventsChecked   int                      `json:"preserved_events_checked"`
	RequiredFilesChecked     int                      `json:"required_files_checked"`
}

func newSelectedCustomerServiceModuleDomainChecks() *ModuleRecoveryRowInspection {
	limits := defaultModuleRecoveryRowInspectionLimits()
	return &ModuleRecoveryRowInspection{Version: 1, Complete: true, RowLimitPerSide: limits.Rows, ProjectionByteLimit: limits.ProjectionBytes, IssueSampleLimit: limits.IssueSamples, IssueCounts: map[string]int{}, ConversationReferenceChecks: &CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection{Version: 1}}
}

// Full rows remain package authority, but only registered relationship fields
// enter the bounded graph. Their digest is private bookkeeping, not a database
// fingerprint and never a substitute for reauthenticating the encrypted source.
func buildSelectedCustomerServiceModuleSnapshotRelationshipRows(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, checks *ModuleRecoveryRowInspection) (map[string]map[string]*moduleRecoveryDependencyRow, error) {
	result := map[string]map[string]*moduleRecoveryDependencyRow{}
	limits := defaultModuleRecoveryRowInspectionLimits()
	count, bytes := 0, 0
	for _, table := range []string{"tickets", "ticket_messages", "customer_service_inbox_states", "ticket_message_media_attachments"} {
		result[table] = map[string]*moduleRecoveryDependencyRow{}
		for _, fullRow := range snapshot.Tables[table].Rows {
			var fields map[string]json.RawMessage
			if json.Unmarshal(fullRow, &fields) != nil {
				checks.Complete = false
				checks.recordIssue("module_snapshot_projection_invalid", "blocking", "package", table, "", "模块记录无法形成引用投影。")
				return result, nil
			}
			values := map[string]json.RawMessage{}
			for _, column := range moduleRecoveryDependencyProjectionColumns(observation, table) {
				values[column] = fields[column]
			}
			if table == "ticket_messages" {
				for _, column := range []string{"message_type", "metadata"} {
					if value, exists := fields[column]; exists {
						values["_structured_"+column] = value
					}
				}
			}
			encoded, err := json.Marshal(values)
			identity := moduleRecoveryDependencyRowKey(values, observation.PrimaryKeys[table])
			if err != nil || identity == "" || len(encoded) > 65536 || bytes+len(encoded) > limits.ProjectionBytes || count >= limits.Rows {
				checks.Complete = false
				checks.recordIssue("module_snapshot_projection_resource_limit", "blocking", "package", table, "", "模块聊天引用投影达到行数或容量上限，覆盖检查未完成。")
				return result, nil
			}
			count++
			bytes += len(encoded)
			var digest string
			if err := tx.QueryRowContext(ctx, `SELECT encode(sha256(convert_to(($1::jsonb)::text,'UTF8')),'hex')`, string(fullRow)).Scan(&digest); err != nil {
				return nil, err
			}
			result[table][identity] = &moduleRecoveryDependencyRow{Values: values, Identity: identity, Digest: digest}
		}
	}
	return result, nil
}

func (cfg Configuration) inspectSelectedCustomerServiceModuleBoundedOverwriteStrategies(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation, missing map[string][]json.RawMessage, manifest *Manifest, verifiedDirectory string, restoreReport *SelectedBusinessModuleIsolatedRestorePreview) (result *SelectedCustomerServiceModuleBoundedOverwritePreview, resultErr error) {
	result = &SelectedCustomerServiceModuleBoundedOverwritePreview{State: "prerequisites_not_satisfied"}
	// Shared differences, unsupported writes and absent execution facts cannot
	// be resolved by selecting a different conversation content strategy.
	for _, table := range restoreReport.Tables {
		if table.ProtectedNewRows > 0 || table.UnsupportedReason != "" || table.ConflictingRows > 0 && !customerServiceRecoveryPlanProposedRowReplacementAllowed(table.Table) {
			return result, nil
		}
	}
	if restoreReport.DependencyConflicts > 0 || restoreReport.FileConflicts > 0 {
		return result, nil
	}
	result.PrerequisitesSatisfied = true
	checks := newSelectedCustomerServiceModuleDomainChecks()
	packageRows, err := buildSelectedCustomerServiceModuleSnapshotRelationshipRows(ctx, tx, snapshot, observation, checks)
	if err != nil {
		return nil, err
	}
	current := newModuleRecoveryRowDependencyGraph(tx, observation, "target", checks, defaultModuleRecoveryRowInspectionLimits())
	if err := current.selectCustomerServiceConversationScopeRows(ctx); err != nil {
		return nil, err
	}
	// Look up package primary keys outside the current category too. A primary
	// key reassigned to a general ticket must never be interpreted as absent.
	for table, rows := range packageRows {
		var keys []map[string]json.RawMessage
		for _, row := range rows {
			keys = append(keys, row.Values)
		}
		for start := 0; start < len(keys) && !current.Exhausted; start += 200 {
			matched, err := current.readRowsMatchingReferenceKeys(ctx, table, observation.PrimaryKeys[table], keys[start:min(start+200, len(keys))], current.Limits.Rows)
			if err != nil {
				return nil, err
			}
			current.includeRowsInDependencyGraph(matched, table, false)
		}
	}
	// Only missing shared metadata is materialized, in the already disposable
	// clone and inside a savepoint. Events/receipts are never inserted. Every
	// explicit-column insert is checked against the existing write contract.
	if _, err := tx.ExecContext(ctx, `SAVEPOINT selected_customer_service_shared_metadata`); err != nil {
		return nil, err
	}
	defer func() {
		_, rollbackErr := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT selected_customer_service_shared_metadata`)
		_, releaseErr := tx.ExecContext(ctx, `RELEASE SAVEPOINT selected_customer_service_shared_metadata`)
		resultErr = errors.Join(resultErr, rollbackErr, releaseErr)
	}()
	if _, err := tx.ExecContext(ctx, `SET LOCAL session_replication_role='replica'`); err != nil {
		return nil, err
	}
	for table, rows := range missing {
		if customerServiceRecoveryPlanProposedRowReplacementAllowed(table) || len(rows) == 0 {
			continue
		}
		if selectedBusinessModuleTableRequiresDedicatedMissingRowRecovery(table) {
			return nil, errors.New("聊天覆盖检查不能插入共享执行事实")
		}
		if err := validateSelectedBusinessModuleIsolatedInsertWriteContract(ctx, tx, table, observation); err != nil {
			return nil, err
		}
		var columns []string
		for column := range observation.Columns[table] {
			columns = append(columns, pq.QuoteIdentifier(column))
		}
		sort.Strings(columns)
		list := strings.Join(columns, ",")
		for _, row := range rows {
			query := `INSERT INTO public.` + pq.QuoteIdentifier(table) + ` (` + list + `) SELECT ` + list + ` FROM jsonb_populate_record(NULL::public.` + pq.QuoteIdentifier(table) + `,$1::jsonb)` // #nosec G202 -- Table and observed column identifiers use pq.QuoteIdentifier; complete row values are bound parameters.
			if _, err := tx.ExecContext(ctx, query, string(row)); err != nil {
				return nil, errors.New("聊天覆盖检查的共享元数据违反目标约束")
			}
		}
	}
	for _, preserveCurrent := range []bool{false, true} {
		inspection, err := cfg.inspectSelectedCustomerServiceModuleEffectiveStrategy(ctx, current, packageRows, checks, preserveCurrent, manifest, verifiedDirectory)
		if err != nil {
			return nil, err
		}
		if preserveCurrent {
			result.PreserveCurrentRows = inspection
		} else {
			result.PackageRows = inspection
		}
	}
	result.State = "bounded_strategies_inspected_not_executable"
	return result, nil
}

func (cfg Configuration) inspectSelectedCustomerServiceModuleEffectiveStrategy(ctx context.Context, current *moduleRecoveryRowDependencyGraph, packageRows map[string]map[string]*moduleRecoveryDependencyRow, seedChecks *ModuleRecoveryRowInspection, preserveCurrent bool, manifest *Manifest, verifiedDirectory string) (*SelectedCustomerServiceModuleOverwriteStrategyInspection, error) {
	checks := newSelectedCustomerServiceModuleDomainChecks()
	result := &SelectedCustomerServiceModuleOverwriteStrategyInspection{Policy: "use_package_conversation_rows"}
	if preserveCurrent {
		result.Policy = "preserve_current_conflicting_conversation_rows"
	}
	if !seedChecks.Complete {
		checks.Complete = false
		checks.recordIssue("module_snapshot_or_current_scope_incomplete", "blocking", "target", "", "", "包内投影或当前会话范围检查未完成，不能据此采用任何覆盖策略。")
	}
	effective := newModuleRecoveryRowDependencyGraph(current.Transaction, current.Observation, "target", checks, current.Limits)
	effective.SelectedCustomerServiceRows = map[string]map[string]*moduleRecoveryDependencyRow{}
	for table, rows := range current.Rows {
		for identity, row := range rows {
			if current.Scoped[table][identity] {
				effective.includeBoundedCustomerServiceEffectiveSelectionSeed(table, row)
				if customerServiceRecoveryPlanProposedRowReplacementAllowed(table) && packageRows[table][identity] == nil {
					result.TargetOnlyRowsPreserved++
				}
			}
		}
	}
	for table, rows := range packageRows {
		effective.SelectedCustomerServiceRows[table] = map[string]*moduleRecoveryDependencyRow{}
		for identity, row := range rows {
			previous := current.Rows[table][identity]
			if previous != nil && customerServiceConversationRowOwnershipConflicts(table, row, previous) {
				checks.recordIssue("module_conversation_row_ownership_conflict", "blocking", "target", table, "", "同一主键的会话、账号或附件归属发生变化；内容保留策略不能豁免身份冲突。")
			}
			// Compare the actual complete row rather than the relationship digest:
			// message content is deliberately absent from the private projection.
			var same bool
			if previous != nil {
				// Both digests use PostgreSQL jsonb text, regardless of package
				// indentation or JSON object key order.
				same = previous.Digest == row.Digest
			}
			if preserveCurrent && previous != nil && !same {
				result.ConflictingRowsPreserved++
				continue
			}
			effective.SelectedCustomerServiceRows[table][identity] = row
			effective.includeBoundedCustomerServiceEffectiveSelectionSeed(table, row)
			result.PackageRowsSelected++
		}
	}
	if err := effective.includeRetainedIncomingRowsReferencingReplacedCustomerServiceKeys(ctx, current); err != nil {
		return nil, err
	}
	if err := effective.inspectCustomerServiceCandidateColumnReferenceSemantics(); err != nil {
		return nil, err
	}
	if err := effective.inspectCustomerServiceRetainedMediaOwnerColumnReferenceSemantics(); err != nil {
		return nil, err
	}
	effective.inspectCustomerServiceCandidateWriteSideEffectContracts()
	if err := effective.expandTransitiveForeignKeyAndAccountDependencies(ctx); err != nil {
		return nil, err
	}
	if err := effective.inspectConversationBusinessReferences(ctx); err != nil {
		return nil, err
	}
	if err := cfg.includeCustomerServiceLegacyAttachmentAssets(ctx, effective); err != nil {
		return nil, err
	}
	if err := effective.expandTransitiveForeignKeyAndAccountDependencies(ctx); err != nil {
		return nil, err
	}
	if err := inspectCustomerServiceConversationNaturalKeyCollisions(ctx, effective, nil); err != nil {
		return nil, err
	}
	if err := inspectCustomerServiceStructuredBusinessSnapshotDependencies(ctx, effective, nil); err != nil {
		return nil, err
	}
	inspectCustomerServiceEffectiveRealtimeVersionAndReadCursorContinuity(effective, current)
	files, err := cfg.inspectCustomerServiceEffectiveSelectionRequiredFiles(ctx, effective, manifest, verifiedDirectory)
	if err != nil {
		return nil, err
	}
	if err := cfg.inspectCustomerServiceDurableEventDependencies(ctx, []*moduleRecoveryRowDependencyGraph{effective}, checks); err != nil {
		return nil, err
	}
	result.Complete, result.BlockingIssues = checks.Complete, checks.BlockingIssues
	result.DomainChecksSatisfied = checks.Complete && checks.BlockingIssues == 0
	result.IssueCounts, result.IssuesTruncated = checks.IssueCounts, checks.IssuesTruncated
	result.Issues = []ModuleRecoveryRowIssue{}
	for _, issue := range checks.Issues {
		if issue.Severity == "blocking" {
			issue.RowIdentity = ""
			result.Issues = append(result.Issues, issue)
		}
	}
	result.RequiredFilesChecked = len(files)
	if checks.DurableEvents != nil {
		result.PreservedEventsChecked = checks.DurableEvents.TargetScannedEvents
	}
	return result, nil
}
