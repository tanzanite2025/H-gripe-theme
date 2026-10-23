package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"

	"github.com/lib/pq"
)

// These columns express recorded ownership, never ownership inferred from a
// table prefix, customer identity, filename or third-party response body.
func registeredOrderTransactionOwnershipColumns() map[string]string {
	ownership := map[string]string{}
	for _, table := range strings.Fields(`orders order_items transactions refunds refund_line_items
after_sales_cases after_sales_case_items order_attributions order_evidence_export_snapshots
order_evidence_items order_evidence_packages order_evidence_snapshots order_evidence_submission_snapshots
order_idempotencies order_policy_disclosures payment_refund_executions payment_refund_recommendations
payment_reviews payment_risk_checkout_decisions payment_risk_events paypal_disputes
shipment_records shipping_tracking_shipments shipping_yanwen_waybills stripe_disputes tracking_events`) {
		ownership[table] = "order_id"
	}
	delete(ownership, "orders")
	for _, table := range strings.Fields(`after_sales_case_attachments after_sales_case_events after_sales_case_events_archive after_sales_refund_reviews after_sales_return_shipments`) {
		ownership[table] = "case_id"
	}
	ownership["order_evidence_attachments"] = "evidence_item_id"
	ownership["refund_idempotencies"] = "refund_id"
	ownership["warranty_claims"] = "order_item_id"
	ownership["warranty_service_records"] = "claim_id"
	return ownership
}

func registeredOrderTransactionBusinessReferences() []ModuleRecoveryDependencyReference {
	var references []ModuleRecoveryDependencyReference
	appendReference := func(table, column, parent string) {
		references = append(references, ModuleRecoveryDependencyReference{SourceTable: table, SourceColumns: []string{column}, TargetTable: parent, TargetColumns: []string{"id"}, Kind: "order_business_reference"})
	}
	for table, column := range registeredOrderTransactionOwnershipColumns() {
		parent := map[string]string{"order_id": "orders", "case_id": "after_sales_cases", "evidence_item_id": "order_evidence_items", "refund_id": "refunds", "order_item_id": "order_items", "claim_id": "warranty_claims"}[column]
		appendReference(table, column, parent)
	}
	for _, entry := range [][3]string{
		{"orders", "user_id", "users"}, {"order_items", "product_id", "products"}, {"order_items", "variant_id", "product_variants"},
		{"orders", "shipping_quote_id", "shipping_quote_snapshots"},
		{"product_variants", "product_id", "products"}, {"refunds", "transaction_id", "transactions"},
		{"refund_line_items", "refund_id", "refunds"}, {"refund_line_items", "order_item_id", "order_items"},
		{"after_sales_case_items", "order_item_id", "order_items"}, {"after_sales_refund_reviews", "linked_refund_id", "refunds"},
		{"payment_refund_executions", "refund_id", "refunds"}, {"payment_refund_executions", "transaction_id", "transactions"},
		{"order_evidence_items", "package_id", "order_evidence_packages"}, {"order_evidence_items", "order_item_id", "order_items"},
		{"order_evidence_items", "snapshot_id", "order_evidence_snapshots"}, {"order_evidence_packages", "snapshot_id", "order_evidence_snapshots"},
		{"coupon_usage", "order_id", "orders"}, {"coupon_usage", "coupon_id", "coupons"}, {"coupon_usage", "user_id", "users"},
		{"referrals", "completed_order_id", "orders"}, {"loyalty_transactions", "source_id", "orders"},
	} {
		appendReference(entry[0], entry[1], entry[2])
	}
	sort.Slice(references, func(i, j int) bool {
		return references[i].SourceTable+references[i].SourceColumns[0] < references[j].SourceTable+references[j].SourceColumns[0]
	})
	return references
}

func orderTransactionDependencyReferences(observation *moduleRecoveryDatabaseObservation) []ModuleRecoveryDependencyReference {
	references := append([]ModuleRecoveryDependencyReference{}, observation.References...)
	for _, reference := range registeredOrderTransactionBusinessReferences() {
		if observation.Columns[reference.SourceTable][reference.SourceColumns[0]] {
			references = append(references, reference)
		}
	}
	return references
}

