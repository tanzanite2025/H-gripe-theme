package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func (graph *moduleRecoveryRowDependencyGraph) selectCustomerServiceConversationScopeRows(ctx context.Context) error {
	if !graph.Observation.Columns["tickets"]["category"] {
		graph.markIncomplete("conversation_category_missing", "tickets", "缺少客服类别字段，不能把其他工单纳入客服恢复范围。")
		return nil
	}
	for _, seed := range []struct{ table, required, predicate string }{
		{"tickets", "id category", `record.category='customer_service'`},
		{"ticket_messages", "ticket_id", `EXISTS(SELECT 1 FROM tickets ticket WHERE ticket.id=record.ticket_id AND ticket.category='customer_service')`},
		{"customer_service_inbox_states", "ticket_id", `EXISTS(SELECT 1 FROM tickets ticket WHERE ticket.id=record.ticket_id AND ticket.category='customer_service')`},
		{"media_upload_sessions", "business_context_id business_context_type", `record.business_context_type='customer_service' AND EXISTS(SELECT 1 FROM tickets ticket WHERE ticket.id=record.business_context_id AND ticket.category='customer_service')`},
		{"ticket_message_media_attachments", "ticket_message_id", `EXISTS(SELECT 1 FROM ticket_messages message JOIN tickets ticket ON ticket.id=message.ticket_id WHERE message.id=record.ticket_message_id AND ticket.category='customer_service')`},
	} {
		valid := true
		for _, column := range strings.Fields(seed.required) {
			if !graph.Observation.Columns[seed.table][column] {
				valid = false
			}
		}
		if seed.table == "ticket_message_media_attachments" && !graph.Observation.Columns["ticket_messages"]["ticket_id"] {
			valid = false
		}
		if !valid {
			graph.markIncomplete("conversation_scope_table_or_column_missing", seed.table, "缺少会话范围所需表或字段。")
			continue
		}
		rows, err := graph.readBoundedRelationshipRows(ctx, seed.table, seed.predicate, nil, graph.Limits.Rows-graph.Count)
		if err != nil {
			return err
		}
		graph.includeRowsInDependencyGraph(rows, seed.table, true)
	}
	return nil
}

func (graph *moduleRecoveryRowDependencyGraph) inspectAndIncludeReferencedParentRows(ctx context.Context, reference ModuleRecoveryDependencyReference, selected []*moduleRecoveryDependencyRow) error {
	if len(reference.SourceColumns) != len(reference.TargetColumns) || len(reference.SourceColumns) == 0 {
		graph.markIncomplete("row_reference_unsupported", reference.SourceTable, "引用列无法配对。")
		return nil
	}
	var keys []map[string]json.RawMessage
	var children []*moduleRecoveryDependencyRow
	for _, row := range selected {
		key := map[string]json.RawMessage{}
		valid := true
		for index, column := range reference.SourceColumns {
			value := row.Values[column]
			if len(value) == 0 || string(value) == "null" || reference.Kind != "foreign_key" && (row.referenceColumnText(column) == "" || row.referenceColumnText(column) == "0") {
				valid = false
				break
			}
			key[reference.TargetColumns[index]] = value
		}
		if valid {
			keys = append(keys, key)
			children = append(children, row)
		}
	}
	for start := 0; start < len(keys) && !graph.Exhausted; start += 200 {
		end := min(start+200, len(keys))
		parents, err := graph.readRowsMatchingReferenceKeys(ctx, reference.TargetTable, reference.TargetColumns, keys[start:end], graph.Limits.Rows-graph.Count)
		if err != nil {
			return err
		}
		graph.includeRowsInDependencyGraph(parents, reference.TargetTable, false)
		if graph.Exhausted {
			break
		}
		found := map[string]bool{}
		for _, parent := range parents {
			found[moduleRecoveryDependencyRowKey(parent.Values, reference.TargetColumns)] = true
		}
		for index, key := range keys[start:end] {
			if !found[moduleRecoveryDependencyRowKey(key, reference.TargetColumns)] {
				child := children[start+index]
				graph.Report.recordIssue("referenced_row_missing", "blocking", graph.Side, reference.SourceTable, moduleRecoveryDependencyRowPublicIdentity(child, graph.Observation.PrimaryKeys[reference.SourceTable]), "引用的 "+reference.TargetTable+" 行不存在；不能用整表覆盖补救。")
			}
		}
	}
	return nil
}

