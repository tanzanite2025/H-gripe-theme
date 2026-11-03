package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"

	"github.com/lib/pq"
)

// Existing immutable order snapshots have UPDATE-only guards. They cannot run
// during these explicit INSERTs. INSERT triggers and rules remain prohibited.
func validateOrderModuleInsertOnlyWriteContract(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, table string, observation *moduleRecoveryDatabaseObservation) error {
	if !reflect.DeepEqual(snapshot.Modules, []string{"orders"}) {
		return validateSelectedBusinessModuleIsolatedInsertWriteContract(ctx, tx, table, observation)
	}
	return validateBusinessModuleInsertOnlyTableContractWithUpdateOnlyGuardsPreserved(ctx, tx, table, observation)
}

// Completed facts are copied exactly from the authenticated package. Current
// records, including newer notes, are retained; execution is never replayed.
type OrderModuleHistoryRestoreEligibility struct {
	Policy                   string   `json:"policy"`
	Eligible                 bool     `json:"eligible"`
	Blockers                 []string `json:"blockers"`
	SettledActiveOrderOwners int      `json:"settled_active_order_owners,omitempty"`
}

func inspectOrderModuleMissingOrderRecordsRestoreEligibility(snapshot selectedBusinessModuleCompleteSnapshot, missing map[string][]json.RawMessage, report *SelectedBusinessModuleIsolatedRestorePreview, missingFiles []FileIdentity) *OrderModuleHistoryRestoreEligibility {
	result := &OrderModuleHistoryRestoreEligibility{Policy: orderModuleBoundedOrderRestorePolicy, Eligible: true, Blockers: []string{}}
	record := func(reason string) {
		for _, previous := range result.Blockers {
			if previous == reason {
				return
			}
		}
		result.Blockers = append(result.Blockers, reason)
		result.Eligible = false
	}
	if !reflect.DeepEqual(snapshot.Modules, []string{"orders"}) {
		record("only_a_separate_order_package_is_supported")
		return result
	}
	allowed := map[string]bool{}
	for _, table := range strings.Fields("orders order_items order_attributions order_policy_disclosures order_evidence_attachments order_evidence_export_snapshots order_evidence_items order_evidence_packages order_evidence_snapshots order_evidence_submission_snapshots") {
		allowed[table] = true
	}
	index := indexOrderModuleHistoricalRowsByTableAndID(snapshot)
	ownerBlockers := map[string]string{}
	activeOwners := map[string]bool{}
	recordOwnerBlocker := func(orderID, table string) {
		blocker, inspected := ownerBlockers[orderID]
		if !inspected {
			blocker = inspectOrderModuleTerminalOrSettledActiveOwnerRestoreBlocker(index, orderID)
			status, _ := readBusinessModuleSnapshotRowColumnText(index["orders"][orderID], "status")
			if blocker == "" && !orderModuleOrderStatusHasTerminalHistoryRestoreContract(status) {
				for _, item := range snapshot.Tables["order_items"].Rows {
					itemOrderID, _ := readBusinessModuleSnapshotRowColumnText(item, "order_id")
					if itemOrderID != orderID {
						continue
					}
					allocations, _ := readOrderModuleConfiguredComponentInventoryAllocationsForRestore(item)
					if !orderModuleComponentInventoryAllocationsHaveRecordedVariantDependencies(snapshot, allocations) {
						blocker = "active_order_component_inventory_original_dependency_missing"
						break
					}
				}
			}
			ownerBlockers[orderID] = blocker
		}
		if blocker != "" {
			record(blocker + ":" + table)
		} else {
			status, _ := readBusinessModuleSnapshotRowColumnText(index["orders"][orderID], "status")
			if !orderModuleOrderStatusHasTerminalHistoryRestoreContract(status) {
				activeOwners[orderID] = true
			}
		}
	}
	itemOrders := map[string]string{}
	for _, row := range snapshot.Tables["order_evidence_items"].Rows {
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
		orderID, _ := readBusinessModuleSnapshotRowColumnText(row, "order_id")
		itemOrders[id] = orderID
	}
	evidenceFileOwners := map[string]string{}
	for _, row := range snapshot.Tables["order_evidence_attachments"].Rows {
		itemID, _ := readBusinessModuleSnapshotRowColumnText(row, "evidence_item_id")
		key, _ := readBusinessModuleSnapshotRowColumnText(row, "storage_key")
		evidenceFileOwners["media/uploads_private/"+key] = itemOrders[itemID]
	}
	for _, file := range missingFiles {
		if owner := evidenceFileOwners[file.Path]; owner != "" {
			recordOwnerBlocker(owner, "order_evidence_attachments")
		} else {
			record("missing_attachment_requires_registered_order_evidence_owner")
		}
	}
	changed := len(missingFiles) > 0
	for table, rows := range missing {
		if len(rows) == 0 {
			continue
		}
		changed = true
		if table == "payment_operation_idempotencies" {
			for _, row := range rows {
				if reason := inspectOrderModuleCompletedGlobalPaymentReceiptRestoreBlocker(row); reason != "" {
					record(reason + ":" + table)
				}
			}
			continue
		}
		if !allowed[table] {
			for _, row := range rows {
				owner, reason := inspectOrderModuleCompletedHistoricalFactOwnerAndBlocker(index, table, row)
				if reason != "" {
					record(reason + ":" + table)
				}
				recordOwnerBlocker(owner, table)
			}
			continue
		}
		for _, row := range rows {
			column := "order_id"
			if table == "orders" {
				column = "id"
			}
			id, _ := readBusinessModuleSnapshotRowColumnText(row, column)
			if table == "order_evidence_attachments" {
				itemID, _ := readBusinessModuleSnapshotRowColumnText(row, "evidence_item_id")
				for _, item := range snapshot.Tables["order_evidence_items"].Rows {
					candidate, _ := readBusinessModuleSnapshotRowColumnText(item, "id")
					if candidate == itemID {
						id, _ = readBusinessModuleSnapshotRowColumnText(item, "order_id")
					}
				}
			}
			recordOwnerBlocker(id, table)
		}
	}
	if changed {
		for _, row := range snapshot.Tables["outbox_events"].Rows {
			status, _ := readBusinessModuleSnapshotRowColumnText(row, "status")
			if status != "processed" {
				record("nonterminal_order_events_require_separate_reconciliation")
			}
		}
	}
	if report.DependencyConflicts > 0 || report.FileConflicts > 0 {
		record("dependencies_or_files_conflict")
	}
	if report.OrderHistoryCurrentExecutionConflicts > 0 {
		record("current_related_execution_records_not_in_package")
	}
	for _, table := range report.Tables {
		if table.ConflictingRows > 0 || table.ProtectedNewRows > 0 || table.UnsupportedReason != "" {
			record("current_content_or_protected_record_conflict:" + table.Table)
		}
	}
	result.SettledActiveOrderOwners = len(activeOwners)
	return result
}

