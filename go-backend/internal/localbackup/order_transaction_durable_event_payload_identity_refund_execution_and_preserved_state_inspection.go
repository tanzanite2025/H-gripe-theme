package localbackup

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/outbox"
	"github.com/lib/pq"
)

// These fields identify relationships, never email bodies, addresses or complete
// provider responses. PostgreSQL hashes the complete payload without returning it.
func orderTransactionDurableEventPayloadProjectionEntries() []string {
	payload := `(to_jsonb(record)->'payload')`
	expressions := map[string]string{
		"_payload_sha256": `encode(sha256(convert_to((` + payload + `)::text,'UTF8')),'hex')`,
		"_payload_shape":  `jsonb_typeof(` + payload + `)`,
	}
	for _, field := range strings.Fields("order_id order_number refund_id transaction_id case_id shipment_id schema_version idempotency_key attempt provider user_id new_status") {
		expressions["_payload_"+field] = payload + `->` + pq.QuoteLiteral(field)
	}
	var names []string
	for name := range expressions {
		names = append(names, name)
	}
	sort.Strings(names)
	var entries []string
	for _, name := range names {
		entries = append(entries, pq.QuoteLiteral(name), expressions[name])
	}
	return entries
}

// The expected aggregate follows the actual producer. In particular payment
// aggregates identify refunds, not transaction IDs or order IDs.
func registeredOrderTransactionDurableEventAggregateTypes() map[string]string {
	registered := map[string]string{}
	for _, eventType := range []string{outbox.EventTypeOrderPaid, outbox.EventTypeOrderPaymentSucceeded,
		outbox.EventTypeOrderPaymentExpired, outbox.EventTypeOrderCancelled, outbox.EventTypeOrderShipped,
		outbox.EventTypeOrderDelivered, outbox.EventTypeOrderRefunded, outbox.EventTypeOrderCompleted,
		outbox.EventTypeOrderDisputeContactEmail, outbox.EventTypeTrackingShipmentRegistration,
		outbox.EventTypeReferralOrderPaid, outbox.EventTypeReferralOrderDelivered,
		outbox.EventTypeReferralOrderInvalidated, outbox.EventTypeVerifiedConversion} {
		registered[eventType] = outbox.AggregateTypeOrder
	}
	for _, eventType := range []string{outbox.EventTypePaymentRefundPending, outbox.EventTypePaymentRefundCompleted,
		outbox.EventTypePaymentRefundFailed, outbox.EventTypePaymentRefundExecutionRequested} {
		registered[eventType] = outbox.AggregateTypePayment
	}
	registered[outbox.EventTypeAfterSalesStatusChanged] = outbox.AggregateTypeAfterSalesCase
	return registered
}

func (cfg Configuration) inspectOrderTransactionDurableEventDependencies(ctx context.Context, graphs []*moduleRecoveryRowDependencyGraph, report *ModuleRecoveryRowInspection) error {
	inspection := &ModuleRecoveryDurableEventInspection{Version: 1, StatePolicy: "preserve_existing_event_state_no_replay",
		ComparisonSampleLimit: report.IssueSampleLimit, SourceStatuses: map[string]int{}, TargetStatuses: map[string]int{}, Comparisons: []ModuleRecoveryDurableEventComparison{}}
	report.DurableEvents = inspection
	registered := registeredOrderTransactionDurableEventAggregateTypes()
	var eventTypes []string
	for eventType := range registered {
		eventTypes = append(eventTypes, eventType)
	}
	sort.Strings(eventTypes)
	var snapshots []*moduleRecoveryDurableEventSnapshot
	for _, graph := range graphs {
		graph.InspectOrderEventPayload = true
		snapshot := &moduleRecoveryDurableEventSnapshot{Graph: graph, Events: map[string]*moduleRecoveryDependencyRow{}, Related: map[string]bool{}}
		snapshots = append(snapshots, snapshot)
		valid := true
		for _, column := range strings.Fields("id event_key event_type aggregate_type aggregate_id payload status attempts max_attempts available_at locked_at locked_by processed_at last_attempt_at uncertain_at reconcile_after deleted_at") {
			if !graph.Observation.Columns["outbox_events"][column] {
				valid = false
			}
		}
		if !valid {
			graph.markIncomplete("order_event_inspection_schema_missing", "outbox_events", "缺少订单事件身份、载荷、执行状态或调度字段，无法完成重放风险诊断。")
			continue
		}
		rows, err := graph.readBoundedRelationshipRows(ctx, "outbox_events", `record.event_type=ANY($1::text[]) OR record.aggregate_type=ANY($2::text[])`, []any{pq.Array(eventTypes), pq.Array([]string{outbox.AggregateTypeOrder, outbox.AggregateTypePayment, outbox.AggregateTypeAfterSalesCase})}, graph.Limits.Rows)
		if err != nil {
			return err
		}
		for _, event := range rows {
			snapshot.Events[event.Identity] = event
			snapshot.Related[event.Identity] = true
			if err := inspectOrderTransactionDurableEventPayloadAndExecutionState(ctx, graph, event, registered); err != nil {
				return err
			}
			status := moduleRecoveryDurableEventPublicStatus(event)
			if graph.Side == "package" {
				inspection.SourceRelatedEvents++
				inspection.SourceStatuses[status]++
			} else {
				inspection.TargetRelatedEvents++
				inspection.TargetStatuses[status]++
			}
		}
		if graph.Side == "package" {
			inspection.SourceScannedEvents = len(rows)
		} else {
			inspection.TargetScannedEvents = len(rows)
		}
	}
	if len(snapshots) == 2 {
		if err := compareModuleRecoveryDurableEventSnapshots(ctx, snapshots[0], snapshots[1], inspection); err != nil {
			return err
		}
	}
	inspection.Complete = report.Complete
	return nil
}

