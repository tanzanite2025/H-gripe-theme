package localbackup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"

	"github.com/lib/pq"
)

// This versioned contract is advisory. It never authorizes insertion, resetting
// a receipt, invoking a provider or replaying an event during module recovery.
type ModuleRecoverySideEffectReceiptInspection struct {
	Version                              int                                         `json:"version"`
	Complete                             bool                                        `json:"complete"`
	StatePolicy                          string                                      `json:"state_policy"`
	ComparedIdentities                   int                                         `json:"compared_identities"`
	ComparisonSampleLimit                int                                         `json:"comparison_sample_limit"`
	ComparisonsTruncated                 bool                                        `json:"comparisons_truncated"`
	SourceNotificationEventsWithoutAudit int                                         `json:"source_notification_events_without_audit"`
	TargetNotificationEventsWithoutAudit int                                         `json:"target_notification_events_without_audit"`
	NotificationAuditPolicy              string                                      `json:"notification_audit_policy"`
	Tables                               []ModuleRecoverySideEffectReceiptTable      `json:"tables"`
	Comparisons                          []ModuleRecoverySideEffectReceiptComparison `json:"comparisons"`
}

type ModuleRecoverySideEffectReceiptTable struct {
	Table             string `json:"table"`
	SourceScanned     int    `json:"source_scanned"`
	TargetScanned     int    `json:"target_scanned"`
	Matching          int    `json:"matching"`
	SourceOnly        int    `json:"source_only"`
	TargetOnly        int    `json:"target_only"`
	Conflicts         int    `json:"conflicts"`
	BlockingIssues    int    `json:"blocking_issues"`
	PreservationScope string `json:"preservation_scope"`
}

type ModuleRecoverySideEffectReceiptComparison struct {
	Table           string `json:"table"`
	IdentitySHA256  string `json:"identity_sha256"`
	SourceReceiptID string `json:"source_receipt_id,omitempty"`
	TargetReceiptID string `json:"target_receipt_id,omitempty"`
	Result          string `json:"result"`
}

type moduleRecoverySideEffectReceiptContract struct {
	Table, RequiredColumns      string
	NaturalKey, IdentityColumns []string
}

func registeredModuleRecoverySideEffectReceiptPreservationContracts() []moduleRecoverySideEffectReceiptContract {
	return []moduleRecoverySideEffectReceiptContract{
		{"email_delivery_records", "id outbox_event_id event_key event_type template_code template_locale template_version recipient_email subject reference_type reference_number status attempt_count provider_code last_error first_attempt_at last_attempt_at sent_at created_at updated_at", []string{"event_key"}, strings.Fields("outbox_event_id event_type template_code template_locale template_version reference_type reference_number")},
		{"payment_operation_idempotencies", "id user_id scope idempotency_key request_hash status status_code content_type response_body claim_token lease_expires_at reconciliation_started_at created_at updated_at", strings.Fields("user_id scope idempotency_key"), []string{"request_hash"}},
		{"refund_idempotencies", "id admin_user_id scope idempotency_key request_hash refund_id created_at updated_at", strings.Fields("scope idempotency_key"), strings.Fields("admin_user_id request_hash refund_id")},
	}
}