func (graph *moduleRecoveryRowDependencyGraph) expandTransitiveForeignKeyAndAccountDependencies(ctx context.Context) error {
	references := append([]ModuleRecoveryDependencyReference{}, graph.Observation.References...)
	for _, entry := range [][4]string{
		{"tickets", "user_id", "users", "id"},
		{"tickets", "customer_user_id", "users", "id"}, {"tickets", "assigned_to", "users", "id"},
		{"ticket_messages", "user_id", "users", "id"}, {"media_upload_sessions", "upload_actor_account_id", "users", "id"},
		{"media_assets", "uploader_id", "users", "id"}, {"visitor_profiles", "user_id", "users", "id"},
	} {
		if graph.Observation.Columns[entry[0]][entry[1]] && graph.Observation.Columns[entry[2]][entry[3]] {
			references = append(references, ModuleRecoveryDependencyReference{SourceTable: entry[0], SourceColumns: []string{entry[1]}, TargetTable: entry[2], TargetColumns: []string{entry[3]}, Kind: "business_reference"})
		}
	}
	processed := map[string]bool{}
	for pass := 0; pass < 32 && !graph.Exhausted; pass++ {
		previousCount := graph.Count
		for referenceIndex, reference := range references {
			var selected []*moduleRecoveryDependencyRow
			for _, row := range graph.Rows[reference.SourceTable] {
				key := fmt.Sprintf("%d:%s", referenceIndex, row.Identity)
				if !processed[key] {
					selected = append(selected, row)
					processed[key] = true
				}
			}
			sort.Slice(selected, func(i, j int) bool { return selected[i].Identity < selected[j].Identity })
			if err := graph.inspectAndIncludeReferencedParentRows(ctx, reference, selected); err != nil {
				return err
			}
		}
		// Derivatives are incoming children, not reachable through outgoing FKs.
		var keys []map[string]json.RawMessage
		for _, asset := range graph.Rows["media_assets"] {
			key := "derivatives:" + asset.Identity
			if !processed[key] {
				keys = append(keys, map[string]json.RawMessage{"media_asset_id": asset.Values["id"]})
				processed[key] = true
			}
		}
		for start := 0; start < len(keys) && !graph.Exhausted; start += 200 {
			children, err := graph.readRowsMatchingReferenceKeys(ctx, "media_asset_derivatives", []string{"media_asset_id"}, keys[start:min(start+200, len(keys))], graph.Limits.Rows-graph.Count)
			if err != nil {
				return err
			}
			graph.includeRowsInDependencyGraph(children, "media_asset_derivatives", false)
		}
		if graph.Count == previousCount {
			return nil
		}
	}
	if !graph.Exhausted {
		graph.markIncomplete("row_reference_depth_limit", "", "依赖超过 32 轮，本版检查未完成闭包。")
	}
	return nil
}

