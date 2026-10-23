package localbackup

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/lib/pq"
)

func (graph *moduleRecoveryRowDependencyGraph) inspectConversationBusinessReferences(ctx context.Context) error {
	for _, ticket := range graph.Rows["tickets"] {
		if ticket.referenceColumnText("category") != "customer_service" {
			graph.Report.recordIssue("dependency_reaches_other_ticket_category", "blocking", graph.Side, "tickets", moduleRecoveryDependencyRowPublicIdentity(ticket, graph.Observation.PrimaryKeys["tickets"]), "客服引用链进入了其他类别工单，不能扩大覆盖范围。")
		}
	}
	for _, table := range []string{"tickets", "media_upload_sessions"} {
		column := "visitor_session_hash"
		if table == "media_upload_sessions" {
			column = "upload_actor_visitor_session_hash"
		}
		var keys []map[string]json.RawMessage
		var owners []*moduleRecoveryDependencyRow
		for _, row := range graph.Rows[table] {
			if row.referenceColumnText(column) != "" {
				keys = append(keys, map[string]json.RawMessage{"customer_service_visitor_hash": row.Values[column]})
				owners = append(owners, row)
			}
		}
		for start := 0; start < len(keys) && !graph.Exhausted; start += 200 {
			end := min(start+200, len(keys))
			profiles, err := graph.readRowsMatchingReferenceKeys(ctx, "visitor_profiles", []string{"customer_service_visitor_hash"}, keys[start:end], graph.Limits.Rows-graph.Count)
			if err != nil {
				return err
			}
			graph.includeRowsInDependencyGraph(profiles, "visitor_profiles", false)
			if graph.Exhausted {
				break
			}
			found := map[string]bool{}
			for _, profile := range profiles {
				found[profile.referenceColumnText("customer_service_visitor_hash")] = true
			}
			for index := start; index < end; index++ {
				if !found[owners[index].referenceColumnText(column)] {
					graph.Report.recordIssue("visitor_context_profile_missing", "review", graph.Side, table, moduleRecoveryDependencyRowPublicIdentity(owners[index], graph.Observation.PrimaryKeys[table]), "找不到访客浏览档案；会话身份以自身访客哈希和签名密钥为准，不能据此删除会话。")
				}
			}
		}
	}
	if err := graph.inspectCustomerServiceInboxReadCursors(ctx); err != nil {
		return err
	}
	graph.inspectCustomerServicePrivateAttachmentUploadReceiptAndRegisteredMediaActorIdentities()
	for _, session := range graph.Rows["media_upload_sessions"] {
		identity := moduleRecoveryDependencyRowPublicIdentity(session, graph.Observation.PrimaryKeys["media_upload_sessions"])
		if session.referenceColumnText("status") != "bound" {
			graph.Report.recordIssue("unfinished_conversation_upload", "blocking", graph.Side, "media_upload_sessions", identity, "上传收据尚未绑定消息，须先确认未完成上传及其文件的处置规则。")
		}
		ticket := findCustomerServiceDependencyRowByID(graph, "tickets", session.Values["business_context_id"])
		if ticket == nil {
			continue
		}
		conflict := session.referenceColumnText("business_context_type") != "customer_service" || ticket.referenceColumnText("category") != "customer_service"
		switch session.referenceColumnText("upload_actor_type") {
		case "visitor":
			conflict = conflict || session.referenceColumnText("upload_actor_visitor_session_hash") == "" || session.referenceColumnText("upload_actor_visitor_session_hash") != ticket.referenceColumnText("visitor_session_hash")
		case "customer":
			customer := ticket.referenceColumnText("customer_user_id")
			if customer == "" || customer == "0" {
				customer = ticket.referenceColumnText("user_id")
			}
			conflict = conflict || session.referenceColumnText("upload_actor_account_id") != customer
		case "administrator":
		default:
			conflict = true
		}
		if conflict {
			graph.Report.recordIssue("upload_actor_conversation_identity_conflict", "blocking", graph.Side, "media_upload_sessions", identity, "上传主体或业务上下文与会话身份不匹配。")
		}
	}
	for _, binding := range graph.Rows["ticket_message_media_attachments"] {
		identity := moduleRecoveryDependencyRowPublicIdentity(binding, graph.Observation.PrimaryKeys["ticket_message_media_attachments"])
		message := findCustomerServiceDependencyRowByID(graph, "ticket_messages", binding.Values["ticket_message_id"])
		if message != nil && message.referenceColumnText("attachment_reference_format") != "media_asset_bindings" {
			graph.Report.recordIssue("attachment_binding_format_conflict", "blocking", graph.Side, "ticket_message_media_attachments", identity, "媒体绑定与消息附件格式标记不一致。")
		}
		if binding.referenceColumnText("upload_session_id") == "" {
			continue
		} // Reusable replies may have no upload receipt.
		session := findCustomerServiceDependencyRowByID(graph, "media_upload_sessions", binding.Values["upload_session_id"])
		if message != nil && session != nil && (session.referenceColumnText("business_context_id") != message.referenceColumnText("ticket_id") || session.referenceColumnText("media_asset_id") != binding.referenceColumnText("media_asset_id") || session.referenceColumnText("status") != "bound") {
			graph.Report.recordIssue("attachment_binding_conversation_conflict", "blocking", graph.Side, "ticket_message_media_attachments", identity, "附件绑定、上传收据和消息不属于同一个会话或媒体对象。")
		}
	}
	graph.Report.recordIssue("configuration_and_event_state_preserved", "review", graph.Side, "outbox_events", "", "共享事件、客服配置与全局媒体任务保持原状态；事件诊断不授权重放、清理或重置幂等记录。")
	return nil
}