// A missing historical order must not be resurrected over newer financial or
// fulfillment facts for that same order. Unrelated new orders remain intact.
func countOrderHistoryCurrentExecutionScopesWithRowsAbsentFromPackage(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, observation *moduleRecoveryDatabaseObservation) (int, error) {
	identities := map[string][]string{}
	for _, table := range []string{"orders", "refunds", "after_sales_cases", "order_items", "order_evidence_items", "warranty_claims", "outbox_events"} {
		for _, row := range snapshot.Tables[table].Rows {
			id, err := readBusinessModuleSnapshotRowColumnText(row, "id")
			if err != nil {
				return 0, err
			}
			identities[table] = append(identities[table], id)
		}
	}
	var names []string
	for table := range snapshot.Tables {
		names = append(names, table)
	}
	sort.Strings(names)
	conflicts := 0
	for _, table := range names {
		predicate := ""
		ids := identities["orders"]
		if observation.Columns[table]["order_id"] {
			predicate = "record.order_id=ANY($2::bigint[])"
		}
		if ownerColumn := registeredOrderTransactionOwnershipColumns()[table]; ownerColumn != "" && ownerColumn != "order_id" {
			parent := map[string]string{"case_id": "after_sales_cases", "evidence_item_id": "order_evidence_items", "refund_id": "refunds", "order_item_id": "order_items", "claim_id": "warranty_claims"}[ownerColumn]
			ids = identities[parent]
			predicate = "record." + pq.QuoteIdentifier(ownerColumn) + "=ANY($2::bigint[])"
		}
		if table == "outbox_events" {
			predicate = "((record.aggregate_type='order' AND record.aggregate_id=ANY($2::text[])) OR (record.aggregate_type='payment' AND record.aggregate_id=ANY($3::text[])) OR (record.aggregate_type='after_sales_case' AND record.aggregate_id=ANY($4::text[])))"
		}
		if table == "email_delivery_records" {
			ids = identities["outbox_events"]
			predicate = "record.outbox_event_id=ANY($2::bigint[])"
		}
		if table == "orders" {
			continue
		}
		if predicate == "" || len(ids) == 0 {
			continue
		}
		var matches []string
		for _, key := range observation.PrimaryKeys[table] {
			matches = append(matches, "package_record."+pq.QuoteIdentifier(key)+"=record."+pq.QuoteIdentifier(key))
		}
		if len(matches) == 0 {
			return 0, errors.New("order execution table has no verifiable primary key")
		}
		rows, err := json.Marshal(append([]json.RawMessage{}, snapshot.Tables[table].Rows...))
		if err != nil {
			return 0, err
		}
		query := `SELECT EXISTS(SELECT 1 FROM public.` + pq.QuoteIdentifier(table) + ` record WHERE ` + predicate + ` AND NOT EXISTS(SELECT 1 FROM jsonb_populate_recordset(NULL::public.` + pq.QuoteIdentifier(table) + `,$1::jsonb) package_record WHERE ` + strings.Join(matches, " AND ") + `))`
		arguments := []any{string(rows), pq.Array(ids)}
		if table == "outbox_events" {
			arguments = append(arguments, pq.Array(identities["refunds"]), pq.Array(identities["after_sales_cases"]))
		}
		var extra bool
		if err := tx.QueryRowContext(ctx, query, arguments...).Scan(&extra); err != nil {
			return 0, err
		}
		if extra {
			conflicts++
		}
	}
	return conflicts, nil
}
