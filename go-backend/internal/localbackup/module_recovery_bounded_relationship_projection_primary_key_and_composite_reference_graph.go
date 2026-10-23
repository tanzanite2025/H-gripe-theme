package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

// This is an advisory inventory, never an executable partial-restore plan.
// Complete means the bounded inspection finished, not that replacement is safe.
type ModuleRecoveryRowInspection struct {
	StatePolicy                 string                                                               `json:"state_policy,omitempty"`
	Version                     int                                                                  `json:"version"`
	Complete                    bool                                                                 `json:"complete"`
	RowLimitPerSide             int                                                                  `json:"row_limit_per_side"`
	ProjectionByteLimit         int                                                                  `json:"projection_byte_limit_per_side"`
	IssueSampleLimit            int                                                                  `json:"issue_sample_limit"`
	IssuesTruncated             bool                                                                 `json:"issues_truncated"`
	SourceSelectedRows          int                                                                  `json:"source_selected_rows"`
	TargetSelectedRows          int                                                                  `json:"target_selected_rows"`
	SourceInspectedRows         int                                                                  `json:"source_inspected_rows"`
	TargetInspectedRows         int                                                                  `json:"target_inspected_rows"`
	BlockingIssues              int                                                                  `json:"blocking_issues"`
	IssueCounts                 map[string]int                                                       `json:"issue_counts"`
	BlockingIssueCounts         map[string]int                                                       `json:"blocking_issue_counts,omitempty"`
	Tables                      []ModuleRecoveryRowComparison                                        `json:"tables"`
	Issues                      []ModuleRecoveryRowIssue                                             `json:"issues"`
	Files                       ModuleRecoveryFileInspection                                         `json:"files"`
	DurableEvents               *ModuleRecoveryDurableEventInspection                                `json:"durable_events,omitempty"`
	SideEffectReceipts          *ModuleRecoverySideEffectReceiptInspection                           `json:"side_effect_receipts,omitempty"`
	ConversationRecoveryPlan    *CustomerServiceConversationRecoveryPlan                             `json:"conversation_recovery_plan,omitempty"`
	ConversationReferenceChecks *CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection `json:"conversation_reference_checks,omitempty"`
}

type ModuleRecoveryRowComparison struct {
	Table             string `json:"table"`
	SourceSelected    int    `json:"source_selected"`
	TargetSelected    int    `json:"target_selected"`
	SourceScoped      int    `json:"source_scoped"`
	TargetScoped      int    `json:"target_scoped"`
	Matching          int    `json:"matching"`
	TargetMissing     int    `json:"target_missing"`
	ContentDiffers    int    `json:"content_differs"`
	IdentityConflicts int    `json:"identity_conflicts"`
	TargetOnlyScoped  int    `json:"target_only_scoped"`
	TargetOnlyShared  int    `json:"target_only_shared,omitempty"`
}

type ModuleRecoveryRowIssue struct {
	Code        string `json:"code"`
	Severity    string `json:"severity"`
	Side        string `json:"side"`
	Table       string `json:"table,omitempty"`
	RowIdentity string `json:"row_identity,omitempty"`
	Message     string `json:"message"`
}

type moduleRecoveryRowInspectionLimits struct{ Rows, ProjectionBytes, IssueSamples int }

func defaultModuleRecoveryRowInspectionLimits() moduleRecoveryRowInspectionLimits {
	return moduleRecoveryRowInspectionLimits{Rows: 20000, ProjectionBytes: 16 << 20, IssueSamples: 200}
}

func (report *ModuleRecoveryRowInspection) recordIssue(code, severity, side, table, identity, message string) {
	report.IssueCounts[code]++
	if severity == "blocking" {
		report.BlockingIssues++
		if report.BlockingIssueCounts == nil {
			report.BlockingIssueCounts = map[string]int{}
		}
		report.BlockingIssueCounts[code]++
	}
	if len(report.Issues) < report.IssueSampleLimit {
		report.Issues = append(report.Issues, ModuleRecoveryRowIssue{code, severity, side, table, identity, message})
	} else {
		report.IssuesTruncated = true
	}
}

type moduleRecoveryDependencyRow struct {
	Values           map[string]json.RawMessage
	Identity, Digest string
}

func (row *moduleRecoveryDependencyRow) referenceColumnText(column string) string {
	value := row.Values[column]
	if len(value) == 0 || string(value) == "null" {
		return ""
	}
	var decoded string
	if json.Unmarshal(value, &decoded) == nil {
		return decoded
	}
	return string(value)
}