func inspectOrderTransactionDurableEventPayloadAndExecutionState(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, event *moduleRecoveryDependencyRow, registered map[string]string) error {
	status := moduleRecoveryDurableEventPublicStatus(event)
	if event.referenceColumnText("event_key") == "" {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_event_key_missing", "blocking", "事件缺少稳定幂等键，不能补入后自动执行。", false)
	}
	if status == "invalid" {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_event_status_unregistered", "blocking", "事件处理状态未登记，不能推断执行结果。", true)
	}
	if status == outbox.EventStatusUnknown || event.referenceColumnText("uncertain_at") != "" {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_event_outcome_unknown", "blocking", "事件存在未知外部结果标记，须核对支付、物流或通知结果，不能直接重试。", false)
	}
	if event.referenceColumnText("deleted_at") == "" && (status == outbox.EventStatusProcessing || event.referenceColumnText("locked_at") != "" || event.referenceColumnText("locked_by") != "") {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_event_execution_lease_present", "blocking", "事件仍有执行或对账租约，须保留围栏并核对执行器。", false)
	}
	eventType := event.referenceColumnText("event_type")
	expectedAggregate, exists := registered[eventType]
	if !exists {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_event_type_unregistered", "blocking", "订单、支付或售后聚合出现未登记事件类型，须补充恢复规则。", true)
		return nil
	}
	if event.referenceColumnText("_payload_shape") != "object" {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_event_payload_not_an_object", "blocking", "订单事件载荷不是对象，无法确认其引用及副作用范围。", true)
		return nil
	}
	aggregateColumn := "_payload_order_id"
	if expectedAggregate == outbox.AggregateTypePayment {
		aggregateColumn = "_payload_refund_id"
	}
	if expectedAggregate == outbox.AggregateTypeAfterSalesCase {
		aggregateColumn = "_payload_case_id"
	}
	if event.referenceColumnText("aggregate_type") != expectedAggregate || moduleRecoveryDurableEventPayloadPositiveID(event, aggregateColumn) == "" || event.referenceColumnText("aggregate_id") != moduleRecoveryDurableEventPayloadPositiveID(event, aggregateColumn) {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_event_aggregate_identity_conflict", "blocking", "事件聚合身份与登记类型或载荷业务编号不一致。", false)
	}
	canonical := eventType == outbox.EventTypeOrderPaymentSucceeded || eventType == outbox.EventTypeOrderPaymentExpired || eventType == outbox.EventTypeOrderCancelled || eventType == outbox.EventTypeOrderShipped || eventType == outbox.EventTypeOrderDelivered || eventType == outbox.EventTypeOrderRefunded || eventType == outbox.EventTypeOrderCompleted || eventType == outbox.EventTypeAfterSalesStatusChanged
	if canonical && (string(event.Values["_payload_schema_version"]) != "1" || orderTransactionDurableEventPayloadString(event, "_payload_idempotency_key") == "") {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_event_canonical_metadata_invalid", "blocking", "规范业务事件缺少已支持的版本或载荷幂等身份。", true)
	}
	var refund *moduleRecoveryDependencyRow
	orderID := moduleRecoveryDurableEventPayloadPositiveID(event, "_payload_order_id")
	if eventType == outbox.EventTypePaymentRefundExecutionRequested {
		var err error
		refund, err = inspectOrderEventReferencedBusinessRow(ctx, graph, event, "_payload_refund_id", "refunds", "")
		if err != nil {
			return err
		}
		if refund != nil {
			orderID = refund.referenceColumnText("order_id")
		}
	}
	if orderID == "" && !(eventType == outbox.EventTypePaymentRefundExecutionRequested && refund == nil) {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_event_order_identity_invalid", "blocking", "事件无法解析出数值型订单身份，不能按字符串编号或客户身份猜测。", false)
	} else if orderID != "" {
		orders, err := graph.readRowsMatchingReferenceKeys(ctx, "orders", []string{"id"}, []map[string]json.RawMessage{{"id": json.RawMessage(orderID)}}, 1)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			recordOrderEventMissingBusinessReference(graph, event)
		} else {
			if number := event.referenceColumnText("_payload_order_number"); number != "" && number != orders[0].referenceColumnText("order_number") {
				recordModuleRecoveryDurableEventIssue(graph, event, "order_event_order_number_identity_conflict", "blocking", "事件记录的订单号与当前订单身份不同。", false)
			}
			if userID := moduleRecoveryDurableEventPayloadPositiveID(event, "_payload_user_id"); userID != "" && userID != orders[0].referenceColumnText("user_id") {
				recordModuleRecoveryDurableEventIssue(graph, event, "order_event_customer_identity_conflict", "blocking", "事件客户编号与所引用订单不一致。", false)
			}
		}
	}
	for _, reference := range []struct{ column, table string }{{"refund_id", "refunds"}, {"transaction_id", "transactions"}, {"case_id", "after_sales_cases"}, {"shipment_id", "shipment_records"}} {
		required := reference.column == "refund_id" && expectedAggregate == outbox.AggregateTypePayment || reference.column == "transaction_id" && expectedAggregate == outbox.AggregateTypePayment && eventType != outbox.EventTypePaymentRefundExecutionRequested || reference.column == "case_id" && expectedAggregate == outbox.AggregateTypeAfterSalesCase || reference.column == "shipment_id" && eventType == outbox.EventTypeOrderShipped || (reference.column == "refund_id" || reference.column == "transaction_id") && eventType == outbox.EventTypeOrderRefunded
		column := "_payload_" + reference.column
		if !required && (len(event.Values[column]) == 0 || string(event.Values[column]) == "null" || string(event.Values[column]) == "0") {
			continue
		}
		row, err := inspectOrderEventReferencedBusinessRow(ctx, graph, event, column, reference.table, orderID)
		if err != nil {
			return err
		}
		if reference.column == "refund_id" && row != nil {
			refund = row
		}
		if reference.column == "transaction_id" && row != nil && refund != nil && row.referenceColumnText("id") != refund.referenceColumnText("transaction_id") {
			recordModuleRecoveryDurableEventIssue(graph, event, "order_event_refund_transaction_identity_conflict", "blocking", "退款事件的交易编号与退款登记不一致。", false)
		}
	}
	if eventType == outbox.EventTypePaymentRefundExecutionRequested && refund != nil {
		if err := inspectOrderRefundExecutionCommandAttemptAndIdempotency(ctx, graph, event, refund); err != nil {
			return err
		}
	}
	return nil
}