func inspectModuleRecoveryPaymentRefundAndNotificationReceiptPreservation(ctx context.Context, graphs []*moduleRecoveryRowDependencyGraph, report *ModuleRecoveryRowInspection) error {
	inspection := &ModuleRecoverySideEffectReceiptInspection{Version: 1, StatePolicy: "preserve_all_target_receipts_no_automatic_import_or_replay", NotificationAuditPolicy: "best_effort_audit_is_not_a_delivery_idempotency_lock", ComparisonSampleLimit: report.IssueSampleLimit, Tables: []ModuleRecoverySideEffectReceiptTable{}, Comparisons: []ModuleRecoverySideEffectReceiptComparison{}}
	report.SideEffectReceipts = inspection
	for _, contract := range registeredModuleRecoverySideEffectReceiptPreservationContracts() {
		blockingIssuesBeforeTable := report.BlockingIssues
		summary := ModuleRecoverySideEffectReceiptTable{Table: contract.Table, PreservationScope: "entire_target_table_including_unbound_and_other_users"}
		var snapshots [][]*moduleRecoveryDependencyRow
		for _, graph := range graphs {
			valid := true
			for _, column := range strings.Fields(contract.RequiredColumns) {
				if !graph.Observation.Columns[contract.Table][column] {
					valid = false
				}
			}
			if !valid {
				graph.markIncomplete("side_effect_receipt_schema_missing", contract.Table, "缺少副作用凭据身份、状态或返回结果字段，不能完成保留检查。")
				snapshots = append(snapshots, nil)
				continue
			}
			rows, err := graph.readBoundedRelationshipRows(ctx, contract.Table, "TRUE", nil, graph.Limits.Rows)
			if err != nil {
				return err
			}
			snapshots = append(snapshots, rows)
			if graph.Side == "package" {
				summary.SourceScanned = len(rows)
			} else {
				summary.TargetScanned = len(rows)
			}
			for _, row := range rows {
				if err := inspectModuleRecoverySideEffectReceiptReferencesAndUncertainState(ctx, graph, contract, row); err != nil {
					return err
				}
			}
		}
		if len(graphs) == 2 {
			compareModuleRecoverySideEffectReceiptNaturalIdentities(graphs, snapshots, contract, &summary, inspection)
		}
		summary.BlockingIssues = report.BlockingIssues - blockingIssuesBeforeTable
		inspection.Tables = append(inspection.Tables, summary)
	}
	for _, graph := range graphs {
		count, err := inspectModuleRecoveryNotificationEventsWithoutDeliveryAudit(ctx, graph)
		if err != nil {
			return err
		}
		if graph.Side == "package" {
			inspection.SourceNotificationEventsWithoutAudit = count
		} else {
			inspection.TargetNotificationEventsWithoutAudit = count
		}
	}
	inspection.Complete = report.Complete && len(graphs) == 2
	return nil
}

func moduleRecoverySideEffectReceiptNaturalIdentity(row *moduleRecoveryDependencyRow, columns []string) string {
	values := make([]json.RawMessage, 0, len(columns))
	for _, column := range columns {
		if strings.TrimSpace(row.referenceColumnText(column)) == "" {
			return ""
		}
		values = append(values, row.Values[column])
	}
	encoded, _ := json.Marshal(values)
	return string(encoded)
}

func recordModuleRecoverySideEffectReceiptIssue(graph *moduleRecoveryRowDependencyGraph, table string, row *moduleRecoveryDependencyRow, code, severity, message string) {
	graph.Report.recordIssue(code, severity, graph.Side, table, moduleRecoverySideEffectReceiptPublicNumericID(row), message)
}

func moduleRecoverySideEffectReceiptPublicNumericID(row *moduleRecoveryDependencyRow) string {
	if value := row.referenceColumnText("id"); canonicalModuleRecoveryDurableEventPositiveID(value) {
		return value
	}
	return "invalid"
}

