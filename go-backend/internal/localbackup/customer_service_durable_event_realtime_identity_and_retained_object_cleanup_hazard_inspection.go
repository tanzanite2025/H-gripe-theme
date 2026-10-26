package localbackup

import (
	"context"
	"encoding/json"
	"strings"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
)

func (cfg Configuration) inspectCustomerServiceDurableEventPayloadReferences(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, event *moduleRecoveryDependencyRow, scope *customerServiceDurableEventScope) (bool, error) {
	eventType, aggregateType := event.referenceColumnText("event_type"), event.referenceColumnText("aggregate_type")
	cleanupOverlaps := false
	if graph.SelectedCustomerServiceRows != nil && isCustomerServiceRecoveryAuditedCleanupEvent(eventType) && graph.Report.DurableEvents.CleanupDispositions != nil {
		previousBlockingIssues := graph.Report.BlockingIssues
		defer func() {
			recordCustomerServiceRecoveryCleanupEventDisposition(graph, event, cleanupOverlaps, previousBlockingIssues)
		}()
		if eventType == outbox.EventTypeObjectStorageCleanup && !customerServiceRecoveryObjectCleanupResourceContractRegistered(event.referenceColumnText("_payload_resource_type")) {
			recordModuleRecoveryDurableEventIssue(graph, event, "cleanup_disposition_resource_contract_unregistered", "blocking", "当前对象清理事件的资源类型尚无恢复处置契约，不能仅凭对象键判为安全保留。", true)
		}
	}
	declaredCustomerService := eventType == outbox.EventTypeCustomerServiceRealtime || eventType == outbox.EventTypeCustomerServiceRetentionCleanup || eventType == outbox.EventTypeCustomerServiceAvatarCleanup || aggregateType == outbox.AggregateTypeCustomerServiceConversation || aggregateType == outbox.AggregateTypeCustomerServiceAgentProfile
	if event.referenceColumnText("_payload_shape") != "object" {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_event_payload_not_an_object", "blocking", "事件载荷不是可识别对象，无法证明其会话或清理范围。", true)
		return true, nil
	}
	// An effective selection must audit every retained cleanup receipt, including
	// a different module's disjoint receipt and its execution/reconciliation lease.
	related := declaredCustomerService || graph.SelectedCustomerServiceRows != nil && isCustomerServiceRecoveryAuditedCleanupEvent(eventType)
	var cleanupPaths []string
	var cleanupKeys []string
	parseReferences := func(column string, required bool) []string {
		raw := event.Values[column]
		if len(raw) == 0 || string(raw) == "null" {
			if required {
				recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_reference_list_invalid", "blocking", "清理对象列表缺失，无法确认删除范围。", true)
			}
			return nil
		}
		var references []string
		if raw[0] != '[' || json.Unmarshal(raw, &references) != nil {
			recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_reference_list_invalid", "blocking", "清理对象列表格式不受支持，无法确认删除范围。", true)
			related = true
			return nil
		}
		return references
	}
	switch eventType {
	case outbox.EventTypeCustomerServiceRealtime:
		if err := inspectCustomerServiceRealtimeEventIdentity(ctx, graph, event); err != nil {
			return true, err
		}
	case outbox.EventTypeCustomerServiceRetentionCleanup:
		ticketID := moduleRecoveryDurableEventPayloadPositiveID(event, "_payload_ticket_id")
		if aggregateType != outbox.AggregateTypeCustomerServiceConversation || ticketID == "" || ticketID != event.referenceColumnText("aggregate_id") {
			recordModuleRecoveryDurableEventIssue(graph, event, "retention_cleanup_conversation_identity_conflict", "blocking", "保留期清理载荷的会话身份与事件聚合身份不一致。", false)
		}
		// A purged ticket may legitimately be absent: its durable references
		// survive in the event and still need to be compared with both snapshots.
		cleanupOverlaps = scope.Tickets[ticketID]
		for _, reference := range parseReferences("_payload_attachment_references", true) {
			root, key := cfg.resolveLocalMediaReference(reference)
			if key == "" {
				recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_reference_unresolved", "blocking", "历史清理引用无法映射到本地对象，须核对原存储及 CDN 处置范围。", true)
				continue
			}
			cleanupPaths = append(cleanupPaths, "media/"+root+"/"+key)
		}
	case outbox.EventTypeObjectStorageCleanup:
		cleanupKeys = parseReferences("_payload_object_keys", false)
		resourceType, resourceID := event.referenceColumnText("_payload_resource_type"), event.referenceColumnText("_payload_resource_id")
		for _, column := range []string{"_payload_resource_type", "_payload_resource_id"} {
			raw := event.Values[column]
			if len(raw) == 0 || raw[0] != '"' {
				recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_resource_identity_invalid", "blocking", "对象清理的资源身份必须为字符串，无法确认其执行范围。", true)
				related = true
			}
		}
		if resourceType != strings.TrimSpace(resourceType) || resourceID != strings.TrimSpace(resourceID) {
			recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_resource_identity_noncanonical", "blocking", "对象清理资源身份不是规范形式，须核对规范化后的实际对象。", true)
			related = true
		}
		resourceType, resourceID = strings.TrimSpace(resourceType), strings.TrimSpace(resourceID)
		if resourceType == "" || resourceID == "" {
			recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_resource_identity_missing", "blocking", "对象清理缺少资源身份，不能证明与恢复对象互不影响。", true)
			related = true
		}
		if resourceType == "customer_service_unregistered_upload" {
			related = true
			if aggregateType != outbox.AggregateTypeCustomerServiceConversation || !canonicalModuleRecoveryDurableEventPositiveID(event.referenceColumnText("aggregate_id")) {
				recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_resource_identity_conflict", "blocking", "客服上传清理的聚合身份不匹配。", false)
			}
		}
		for _, key := range cleanupKeys {
			normalized, valid := storage.NormalizeObjectKey(key)
			if !valid || !safePackagePath("media/uploads/"+normalized) {
				recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_object_key_unsafe", "blocking", "清理键包含不安全路径，未读取或删除任何对象。", true)
				related = true
				continue
			}
			if key != normalized {
				recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_object_key_noncanonical", "blocking", "清理键需规范化后执行，已按实际规范形式检查冲突；须先核对原始收据。", true)
				related = true
			}
			if scope.ObjectKeys[normalized] {
				related = true
				cleanupOverlaps = true
			}
			if strings.HasPrefix(normalized, "customer-service/") {
				related = true
			} // Hazard selection only; never an ownership assignment.
		}
		mediaOverlaps, err := cfg.inspectCustomerServiceMediaCleanupResourceAndRetiredDerivativeRegistrationIdentity(ctx, graph, event, scope, resourceType, resourceID, cleanupKeys, related)
		if err != nil {
			return true, err
		}
		related = related || mediaOverlaps
		cleanupOverlaps = cleanupOverlaps || mediaOverlaps
	case outbox.EventTypeCustomerServiceAvatarCleanup:
		if aggregateType != outbox.AggregateTypeCustomerServiceAgentProfile || !canonicalModuleRecoveryDurableEventPositiveID(event.referenceColumnText("aggregate_id")) {
			recordModuleRecoveryDurableEventIssue(graph, event, "avatar_cleanup_profile_identity_conflict", "blocking", "头像清理事件缺少正确的客服档案聚合身份。", false)
		}
		root, key := cfg.resolveLocalMediaReference(event.referenceColumnText("_payload_avatar_url"))
		if key == "" || root != "uploads" || !strings.HasPrefix(key, "customer-service/avatars/") {
			recordModuleRecoveryDurableEventIssue(graph, event, "avatar_cleanup_reference_unresolved", "blocking", "头像清理引用不属于可核对的本地头像命名空间。", true)
		} else {
			cleanupPaths = append(cleanupPaths, "media/"+root+"/"+key)
		}
	default:
		if related {
			recordModuleRecoveryDurableEventIssue(graph, event, "customer_service_event_type_unregistered", "blocking", "客服聚合出现未登记事件类型，须先补充恢复规则。", true)
		}
	}
	for _, path := range cleanupPaths {
		if scope.ExactPaths[path] || scope.AvatarPaths[path] {
			cleanupOverlaps = true
		}
	}
	if !related {
		return false, nil
	}
	if event.referenceColumnText("event_key") == "" {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_event_key_missing", "blocking", "事件缺少稳定幂等键，不能自动补入或消费。", false)
	}
	status := moduleRecoveryDurableEventPublicStatus(event)
	if status == "invalid" {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_event_status_unregistered", "blocking", "事件状态未登记，不能自动解释其执行结果。", true)
	}
	if event.referenceColumnText("deleted_at") == "" && (status == outbox.EventStatusProcessing || event.referenceColumnText("locked_at") != "" || event.referenceColumnText("locked_by") != "") {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_event_execution_lease_present", "blocking", "事件仍持有执行或原始收据对账租约，须先确认执行器与永久写入围栏。", false)
	}
	if status == outbox.EventStatusUnknown {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_event_outcome_unknown", "blocking", "事件外部结果未知，不能回退到待执行状态。", false)
	}
	if cleanupOverlaps {
		if status == outbox.EventStatusIgnored || event.referenceColumnText("deleted_at") != "" {
			recordModuleRecoveryDurableEventIssue(graph, event, "suppressed_cleanup_state_must_be_preserved", "review", "已忽略或已消费的清理收据命中恢复对象，必须保留该状态和原始围栏。", false)
		} else {
			// Processed object/avatar receipts participate in durable periodic
			// reconciliation. They cannot be treated as permanently inactive.
			recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_overlaps_retained_recovery_object", "blocking", "清理事件命中包内或目标仍需保留的对象；已处理的原始对象/头像收据仍可能再次对账，必须先核对处置。", false)
		}
	}
	return true, nil
}