func (cfg Configuration) inspectCustomerServiceConversationRows(ctx context.Context, source, target *sql.Tx, sourceObservation, targetObservation *moduleRecoveryDatabaseObservation, manifest *Manifest, verifiedDirectory string, limits moduleRecoveryRowInspectionLimits, observedGraphs ...func(*moduleRecoveryRowDependencyGraph, *moduleRecoveryRowDependencyGraph)) (*ModuleRecoveryRowInspection, error) {
	report := &ModuleRecoveryRowInspection{Version: 1, Complete: true, RowLimitPerSide: limits.Rows, ProjectionByteLimit: limits.ProjectionBytes, IssueSampleLimit: limits.IssueSamples, IssueCounts: map[string]int{}, Tables: []ModuleRecoveryRowComparison{}, Issues: []ModuleRecoveryRowIssue{}}
	report.ConversationReferenceChecks = &CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection{Version: 1}
	sourceGraph := newModuleRecoveryRowDependencyGraph(source, sourceObservation, "package", report, limits)
	graphs := []*moduleRecoveryRowDependencyGraph{sourceGraph}
	var targetGraph *moduleRecoveryRowDependencyGraph
	if target != nil && targetObservation != nil {
		targetGraph = newModuleRecoveryRowDependencyGraph(target, targetObservation, "target", report, limits)
		graphs = append(graphs, targetGraph)
	} else {
		report.Complete = false
	}
	for _, graph := range graphs {
		if err := graph.inspectCustomerServiceCandidateColumnReferenceSemantics(); err != nil {
			return nil, err
		}
		if err := graph.inspectCustomerServiceRetainedMediaOwnerColumnReferenceSemantics(); err != nil {
			return nil, err
		}
		graph.inspectCustomerServiceCandidateWriteSideEffectContracts()
		if err := graph.selectCustomerServiceConversationScopeRows(ctx); err != nil {
			return nil, err
		}
		if err := graph.expandTransitiveForeignKeyAndAccountDependencies(ctx); err != nil {
			return nil, err
		}
		if err := graph.inspectConversationBusinessReferences(ctx); err != nil {
			return nil, err
		}
		if err := cfg.includeCustomerServiceLegacyAttachmentAssets(ctx, graph); err != nil {
			return nil, err
		}
		if err := graph.expandTransitiveForeignKeyAndAccountDependencies(ctx); err != nil {
			return nil, err
		}
		if err := graph.inspectPreservedIncomingConversationReferences(ctx); err != nil {
			return nil, err
		}
	}
	if err := inspectCustomerServiceConversationNaturalKeyCollisions(ctx, sourceGraph, targetGraph); err != nil {
		return nil, err
	}
	if err := inspectCustomerServiceStructuredBusinessSnapshotDependencies(ctx, sourceGraph, targetGraph); err != nil {
		return nil, err
	}
	report.SourceSelectedRows = sourceGraph.Count
	if targetGraph != nil {
		report.TargetSelectedRows = targetGraph.Count
		if err := compareCustomerServiceConversationRowGraphs(ctx, sourceGraph, targetGraph); err != nil {
			return nil, err
		}
	}
	if err := cfg.inspectCustomerServiceConversationAttachmentFiles(ctx, graphs, manifest, verifiedDirectory, report); err != nil {
		return nil, err
	}
	if err := cfg.inspectCustomerServiceDurableEventDependencies(ctx, graphs, report); err != nil {
		return nil, err
	}
	report.SourceInspectedRows = len(sourceGraph.ObservedRows)
	if targetGraph != nil {
		report.TargetInspectedRows = len(targetGraph.ObservedRows)
	}
	var err error
	report.ConversationReferenceChecks.Complete = report.Complete
	report.ConversationRecoveryPlan, err = buildCustomerServiceConversationRecoveryPlanDraft(ctx, sourceGraph, targetGraph)
	if err != nil {
		return nil, err
	}
	for _, observe := range observedGraphs {
		observe(sourceGraph, targetGraph)
	}
	return report, nil
}

func compareCustomerServiceConversationRowGraphs(ctx context.Context, source, target *moduleRecoveryRowDependencyGraph) error {
	return compareModuleRecoveryDependencyRowGraphs(ctx, source, target, moduleRecoveryRowComparisonPolicy{
		ScopedContentDifferenceCode: "conversation_row_content_differs", ScopedContentDifferenceMessage: "会话范围内的行与恢复点不同，须明确覆盖后会丢失的后续记录。",
		IdentityConflictMessage: "同一主键已属于其他会话、客户身份或附件对象，不能直接覆盖。",
		TargetOnlyCode:          "target_only_conversation_row", TargetOnlyMessage: "当前会话行不在包内，局部覆盖必须先明确其保留或回退规则。", OwnershipConflicts: customerServiceConversationRowOwnershipConflicts,
	})
}

func customerServiceConversationRowOwnershipConflicts(table string, source, target *moduleRecoveryDependencyRow) bool {
	fields := map[string]string{
		"tickets":         "category conversation_id user_id customer_user_id visitor_session_hash",
		"ticket_messages": "ticket_id", "customer_service_inbox_states": "ticket_id recipient_user_id",
		"media_upload_sessions":            "business_context_type business_context_id upload_actor_type upload_actor_account_id upload_actor_visitor_session_hash media_asset_id object_key",
		"ticket_message_media_attachments": "ticket_message_id media_asset_id upload_session_id",
		"visitor_profiles":                 "customer_service_visitor_hash user_id", "media_assets": "storage_key visibility library_partition customer_service_media_purpose",
		"media_asset_derivatives": "media_asset_id storage_key",
	}
	for _, column := range strings.Fields(fields[table]) {
		if source.referenceColumnText(column) != target.referenceColumnText(column) {
			return true
		}
	}
	return false
}
