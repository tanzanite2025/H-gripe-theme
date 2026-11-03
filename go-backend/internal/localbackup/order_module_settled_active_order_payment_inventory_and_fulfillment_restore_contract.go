package localbackup

import (
	"encoding/json"
	"strconv"
	"strings"
)

func orderModuleOrderStatusHasTerminalHistoryRestoreContract(status string) bool {
	return status == "completed" || status == "cancelled" || status == "refunded"
}

// Paid orders cannot enter the pending/unpaid expiration and reservation-release
// path. Restore only recorded settled facts; neither payment nor fulfillment
// services are called by the importer. Current dependencies must still match.
func inspectOrderModuleTerminalOrSettledActiveOwnerRestoreBlocker(index map[string]map[string]json.RawMessage, orderID string) string {
	orderRow := index["orders"][orderID]
	readOrderRecoveryRecordColumnText := func(row json.RawMessage, column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	status := readOrderRecoveryRecordColumnText(orderRow, "status")
	if orderModuleOrderStatusHasTerminalHistoryRestoreContract(status) {
		return ""
	}
	if status == "pending" {
		return "pending_order_inventory_reservation_requires_durable_release_reconciliation"
	}
	if status != "paid" && status != "processing" && status != "shipped" {
		return "order_state_requires_separate_payment_or_dispute_reconciliation"
	}
	if readOrderRecoveryRecordColumnText(orderRow, "payment_status") != "paid" || !orderModuleHistoricalRowHasRecordedTimestamp(orderRow, "paid_at") || readOrderRecoveryRecordColumnText(orderRow, "fulfillment_hold") != "false" || readOrderRecoveryRecordColumnText(orderRow, "dispute_previous_status") != "" {
		return "active_order_requires_recorded_payment_without_fulfillment_hold"
	}
	for _, column := range strings.Fields("completed_at cancelled_at deleted_at") {
		if readOrderRecoveryRecordColumnText(orderRow, column) != "" {
			return "active_order_lifecycle_facts_are_inconsistent"
		}
	}
	shippingStatus := readOrderRecoveryRecordColumnText(orderRow, "shipping_status")
	if status == "shipped" {
		if (shippingStatus != "shipped" && shippingStatus != "delivered") || !orderModuleHistoricalRowHasRecordedTimestamp(orderRow, "shipped_at") || (shippingStatus == "delivered" && !orderModuleHistoricalRowHasRecordedTimestamp(orderRow, "delivered_at")) || (shippingStatus == "shipped" && readOrderRecoveryRecordColumnText(orderRow, "delivered_at") != "") {
			return "active_order_requires_matching_shipping_facts"
		}
	} else if (shippingStatus != "pending" && shippingStatus != "processing") || readOrderRecoveryRecordColumnText(orderRow, "shipped_at") != "" || readOrderRecoveryRecordColumnText(orderRow, "delivered_at") != "" {
		return "active_order_requires_matching_shipping_facts"
	}
	amount, amountError := strconv.ParseInt(readOrderRecoveryRecordColumnText(orderRow, "payment_amount_minor"), 10, 64)
	if amountError != nil || amount <= 0 || readOrderRecoveryRecordColumnText(orderRow, "payment_currency") == "" {
		return "active_order_requires_one_matching_settled_transaction"
	}
	transactions := 0
	for _, transaction := range index["transactions"] {
		if readOrderRecoveryRecordColumnText(transaction, "order_id") != orderID {
			continue
		}
		transactions++
		_, blocker := inspectOrderModuleCompletedHistoricalFactOwnerAndBlocker(index, "transactions", transaction)
		if blocker != "" || readOrderRecoveryRecordColumnText(transaction, "status") != "completed" || readOrderRecoveryRecordColumnText(transaction, "deleted_at") != "" || readOrderRecoveryRecordColumnText(transaction, "amount_minor") != readOrderRecoveryRecordColumnText(orderRow, "payment_amount_minor") || readOrderRecoveryRecordColumnText(transaction, "currency") != readOrderRecoveryRecordColumnText(orderRow, "payment_currency") || readOrderRecoveryRecordColumnText(transaction, "payment_method") == "" || readOrderRecoveryRecordColumnText(transaction, "payment_method") != readOrderRecoveryRecordColumnText(orderRow, "payment_method") {
			return "active_order_requires_one_matching_settled_transaction"
		}
	}
	if transactions != 1 {
		return "active_order_requires_one_matching_settled_transaction"
	}
	// Refunds, disputes and after-sales can change external liabilities while an
	// order remains open. They need their own active-workflow recovery contract.
	for _, table := range strings.Fields("refunds payment_refund_executions after_sales_cases paypal_disputes stripe_disputes payment_reviews") {
		for _, row := range index[table] {
			if readOrderRecoveryRecordColumnText(row, "order_id") == orderID {
				return "active_order_refund_dispute_or_review_workflow_requires_reconciliation"
			}
		}
	}
	items := 0
	for _, item := range index["order_items"] {
		if readOrderRecoveryRecordColumnText(item, "order_id") != orderID {
			continue
		}
		items++
		quantity, err := strconv.ParseInt(readOrderRecoveryRecordColumnText(item, "quantity"), 10, 64)
		if err != nil || quantity <= 0 || !canonicalModuleRecoveryDurableEventPositiveID(readOrderRecoveryRecordColumnText(item, "product_id")) || !canonicalModuleRecoveryDurableEventPositiveID(readOrderRecoveryRecordColumnText(item, "variant_id")) {
			return "active_order_requires_complete_original_item_identities"
		}
		if _, blocker := readOrderModuleConfiguredComponentInventoryAllocationsForRestore(item); blocker != "" {
			return blocker
		}
	}
	if items == 0 {
		return "active_order_requires_complete_original_item_identities"
	}
	shippingEvidence := false
	for _, shipment := range index["shipment_records"] {
		if readOrderRecoveryRecordColumnText(shipment, "order_id") == orderID {
			_, blocker := inspectOrderModuleCompletedHistoricalFactOwnerAndBlocker(index, "shipment_records", shipment)
			if status != "shipped" || blocker != "" {
				return "active_order_requires_matching_shipping_facts"
			}
			shippingEvidence = true
		}
	}
	if status == "shipped" && !shippingEvidence {
		return "active_order_requires_matching_shipping_facts"
	}
	for _, event := range index["outbox_events"] {
		if readOrderRecoveryRecordColumnText(event, "aggregate_type") != "order" || readOrderRecoveryRecordColumnText(event, "aggregate_id") != orderID {
			continue
		}
		switch readOrderRecoveryRecordColumnText(event, "event_type") {
		case "order.completed", "order.cancelled", "order.payment_expired", "order.refunded":
			return "active_order_lifecycle_facts_are_inconsistent"
		case "order.shipped", "order.delivered":
			if status != "shipped" {
				return "active_order_lifecycle_facts_are_inconsistent"
			}
		}
	}
	return ""
}