func recordOrderEventMissingBusinessReference(graph *moduleRecoveryRowDependencyGraph, event *moduleRecoveryDependencyRow) {
	severity := "blocking"
	if event.referenceColumnText("deleted_at") != "" || moduleRecoveryDurableEventPublicStatus(event) == outbox.EventStatusProcessed || moduleRecoveryDurableEventPublicStatus(event) == outbox.EventStatusIgnored {
		severity = "review"
	}
	recordModuleRecoveryDurableEventIssue(graph, event, "order_event_business_reference_unavailable", severity, "事件引用的业务行不存在；历史已处理记录须保留状态，不作为重新建立交易或重放的依据。", false)
}

func inspectOrderEventReferencedBusinessRow(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, event *moduleRecoveryDependencyRow, column, table, orderID string) (*moduleRecoveryDependencyRow, error) {
	if moduleRecoveryDurableEventPayloadPositiveID(event, column) == "" {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_event_business_identity_invalid", "blocking", "事件业务引用不是有效数值编号。", false)
		return nil, nil
	}
	for _, requiredColumn := range []string{"id", "order_id"} {
		if !graph.Observation.Columns[table][requiredColumn] {
			graph.markIncomplete("order_event_reference_schema_missing", table, "缺少事件所引用业务表的身份或归属字段。")
			return nil, nil
		}
	}
	rows, err := graph.readRowsMatchingReferenceKeys(ctx, table, []string{"id"}, []map[string]json.RawMessage{{"id": event.Values[column]}}, 1)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		recordOrderEventMissingBusinessReference(graph, event)
		return nil, nil
	}
	if orderID != "" && rows[0].referenceColumnText("order_id") != orderID {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_event_reference_crosses_orders", "blocking", "事件退款、交易、售后或发货引用属于另一笔订单。", false)
	}
	return rows[0], nil
}

