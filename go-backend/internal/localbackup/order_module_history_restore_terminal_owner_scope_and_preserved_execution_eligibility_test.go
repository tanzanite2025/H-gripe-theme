package localbackup

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderModuleHistoryRestoreGlobalPaymentReceiptRequiresDurableCompletedRetryResponse(t *testing.T) {
	for _, scenario := range []string{"paypal_capture", "alipay_confirm", "wechat_confirm", "cached_client_error", "completed_after_reconciliation", "pending", "reconciling", "claim_present", "lease_present", "unknown_scope", "invalid_user", "empty_key", "untrimmed_key", "oversized_key", "invalid_digest", "uppercase_digest", "missing_created_time", "updated_before_created", "invalid_reconciliation_time", "future_reconciliation_time", "uncached_server_error", "redirect_response", "invalid_status", "non_json_content_type", "unsafe_content_type", "trailing_content_type_newline", "unknown_content_parameter", "invalid_json", "null_json", "array_json", "empty_body", "oversized_body"} {
		t.Run(scenario, func(t *testing.T) {
			fields := map[string]any{"id": 1, "user_id": 3, "scope": "paypal_capture", "idempotency_key": "original-payment-request", "request_hash": strings.Repeat("a", 64), "status": "completed", "claim_token": "", "lease_expires_at": nil, "reconciliation_started_at": nil, "status_code": 200, "content_type": "application/json; charset=utf-8", "response_body": `{"recorded":"response"}`, "created_at": "2026-10-27T01:00:00Z", "updated_at": "2026-10-27T02:00:00Z"}
			valid := false
			switch scenario {
			case "paypal_capture", "alipay_confirm", "wechat_confirm":
				fields["scope"], valid = scenario, true
			case "cached_client_error":
				fields["status_code"], valid = 422, true
			case "completed_after_reconciliation":
				fields["reconciliation_started_at"], valid = "2026-10-27T01:30:00Z", true
			case "pending", "reconciling":
				fields["status"] = scenario
			case "claim_present":
				fields["claim_token"] = "private-claim-token"
			case "lease_present":
				fields["lease_expires_at"] = "2026-10-27T01:30:00Z"
			case "unknown_scope":
				fields["scope"] = "new_provider_mutation"
			case "invalid_user":
				fields["user_id"] = 0
			case "empty_key":
				fields["idempotency_key"] = ""
			case "untrimmed_key":
				fields["idempotency_key"] = " original-payment-request "
			case "oversized_key":
				fields["idempotency_key"] = strings.Repeat("a", 256)
			case "invalid_digest":
				fields["request_hash"] = "invalid"
			case "uppercase_digest":
				fields["request_hash"] = strings.Repeat("A", 64)
			case "missing_created_time":
				fields["created_at"] = nil
			case "updated_before_created":
				fields["updated_at"] = "2026-10-26T01:00:00Z"
			case "invalid_reconciliation_time":
				fields["reconciliation_started_at"] = "invalid"
			case "future_reconciliation_time":
				fields["reconciliation_started_at"] = "2026-10-28T01:00:00Z"
			case "uncached_server_error":
				fields["status_code"] = 503
			case "redirect_response":
				fields["status_code"] = 302
			case "invalid_status":
				fields["status_code"] = 0
			case "non_json_content_type":
				fields["content_type"] = "text/html"
			case "unsafe_content_type":
				fields["content_type"] = "application/json\r\nInjected: secret"
			case "trailing_content_type_newline":
				fields["content_type"] = "application/json\r\n"
			case "unknown_content_parameter":
				fields["content_type"] = "application/json; unknown=utf-8"
			case "invalid_json":
				fields["response_body"] = "private truncated response"
			case "null_json":
				fields["response_body"] = "null"
			case "array_json":
				fields["response_body"] = "[]"
			case "empty_body":
				fields["response_body"] = ""
			case "oversized_body":
				fields["response_body"] = `{"value":"` + strings.Repeat("a", 256<<10) + `"}`
			}
			row, err := json.Marshal(fields)
			require.NoError(t, err)
			snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"orders"}, Tables: map[string]businessModuleCompleteTableSnapshot{}}
			result := inspectOrderModuleMissingOrderRecordsRestoreEligibility(snapshot, map[string][]json.RawMessage{"payment_operation_idempotencies": {row}}, &SelectedBusinessModuleIsolatedRestorePreview{}, nil)
			require.Equal(t, valid, result.Eligible, "%v", result.Blockers)
			for _, blocker := range result.Blockers {
				require.NotContains(t, blocker, "private-")
				require.NotContains(t, blocker, "original-payment-request")
			}
		})
	}
}