func compareModuleRecoverySideEffectReceiptNaturalIdentities(graphs []*moduleRecoveryRowDependencyGraph, snapshots [][]*moduleRecoveryDependencyRow, contract moduleRecoverySideEffectReceiptContract, summary *ModuleRecoverySideEffectReceiptTable, inspection *ModuleRecoverySideEffectReceiptInspection) {
	byKey := []map[string]*moduleRecoveryDependencyRow{{}, {}}
	byID := []map[string]*moduleRecoveryDependencyRow{{}, {}}
	keySet := map[string]bool{}
	for index, rows := range snapshots {
		for _, row := range rows {
			byID[index][row.Identity] = row
			key := moduleRecoverySideEffectReceiptNaturalIdentity(row, contract.NaturalKey)
			if key == "" {
				recordModuleRecoverySideEffectReceiptIssue(graphs[index], contract.Table, row, "side_effect_receipt_natural_key_invalid", "blocking", "副作用凭据缺少完整自然幂等身份，不能导入或自动重建。")
				continue
			}
			if previous := byKey[index][key]; previous != nil && previous.Identity != row.Identity {
				graphs[index].Report.Complete = false
				recordModuleRecoverySideEffectReceiptIssue(graphs[index], contract.Table, row, "side_effect_receipt_natural_key_not_unique", "blocking", "同一自然幂等身份对应多个凭据，不能自动选择。")
			}
			byKey[index][key] = row
			keySet[key] = true
		}
	}
	var keys []string
	for key := range keySet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		left, right := byKey[0][key], byKey[1][key]
		digest := sha256.Sum256([]byte(key))
		comparison := ModuleRecoverySideEffectReceiptComparison{Table: contract.Table, IdentitySHA256: hex.EncodeToString(digest[:])}
		if left != nil {
			comparison.SourceReceiptID = moduleRecoverySideEffectReceiptPublicNumericID(left)
		}
		if right != nil {
			comparison.TargetReceiptID = moduleRecoverySideEffectReceiptPublicNumericID(right)
		}
		switch {
		case left == nil:
			summary.TargetOnly++
			comparison.Result = "target_only_preserved"
		case right == nil:
			summary.SourceOnly++
			comparison.Result = "source_only_requires_reconciliation"
			recordModuleRecoverySideEffectReceiptIssue(graphs[0], contract.Table, left, "source_side_effect_receipt_missing_from_target", "blocking", "包内副作用凭据在目标缺失，结果或去重边界未知；不得补入后重放。")
		default:
			identityDiffers := false
			for _, column := range contract.IdentityColumns {
				identityDiffers = identityDiffers || left.referenceColumnText(column) != right.referenceColumnText(column)
			}
			switch {
			case identityDiffers:
				comparison.Result = "natural_key_identity_conflict"
				recordModuleRecoverySideEffectReceiptIssue(graphs[1], contract.Table, right, "side_effect_receipt_natural_key_identity_conflict", "blocking", "同一自然幂等键对应不同请求、事件、模板或退款身份，禁止覆盖。")
			case left.Identity != right.Identity:
				comparison.Result = "receipt_identifier_moved"
				recordModuleRecoverySideEffectReceiptIssue(graphs[1], contract.Table, right, "side_effect_receipt_identifier_moved", "blocking", "相同副作用凭据在两侧主键不同，须保留目标身份并核对引用。")
			case left.Digest != right.Digest:
				comparison.Result = "target_state_preserved_requires_reconciliation"
				recordModuleRecoverySideEffectReceiptIssue(graphs[1], contract.Table, right, "side_effect_receipt_current_state_differs", "blocking", "副作用凭据状态、租约、返回结果或其他内容变化，须保留目标完整行，不能被旧包覆盖。")
			default:
				comparison.Result = "same_record_preserved"
				summary.Matching++
			}
			if comparison.Result != "same_record_preserved" {
				summary.Conflicts++
			}
		}
		for index, row := range []*moduleRecoveryDependencyRow{left, right} {
			if row == nil {
				continue
			}
			other := byID[1-index][row.Identity]
			if other != nil && moduleRecoverySideEffectReceiptNaturalIdentity(other, contract.NaturalKey) != key {
				recordModuleRecoverySideEffectReceiptIssue(graphs[1-index], contract.Table, other, "side_effect_receipt_primary_key_reassigned", "blocking", "副作用凭据主键在另一侧属于不同自然幂等键，不能按主键覆盖或重编号。")
			}
		}
		inspection.ComparedIdentities++
		if len(inspection.Comparisons) < inspection.ComparisonSampleLimit {
			inspection.Comparisons = append(inspection.Comparisons, comparison)
		} else {
			inspection.ComparisonsTruncated = true
		}
	}
}