func (graph *moduleRecoveryRowDependencyGraph) selectOrderTransactionScopeAndRelatedSharedFacts(ctx context.Context) error {
	for _, column := range []string{"id", "order_number", "user_id"} {
		if !graph.Observation.Columns["orders"][column] {
			graph.markIncomplete("order_root_identity_schema_missing", "orders", "缺少订单主键、订单号或客户身份字段。")
		}
	}
	rows, err := graph.readBoundedRelationshipRows(ctx, "orders", "TRUE", nil, graph.Limits.Rows)
	if err != nil {
		return err
	}
	graph.includeRowsInDependencyGraph(rows, "orders", true)
	ownership := registeredOrderTransactionOwnershipColumns()
	var tables []string
	for table := range ownership {
		tables = append(tables, table)
	}
	sort.Strings(tables)
	for _, table := range tables {
		column := ownership[table]
		if !graph.Observation.Columns[table][column] {
			graph.markIncomplete("order_scope_table_or_column_missing", table, "缺少已登记的订单归属表或字段，逐行范围检查不完整。")
			continue
		}
		// Non-null orphan references are deliberately retained for diagnostics.
		// Unbound global reviews/idempotencies are preserved outside this scope.
		predicate := "record." + pq.QuoteIdentifier(column) + " IS NOT NULL"
		if table == "warranty_service_records" {
			if !graph.Observation.Columns["warranty_claims"]["order_item_id"] {
				continue
			}
			predicate = "EXISTS(SELECT 1 FROM warranty_claims claim WHERE claim.id=record.claim_id AND claim.order_item_id IS NOT NULL)"
		}
		rows, err := graph.readBoundedRelationshipRows(ctx, table, predicate, nil, graph.Limits.Rows)
		if err != nil {
			return err
		}
		graph.includeRowsInDependencyGraph(rows, table, true)
		if int64(len(rows)) < graph.Observation.Tables[table] {
			graph.Report.recordIssue("unbound_order_records_preserved", "review", graph.Side, table, "", "未关联订单的全局或历史记录保留，不按共享客户身份加入候选覆盖范围。")
		}
	}
	for _, seed := range []struct{ table, predicate string }{
		{"coupon_usage", "record.order_id IS NOT NULL"},
		{"referral_records", "record.order_id IS NOT NULL"},
		{"referrals", "record.completed_order_id IS NOT NULL AND record.completed_order_id<>0"},
		{"loyalty_transactions", "record.source='order'"},
		{"referral_reward_logs", "EXISTS(SELECT 1 FROM referral_records parent WHERE parent.id=record.referral_record_id AND parent.order_id IS NOT NULL)"},
		{"referral_transitions", "EXISTS(SELECT 1 FROM referral_records parent WHERE parent.id=record.referral_record_id AND parent.order_id IS NOT NULL)"},
	} {
		if _, exists := graph.Observation.Tables[seed.table]; !exists {
			graph.markIncomplete("order_shared_fact_table_missing", seed.table, "缺少订单关联的优惠、积分或推荐事实表。")
			continue
		}
		column := map[string]string{"coupon_usage": "order_id", "referral_records": "order_id", "referrals": "completed_order_id", "loyalty_transactions": "source", "referral_reward_logs": "referral_record_id", "referral_transitions": "referral_record_id"}[seed.table]
		if !graph.Observation.Columns[seed.table][column] {
			graph.markIncomplete("order_shared_fact_reference_missing", seed.table, "缺少共享事实的订单关联字段。")
			continue
		}
		if strings.HasPrefix(seed.table, "referral_") && seed.table != "referral_records" && !graph.Observation.Columns["referral_records"]["order_id"] {
			graph.markIncomplete("order_shared_fact_reference_missing", seed.table, "缺少推荐奖励所需的订单关联字段。")
			continue
		}
		rows, err := graph.readBoundedRelationshipRows(ctx, seed.table, seed.predicate, nil, graph.Limits.Rows)
		if err != nil {
			return err
		}
		graph.includeRowsInDependencyGraph(rows, seed.table, false)
	}
	return nil
}