func inspectOrderRefundExecutionCommandAttemptAndIdempotency(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, event, refund *moduleRecoveryDependencyRow) error {
	for _, column := range strings.Fields("id refund_id order_id transaction_id status provider idempotency_key attempt_count") {
		if !graph.Observation.Columns["payment_refund_executions"][column] {
			graph.markIncomplete("order_refund_execution_schema_missing", "payment_refund_executions", "缺少退款执行批次、提供方或幂等身份字段。")
			return nil
		}
	}
	rows, err := graph.readRowsMatchingReferenceKeys(ctx, "payment_refund_executions", []string{"refund_id"}, []map[string]json.RawMessage{{"refund_id": refund.Values["id"]}}, 2)
	if err != nil {
		return err
	}
	if len(rows) != 1 {
		recordOrderEventMissingBusinessReference(graph, event)
		return nil
	}
	execution := rows[0]
	switch execution.referenceColumnText("status") {
	case "processing", "succeeded", "failed":
	default:
		recordModuleRecoveryDurableEventIssue(graph, event, "order_refund_execution_status_unregistered", "blocking", "退款执行状态未登记，不能把其外部结果解释为失败或允许重试。", true)
	}
	attempt, err := strconv.Atoi(moduleRecoveryDurableEventPayloadPositiveID(event, "_payload_attempt"))
	currentAttempt, _ := strconv.Atoi(execution.referenceColumnText("attempt_count"))
	if err != nil || attempt <= 0 || currentAttempt <= 0 {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_refund_command_attempt_invalid", "blocking", "退款执行命令或当前记录缺少有效批次，不能推断其是否已被后续批次替代。", false)
	} else if attempt != currentAttempt {
		severity := "blocking"
		if attempt < currentAttempt {
			severity = "review"
		}
		recordModuleRecoveryDurableEventIssue(graph, event, "order_refund_command_attempt_differs", severity, "退款执行批次不同；较旧命令依赖当前退款记录保持无效，禁止把当前批次回退。", false)
	}
	if attempt == currentAttempt && (orderTransactionDurableEventPayloadString(event, "_payload_idempotency_key") == "" || orderTransactionDurableEventPayloadString(event, "_payload_idempotency_key") != execution.referenceColumnText("idempotency_key") || !strings.EqualFold(strings.TrimSpace(orderTransactionDurableEventPayloadString(event, "_payload_provider")), strings.TrimSpace(execution.referenceColumnText("provider")))) {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_refund_command_idempotency_conflict", "blocking", "同批次退款命令与执行记录的提供方或幂等身份不同。", false)
	}
	if execution.referenceColumnText("order_id") != refund.referenceColumnText("order_id") || execution.referenceColumnText("transaction_id") != refund.referenceColumnText("transaction_id") {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_refund_execution_business_identity_conflict", "blocking", "退款执行记录与退款的订单或交易身份不同。", false)
	}
	if execution.referenceColumnText("status") == "succeeded" || refund.referenceColumnText("status") == "completed" {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_refund_completed_execution_preserved", "review", "退款已经完成；命令保持无副作用依赖当前退款及执行记录，不能恢复为待退款。", false)
	} else if execution.referenceColumnText("status") == "processing" && refund.referenceColumnText("status") == "pending" && event.referenceColumnText("deleted_at") == "" && moduleRecoveryDurableEventPublicStatus(event) != outbox.EventStatusIgnored {
		recordModuleRecoveryDurableEventIssue(graph, event, "order_refund_external_execution_requires_reconciliation", "blocking", "退款记录仍允许外部执行，须核对平台结果并保留执行围栏，不能自动重放。", false)
	}
	return nil
}

func orderTransactionDurableEventPayloadString(event *moduleRecoveryDependencyRow, column string) string {
	var value string
	if json.Unmarshal(event.Values[column], &value) != nil {
		return ""
	}
	return value
}