func inspectCustomerServiceRealtimeEventIdentity(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, event *moduleRecoveryDependencyRow) error {
	if handled, err := inspectAuditedIgnoredLegacyCustomerServiceCreationNotificationPreservation(ctx, graph, event); handled || err != nil {
		return err
	}
	ticketID := moduleRecoveryDurableEventPayloadPositiveID(event, "_payload_ticket_id")
	if ticketID == "" || event.referenceColumnText("aggregate_type") != outbox.AggregateTypeCustomerServiceConversation || event.referenceColumnText("aggregate_id") != ticketID {
		recordModuleRecoveryDurableEventIssue(graph, event, "realtime_event_conversation_identity_conflict", "blocking", "实时事件的载荷与聚合会话身份不一致。", false)
	}
	if event.referenceColumnText("event_key") == "" || event.referenceColumnText("event_key") != event.referenceColumnText("_payload_event_id") {
		recordModuleRecoveryDurableEventIssue(graph, event, "realtime_event_idempotency_identity_conflict", "blocking", "实时事件载荷身份与幂等键不一致。", false)
	}
	var ticket *moduleRecoveryDependencyRow
	for _, candidate := range graph.Rows["tickets"] {
		if candidate.referenceColumnText("id") == ticketID {
			ticket = candidate
			break
		}
	}
	if ticket != nil && ticket.referenceColumnText("category") == "customer_service" {
		if event.referenceColumnText("_payload_conversation_id") != "" && event.referenceColumnText("_payload_conversation_id") != ticket.referenceColumnText("conversation_id") {
			recordModuleRecoveryDurableEventIssue(graph, event, "realtime_event_public_conversation_identity_conflict", "blocking", "实时事件的公开会话身份与会话记录不同。", false)
		}
	} else {
		severity := "blocking"
		if event.referenceColumnText("deleted_at") != "" || moduleRecoveryDurableEventPublicStatus(event) == outbox.EventStatusProcessed || moduleRecoveryDurableEventPublicStatus(event) == outbox.EventStatusIgnored {
			severity = "review"
		}
		recordModuleRecoveryDurableEventIssue(graph, event, "realtime_event_conversation_unavailable", severity, "实时事件所指会话已不存在或不在客服范围；历史已处理事件不作为删除会话或重放的依据。", false)
	}
	payloadType := event.referenceColumnText("_payload_type")
	messageColumn := ""
	switch payloadType {
	case "conversation.message.created":
		messageColumn = "_payload_message_id"
	case "conversation.messages.read":
		messageColumn = "_payload_last_read_message_id"
	case "conversation.created", "conversation.assigned", "conversation.status.changed", "conversation.inbox_state.changed":
	default:
		recordModuleRecoveryDurableEventIssue(graph, event, "realtime_event_payload_type_unregistered", "blocking", "实时事件载荷类型尚无本地恢复检查规则。", true)
	}
	if messageColumn != "" {
		messageID := moduleRecoveryDurableEventPayloadPositiveID(event, messageColumn)
		if messageColumn == "_payload_last_read_message_id" && string(event.Values[messageColumn]) == "0" {
			return nil
		}
		if messageID == "" {
			recordModuleRecoveryDurableEventIssue(graph, event, "realtime_event_message_identity_invalid", "blocking", "实时消息事件缺少合法的消息身份。", false)
			return nil
		}
		messages, err := graph.readRowsMatchingReferenceKeys(ctx, "ticket_messages", []string{"id"}, []map[string]json.RawMessage{{"id": event.Values[messageColumn]}}, 1)
		if err != nil {
			return err
		}
		if graph.Exhausted {
			return nil
		}
		if len(messages) == 0 || messages[0].referenceColumnText("ticket_id") != ticketID {
			severity := "blocking"
			if len(messages) == 0 && (event.referenceColumnText("deleted_at") != "" || moduleRecoveryDurableEventPublicStatus(event) == outbox.EventStatusProcessed || moduleRecoveryDurableEventPublicStatus(event) == outbox.EventStatusIgnored) {
				severity = "review"
			}
			recordModuleRecoveryDurableEventIssue(graph, event, "realtime_event_message_outside_conversation", severity, "实时消息事件指向缺失消息或其他会话，禁止自动重放。", false)
		}
	}
	return nil
}