func (graph *moduleRecoveryRowDependencyGraph) expandOrderTransactionForeignKeyAndBusinessDependencies(ctx context.Context) error {
	processed := map[string]bool{}
	references := orderTransactionDependencyReferences(graph.Observation)
	for pass := 0; pass < 32 && !graph.Exhausted; pass++ {
		previousCount := graph.Count
		for _, reference := range references {
			var selected []*moduleRecoveryDependencyRow
			for _, row := range graph.Rows[reference.SourceTable] {
				if reference.SourceTable == "loyalty_transactions" && reference.SourceColumns[0] == "source_id" && row.referenceColumnText("source") != "order" {
					continue
				}
				key := reference.SourceTable + ":" + row.Identity + ":" + strings.Join(reference.SourceColumns, ",") + ":" + reference.TargetTable + ":" + strings.Join(reference.TargetColumns, ",")
				if !processed[key] {
					selected = append(selected, row)
					processed[key] = true
				}
			}
			sort.Slice(selected, func(i, j int) bool { return selected[i].Identity < selected[j].Identity })
			if err := graph.inspectAndIncludeReferencedParentRows(ctx, reference, selected); err != nil {
				return err
			}
		}
		if graph.Count == previousCount {
			return nil
		}
	}
	if !graph.Exhausted {
		graph.markIncomplete("row_reference_depth_limit", "", "订单依赖超过 32 轮，闭包检查未完成。")
	}
	return nil
}

