package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type CustomerServiceEffectiveSelectionInspectionRequest struct {
	PlanSHA256   string `json:"plan_sha256"`
	ReviewSHA256 string `json:"review_sha256"`
}

// This is an observation of a mixed selection, never a candidate or authority
// to restore. Original repair blockers survive even if a selection avoids them.
type CustomerServiceEffectiveSelectionInspection struct {
	Version                        int                                            `json:"version"`
	State                          string                                         `json:"state"`
	ExecutionAvailable             bool                                           `json:"execution_available"`
	CandidateConstructionAvailable bool                                           `json:"candidate_construction_available"`
	TaskID                         string                                         `json:"task_id"`
	PlanSHA256                     string                                         `json:"plan_sha256"`
	ReviewSHA256                   string                                         `json:"review_sha256"`
	SHA256                         string                                         `json:"sha256"`
	InspectedAt                    time.Time                                      `json:"inspected_at"`
	Binding                        CustomerServiceRecoveryPlanSnapshotBinding     `json:"binding"`
	SelectionDependenciesSatisfied bool                                           `json:"selection_dependencies_satisfied"`
	OriginalInspectionComplete     bool                                           `json:"original_inspection_complete"`
	RepairRequiredIssueCounts      map[string]int                                 `json:"repair_required_issue_counts"`
	DependencyBlockers             []ModuleRecoveryDependencyBlocker              `json:"dependency_blockers"`
	EffectiveRows                  []CustomerServiceRecoveryPlanRowAction         `json:"effective_rows"`
	EffectiveRowActionCounts       map[string]int                                 `json:"effective_row_action_counts"`
	RequiredFiles                  []CustomerServiceRecoveryPlanFileAction        `json:"required_files"`
	TargetPreservation             []CustomerServiceRecoveryPlanTablePreservation `json:"target_preservation"`
	DependencyChecks               *ModuleRecoveryRowInspection                   `json:"dependency_checks"`
}

func validateCustomerServiceSavedDecisionsAgainstLatestPlan(plan *CustomerServiceConversationRecoveryPlan, review *CustomerServiceRecoveryConflictReview, request CustomerServiceEffectiveSelectionInspectionRequest, now time.Time) (map[string]CustomerServiceRecoveryRecordedContentDecision, error) {
	if plan == nil || request.PlanSHA256 != plan.SHA256 {
		return nil, errors.New("recovery plan changed; inspect again before validating selections")
	}
	if err := verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(plan, plan.Binding, now); err != nil {
		return nil, err
	}
	policy := buildCustomerServiceRecoveryConflictDecisionPolicy(plan)
	if !policy.MappingComplete {
		return nil, errors.New("content decision mapping is incomplete")
	}
	decisions := map[string]CustomerServiceRecoveryRecordedContentDecision{}
	// A new plan with no content differences needs no artificial empty review.
	// An older revision remains history and is never inherited by the new plan.
	if policy.RequiredRowDecisions == 0 && request.ReviewSHA256 == "" && (review == nil || review.PlanSHA256 != plan.SHA256) {
		return decisions, nil
	}
	if review == nil {
		if request.ReviewSHA256 != "" || policy.RequiredRowDecisions != 0 {
			return nil, errors.New("save all required content decisions before checking selections")
		}
		return decisions, nil
	}
	copyOfReview := *review
	copyOfReview.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(&copyOfReview)
	if err != nil || digest != review.SHA256 || request.ReviewSHA256 != review.SHA256 || review.Version != 1 || review.State != "recorded_not_executable" || review.ExecutionAvailable || review.TaskID != plan.Binding.TaskID || review.PlanSHA256 != plan.SHA256 {
		return nil, errors.New("saved decisions changed or belong to another plan; reload and review")
	}
	bound, err := customerServiceRecoveryPlanCanonicalJSONSHA256(review.Binding)
	if err != nil {
		return nil, err
	}
	expected, err := customerServiceRecoveryPlanCanonicalJSONSHA256(plan.Binding)
	if err != nil || bound != expected {
		return nil, errors.New("saved decisions snapshot binding changed")
	}
	allowed := map[string]CustomerServiceRecoveryPlanRowAction{}
	for _, row := range plan.RowActions {
		if customerServiceRecoveryRowRequiresContentDecision(row) {
			allowed[customerServiceRecoveryContentDecisionKey(row.Table, row.IdentitySHA256)] = row
		}
	}
	if !review.ContentDecisionsComplete || review.PendingRowDecisions != 0 || review.RequiredRowDecisions != len(allowed) || len(review.Decisions) != len(allowed) {
		return nil, errors.New("content decisions are incomplete; save every required row first")
	}
	for _, decision := range review.Decisions {
		key := customerServiceRecoveryContentDecisionKey(decision.Table, decision.IdentitySHA256)
		row, exists := allowed[key]
		_, duplicate := decisions[key]
		if !exists || duplicate || decision.SourceSHA256 != row.SourceSHA256 || decision.TargetSHA256 != row.TargetSHA256 || (decision.Choice != "use_package_row" && (decision.Choice != "preserve_current_row" || row.Action != "propose_replace_after_review")) {
			return nil, errors.New("saved selection contains a duplicate, unsupported or mismatched row")
		}
		decisions[key] = decision
	}
	return decisions, nil
}