func moduleRecoveryDependencyRowKey(values map[string]json.RawMessage, columns []string) string {
	key := make(map[string]json.RawMessage, len(columns))
	for _, column := range columns {
		if len(values[column]) == 0 || string(values[column]) == "null" {
			return ""
		}
		key[column] = values[column]
	}
	encoded, _ := json.Marshal(key)
	return string(encoded)
}

func moduleRecoveryDependencyRowPublicIdentity(row *moduleRecoveryDependencyRow, primaryKeys []string) string {
	if len(primaryKeys) == 1 && primaryKeys[0] == "id" {
		identity := row.referenceColumnText("id")
		if numeric, err := strconv.ParseUint(identity, 10, 64); err == nil && strconv.FormatUint(numeric, 10) == identity {
			return identity
		}
	}
	digest := sha256.Sum256([]byte(row.Identity))
	return "sha256:" + hex.EncodeToString(digest[:16])
}

type moduleRecoveryRowDependencyGraph struct {
	Transaction                        *sql.Tx
	Observation                        *moduleRecoveryDatabaseObservation
	Rows                               map[string]map[string]*moduleRecoveryDependencyRow
	Scoped                             map[string]map[string]bool
	Side                               string
	Count, ProjectionBytes             int
	Exhausted                          bool
	Limits                             moduleRecoveryRowInspectionLimits
	Report                             *ModuleRecoveryRowInspection
	LegacyFiles                        []moduleRecoveryFileReference
	ObservedRows                       map[string]bool
	InspectOrderEventPayload           bool
	ComparedTargetRows                 map[string]map[string]*moduleRecoveryDependencyRow
	InspectedFiles                     map[string]moduleRecoveryInspectedFileSnapshot
	RetainedStructuredFiles            []moduleRecoveryFileReference
	RetainedStructuredAssetIDs         map[string]bool
	BlockedCandidateRows               map[string]map[string]bool
	BusinessSnapshots                  []customerServiceStructuredBusinessSnapshot
	BusinessSnapshotReferenceCount     int
	CustomerServiceDecisionIssueCounts map[string]map[string]map[string]int
	SelectedCustomerServiceRows        map[string]map[string]*moduleRecoveryDependencyRow
}

func newModuleRecoveryRowDependencyGraph(transaction *sql.Tx, observation *moduleRecoveryDatabaseObservation, side string, report *ModuleRecoveryRowInspection, limits moduleRecoveryRowInspectionLimits) *moduleRecoveryRowDependencyGraph {
	return &moduleRecoveryRowDependencyGraph{Transaction: transaction, Observation: observation, Side: side, Report: report, Limits: limits, Rows: map[string]map[string]*moduleRecoveryDependencyRow{}, Scoped: map[string]map[string]bool{}, ObservedRows: map[string]bool{}, ComparedTargetRows: map[string]map[string]*moduleRecoveryDependencyRow{}, InspectedFiles: map[string]moduleRecoveryInspectedFileSnapshot{}, RetainedStructuredAssetIDs: map[string]bool{}, BlockedCandidateRows: map[string]map[string]bool{}, CustomerServiceDecisionIssueCounts: map[string]map[string]map[string]int{}}
}

func (graph *moduleRecoveryRowDependencyGraph) markIncomplete(code, table, message string) {
	graph.Report.Complete = false
	graph.Report.recordIssue(code, "blocking", graph.Side, table, "", message)
}