func inspectModuleRecoverySideEffectReceiptReferencesAndUncertainState(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, contract moduleRecoverySideEffectReceiptContract, row *moduleRecoveryDependencyRow) error {
	switch contract.Table {
	case "payment_operation_idempotencies":
		if !canonicalModuleRecoveryDurableEventPositiveID(row.referenceColumnText("user_id")) || !moduleRecoverySideEffectReceiptRequestDigestIsValid(row.referenceColumnText("request_hash")) {
			recordModuleRecoverySideEffectReceiptIssue(graph, contract.Table, row, "payment_receipt_request_identity_invalid", "blocking", "支付幂等凭据的客户或请求摘要无效，不能从返回正文猜测订单归属。")
		}
		switch row.referenceColumnText("status") {
		case "completed":
			if row.referenceColumnText("_claim_token_present") == "true" || row.referenceColumnText("lease_expires_at") != "" {
				recordModuleRecoverySideEffectReceiptIssue(graph, contract.Table, row, "completed_payment_receipt_execution_lease_present", "blocking", "已完成支付凭据仍有执行租约或认领标记，须核对执行器；不会输出认领凭据。")
			}
		case "pending", "reconciling":
			recordModuleRecoverySideEffectReceiptIssue(graph, contract.Table, row, "payment_receipt_external_outcome_requires_reconciliation", "blocking", "支付幂等请求未完成，租约过期也不能证明外部操作失败，须保留并对账。")
		default:
			graph.Report.Complete = false
			recordModuleRecoverySideEffectReceiptIssue(graph, contract.Table, row, "payment_receipt_status_unregistered", "blocking", "支付幂等状态未登记，无法推断外部执行结果。")
		}
	case "refund_idempotencies":
		if !canonicalModuleRecoveryDurableEventPositiveID(row.referenceColumnText("admin_user_id")) || !moduleRecoverySideEffectReceiptRequestDigestIsValid(row.referenceColumnText("request_hash")) {
			recordModuleRecoverySideEffectReceiptIssue(graph, contract.Table, row, "refund_receipt_request_identity_invalid", "blocking", "退款请求凭据缺少有效操作员身份或请求摘要。")
		}
		if row.referenceColumnText("refund_id") == "" {
			recordModuleRecoverySideEffectReceiptIssue(graph, contract.Table, row, "unbound_refund_receipt_preserved", "review", "未绑定退款的全局幂等键继续保留；缺少退款编号不等于允许重新提交。")
			return nil
		}
		if !canonicalModuleRecoveryDurableEventPositiveID(row.referenceColumnText("refund_id")) {
			recordModuleRecoverySideEffectReceiptIssue(graph, contract.Table, row, "refund_receipt_business_identity_invalid", "blocking", "退款凭据的退款编号无效，不能自动重新绑定。")
			return nil
		}
		if !graph.Observation.Columns["refunds"]["id"] {
			graph.markIncomplete("refund_receipt_reference_schema_missing", "refunds", "缺少退款身份字段，无法核对退款请求凭据。")
			return nil
		}
		rows, err := graph.readRowsMatchingReferenceKeys(ctx, "refunds", []string{"id"}, []map[string]json.RawMessage{{"id": row.Values["refund_id"]}}, 1)
		if err != nil {
			return err
		}
		if len(rows) != 1 {
			recordModuleRecoverySideEffectReceiptIssue(graph, contract.Table, row, "refund_receipt_business_reference_missing", "blocking", "退款凭据指向的退款不存在，不能单独恢复凭据或生成新退款。")
		}
	case "email_delivery_records":
		return inspectModuleRecoveryNotificationReceiptEventAndReplayRisk(ctx, graph, row)
	}
	return nil
}

func moduleRecoverySideEffectReceiptRequestDigestIsValid(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func inspectModuleRecoveryNotificationReceiptEventAndReplayRisk(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, row *moduleRecoveryDependencyRow) error {
	const table = "email_delivery_records"
	status := row.referenceColumnText("status")
	switch status {
	case "sent":
		if row.referenceColumnText("sent_at") == "" {
			recordModuleRecoverySideEffectReceiptIssue(graph, table, row, "notification_receipt_sent_time_missing", "blocking", "邮件标记已发送但缺少完成时间，须核对审计与服务商结果。")
		}
	case "failed":
	case "sending", "unknown":
		recordModuleRecoverySideEffectReceiptIssue(graph, table, row, "notification_receipt_outcome_requires_reconciliation", "blocking", "邮件发送中或结果未知，必须核对外部结果，不能自动重发。")
	default:
		graph.Report.Complete = false
		recordModuleRecoverySideEffectReceiptIssue(graph, table, row, "notification_receipt_status_unregistered", "blocking", "邮件发送状态未登记，无法推断投递结果。")
	}
	for _, column := range strings.Fields("id event_key event_type payload status deleted_at") {
		if !graph.Observation.Columns["outbox_events"][column] {
			graph.markIncomplete("notification_receipt_event_schema_missing", "outbox_events", "缺少邮件事件关联或消费状态字段。")
			return nil
		}
	}
	if !canonicalModuleRecoveryDurableEventPositiveID(row.referenceColumnText("outbox_event_id")) {
		recordModuleRecoverySideEffectReceiptIssue(graph, table, row, "notification_receipt_event_identity_invalid", "blocking", "邮件审计缺少有效事件编号。")
		return nil
	}
	graph.InspectOrderEventPayload = true
	events, err := graph.readRowsMatchingReferenceKeys(ctx, "outbox_events", []string{"id"}, []map[string]json.RawMessage{{"id": row.Values["outbox_event_id"]}}, 1)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		severity := "blocking"
		if status == "sent" || status == "failed" {
			severity = "review"
		}
		recordModuleRecoverySideEffectReceiptIssue(graph, table, row, "notification_receipt_event_unavailable", severity, "审计关联事件不存在；保留历史收据，不能据此创建或重发事件。")
		return nil
	}
	event := events[0]
	template := moduleRecoveryNotificationReceiptExpectedTemplateCode(event)
	idempotencyKey := strings.TrimSpace(orderTransactionDurableEventPayloadString(event, "_payload_idempotency_key"))
	if idempotencyKey == "" {
		idempotencyKey = strings.TrimSpace(event.referenceColumnText("event_key"))
	}
	if idempotencyKey == "" {
		idempotencyKey = "event:" + event.referenceColumnText("id")
	}
	expectedKey := "notification:" + idempotencyKey + ":" + template
	if template == "" || template != row.referenceColumnText("template_code") || row.referenceColumnText("event_type") != event.referenceColumnText("event_type") || row.referenceColumnText("event_key") != expectedKey {
		recordModuleRecoverySideEffectReceiptIssue(graph, table, row, "notification_receipt_event_identity_conflict", "blocking", "邮件审计的事件类型、模板或通知幂等键与事件不一致；通知键不等同于 Outbox 键。")
	}
	expectedReferenceType, expectedReferenceNumber := "order", strings.TrimSpace(orderTransactionDurableEventPayloadString(event, "_payload_order_number"))
	if event.referenceColumnText("event_type") == "after_sales.status_changed" {
		expectedReferenceType, expectedReferenceNumber = "after_sales", moduleRecoveryDurableEventPayloadPositiveID(event, "_payload_case_id")
	}
	if expectedReferenceNumber != "" && (row.referenceColumnText("reference_type") != expectedReferenceType || row.referenceColumnText("reference_number") != expectedReferenceNumber) {
		recordModuleRecoverySideEffectReceiptIssue(graph, table, row, "notification_receipt_business_reference_conflict", "blocking", "邮件审计的业务引用与事件订单号或售后编号不一致。")
	}
	if status == "sent" && event.referenceColumnText("deleted_at") == "" && moduleRecoveryDurableEventPublicStatus(event) != "processed" && moduleRecoveryDurableEventPublicStatus(event) != "ignored" {
		recordModuleRecoverySideEffectReceiptIssue(graph, table, row, "notification_sent_receipt_has_replayable_event", "blocking", "邮件已发送但事件仍可能执行；审计行不是发送去重锁，须核对并处置事件后才能开放任务。")
	}
	return nil
}