func TestOrderModuleHistoryRestoreEligibilityRequiresTerminalOwnersAndPreservesExecutionFacts(t *testing.T) {
	for _, scenario := range []string{"missing_terminal_history", "terminal_evidence_file_only", "nonterminal_order", "nonterminal_evidence_owner", "unregistered_missing_file", "missing_payment_execution", "missing_transaction", "pending_durable_event", "conflicting_current_order", "missing_shared_dependency", "conflicting_file", "mixed_module_package"} {
		t.Run(scenario, func(t *testing.T) {
			snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"orders"}, Tables: map[string]businessModuleCompleteTableSnapshot{
				"orders":                     {Rows: []json.RawMessage{json.RawMessage(`{"id":1,"status":"completed"}`)}},
				"order_evidence_items":       {Rows: []json.RawMessage{json.RawMessage(`{"id":7,"order_id":1}`)}},
				"order_evidence_attachments": {Rows: []json.RawMessage{json.RawMessage(`{"id":8,"evidence_item_id":7,"storage_key":"order-evidence/1/7/proof.bin"}`)}},
				"outbox_events":              {Rows: []json.RawMessage{json.RawMessage(`{"id":9,"status":"processed"}`)}},
			}}
			missing := map[string][]json.RawMessage{"orders": snapshot.Tables["orders"].Rows}
			report := &SelectedBusinessModuleIsolatedRestorePreview{}
			files := []FileIdentity{{Path: "media/uploads_private/order-evidence/1/7/proof.bin"}}
			switch scenario {
			case "terminal_evidence_file_only":
				missing = nil
			case "nonterminal_order", "nonterminal_evidence_owner":
				snapshot.Tables["orders"] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{json.RawMessage(`{"id":1,"status":"processing"}`)}}
				if scenario == "nonterminal_evidence_owner" {
					missing = nil
				}
			case "unregistered_missing_file":
				files[0].Path = "media/uploads_private/after-sales/1/proof.bin"
			case "missing_payment_execution":
				missing["payment_refund_executions"] = []json.RawMessage{json.RawMessage(`{"id":1,"order_id":1}`)}
			case "missing_transaction":
				missing["transactions"] = []json.RawMessage{json.RawMessage(`{"id":1,"order_id":1}`)}
			case "pending_durable_event":
				snapshot.Tables["outbox_events"] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{json.RawMessage(`{"id":9,"status":"pending"}`)}}
			case "conflicting_current_order":
				report.Tables = []SelectedBusinessModuleTableRestorePreview{{Table: "orders", ConflictingRows: 1}}
			case "missing_shared_dependency":
				report.DependencyConflicts = 1
			case "conflicting_file":
				report.FileConflicts = 1
			case "mixed_module_package":
				snapshot.Modules = []string{"customer_service", "orders"}
			}
			result := inspectOrderModuleMissingOrderRecordsRestoreEligibility(snapshot, missing, report, files)
			require.Equal(t, scenario == "missing_terminal_history" || scenario == "terminal_evidence_file_only", result.Eligible, "%v", result.Blockers)
			if !result.Eligible {
				require.NotEmpty(t, result.Blockers)
			}
		})
	}
}