func findCustomerServiceDependencyRowByID(graph *moduleRecoveryRowDependencyGraph, table string, id json.RawMessage) *moduleRecoveryDependencyRow {
	return graph.Rows[table][moduleRecoveryDependencyRowKey(map[string]json.RawMessage{"id": id}, []string{"id"})]
}

func (graph *moduleRecoveryRowDependencyGraph) inspectCustomerServiceInboxReadCursors(ctx context.Context) error {
	var keys []map[string]json.RawMessage
	var inboxes []*moduleRecoveryDependencyRow
	for _, inbox := range graph.Rows["customer_service_inbox_states"] {
		if inbox.referenceColumnText("last_read_message_id") != "" && inbox.referenceColumnText("last_read_message_id") != "0" {
			keys = append(keys, map[string]json.RawMessage{"id": inbox.Values["last_read_message_id"]})
			inboxes = append(inboxes, inbox)
		}
	}
	for start := 0; start < len(keys) && !graph.Exhausted; start += 200 {
		end := min(start+200, len(keys))
		messages, err := graph.readRowsMatchingReferenceKeys(ctx, "ticket_messages", []string{"id"}, keys[start:end], end-start)
		if err != nil {
			return err
		}
		if graph.Exhausted {
			break
		}
		found := map[string]*moduleRecoveryDependencyRow{}
		for _, message := range messages {
			found[message.Identity] = message
		}
		for index := start; index < end; index++ {
			inbox := inboxes[index]
			message := found[moduleRecoveryDependencyRowKey(keys[index], []string{"id"})]
			if message == nil || message.referenceColumnText("ticket_id") != inbox.referenceColumnText("ticket_id") {
				graph.Report.recordIssue("inbox_read_cursor_outside_conversation", "blocking", graph.Side, "customer_service_inbox_states", moduleRecoveryDependencyRowPublicIdentity(inbox, graph.Observation.PrimaryKeys["customer_service_inbox_states"]), "阅读位置指向缺失消息或其他会话，不能直接恢复该游标。")
			}
		}
	}
	return nil
}

// Report actual incoming references from preserved tables without adding those
// rows to the recovery scope. Composite foreign keys are paired in order.
func (graph *moduleRecoveryRowDependencyGraph) inspectPreservedIncomingConversationReferences(ctx context.Context) error {
	return graph.inspectPreservedIncomingModuleObjectReferences(ctx, "preserved_row_references_conversation_object", "保留行引用会话或附件对象，须保护该行及其引用的现有内容。", false)
}

func (graph *moduleRecoveryRowDependencyGraph) inspectPreservedIncomingModuleObjectReferences(ctx context.Context, code, message string, includedSharedReferencesAreReview bool) error {
	var references []ModuleRecoveryDependencyReference
	for _, reference := range graph.Observation.References {
		if len(graph.Scoped[reference.TargetTable]) > 0 || reference.TargetTable == "media_assets" {
			references = append(references, reference)
		}
	}
	sort.Slice(references, func(i, j int) bool { return references[i].SourceTable < references[j].SourceTable })
	for _, reference := range references {
		if graph.Exhausted {
			break
		}
		var keys []map[string]json.RawMessage
		for _, row := range graph.Rows[reference.TargetTable] {
			key := map[string]json.RawMessage{}
			for index, column := range reference.TargetColumns {
				key[reference.SourceColumns[index]] = row.Values[column]
			}
			keys = append(keys, key)
		}
		for start := 0; start < len(keys) && !graph.Exhausted; start += 200 {
			data, err := json.Marshal(keys[start:min(start+200, len(keys))])
			if err != nil {
				return err
			}
			var equalities string
			for _, column := range reference.SourceColumns {
				if equalities != "" {
					equalities += " AND "
				}
				equalities += "record." + pq.QuoteIdentifier(column) + "=requested." + pq.QuoteIdentifier(column)
			}
			predicate := `EXISTS(SELECT 1 FROM jsonb_populate_recordset(NULL::public.` + pq.QuoteIdentifier(reference.SourceTable) + `,$1::jsonb) requested WHERE ` + equalities + `)`
			rows, err := graph.readBoundedRelationshipRows(ctx, reference.SourceTable, predicate, []any{string(data)}, graph.Limits.Rows)
			if err != nil {
				return err
			}
			for _, row := range rows {
				if graph.Scoped[reference.SourceTable][row.Identity] || reference.SourceTable == "media_asset_derivatives" {
					continue
				}
				severity := "blocking"
				if includedSharedReferencesAreReview && graph.Rows[reference.SourceTable][row.Identity] != nil {
					severity = "review"
				}
				graph.Report.recordIssue(code, severity, graph.Side, reference.SourceTable, moduleRecoveryDependencyRowPublicIdentity(row, graph.Observation.PrimaryKeys[reference.SourceTable]), message)
			}
		}
	}
	return nil
}
