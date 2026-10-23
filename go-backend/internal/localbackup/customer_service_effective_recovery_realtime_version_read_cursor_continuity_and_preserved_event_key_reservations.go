package localbackup

import (
	"encoding/json"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/outbox"
)

type customerServiceEffectiveRealtimeVersionIdentityIndex struct {
	Tickets map[string]*moduleRecoveryDependencyRow
	Inboxes map[string]*moduleRecoveryDependencyRow
}

func customerServiceRecoveryCanonicalNonnegativeInteger(raw json.RawMessage, maximum int64) (int64, bool) {
	value, err := strconv.ParseInt(string(raw), 10, 64)
	return value, err == nil && value >= 0 && value <= maximum && strconv.FormatInt(value, 10) == string(raw)
}

func customerServiceRecoveryCanonicalPositiveVersion(raw json.RawMessage) (int64, bool) {
	value, valid := customerServiceRecoveryCanonicalNonnegativeInteger(raw, 2147483647)
	return value, valid && value > 0
}

func buildCustomerServiceEffectiveRealtimeVersionIdentityIndex(graph *moduleRecoveryRowDependencyGraph) customerServiceEffectiveRealtimeVersionIdentityIndex {
	index := customerServiceEffectiveRealtimeVersionIdentityIndex{Tickets: map[string]*moduleRecoveryDependencyRow{}, Inboxes: map[string]*moduleRecoveryDependencyRow{}}
	for identity, row := range graph.Rows["tickets"] {
		if graph.Scoped["tickets"][identity] && row.referenceColumnText("category") == "customer_service" {
			index.Tickets[row.referenceColumnText("id")] = row
		}
	}
	for identity, row := range graph.Rows["customer_service_inbox_states"] {
		if graph.Scoped["customer_service_inbox_states"][identity] {
			index.Inboxes[row.referenceColumnText("ticket_id")+":"+row.referenceColumnText("recipient_user_id")] = row
		}
	}
	return index
}

// A retained Outbox key is still unique when ignored or soft deleted. Restoring
// an older per-row counter can therefore suppress a later, genuinely new event
// without rewinding any PostgreSQL sequence. Never silently rewrite full rows
// or clear receipts to compensate; reject unsafe effective selections instead.
func inspectCustomerServiceEffectiveRealtimeVersionAndReadCursorContinuity(effective, current *moduleRecoveryRowDependencyGraph) {
	for _, contract := range []struct{ table, version string }{{"tickets", "status_version"}, {"customer_service_inbox_states", "assignment_version"}} {
		if len(effective.Scoped[contract.table]) == 0 {
			continue
		}
		if !effective.Observation.Columns[contract.table][contract.version] {
			effective.markIncomplete("realtime_version_identity_schema_missing", contract.table, "缺少客服事件版本字段，无法证明恢复后不会复用历史事件身份。")
			continue
		}
		for identity := range effective.Scoped[contract.table] {
			row := effective.Rows[contract.table][identity]
			version, valid := customerServiceRecoveryCanonicalPositiveVersion(row.Values[contract.version])
			publicIdentity := moduleRecoveryDependencyRowPublicIdentity(row, effective.Observation.PrimaryKeys[contract.table])
			if !valid || version == 2147483647 {
				effective.Report.recordIssue("realtime_version_identity_invalid_or_exhausted", "blocking", effective.Side, contract.table, publicIdentity, "事件版本不是可继续递增的规范正整数，不能保证后续事件身份连续。")
				continue
			}
			previous := current.Rows[contract.table][identity]
			if previous == nil {
				continue
			}
			previousVersion, previousValid := customerServiceRecoveryCanonicalPositiveVersion(previous.Values[contract.version])
			if !previousValid {
				effective.markIncomplete("realtime_current_version_identity_invalid", contract.table, "当前事件版本无法核对，不能用包内版本覆盖掩盖问题。")
				continue
			}
			if version < previousVersion {
				effective.Report.recordIssue("realtime_version_counter_would_rewind", "blocking", effective.Side, contract.table, publicIdentity, "所选会话状态或收件箱分配版本低于当前值，会复用保留事件的幂等身份；须保留当前行或另行核对。")
			}
			if contract.table != "customer_service_inbox_states" || version != previousVersion {
				continue
			}
			cursor, cursorValid := customerServiceRecoveryCanonicalNonnegativeInteger(row.Values["last_read_message_id"], 9223372036854775807)
			previousCursor, previousCursorValid := customerServiceRecoveryCanonicalNonnegativeInteger(previous.Values["last_read_message_id"], 9223372036854775807)
			unread, unreadValid := customerServiceRecoveryCanonicalNonnegativeInteger(row.Values["unread_count"], 2147483647)
			previousUnread, previousUnreadValid := customerServiceRecoveryCanonicalNonnegativeInteger(previous.Values["unread_count"], 2147483647)
			if !cursorValid || !previousCursorValid || !unreadValid || !previousUnreadValid {
				effective.markIncomplete("realtime_read_identity_fields_invalid", contract.table, "阅读位置或未读数量无法核对，不能确认阅读事件身份连续。")
			} else if cursor < previousCursor || (cursor == previousCursor && unread > 0 && previousUnread == 0) {
				effective.Report.recordIssue("realtime_read_event_identity_would_be_reused", "blocking", effective.Side, contract.table, publicIdentity, "同一分配版本内回退阅读位置或重新激活已读状态，会再次使用已有阅读事件键；须保留当前收件箱行。")
			}
		}
	}
}