func (cfg Configuration) inspectOrderTransactionRows(ctx context.Context, source, target *sql.Tx, sourceObservation, targetObservation *moduleRecoveryDatabaseObservation, manifest *Manifest, verifiedDirectory string, limits moduleRecoveryRowInspectionLimits) (*ModuleRecoveryRowInspection, error) {
	report := &ModuleRecoveryRowInspection{Version: 1, Complete: true, StatePolicy: "preserve_financial_inventory_and_idempotency_facts_no_replay", RowLimitPerSide: limits.Rows, ProjectionByteLimit: limits.ProjectionBytes, IssueSampleLimit: limits.IssueSamples, IssueCounts: map[string]int{}, Tables: []ModuleRecoveryRowComparison{}, Issues: []ModuleRecoveryRowIssue{}}
	graphs := []*moduleRecoveryRowDependencyGraph{newModuleRecoveryRowDependencyGraph(source, sourceObservation, "package", report, limits)}
	if target != nil && targetObservation != nil {
		graphs = append(graphs, newModuleRecoveryRowDependencyGraph(target, targetObservation, "target", report, limits))
	} else {
		report.Complete = false
	}
	for _, graph := range graphs {
		if err := graph.selectOrderTransactionScopeAndRelatedSharedFacts(ctx); err != nil {
			return nil, err
		}
		if err := graph.expandOrderTransactionForeignKeyAndBusinessDependencies(ctx); err != nil {
			return nil, err
		}
		if err := graph.inspectOrderTransactionCrossOrderBusinessIdentities(ctx); err != nil {
			return nil, err
		}
		if err := graph.inspectPreservedIncomingModuleObjectReferences(ctx, "preserved_row_references_order_object", "保留行引用订单或证据对象，须保留其身份与现有引用，不能随订单局部覆盖删除。", true); err != nil {
			return nil, err
		}
		graph.Report.recordIssue("financial_inventory_and_event_state_preserved", "review", graph.Side, "outbox_events", "", "支付退款、库存、共享余额和全局幂等/事件状态继续保留；本报告不执行扣库存、退还优惠、退款、发货或重放。")
	}
	report.SourceSelectedRows = graphs[0].Count
	if len(graphs) == 2 {
		report.TargetSelectedRows = graphs[1].Count
		if err := compareModuleRecoveryDependencyRowGraphs(ctx, graphs[0], graphs[1], moduleRecoveryRowComparisonPolicy{
			ScopedContentDifferenceCode: "order_transaction_row_content_differs", ScopedContentDifferenceMessage: "订单交易行与恢复点不同，须先核对后续支付、退款、发货及证据事实。",
			IdentityConflictMessage: "同一主键的订单、客户、商品或证据归属发生变化，不能按主键直接覆盖。",
			TargetOnlyCode:          "target_only_order_transaction_row", TargetOnlyMessage: "当前新增交易行不在恢复点，须明确保留与外部对账规则，不能静默删除。", OwnershipConflicts: orderTransactionRowOwnershipConflicts,
		}); err != nil {
			return nil, err
		}
		for index := range report.Tables {
			comparison := &report.Tables[index]
			switch comparison.Table {
			case "coupon_usage", "loyalty_transactions", "referral_records", "referrals", "referral_reward_logs", "referral_transitions":
			default:
				continue
			}
			for identity, row := range graphs[1].Rows[comparison.Table] {
				if graphs[0].Rows[comparison.Table][identity] == nil {
					comparison.TargetOnlyShared++
					report.recordIssue("target_only_shared_order_fact", "blocking", "target", comparison.Table, moduleRecoveryDependencyRowPublicIdentity(row, graphs[1].Observation.PrimaryKeys[comparison.Table]), "目标新增的优惠、积分或推荐流水不在恢复点，须保留并对账，不能随旧订单恢复回退。")
				}
			}
		}
		var sourceOrders []*moduleRecoveryDependencyRow
		sourceByOrderNumber := map[string]*moduleRecoveryDependencyRow{}
		for _, orderRow := range graphs[0].Rows["orders"] {
			if graphs[0].Observation.Shapes["orders"] != graphs[1].Observation.Shapes["orders"] {
				continue
			}
			if orderRow.referenceColumnText("order_number") != "" {
				sourceOrders = append(sourceOrders, orderRow)
				sourceByOrderNumber[orderRow.referenceColumnText("order_number")] = orderRow
			}
		}
		sort.Slice(sourceOrders, func(i, j int) bool { return sourceOrders[i].Identity < sourceOrders[j].Identity })
		for start := 0; start < len(sourceOrders) && !graphs[1].Exhausted; start += 200 {
			var keys []map[string]json.RawMessage
			for _, orderRow := range sourceOrders[start:min(start+200, len(sourceOrders))] {
				keys = append(keys, map[string]json.RawMessage{"order_number": orderRow.Values["order_number"]})
			}
			matches, err := graphs[1].readRowsMatchingReferenceKeys(ctx, "orders", []string{"order_number"}, keys, limits.Rows)
			if err != nil {
				return nil, err
			}
			for _, match := range matches {
				if sourceByOrderNumber[match.referenceColumnText("order_number")].Identity != match.Identity {
					report.recordIssue("order_number_primary_key_identity_conflict", "blocking", "target", "orders", match.referenceColumnText("id"), "相同订单号对应不同主键，不能重编号、补入或按旧主键覆盖。")
				}
			}
		}
	}
	if err := cfg.inspectModuleRecoveryAttachmentFiles(ctx, graphs, manifest, verifiedDirectory, report, cfg.collectOrderTransactionEvidenceAndAfterSalesFileReferences); err != nil {
		return nil, err
	}
	if err := cfg.inspectOrderTransactionDurableEventDependencies(ctx, graphs, report); err != nil {
		return nil, err
	}
	if err := inspectModuleRecoveryPaymentRefundAndNotificationReceiptPreservation(ctx, graphs, report); err != nil {
		return nil, err
	}
	report.SourceInspectedRows = len(graphs[0].ObservedRows)
	if len(graphs) == 2 {
		report.TargetInspectedRows = len(graphs[1].ObservedRows)
	}
	return report, nil
}

func orderTransactionRowOwnershipConflicts(table string, source, target *moduleRecoveryDependencyRow) bool {
	columns := []string{"order_id", "order_item_id", "case_id", "refund_id", "evidence_item_id", "package_id", "snapshot_id", "claim_id", "product_id", "variant_id"}
	if table == "orders" {
		columns = append(columns, "order_number", "user_id")
	}
	for _, column := range columns {
		if source.referenceColumnText(column) != target.referenceColumnText(column) {
			return true
		}
	}
	return false
}

func findOrderTransactionDependencyRowByID(graph *moduleRecoveryRowDependencyGraph, table string, identifier json.RawMessage) *moduleRecoveryDependencyRow {
	return graph.Rows[table][moduleRecoveryDependencyRowKey(map[string]json.RawMessage{"id": identifier}, []string{"id"})]
}