// Column values are a private relationship projection; message bodies, names,
// email addresses and full event payloads never leave PostgreSQL. Oversized projections
// are rejected before transfer, and the whole graph has an additional byte cap.
func moduleRecoveryDependencyProjectionColumns(observation *moduleRecoveryDatabaseObservation, table string) []string {
	fields := map[string]bool{}
	for _, reference := range registeredOrderTransactionBusinessReferences() {
		if reference.SourceTable == table {
			for _, column := range reference.SourceColumns {
				if observation.Columns[table][column] {
					fields[column] = true
				}
			}
		}
	}
	for _, column := range observation.PrimaryKeys[table] {
		fields[column] = true
	}
	for _, reference := range observation.References {
		if reference.SourceTable == table {
			for _, column := range reference.SourceColumns {
				fields[column] = true
			}
		}
		if reference.TargetTable == table {
			for _, column := range reference.TargetColumns {
				fields[column] = true
			}
		}
	}
	extra := map[string]string{
		"orders":                            "order_number user_id checkout_cart_id shipping_quote_id",
		"order_items":                       "order_id product_id variant_id quantity currency price_minor subtotal_minor tax_amount_minor discount_minor total_minor",
		"products":                          "id sku slug",
		"product_variants":                  "product_id stock sku",
		"faqs":                              "page_id locale",
		"selection_assistant_flows":         "slug",
		"selection_assistant_flow_versions": "flow_id version_number",
		"transactions":                      "order_id status amount_minor currency",
		"refunds":                           "order_id transaction_id amount_minor currency status",
		"refund_line_items":                 "refund_id order_id order_item_id product_id variant_id quantity amount_minor currency",
		"payment_refund_executions":         "refund_id order_id transaction_id status provider idempotency_key attempt_count",
		"after_sales_cases":                 "order_id",
		"after_sales_case_items":            "case_id order_id order_item_id product_id variant_id quantity",
		"after_sales_refund_reviews":        "case_id linked_refund_id",
		"order_evidence_packages":           "order_id snapshot_id",
		"order_evidence_items":              "order_id package_id order_item_id snapshot_id",
		"order_evidence_snapshots":          "order_id",
		"order_evidence_attachments":        "evidence_item_id storage_key size_bytes sha256",
		"after_sales_case_attachments":      "case_id storage_url size_bytes",
		"order_idempotencies":               "order_id user_id",
		"refund_idempotencies":              "admin_user_id scope idempotency_key request_hash refund_id",
		"payment_operation_idempotencies":   "user_id scope idempotency_key request_hash status status_code lease_expires_at reconciliation_started_at",
		"email_delivery_records":            "outbox_event_id event_key event_type template_code template_locale template_version reference_type reference_number status attempt_count provider_code first_attempt_at last_attempt_at sent_at",
		"loyalty_transactions":              "source source_id user_id",
		"coupon_usage":                      "order_id user_id coupon_id",
		"referrals":                         "completed_order_id",
		"tickets":                           "category ticket_number conversation_id visitor_session_hash user_id customer_user_id assigned_to group_id status_version",
		"ticket_messages":                   "ticket_id user_id attachments attachment_reference_format",
		"ticket_auto_replies":               "attachments",
		"customer_service_inbox_states":     "ticket_id recipient_user_id last_read_message_id assignment_version unread_count",
		"visitor_profiles":                  "customer_service_visitor_hash user_id deleted_at",
		"media_upload_sessions":             "business_context_type business_context_id upload_actor_type upload_actor_account_id upload_actor_visitor_session_hash media_asset_id object_key status",
		"ticket_message_media_attachments":  "ticket_message_id media_asset_id upload_session_id attachment_index",
		"media_assets":                      "storage_key url visibility size content_sha256 library_partition customer_service_media_purpose classification_method upload_origin uploader_id uploader_actor_type uploader_visitor_session_hash deleted_at customer_service_derivative_write_token customer_service_derivative_write_expires_at",
		"media_asset_derivatives":           "media_asset_id storage_key url size deleted_at",
		"outbox_events":                     "event_key event_type aggregate_type aggregate_id status attempts max_attempts available_at last_attempt_at locked_at locked_by processed_at uncertain_at reconcile_after deleted_at",
		"customer_service_agent_profiles":   "avatar",
	}
	for _, column := range strings.Fields(extra[table]) {
		if observation.Columns[table][column] {
			fields[column] = true
		}
	}
	columns := make([]string, 0, len(fields))
	for column := range fields {
		columns = append(columns, column)
	}
	sort.Strings(columns)
	return columns
}

