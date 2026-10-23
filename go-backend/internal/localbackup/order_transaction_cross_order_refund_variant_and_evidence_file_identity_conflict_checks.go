package localbackup

import (
	"context"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

func (graph *moduleRecoveryRowDependencyGraph) inspectOrderTransactionCrossOrderBusinessIdentities(ctx context.Context) error {
	ownershipColumns := registeredOrderTransactionOwnershipColumns()
	var tables []string
	for table := range graph.Rows {
		tables = append(tables, table)
	}
	sort.Strings(tables)
	for _, table := range tables {
		var identities []string
		for identity := range graph.Rows[table] {
			identities = append(identities, identity)
		}
		sort.Strings(identities)
		for _, rowIdentity := range identities {
			row := graph.Rows[table][rowIdentity]
			if err := ctx.Err(); err != nil {
				return err
			}
			identity := moduleRecoveryDependencyRowPublicIdentity(row, graph.Observation.PrimaryKeys[table])
			if table == "orders" && row.referenceColumnText("order_number") == "" {
				graph.Report.recordIssue("order_number_identity_missing", "blocking", graph.Side, table, identity, "订单缺少稳定订单号，不能通过重编号修补历史身份。")
			}
			if column := ownershipColumns[table]; column != "" && graph.Scoped[table][row.Identity] && row.referenceColumnText(column) == "0" {
				graph.Report.recordIssue("order_owner_identity_invalid", "blocking", graph.Side, table, identity, "订单关联行的归属 ID 为零，不能视为合法的未绑定全局记录。")
			}
			for column, value := range row.Values {
				if !strings.HasSuffix(column, "_minor") || string(value) == "null" {
					continue
				}
				amount, err := strconv.ParseInt(string(value), 10, 64)
				if err != nil || amount < 0 {
					graph.Report.recordIssue("order_monetary_minor_snapshot_invalid", "blocking", graph.Side, table, identity, "交易金额快照不是合法的非负最小货币单位，不能自动修正或重算。")
				}
			}
			orderID := row.referenceColumnText("order_id")
			for _, link := range []struct{ column, parent string }{
				{"transaction_id", "transactions"}, {"refund_id", "refunds"}, {"order_item_id", "order_items"},
				{"case_id", "after_sales_cases"}, {"package_id", "order_evidence_packages"}, {"snapshot_id", "order_evidence_snapshots"},
			} {
				// Third-party transaction identifiers are strings on transactions;
				// only registered numeric transaction links are compared here.
				if orderID == "" || table == link.parent || !graph.Observation.Columns[table][link.column] {
					continue
				}
				parent := findOrderTransactionDependencyRowByID(graph, link.parent, row.Values[link.column])
				if parent != nil && parent.referenceColumnText("order_id") != orderID {
					graph.Report.recordIssue("transaction_reference_crosses_orders", "blocking", graph.Side, table, identity, "交易、退款、售后或证据引用属于另一笔订单，不能仅凭独立外键存在就放行。")
				}
			}
			if table == "order_items" {
				variant := findOrderTransactionDependencyRowByID(graph, "product_variants", row.Values["variant_id"])
				if variant != nil && variant.referenceColumnText("product_id") != row.referenceColumnText("product_id") {
					graph.Report.recordIssue("order_variant_product_identity_conflict", "blocking", graph.Side, table, identity, "订单 SKU 与商品归属不一致，不能用当前库存或商品快照替换历史交易事实。")
				}
			}
			if table == "order_items" || table == "refund_line_items" || table == "after_sales_case_items" {
				quantity, err := strconv.ParseInt(row.referenceColumnText("quantity"), 10, 64)
				if err != nil || quantity <= 0 {
					graph.Report.recordIssue("order_line_quantity_invalid", "blocking", graph.Side, table, identity, "交易行数量无效，不能推算库存返还或退款数量。")
				}
				if table != "order_items" {
					original := findOrderTransactionDependencyRowByID(graph, "order_items", row.Values["order_item_id"])
					if original != nil {
						originalQuantity, _ := strconv.ParseInt(original.referenceColumnText("quantity"), 10, 64)
						if quantity > originalQuantity {
							graph.Report.recordIssue("refund_or_after_sales_quantity_exceeds_order_line", "blocking", graph.Side, table, identity, "退款或售后行数量超过原订单行，不能自动返库存。")
						}
						for _, column := range []string{"product_id", "variant_id"} {
							if row.referenceColumnText(column) != "" && row.referenceColumnText(column) != original.referenceColumnText(column) {
								graph.Report.recordIssue("refund_or_after_sales_product_identity_conflict", "blocking", graph.Side, table, identity, "退款或售后商品/SKU 快照与所引用订单行不符。")
							}
						}
					}
				}
			}
			if table == "after_sales_refund_reviews" {
				caseRow := findOrderTransactionDependencyRowByID(graph, "after_sales_cases", row.Values["case_id"])
				refund := findOrderTransactionDependencyRowByID(graph, "refunds", row.Values["linked_refund_id"])
				if caseRow != nil && refund != nil && caseRow.referenceColumnText("order_id") != refund.referenceColumnText("order_id") {
					graph.Report.recordIssue("after_sales_refund_review_crosses_orders", "blocking", graph.Side, table, identity, "售后审批关联退款属于另一笔订单，须先核对付款与审批事实。")
				}
			}
		}
	}
	return nil
}

func (cfg Configuration) collectOrderTransactionEvidenceAndAfterSalesFileReferences(graph *moduleRecoveryRowDependencyGraph) []moduleRecoveryFileReference {
	var references []moduleRecoveryFileReference
	for _, table := range []string{"order_evidence_attachments", "after_sales_case_attachments"} {
		for _, row := range graph.Rows[table] {
			identity := moduleRecoveryDependencyRowPublicIdentity(row, graph.Observation.PrimaryKeys[table])
			root, key := "uploads_private", row.referenceColumnText("storage_key")
			digest := strings.ToLower(strings.TrimSpace(row.referenceColumnText("sha256")))
			if table == "after_sales_case_attachments" {
				root, key = cfg.resolveLocalMediaReference(row.referenceColumnText("storage_url"))
			}
			if key == "" || !safePackagePath("media/"+root+"/"+key) {
				graph.markIncomplete("order_attachment_reference_unresolved", table, "订单附件无法映射到安全本地对象，不能证明备份文件范围完整。")
				continue
			}
			if table == "order_evidence_attachments" {
				item := findOrderTransactionDependencyRowByID(graph, "order_evidence_items", row.Values["evidence_item_id"])
				if item == nil || !strings.HasPrefix(key, "order-evidence/"+item.referenceColumnText("order_id")+"/"+item.referenceColumnText("id")+"/") {
					graph.Report.recordIssue("order_evidence_attachment_owner_identity_conflict", "blocking", graph.Side, table, identity, "证据附件路径与显式登记的订单及证据项身份不符，不能按路径认领文件。")
				}
				decoded, err := hex.DecodeString(digest)
				if err != nil || len(decoded) != 32 {
					graph.markIncomplete("order_evidence_attachment_digest_invalid", table, "证据附件缺少合法 SHA-256，无法核对不可变原件。")
				}
			}
			size, err := strconv.ParseInt(row.referenceColumnText("size_bytes"), 10, 64)
			if err != nil || size < 0 {
				graph.markIncomplete("order_attachment_size_invalid", table, "附件大小登记无效，无法完成文件一致性核对。")
				continue
			}
			references = append(references, moduleRecoveryFileReference{Root: root, Key: key, Table: table, RowIdentity: identity, Digest: digest, Size: size, SizeKnown: true})
		}
	}
	return references
}