func inspectCustomerServicePreservedRealtimeEventVersionReservation(graph *moduleRecoveryRowDependencyGraph, index customerServiceEffectiveRealtimeVersionIdentityIndex, event *moduleRecoveryDependencyRow) bool {
	key := event.referenceColumnText("event_key")
	parts := strings.SplitN(key, ":", 6)
	typeContract := map[string]struct {
		eventType, versionColumn, recipientColumn string
		parts                                     int
	}{
		"customer_service.conversation.status.changed":      {"conversation.status.changed", "_payload_status_version", "", 3},
		"customer_service.conversation.assigned":            {"conversation.assigned", "_payload_assignment_version", "_payload_assigned_to", 4},
		"customer_service.messages.read":                    {"conversation.messages.read", "_payload_assignment_version", "_payload_read_by_user_id", 5},
		"customer_service.conversation.inbox_state.changed": {"conversation.inbox_state.changed", "_payload_assignment_version", "_payload_recipient_user_id", 6},
	}
	contract, registered := typeContract[parts[0]]
	payloadTicket := moduleRecoveryDurableEventPayloadPositiveID(event, "_payload_ticket_id")
	keyTicket := ""
	if len(parts) > 1 {
		keyTicket = parts[1]
	}
	if index.Tickets[payloadTicket] == nil && index.Tickets[keyTicket] == nil {
		return false
	}
	if !registered {
		for _, candidate := range typeContract {
			if event.referenceColumnText("_payload_type") == candidate.eventType {
				recordModuleRecoveryDurableEventIssue(graph, event, "realtime_version_event_key_contract_unrecognized", "blocking", "版本化实时事件没有可核对的确定性事件键，须核对历史契约。", false)
				return true
			}
		}
		return false
	}
	valid := len(parts) == contract.parts && canonicalModuleRecoveryDurableEventPositiveID(keyTicket)
	versionIndex := 2
	if contract.recipientColumn != "" {
		versionIndex = 3
		valid = valid && canonicalModuleRecoveryDurableEventPositiveID(parts[2])
	}
	var reservedVersion int64
	if valid {
		reservedVersion, valid = customerServiceRecoveryCanonicalPositiveVersion(json.RawMessage(parts[versionIndex]))
	}
	payloadVersion, payloadVersionValid := customerServiceRecoveryCanonicalPositiveVersion(event.Values[contract.versionColumn])
	valid = valid && payloadVersionValid && reservedVersion == payloadVersion && payloadTicket == keyTicket && event.referenceColumnText("event_type") == outbox.EventTypeCustomerServiceRealtime && event.referenceColumnText("aggregate_type") == outbox.AggregateTypeCustomerServiceConversation && event.referenceColumnText("aggregate_id") == keyTicket && event.referenceColumnText("_payload_type") == contract.eventType && event.referenceColumnText("_payload_event_id") == key
	if valid && contract.recipientColumn != "" {
		valid = moduleRecoveryDurableEventPayloadPositiveID(event, contract.recipientColumn) == parts[2]
	}
	if valid && contract.eventType == "conversation.messages.read" {
		cursor, cursorValid := customerServiceRecoveryCanonicalNonnegativeInteger(event.Values["_payload_last_read_message_id"], 9223372036854775807)
		valid = cursorValid && strconv.FormatInt(cursor, 10) == parts[4]
	}
	if !valid {
		recordModuleRecoveryDurableEventIssue(graph, event, "realtime_version_event_key_payload_conflict", "blocking", "保留事件的版本、收件人、会话或路由与确定性事件键不一致，不能确认后续幂等身份。", false)
		return true
	}
	owner, versionColumn := index.Tickets[keyTicket], "status_version"
	if contract.recipientColumn != "" {
		owner, versionColumn = index.Inboxes[keyTicket+":"+parts[2]], "assignment_version"
	}
	if owner == nil {
		recordModuleRecoveryDurableEventIssue(graph, event, "realtime_version_event_owner_missing", "blocking", "保留版本化事件的收件箱行缺失；后续重建可能从首个版本复用事件键。", false)
		return true
	}
	version, versionValid := customerServiceRecoveryCanonicalPositiveVersion(owner.Values[versionColumn])
	if !versionValid || version < reservedVersion {
		recordModuleRecoveryDurableEventIssue(graph, event, "realtime_preserved_event_version_ahead_of_effective_row", "blocking", "所选行版本低于保留事件已占用的版本，后续变更可能被历史事件去重。", false)
	}
	if contract.eventType == "conversation.messages.read" && version == reservedVersion {
		cursor, valid := customerServiceRecoveryCanonicalNonnegativeInteger(owner.Values["last_read_message_id"], 9223372036854775807)
		reservedCursor, _ := strconv.ParseInt(parts[4], 10, 64)
		if !valid || cursor < reservedCursor {
			recordModuleRecoveryDurableEventIssue(graph, event, "realtime_preserved_read_cursor_ahead_of_effective_row", "blocking", "同一分配版本内的阅读位置低于保留阅读事件，后续阅读会复用已占用事件键。", false)
		}
	}
	return true
}