func (cfg Configuration) inspectCustomerServiceEffectiveSelectedRows(ctx context.Context, source, target *moduleRecoveryRowDependencyGraph, plan *CustomerServiceConversationRecoveryPlan, review *CustomerServiceRecoveryConflictReview, request CustomerServiceEffectiveSelectionInspectionRequest, manifest *Manifest, verifiedDirectory string) (*CustomerServiceEffectiveSelectionInspection, error) {
	decisions, err := validateCustomerServiceSavedDecisionsAgainstLatestPlan(plan, review, request, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if target == nil || target.Transaction == nil {
		return nil, errors.New("effective selection requires an existing current target database")
	}
	checks := &ModuleRecoveryRowInspection{Version: 1, Complete: true, RowLimitPerSide: target.Limits.Rows, ProjectionByteLimit: target.Limits.ProjectionBytes, IssueSampleLimit: target.Limits.IssueSamples, IssueCounts: map[string]int{}, Tables: []ModuleRecoveryRowComparison{}, Issues: []ModuleRecoveryRowIssue{}, ConversationReferenceChecks: &CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection{Version: 1}}
	effective := newModuleRecoveryRowDependencyGraph(target.Transaction, target.Observation, "target", checks, target.Limits)
	effective.SelectedCustomerServiceRows = map[string]map[string]*moduleRecoveryDependencyRow{}
	result := &CustomerServiceEffectiveSelectionInspection{Version: 1, State: "inspected_not_executable", TaskID: plan.Binding.TaskID, PlanSHA256: plan.SHA256, ReviewSHA256: request.ReviewSHA256, InspectedAt: time.Now().UTC(), Binding: plan.Binding, OriginalInspectionComplete: plan.InspectionComplete, RepairRequiredIssueCounts: buildCustomerServiceRecoveryConflictDecisionPolicy(plan).RepairRequiredIssueCounts, DependencyBlockers: plan.DependencyBlockers, EffectiveRows: []CustomerServiceRecoveryPlanRowAction{}, EffectiveRowActionCounts: map[string]int{}, RequiredFiles: []CustomerServiceRecoveryPlanFileAction{}, TargetPreservation: []CustomerServiceRecoveryPlanTablePreservation{}, DependencyChecks: checks}
	// Start with current scoped rows, including target-only rows and shared upload
	// receipts. Recompute their dependencies instead of carrying stale source-only
	// media into the effective closure.
	for table, rows := range target.Rows {
		for identity, row := range rows {
			if target.Scoped[table][identity] {
				effective.includeBoundedCustomerServiceEffectiveSelectionSeed(table, row)
			}
		}
	}
	lookup := map[string]*moduleRecoveryDependencyRow{}
	for table, rows := range source.Rows {
		for identity, row := range rows {
			hash, err := customerServiceRecoveryPlanCanonicalJSONSHA256(json.RawMessage(identity))
			if err != nil {
				return nil, err
			}
			lookup[customerServiceRecoveryContentDecisionKey(table, hash)] = row
		}
	}
	replacements := map[string][]json.RawMessage{}
	for _, action := range plan.RowActions {
		if decision, exists := decisions[customerServiceRecoveryContentDecisionKey(action.Table, action.IdentitySHA256)]; exists {
			if decision.Choice == "use_package_row" {
				row := lookup[customerServiceRecoveryContentDecisionKey(action.Table, action.IdentitySHA256)]
				if row == nil || row.Digest != action.SourceSHA256 || !source.Scoped[action.Table][row.Identity] {
					return nil, errors.New("selected source row is outside the observed conversation scope")
				}
				if effective.SelectedCustomerServiceRows[action.Table] == nil {
					effective.SelectedCustomerServiceRows[action.Table] = map[string]*moduleRecoveryDependencyRow{}
				}
				effective.SelectedCustomerServiceRows[action.Table][row.Identity] = row
				effective.includeBoundedCustomerServiceEffectiveSelectionSeed(action.Table, row)
				if action.TargetSHA256 != "" {
					replacements[action.Table] = append(replacements[action.Table], json.RawMessage(row.Identity))
					action.Action = "selected_package_replacement"
				} else {
					action.Action = "selected_package_insert"
				}
			} else {
				action.Action = "selected_preserve_current"
			}
		}
		result.EffectiveRows = append(result.EffectiveRows, action)
		result.EffectiveRowActionCounts[action.Action]++
	}
	if err := effective.includeRetainedIncomingRowsReferencingReplacedCustomerServiceKeys(ctx, target); err != nil {
		return nil, err
	}
	if err := effective.inspectCustomerServiceCandidateColumnReferenceSemantics(); err != nil {
		return nil, err
	}
	if err := effective.inspectCustomerServiceRetainedMediaOwnerColumnReferenceSemantics(); err != nil {
		return nil, err
	}
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
	inspectCustomerServiceEffectiveRealtimeVersionAndReadCursorContinuity(effective, target)
	result.RequiredFiles, err = cfg.inspectCustomerServiceEffectiveSelectionRequiredFiles(ctx, effective, manifest, verifiedDirectory)
	if err != nil {
		return nil, err
	}
	if err := cfg.inspectCustomerServiceDurableEventDependencies(ctx, []*moduleRecoveryRowDependencyGraph{effective}, checks); err != nil {
		return nil, err
	}
	var tables []string
	for table := range target.Observation.Tables {
		tables = append(tables, table)
	}
	sort.Strings(tables)
	for _, table := range tables {
		keys := replacements[table]
		sort.Slice(keys, func(i, j int) bool { return string(keys[i]) < string(keys[j]) })
		preserved, err := fingerprintCustomerServiceRecoveryPlanRetainedTargetTable(ctx, target.Transaction, table, target.Observation.Shapes[table], keys)
		if err != nil {
			return nil, err
		}
		result.TargetPreservation = append(result.TargetPreservation, preserved)
	}
	checks.TargetSelectedRows, checks.TargetInspectedRows = effective.Count, len(effective.ObservedRows)
	checks.ConversationReferenceChecks.Complete = checks.Complete
	result.SelectionDependenciesSatisfied = plan.InspectionComplete && checks.Complete && checks.BlockingIssues == 0 && len(result.RepairRequiredIssueCounts) == 0
	if !time.Now().Before(plan.Binding.WorkCopyExpiresAt) {
		return nil, errors.New("work copy expired while inspecting selections")
	}
	result.SHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(result)
	return result, err
}

func (graph *moduleRecoveryRowDependencyGraph) includeBoundedCustomerServiceEffectiveSelectionSeed(table string, row *moduleRecoveryDependencyRow) {
	encoded, err := json.Marshal(row.Values)
	key := table + ":" + row.Identity
	if err != nil || len(encoded) > 65536 || graph.ProjectionBytes+len(encoded) > graph.Limits.ProjectionBytes || !graph.ObservedRows[key] && len(graph.ObservedRows) >= graph.Limits.Rows {
		graph.Exhausted = true
		graph.markIncomplete("effective_selection_seed_resource_limit", table, "混合行集合的行数或引用投影超过上限，检查未完成。")
		return
	}
	graph.ProjectionBytes += len(encoded)
	graph.ObservedRows[key] = true
	graph.includeRowsInDependencyGraph([]*moduleRecoveryDependencyRow{row}, table, true)
	graph.Rows[table][row.Identity] = row
}

// Incoming references must be checked against OLD parent keys as well: a
// replacement can remove a referenced non-primary unique value without deleting
// the parent primary key. Read the effective children, then recheck their FKs.
func (graph *moduleRecoveryRowDependencyGraph) includeRetainedIncomingRowsReferencingReplacedCustomerServiceKeys(ctx context.Context, current *moduleRecoveryRowDependencyGraph) error {
	for _, reference := range graph.Observation.References {
		var keys []map[string]json.RawMessage
		for identity := range graph.SelectedCustomerServiceRows[reference.TargetTable] {
			parent := current.Rows[reference.TargetTable][identity]
			if parent == nil {
				parent = current.ComparedTargetRows[reference.TargetTable][identity]
			}
			if parent == nil {
				continue
			}
			key := map[string]json.RawMessage{}
			for index, column := range reference.TargetColumns {
				key[reference.SourceColumns[index]] = parent.Values[column]
			}
			keys = append(keys, key)
		}
		for start := 0; start < len(keys) && !graph.Exhausted; start += 200 {
			rows, err := graph.readRowsMatchingReferenceKeys(ctx, reference.SourceTable, reference.SourceColumns, keys[start:min(start+200, len(keys))], graph.Limits.Rows-graph.Count)
			if err != nil {
				return err
			}
			graph.includeRowsInDependencyGraph(rows, reference.SourceTable, false)
		}
	}
	return nil
}

func (cfg Configuration) inspectCustomerServiceEffectiveSelectionRequiredFiles(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, manifest *Manifest, verifiedDirectory string) ([]CustomerServiceRecoveryPlanFileAction, error) {
	manifestFiles := map[string]FileIdentity{}
	for _, file := range manifest.Files {
		manifestFiles[file.Path] = file
	}
	references := append(cfg.collectCustomerServiceConversationFileReferences(graph), graph.RetainedStructuredFiles...)
	files := map[string]CustomerServiceRecoveryPlanFileAction{}
	actualFiles := map[string]FileIdentity{}
	for _, reference := range references {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		path := reference.packagePath()
		existing, inspected := files[path]
		if !inspected && len(files) >= graph.Limits.Rows {
			graph.markIncomplete("effective_selection_file_resource_limit", reference.Table, "选择后的文件集合超过上限，检查未完成。")
			break
		}
		if inspected {
			actual := actualFiles[path]
			if actual.SHA256 != "" && (reference.Digest != "" && reference.Digest != actual.SHA256 || (reference.SizeKnown || reference.Size > 0) && reference.Size != actual.Size) {
				existing.Action = "blocked_effective_file_registration_conflict"
				files[path] = existing
				graph.Report.recordIssue(existing.Action, "blocking", graph.Side, reference.Table, reference.RowIdentity, "同路径的另一引用与拟采用文件的大小或摘要不一致。")
			}
			continue
		}
		hash := sha256.Sum256([]byte(path))
		action := CustomerServiceRecoveryPlanFileAction{PathSHA256: hex.EncodeToString(hash[:]), Action: "blocked_effective_file_unavailable"}
		actual, err := inspectBoundedModuleRecoveryRegularFileIdentity(ctx, cfg.StorageRoots[reference.Root], reference.Key, cfg.MaxExpandedBytes)
		if err == nil {
			action.Action, action.TargetSHA256, action.TargetSize = "preserve_current_required_file", actual.SHA256, actual.Size
		} else if errors.Is(err, os.ErrNotExist) {
			if expected, exists := manifestFiles[path]; exists {
				actual, err = inspectBoundedModuleRecoveryRegularFileIdentity(ctx, verifiedDirectory, path, expected.Size)
				if err == nil && actual.SHA256 == expected.SHA256 && actual.Size == expected.Size {
					action.Action, action.SourceSHA256, action.SourceSize = "selected_add_package_file", actual.SHA256, actual.Size
				} else {
					action.Action = "blocked_effective_package_file_changed"
				}
			}
		}
		if action.Action == "blocked_effective_file_unavailable" || action.Action == "blocked_effective_package_file_changed" {
			graph.Report.recordIssue(action.Action, "blocking", graph.Side, reference.Table, reference.RowIdentity, "最终选择需要的文件缺失、不可读或已变化，不能构造媒体候选。")
		} else if reference.Digest != "" && reference.Digest != actual.SHA256 || (reference.SizeKnown || reference.Size > 0) && reference.Size != actual.Size {
			action.Action = "blocked_effective_file_registration_conflict"
			graph.Report.recordIssue(action.Action, "blocking", graph.Side, reference.Table, reference.RowIdentity, "最终引用的媒体登记与拟采用文件的大小或摘要不一致。")
		}
		files[path] = action
		actualFiles[path] = actual
	}
	result := make([]CustomerServiceRecoveryPlanFileAction, 0, len(files))
	for _, action := range files {
		result = append(result, action)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].PathSHA256 < result[j].PathSHA256 })
	return result, nil
}

