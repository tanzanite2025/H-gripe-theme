package localbackup

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func createOrderModuleSettledActiveOrderRestoreContractTestSnapshot() selectedBusinessModuleCompleteSnapshot {
	rows := map[string]string{
		"orders":                     `{"id":1,"order_number":"original-order","user_id":3,"status":"paid","payment_status":"paid","payment_method":"original-provider","payment_currency":"USD","payment_amount_minor":123,"paid_at":"2026-10-27T01:00:00Z","fulfillment_hold":false,"dispute_previous_status":"","shipping_status":"pending"}`,
		"order_items":                `{"id":2,"order_id":1,"product_id":4,"variant_id":5,"quantity":1,"configuration_snapshot":{}}`,
		"transactions":               `{"id":6,"order_id":1,"transaction_id":"original-provider-payment","payment_method":"original-provider","amount_minor":123,"currency":"USD","status":"completed","completed_at":"2026-10-27T01:00:00Z"}`,
		"outbox_events":              `{"id":7,"event_key":"original-paid-event","event_type":"order.payment_succeeded","aggregate_type":"order","aggregate_id":"1","status":"processed","processed_at":"2026-10-27T01:00:00Z"}`,
		"order_evidence_items":       `{"id":8,"order_id":1}`,
		"order_evidence_attachments": `{"id":9,"evidence_item_id":8,"storage_key":"order-evidence/1/8/original.bin"}`,
	}
	snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"orders"}, Tables: map[string]businessModuleCompleteTableSnapshot{}}
	for table, row := range rows {
		snapshot.Tables[table] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{json.RawMessage(row)}}
	}
	return snapshot
}