func (graph *moduleRecoveryRowDependencyGraph) readBoundedRelationshipRows(ctx context.Context, table, predicate string, arguments []any, maximum int) ([]*moduleRecoveryDependencyRow, error) {
	if graph.Transaction == nil || graph.Exhausted {
		return nil, nil
	}
	if len(graph.Observation.PrimaryKeys[table]) == 0 {
		graph.markIncomplete("row_primary_key_or_table_missing", table, "缺少表或主键，无法稳定识别模块依赖行。")
		return nil, nil
	}
	var projectionParts []string
	for _, column := range moduleRecoveryDependencyProjectionColumns(graph.Observation, table) {
		projectionParts = append(projectionParts, pq.QuoteLiteral(column), "record."+pq.QuoteIdentifier(column))
	}
	if table == "outbox_events" && graph.Observation.Columns[table]["payload"] {
		if graph.InspectOrderEventPayload {
			projectionParts = append(projectionParts, orderTransactionDurableEventPayloadProjectionEntries()...)
		} else {
			projectionParts = append(projectionParts, customerServiceDurableEventPayloadProjectionEntries()...)
		}
	}
	if table == "payment_operation_idempotencies" && graph.Observation.Columns[table]["claim_token"] {
		projectionParts = append(projectionParts, "'_claim_token_present'", "COALESCE(record.claim_token,'')<>''")
	}
	if table == "selection_assistant_flow_versions" && graph.Observation.Columns[table]["config"] {
		projectionParts = append(projectionParts, "'_snapshot_configuration_sha256'", "encode(sha256(convert_to(to_jsonb(record.config)::text,'UTF8')),'hex')")
	}
	if table == "ticket_messages" || table == "ticket_auto_replies" {
		for _, column := range []string{"message_type", "metadata"} {
			if graph.Observation.Columns[table][column] {
				projectionParts = append(projectionParts, pq.QuoteLiteral("_structured_"+column), "record."+pq.QuoteIdentifier(column))
			}
		}
	}
	projection := `jsonb_build_object(` + strings.Join(projectionParts, ",") + `)`
	// jsonb_build_object accepts at most 100 function arguments.
	if len(projectionParts) > 100 {
		graph.markIncomplete("row_projection_unsupported", table, "引用字段过多，本版检查无法完整展开。 ")
		return nil, nil
	}
	relation, digest, arguments, err := graph.customerServiceEffectiveSelectionReadOnlyRelation(table, arguments)
	if err != nil {
		return nil, err
	}
	arguments = append(arguments, maximum+1)
	query := `SELECT CASE WHEN octet_length((` + projection + `)::text)<=65536 THEN (` + projection + `)::text ELSE '{}' END, ` + digest + ` FROM ` + relation + ` WHERE ` + predicate + ` ORDER BY `
	var ordering []string
	for _, column := range graph.Observation.PrimaryKeys[table] {
		ordering = append(ordering, "record."+pq.QuoteIdentifier(column))
	}
	query += strings.Join(ordering, ",") + fmt.Sprintf(" LIMIT $%d", len(arguments))
	rows, err := graph.Transaction.QueryContext(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("inspect %s module row references: %w", table, err)
	}
	defer rows.Close()
	result := make([]*moduleRecoveryDependencyRow, 0)
	for rows.Next() {
		var projectionText, digest string
		if err := rows.Scan(&projectionText, &digest); err != nil {
			return nil, err
		}
		if len(result) >= maximum || graph.ProjectionBytes+len(projectionText) > graph.Limits.ProjectionBytes {
			graph.Exhausted = true
			graph.markIncomplete("row_inspection_resource_limit", table, "行数或引用字段容量达到上限，结果不完整；未执行任何恢复。 ")
			break
		}
		graph.ProjectionBytes += len(projectionText)
		row := &moduleRecoveryDependencyRow{Digest: digest}
		if err := json.Unmarshal([]byte(projectionText), &row.Values); err != nil {
			return nil, err
		}
		row.Identity = moduleRecoveryDependencyRowKey(row.Values, graph.Observation.PrimaryKeys[table])
		if row.Identity == "" {
			graph.markIncomplete("row_projection_too_large", table, "单行引用字段超过 64 KiB，未传输或输出其内容。")
			continue
		}
		observedKey := table + ":" + row.Identity
		if !graph.ObservedRows[observedKey] {
			if len(graph.ObservedRows) >= graph.Limits.Rows {
				graph.Exhausted = true
				graph.markIncomplete("row_inspection_resource_limit", table, "引用和保留边界的唯一行数达到上限，检查未完成。")
				break
			}
			graph.ObservedRows[observedKey] = true
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (graph *moduleRecoveryRowDependencyGraph) includeRowsInDependencyGraph(rows []*moduleRecoveryDependencyRow, table string, scoped bool) {
	if graph.Rows[table] == nil {
		graph.Rows[table] = map[string]*moduleRecoveryDependencyRow{}
		graph.Scoped[table] = map[string]bool{}
	}
	for _, row := range rows {
		if _, exists := graph.Rows[table][row.Identity]; !exists {
			graph.Rows[table][row.Identity] = row
			graph.Count++
		}
		if scoped {
			graph.Scoped[table][row.Identity] = true
		}
	}
}

func (graph *moduleRecoveryRowDependencyGraph) readRowsMatchingReferenceKeys(ctx context.Context, table string, columns []string, keys []map[string]json.RawMessage, maximum int) ([]*moduleRecoveryDependencyRow, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	for _, column := range columns {
		if !graph.Observation.Columns[table][column] {
			graph.markIncomplete("row_reference_column_missing", table, "依赖字段不存在，无法完成逐行比较。")
			return nil, nil
		}
	}
	data, err := json.Marshal(keys)
	if err != nil {
		return nil, err
	}
	var equalities []string
	for _, column := range columns {
		quoted := pq.QuoteIdentifier(column)
		equalities = append(equalities, "record."+quoted+"=requested."+quoted)
	}
	predicate := `EXISTS(SELECT 1 FROM jsonb_populate_recordset(NULL::public.` + pq.QuoteIdentifier(table) + `,$1::jsonb) requested WHERE ` + strings.Join(equalities, " AND ") + `)`
	return graph.readBoundedRelationshipRows(ctx, table, predicate, []any{string(data)}, maximum)
}