func (manager *Manager) InspectCustomerServiceRecoveryEffectiveSelection(ctx context.Context, id string, request CustomerServiceEffectiveSelectionInspectionRequest) (*CustomerServiceEffectiveSelectionInspection, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	review, err := manager.loadCustomerServiceRecoveryConflictReviewWhileOperationLocked(id)
	if err != nil {
		return nil, err
	}
	directory, err := manager.taskDirectory(id)
	if err != nil {
		return nil, err
	}
	var result *CustomerServiceEffectiveSelectionInspection
	_, err = manager.inspectVerifiedPackageModuleDependenciesWithCustomerServiceSelectionObserverWhileOperationLocked(ctx, id, "customer_service", func(ctx context.Context, report *ModuleRecoveryDependencyReport, source, target *moduleRecoveryRowDependencyGraph) error {
		task, err := manager.GetTask(id)
		if err != nil {
			return err
		}
		result, err = manager.Configuration.inspectCustomerServiceEffectiveSelectedRows(ctx, source, target, report.ConversationRows.ConversationRecoveryPlan, review, request, task.Manifest, filepath.Join(directory, "verified"))
		return err
	})
	if err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	if len(encoded) > 32<<20 {
		return nil, errors.New("effective selection report exceeds its size limit")
	}
	if err := manager.Configuration.checkWorkspaceCapacity(int64(len(encoded))); err != nil {
		return nil, err
	}
	if err := writeDurableFile(filepath.Join(directory, "customer-service-effective-selection-inspection-v1.json"), encoded); err != nil {
		return nil, errors.New("effective selection report could not be persisted")
	}
	return result, nil
}