func TestOrderModuleHistoryRestoreSettledActiveOrdersRequirePaymentInventoryAndShippingFacts(t *testing.T) {
	for _, scenario := range []string{"paid", "processing", "shipped", "delivered", "empty_component_allocations", "attachment_only", "unpaid_pending", "needs_review", "disputed", "unpaid_fact", "missing_paid_time", "fulfillment_hold", "dispute_history", "deleted_order", "completion_time", "paid_but_shipped", "delivery_without_status", "delivered_without_time", "shipping_without_time", "shipping_without_evidence", "wrong_shipment_number", "missing_transaction", "duplicate_transaction", "pending_transaction", "refunded_transaction", "missing_provider_identity", "missing_payment_completion_time", "wrong_payment_amount", "wrong_payment_currency", "wrong_payment_method", "active_refund", "active_dispute", "active_after_sales", "active_review", "missing_items", "invalid_quantity", "missing_variant", "component_allocations", "invalid_configuration", "pending_event", "terminal_event", "current_stock_dependency_changed", "current_order_status_changed", "new_current_execution"} {
		t.Run(scenario, func(t *testing.T) {
			snapshot := createOrderModuleSettledActiveOrderRestoreContractTestSnapshot()
			changeRecordedRowColumn := func(table, column string, value any) {
				var fields map[string]any
				require.NoError(t, json.Unmarshal(snapshot.Tables[table].Rows[0], &fields))
				fields[column] = value
				encoded, err := json.Marshal(fields)
				require.NoError(t, err)
				snapshot.Tables[table] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{encoded}}
			}
			valid := scenario == "paid" || scenario == "processing" || scenario == "shipped" || scenario == "delivered" || scenario == "empty_component_allocations" || scenario == "attachment_only"
			switch scenario {
			case "processing":
				changeRecordedRowColumn("orders", "status", "processing")
			case "shipped", "delivered", "delivery_without_status", "delivered_without_time", "shipping_without_time", "shipping_without_evidence", "wrong_shipment_number":
				changeRecordedRowColumn("orders", "status", "shipped")
				changeRecordedRowColumn("orders", "shipping_status", "shipped")
				changeRecordedRowColumn("orders", "shipped_at", "2026-10-27T02:00:00Z")
				snapshot.Tables["shipment_records"] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{json.RawMessage(`{"id":10,"order_id":1,"order_number":"original-order","shipped_at":"2026-10-27T02:00:00Z"}`)}}
				if scenario == "delivered" || scenario == "delivery_without_status" || scenario == "delivered_without_time" {
					changeRecordedRowColumn("orders", "shipping_status", "delivered")
					changeRecordedRowColumn("orders", "delivered_at", "2026-10-27T03:00:00Z")
				}
				if scenario == "delivery_without_status" {
					changeRecordedRowColumn("orders", "shipping_status", "shipped")
				}
				if scenario == "delivered_without_time" {
					changeRecordedRowColumn("orders", "delivered_at", nil)
				}
				if scenario == "shipping_without_time" {
					changeRecordedRowColumn("orders", "shipped_at", nil)
				}
				if scenario == "shipping_without_evidence" {
					delete(snapshot.Tables, "shipment_records")
				}
				if scenario == "wrong_shipment_number" {
					changeRecordedRowColumn("shipment_records", "order_number", "another-order")
				}
			case "empty_component_allocations":
				changeRecordedRowColumn("order_items", "configuration_snapshot", map[string]any{"inventory_allocations": []any{}})
			case "unpaid_pending":
				changeRecordedRowColumn("orders", "status", "pending")
				changeRecordedRowColumn("orders", "payment_status", "unpaid")
			case "needs_review", "disputed":
				changeRecordedRowColumn("orders", "status", scenario)
			case "unpaid_fact":
				changeRecordedRowColumn("orders", "payment_status", "unpaid")
			case "missing_paid_time":
				changeRecordedRowColumn("orders", "paid_at", nil)
			case "fulfillment_hold":
				changeRecordedRowColumn("orders", "fulfillment_hold", true)
			case "dispute_history":
				changeRecordedRowColumn("orders", "dispute_previous_status", "paid")
			case "deleted_order":
				changeRecordedRowColumn("orders", "deleted_at", "2026-10-27T02:00:00Z")
			case "completion_time":
				changeRecordedRowColumn("orders", "completed_at", "2026-10-27T02:00:00Z")
			case "paid_but_shipped":
				changeRecordedRowColumn("orders", "shipping_status", "shipped")
			case "missing_transaction":
				delete(snapshot.Tables, "transactions")
			case "duplicate_transaction":
				original := snapshot.Tables["transactions"].Rows[0]
				changeRecordedRowColumn("transactions", "id", 11)
				contents := snapshot.Tables["transactions"]
				contents.Rows = append(contents.Rows, original)
				snapshot.Tables["transactions"] = contents
			case "pending_transaction":
				changeRecordedRowColumn("transactions", "status", "pending")
			case "refunded_transaction":
				changeRecordedRowColumn("transactions", "status", "refunded")
			case "missing_provider_identity":
				changeRecordedRowColumn("transactions", "transaction_id", nil)
			case "missing_payment_completion_time":
				changeRecordedRowColumn("transactions", "completed_at", nil)
			case "wrong_payment_amount":
				changeRecordedRowColumn("transactions", "amount_minor", 124)
			case "wrong_payment_currency":
				changeRecordedRowColumn("transactions", "currency", "EUR")
			case "wrong_payment_method":
				changeRecordedRowColumn("transactions", "payment_method", "another-provider")
			case "active_refund", "active_dispute", "active_after_sales", "active_review":
				table := map[string]string{"active_refund": "refunds", "active_dispute": "paypal_disputes", "active_after_sales": "after_sales_cases", "active_review": "payment_reviews"}[scenario]
				snapshot.Tables[table] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{json.RawMessage(`{"id":11,"order_id":1}`)}}
			case "missing_items":
				delete(snapshot.Tables, "order_items")
			case "invalid_quantity":
				changeRecordedRowColumn("order_items", "quantity", 0)
			case "missing_variant":
				changeRecordedRowColumn("order_items", "variant_id", nil)
			case "component_allocations":
				changeRecordedRowColumn("order_items", "configuration_snapshot", map[string]any{"inventory_allocations": []any{map[string]any{"variant_id": 12, "quantity": 1}}})
			case "invalid_configuration":
				changeRecordedRowColumn("order_items", "configuration_snapshot", nil)
			case "pending_event":
				changeRecordedRowColumn("outbox_events", "status", "pending")
			case "terminal_event":
				changeRecordedRowColumn("outbox_events", "event_type", "order.completed")
			}
			missing := map[string][]json.RawMessage{}
			for table, contents := range snapshot.Tables {
				missing[table] = contents.Rows
			}
			if scenario == "attachment_only" {
				missing = nil
			}
			report := &SelectedBusinessModuleIsolatedRestorePreview{}
			if scenario == "current_stock_dependency_changed" {
				report.DependencyConflicts = 1
			}
			if scenario == "current_order_status_changed" {
				report.Tables = []SelectedBusinessModuleTableRestorePreview{{Table: "orders", ConflictingRows: 1}}
			}
			if scenario == "new_current_execution" {
				report.OrderHistoryCurrentExecutionConflicts = 1
			}
			result := inspectOrderModuleMissingOrderRecordsRestoreEligibility(snapshot, missing, report, []FileIdentity{{Path: "media/uploads_private/order-evidence/1/8/original.bin"}})
			require.Equal(t, valid, result.Eligible, "%v", result.Blockers)
			if valid {
				require.Equal(t, 1, result.SettledActiveOrderOwners)
			}
		})
	}
}