func inspectModuleRecoveryNotificationEventsWithoutDeliveryAudit(ctx context.Context, graph *moduleRecoveryRowDependencyGraph) (int, error) {
	if !graph.Observation.Columns["email_delivery_records"]["outbox_event_id"] {
		return 0, nil // The receipt schema validator has already marked this incomplete.
	}
	for _, column := range strings.Fields("id event_type payload") {
		if !graph.Observation.Columns["outbox_events"][column] {
			graph.markIncomplete("notification_receipt_event_schema_missing", "outbox_events", "缺少通知事件字段，无法核对缺失审计。")
			return 0, nil
		}
	}
	graph.InspectOrderEventPayload = true
	eventTypes := strings.Fields("order.payment_succeeded order.payment_expired order.cancelled order.shipped order.delivered order.completed order.refunded after_sales.status_changed")
	events, err := graph.readBoundedRelationshipRows(ctx, "outbox_events", `record.event_type=ANY($1::text[]) AND NOT EXISTS(SELECT 1 FROM public.email_delivery_records receipt WHERE receipt.outbox_event_id=record.id)`, []any{pq.Array(eventTypes)}, graph.Limits.Rows)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, event := range events {
		if moduleRecoveryNotificationReceiptExpectedTemplateCode(event) == "" {
			continue
		}
		count++
		graph.Report.recordIssue("notification_audit_absence_not_delivery_evidence", "review", graph.Side, "outbox_events", event.referenceColumnText("id"), "通知事件没有审计行；审计写入为尽力执行，缺行不能证明未发送，也不授权重发。")
	}
	return count, nil
}

// Version 1 mirrors the current deterministic notification rule. All unknown
// types/statuses block receipt association instead of guessing a template.
func moduleRecoveryNotificationReceiptExpectedTemplateCode(event *moduleRecoveryDependencyRow) string {
	if event.referenceColumnText("event_type") == "after_sales.status_changed" {
		status := orderTransactionDurableEventPayloadString(event, "_payload_new_status")
		for _, registered := range strings.Fields("requested approved awaiting_return return_in_transit received resolving completed rejected") {
			if status == registered {
				return "after_sales_" + status
			}
		}
		return ""
	}
	return map[string]string{"order.payment_succeeded": "order_confirmation", "order.payment_expired": "order_payment_expired", "order.cancelled": "order_cancelled", "order.shipped": "order_shipping_notification", "order.delivered": "order_delivered", "order.completed": "order_completed", "order.refunded": "order_refunded"}[event.referenceColumnText("event_type")]
}
