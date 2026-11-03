package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// The policy version is part of the bound preview. A preview from the earlier
// detail-only or completed-facts implementation cannot authorize new receipts.
const orderModuleBoundedOrderRestorePolicy = "restore_terminal_and_settled_active_orders_with_normal_constraints_and_original_inventory_dependencies_v6"

// Payment operation receipts have no order_id. Restore their original global
// retry identity without inventing an order from the response body. All order
// rows and external user dependencies still undergo the ordinary comparisons.
func inspectOrderModuleCompletedGlobalPaymentReceiptRestoreBlocker(row json.RawMessage) string {
	readPaymentReceiptColumnText := func(column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	if readPaymentReceiptColumnText("status") != "completed" || readPaymentReceiptColumnText("claim_token") != "" || readPaymentReceiptColumnText("lease_expires_at") != "" {
		return "global_payment_receipt_requires_completed_state_without_execution_lease"
	}
	switch readPaymentReceiptColumnText("scope") {
	case "paypal_capture", "alipay_confirm", "wechat_confirm":
	default:
		return "global_payment_receipt_scope_not_registered"
	}
	key, digest := readPaymentReceiptColumnText("idempotency_key"), readPaymentReceiptColumnText("request_hash")
	if !canonicalModuleRecoveryDurableEventPositiveID(readPaymentReceiptColumnText("id")) || !canonicalModuleRecoveryDurableEventPositiveID(readPaymentReceiptColumnText("user_id")) || key == "" || key != strings.TrimSpace(key) || len(key) > 255 || !moduleRecoverySideEffectReceiptRequestDigestIsValid(digest) || digest != strings.ToLower(digest) {
		return "global_payment_receipt_request_identity_invalid"
	}
	created, createdError := time.Parse(time.RFC3339Nano, readPaymentReceiptColumnText("created_at"))
	updated, updatedError := time.Parse(time.RFC3339Nano, readPaymentReceiptColumnText("updated_at"))
	if createdError != nil || updatedError != nil || created.IsZero() || updated.IsZero() || updated.Before(created) {
		return "global_payment_receipt_completion_timestamps_invalid"
	}
	if reconciliation := readPaymentReceiptColumnText("reconciliation_started_at"); reconciliation != "" {
		started, err := time.Parse(time.RFC3339Nano, reconciliation)
		if err != nil || started.Before(created) || started.After(updated) {
			return "global_payment_receipt_completion_timestamps_invalid"
		}
	}
	status, err := strconv.Atoi(readPaymentReceiptColumnText("status_code"))
	cacheable := status >= http.StatusOK && status < http.StatusMultipleChoices
	switch status {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity:
		cacheable = true
	}
	contentType := readPaymentReceiptColumnText("content_type")
	mediaType, parameters, mediaTypeError := mime.ParseMediaType(contentType)
	body := readPaymentReceiptColumnText("response_body")
	var response map[string]json.RawMessage
	if err != nil || !cacheable || mediaTypeError != nil || strings.ContainsAny(contentType, "\r\n") || mediaType != "application/json" || len(parameters) > 1 || (len(parameters) == 1 && !strings.EqualFold(parameters["charset"], "utf-8")) || len(body) == 0 || len(body) > 256<<10 || json.Unmarshal([]byte(body), &response) != nil || response == nil {
		return "global_payment_receipt_recorded_response_invalid"
	}
	return ""
}

func orderModuleMissingGlobalPaymentReceiptConflictsWithCurrentNaturalIdentity(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, table string, row json.RawMessage) (bool, error) {
	if !reflect.DeepEqual(snapshot.Modules, []string{"orders"}) || table != "payment_operation_idempotencies" {
		return false, nil
	}
	var conflict bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.payment_operation_idempotencies record CROSS JOIN jsonb_populate_record(NULL::public.payment_operation_idempotencies,$1::jsonb) package_record WHERE record.user_id=package_record.user_id AND record.scope=package_record.scope AND record.idempotency_key=package_record.idempotency_key AND record.id<>package_record.id)`, string(row)).Scan(&conflict)
	return conflict, err
}

func orderModuleHasDedicatedCompletedHistoryInsertContract(snapshot selectedBusinessModuleCompleteSnapshot, table string) bool {
	return reflect.DeepEqual(snapshot.Modules, []string{"orders"}) && (table == "outbox_events" || table == "email_delivery_records")
}

func indexOrderModuleHistoricalRowsByTableAndID(snapshot selectedBusinessModuleCompleteSnapshot) map[string]map[string]json.RawMessage {
	index := map[string]map[string]json.RawMessage{}
	for table, contents := range snapshot.Tables {
		index[table] = map[string]json.RawMessage{}
		for _, row := range contents.Rows {
			id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
			index[table][id] = row
		}
	}
	return index
}

func orderModuleHistoricalRowHasRecordedTimestamp(row json.RawMessage, column string) bool {
	value, err := readBusinessModuleSnapshotRowColumnText(row, column)
	if err != nil {
		return false
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return err == nil && !parsed.IsZero()
}

// Values are inspected only inside the maintenance service. Reasons exposed in
// the preview contain registered table names, never payloads or provider IDs.
func inspectOrderModuleCompletedHistoricalFactOwnerAndBlocker(index map[string]map[string]json.RawMessage, table string, row json.RawMessage) (string, string) {
	text := func(record json.RawMessage, column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(record, column)
		return value
	}
	orderID := text(row, "order_id")
	completedTransaction := func(transaction json.RawMessage) bool {
		status := text(transaction, "status")
		return len(transaction) > 0 && (status == "completed" || status == "refunded") && text(transaction, "transaction_id") != "" && orderModuleHistoricalRowHasRecordedTimestamp(transaction, "completed_at")
	}
	completedRefund := func(refund json.RawMessage) bool {
		transaction := index["transactions"][text(refund, "transaction_id")]
		return len(refund) > 0 && text(refund, "status") == "completed" && text(refund, "refund_id") != "" && orderModuleHistoricalRowHasRecordedTimestamp(refund, "completed_at") && completedTransaction(transaction) && text(transaction, "order_id") == text(refund, "order_id") && text(transaction, "currency") == text(refund, "currency")
	}
	switch table {
	case "transactions":
		if !completedTransaction(row) {
			return orderID, "transaction_requires_recorded_completion_and_provider_identity"
		}
	case "refunds":
		if !completedRefund(row) {
			return orderID, "refund_requires_recorded_completion_and_matching_transaction"
		}
	case "payment_refund_executions":
		refund := index["refunds"][text(row, "refund_id")]
		transaction := index["transactions"][text(refund, "transaction_id")]
		order := index["orders"][orderID]
		completeReceipt := completedRefund(refund) && text(row, "status") == "succeeded" && orderModuleHistoricalRowHasRecordedTimestamp(row, "completed_at") && text(row, "idempotency_key") != "" && text(row, "provider_refund_id") != ""
		refundFactsMatch := text(row, "provider_refund_id") == text(refund, "refund_id") && text(row, "transaction_id") == text(refund, "transaction_id") && orderID == text(refund, "order_id") && text(row, "currency") == text(refund, "currency") && text(row, "amount_minor") == text(refund, "amount_minor")
		providerIdentityMatches := text(row, "provider_payment_id") == strings.TrimSpace(text(transaction, "transaction_id")) && text(row, "provider_transaction_id") == strings.TrimSpace(text(transaction, "transaction_id")) && text(row, "merchant_order_number") == strings.TrimSpace(text(order, "order_number"))
		if !completeReceipt || !refundFactsMatch || !providerIdentityMatches {
			return orderID, "refund_execution_requires_matching_successful_receipt"
		}
	case "refund_line_items":
		refund := index["refunds"][text(row, "refund_id")]
		item := index["order_items"][text(row, "order_item_id")]
		if !completedRefund(refund) || orderID != text(refund, "order_id") || len(item) == 0 || orderID != text(item, "order_id") {
			return orderID, "refund_line_requires_matching_completed_refund_and_order_item"
		}
	case "order_idempotencies":
		order := index["orders"][orderID]
		if len(order) == 0 || text(row, "user_id") != text(order, "user_id") || text(row, "scope") == "" || text(row, "idempotency_key") == "" || !moduleRecoverySideEffectReceiptRequestDigestIsValid(text(row, "request_hash")) {
			return orderID, "order_request_requires_bound_identity_and_digest"
		}
	case "refund_idempotencies":
		refund := index["refunds"][text(row, "refund_id")]
		orderID = text(refund, "order_id")
		if !completedRefund(refund) || text(row, "scope") == "" || text(row, "idempotency_key") == "" || !canonicalModuleRecoveryDurableEventPositiveID(text(row, "admin_user_id")) || !moduleRecoverySideEffectReceiptRequestDigestIsValid(text(row, "request_hash")) {
			return orderID, "refund_request_requires_bound_completed_refund_and_digest"
		}
	case "shipment_records":
		if !orderModuleHistoricalRowHasRecordedTimestamp(row, "shipped_at") || text(row, "order_number") != text(index["orders"][orderID], "order_number") {
			return orderID, "shipment_evidence_requires_recorded_shipping_and_order_identity"
		}
	case "outbox_events":
		if text(row, "status") != "processed" || !orderModuleHistoricalRowHasRecordedTimestamp(row, "processed_at") || text(row, "event_key") == "" {
			return "", "historical_event_requires_recorded_processed_state"
		}
		for _, column := range strings.Fields("locked_at locked_by uncertain_at reconcile_after") {
			if text(row, column) != "" {
				return "", "historical_event_has_execution_or_reconciliation_state"
			}
		}
		aggregate := text(row, "aggregate_type")
		if registeredOrderTransactionDurableEventAggregateTypes()[text(row, "event_type")] != aggregate || aggregate == "" {
			return "", "historical_event_requires_registered_order_ownership"
		}
		switch aggregate {
		case "order":
			orderID = text(row, "aggregate_id")
		case "payment":
			refund := index["refunds"][text(row, "aggregate_id")]
			orderID = text(refund, "order_id")
			if !completedRefund(refund) {
				return orderID, "historical_payment_event_requires_completed_refund"
			}
		default:
			return "", "historical_event_owner_not_supported"
		}
	case "email_delivery_records":
		event := index["outbox_events"][text(row, "outbox_event_id")]
		var reason string
		orderID, reason = inspectOrderModuleCompletedHistoricalFactOwnerAndBlocker(index, "outbox_events", event)
		if reason != "" || text(row, "status") != "sent" || !orderModuleHistoricalRowHasRecordedTimestamp(row, "sent_at") {
			return orderID, "notification_requires_sent_receipt_and_processed_event"
		}
	default:
		return orderID, "missing_execution_shared_or_unsupported_records"
	}
	if !canonicalModuleRecoveryDurableEventPositiveID(orderID) {
		return orderID, "historical_fact_has_no_recorded_order_owner"
	}
	return orderID, ""
}

// Only newer notes can differ. Every other field, including immutable price,
// address, identity, status, deletion and execution timestamps, must match.
// No UPDATE is issued and the complete current row is preservation-checked.
func canPreserveCurrentRestorableOrderNotesWithOtherwiseIdenticalFacts(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, table string, row json.RawMessage) (bool, error) {
	if !reflect.DeepEqual(snapshot.Modules, []string{"orders"}) || table != "orders" {
		return false, nil
	}
	orderID, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
	if inspectOrderModuleTerminalOrSettledActiveOwnerRestoreBlocker(indexOrderModuleHistoricalRowsByTableAndID(snapshot), orderID) != "" {
		return false, nil
	}
	var preserve bool
	err := tx.QueryRowContext(ctx, `SELECT
		(to_jsonb(record)-'customer_note'-'admin_note'-'updated_at')=($1::jsonb-'customer_note'-'admin_note'-'updated_at')
		AND (record.customer_note IS DISTINCT FROM package_record.customer_note OR record.admin_note IS DISTINCT FROM package_record.admin_note)
		AND record.updated_at IS NOT NULL AND package_record.updated_at IS NOT NULL AND record.updated_at>package_record.updated_at
		FROM public.orders record CROSS JOIN jsonb_populate_record(NULL::public.orders,$1::jsonb) package_record WHERE record.id=package_record.id`, string(row)).Scan(&preserve)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return preserve, err
}