func TestOrderModuleHistoryRestoreCompletedFactsRequireRecordedReceiptsAndTerminalOwnership(t *testing.T) {
	for _, scenario := range []string{"completed_facts", "transaction_in_progress", "missing_provider_payment_id", "refund_in_progress", "refund_wrong_order", "refund_wrong_currency", "refund_execution_in_progress", "refund_execution_wrong_provider_refund", "refund_execution_wrong_amount", "refund_execution_wrong_provider_payment", "refund_execution_wrong_merchant_order", "unbound_order_request", "invalid_refund_request_digest", "pending_event", "unknown_event_outcome", "event_execution_lease", "unregistered_event", "unknown_notification", "missing_sent_time", "unsupported_global_payment_receipt", "unsupported_tracking_job", "nonterminal_owner"} {
		t.Run(scenario, func(t *testing.T) {
			encoded := map[string]string{
				"orders":                    `{"id":1,"order_number":"order-one","user_id":3,"status":"completed"}`,
				"transactions":              `{"id":4,"order_id":1,"status":"completed","currency":"USD","transaction_id":"provider-payment","completed_at":"2026-10-27T01:00:00Z"}`,
				"refunds":                   `{"id":5,"order_id":1,"transaction_id":4,"status":"completed","refund_id":"provider-refund","amount_minor":10,"currency":"USD","completed_at":"2026-10-27T01:00:00Z"}`,
				"payment_refund_executions": `{"id":6,"order_id":1,"refund_id":5,"transaction_id":4,"status":"succeeded","amount_minor":10,"currency":"USD","idempotency_key":"refund-command","provider_refund_id":"provider-refund","provider_payment_id":"provider-payment","provider_transaction_id":"provider-payment","merchant_order_number":"order-one","completed_at":"2026-10-27T01:00:00Z"}`,
				"order_idempotencies":       `{"id":7,"order_id":1,"user_id":3,"scope":"checkout","idempotency_key":"checkout-key","request_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
				"refund_idempotencies":      `{"id":8,"refund_id":5,"admin_user_id":3,"scope":"refund","idempotency_key":"refund-key","request_hash":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`,
				"outbox_events":             `{"id":9,"event_type":"order.completed","aggregate_type":"order","aggregate_id":"1","event_key":"order-completed","status":"processed","processed_at":"2026-10-27T01:00:00Z","locked_at":null,"locked_by":"","uncertain_at":null,"reconcile_after":null}`,
				"email_delivery_records":    `{"id":10,"outbox_event_id":9,"status":"sent","sent_at":"2026-10-27T01:00:00Z"}`,
				"shipment_records":          `{"id":11,"order_id":1,"order_number":"order-one","shipped_at":"2026-10-27T01:00:00Z"}`,
			}
			change := func(table, column, value string) {
				var fields map[string]json.RawMessage
				require.NoError(t, json.Unmarshal([]byte(encoded[table]), &fields))
				fields[column] = json.RawMessage(value)
				row, err := json.Marshal(fields)
				require.NoError(t, err)
				encoded[table] = string(row)
			}
			switch scenario {
			case "transaction_in_progress":
				change("transactions", "status", `"processing"`)
			case "missing_provider_payment_id":
				change("transactions", "transaction_id", `""`)
			case "refund_in_progress":
				change("refunds", "status", `"pending"`)
			case "refund_wrong_order":
				change("refunds", "order_id", `2`)
			case "refund_wrong_currency":
				change("refunds", "currency", `"EUR"`)
			case "refund_execution_in_progress":
				change("payment_refund_executions", "status", `"processing"`)
			case "refund_execution_wrong_provider_refund":
				change("payment_refund_executions", "provider_refund_id", `"another-refund"`)
			case "refund_execution_wrong_amount":
				change("payment_refund_executions", "amount_minor", `20`)
			case "refund_execution_wrong_provider_payment":
				change("payment_refund_executions", "provider_payment_id", `"another-payment"`)
			case "refund_execution_wrong_merchant_order":
				change("payment_refund_executions", "merchant_order_number", `"another-order"`)
			case "unbound_order_request":
				change("order_idempotencies", "order_id", `null`)
			case "invalid_refund_request_digest":
				change("refund_idempotencies", "request_hash", `"invalid"`)
			case "pending_event":
				change("outbox_events", "status", `"pending"`)
			case "unknown_event_outcome":
				change("outbox_events", "uncertain_at", `"2026-10-27T01:00:00Z"`)
			case "event_execution_lease":
				change("outbox_events", "locked_by", `"worker"`)
			case "unregistered_event":
				change("outbox_events", "event_type", `"future-event"`)
			case "unknown_notification":
				change("email_delivery_records", "status", `"unknown"`)
			case "missing_sent_time":
				change("email_delivery_records", "sent_at", `null`)
			case "unsupported_global_payment_receipt":
				encoded["payment_operation_idempotencies"] = `{"id":12,"status":"completed"}`
			case "unsupported_tracking_job":
				encoded["shipping_tracking_shipments"] = `{"id":12,"order_id":1}`
			case "nonterminal_owner":
				change("orders", "status", `"processing"`)
			}
			snapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"orders"}, Tables: map[string]businessModuleCompleteTableSnapshot{}}
			missing := map[string][]json.RawMessage{}
			for table, row := range encoded {
				snapshot.Tables[table] = businessModuleCompleteTableSnapshot{Rows: []json.RawMessage{json.RawMessage(row)}}
				missing[table] = snapshot.Tables[table].Rows
			}
			result := inspectOrderModuleMissingOrderRecordsRestoreEligibility(snapshot, missing, &SelectedBusinessModuleIsolatedRestorePreview{}, nil)
			require.Equal(t, scenario == "completed_facts", result.Eligible, "%v", result.Blockers)
			require.Equal(t, orderModuleBoundedOrderRestorePolicy, result.Policy)
		})
	}
}
