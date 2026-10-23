package localbackup

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
)

type CustomerServiceConversationNaturalKeyAndSnapshotReferenceInspection struct {
	Version                           int                                                       `json:"version"`
	Complete                          bool                                                      `json:"complete"`
	NaturalKeyComparisons             int                                                       `json:"natural_key_comparisons"`
	SnapshotReferenceChecks           int                                                       `json:"snapshot_reference_checks"`
	SharedBusinessRowsPreserved       int                                                       `json:"shared_business_rows_preserved"`
	ColumnContracts                   []CustomerServiceCandidateColumnContractCoverage          `json:"column_contracts,omitempty"`
	RetainedMediaOwnerColumnContracts []CustomerServiceRetainedMediaOwnerColumnContractCoverage `json:"retained_media_owner_column_contracts,omitempty"`
}

type customerServiceStructuredBusinessSnapshot struct {
	OwnerTable                                                                                         string
	Owner                                                                                              *moduleRecoveryDependencyRow
	MessageType, ProductID, VariantID, OrderNumber, FAQID, FAQPageID, FAQLocale, FlowSlug, FlowVersion string
	Lines                                                                                              []customerServiceOrderSnapshotLineReference
}

type customerServiceOrderSnapshotLineReference struct{ ID, ProductID, VariantID string }

func customerServiceSnapshotJSONValueAtKnownPath(object map[string]json.RawMessage, path string) (json.RawMessage, error) {
	parts := strings.SplitN(path, ".", 2)
	raw := object[parts[0]]
	if len(raw) == 0 || string(raw) == "null" || len(parts) == 1 {
		return raw, nil
	}
	var child map[string]json.RawMessage
	if json.Unmarshal(raw, &child) != nil || child == nil {
		return nil, errors.New("snapshot container must be an object")
	}
	return customerServiceSnapshotJSONValueAtKnownPath(child, parts[1])
}

// IDs never pass through float64. Strings and numbers must be canonical decimal
// integers; optional null/zero IDs represent display-only historical snapshots.
func customerServiceSnapshotCanonicalPositiveID(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	value := string(raw)
	if raw[0] == '"' {
		if json.Unmarshal(raw, &value) != nil {
			return "", errors.New("invalid snapshot id")
		}
	}
	if value == "" || value == "0" {
		return "", nil
	}
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil || number <= 0 || strconv.FormatInt(number, 10) != value {
		return "", errors.New("snapshot id must be a canonical positive integer")
	}
	return value, nil
}

func (cfg Configuration) collectCustomerServiceStructuredBusinessSnapshotAndOrderLineMediaReferences(graph *moduleRecoveryRowDependencyGraph, table string, row *moduleRecoveryDependencyRow) []string {
	messageType := strings.ToLower(strings.TrimSpace(row.referenceColumnText("_structured_message_type")))
	switch messageType {
	case "product", "order", "faq", "config_confirm", "wheelset_selection_request":
	default:
		return nil
	}
	metadata := strings.TrimSpace(row.referenceColumnText("_structured_metadata"))
	if metadata == "" || metadata == "null" {
		return nil
	}
	var object map[string]json.RawMessage
	if json.Unmarshal([]byte(metadata), &object) != nil || object == nil {
		graph.markIncomplete("business_snapshot_metadata_invalid", table, "业务卡片元数据不是有效对象，引用检查未完成。")
		return nil
	}
	var parseError error
	readID := func(path string) string {
		raw, err := customerServiceSnapshotJSONValueAtKnownPath(object, path)
		if err != nil {
			parseError = err
			return ""
		}
		value, err := customerServiceSnapshotCanonicalPositiveID(raw)
		if err != nil {
			parseError = err
		}
		return value
	}
	readText := func(path string) string {
		raw, err := customerServiceSnapshotJSONValueAtKnownPath(object, path)
		if err != nil {
			parseError = err
			return ""
		}
		if len(raw) == 0 || string(raw) == "null" {
			return ""
		}
		var value string
		if json.Unmarshal(raw, &value) != nil {
			parseError = errors.New("snapshot identity must be a string")
		}
		return value
	}
	snapshot := customerServiceStructuredBusinessSnapshot{OwnerTable: table, Owner: row, MessageType: messageType}
	var media []string
	switch messageType {
	case "product":
		snapshot.ProductID = readID("product_id")
		snapshot.VariantID = readID("variant_id")
	case "config_confirm":
		snapshot.ProductID = readID("product.id")
		snapshot.VariantID = readID("product.variant_id")
		selectedVariant := readID("selections.variant_id")
		if snapshot.VariantID != "" && selectedVariant != "" && snapshot.VariantID != selectedVariant {
			parseError = errors.New("configuration variant aliases disagree")
		}
		if snapshot.VariantID == "" {
			snapshot.VariantID = selectedVariant
		}
	case "order":
		snapshot.OrderNumber = readText("order_number")
		if raw := object["items"]; len(raw) > 0 && string(raw) != "null" {
			var lines []map[string]json.RawMessage
			if json.Unmarshal(raw, &lines) != nil {
				parseError = errors.New("order snapshot items must be objects")
			} else {
				for _, line := range lines {
					var entry customerServiceOrderSnapshotLineReference
					for column, destination := range map[string]*string{"id": &entry.ID, "product_id": &entry.ProductID, "variant_id": &entry.VariantID} {
						value, err := customerServiceSnapshotCanonicalPositiveID(line[column])
						if err != nil {
							parseError = err
						}
						*destination = value
					}
					snapshot.Lines = append(snapshot.Lines, entry)
					if raw := line["thumbnail"]; len(raw) > 0 && string(raw) != "null" {
						var reference string
						if json.Unmarshal(raw, &reference) != nil {
							parseError = errors.New("order line thumbnail must be a string")
						} else if strings.TrimSpace(reference) != "" {
							media = append(media, reference)
						}
					}
				}
			}
		}
	case "faq":
		snapshot.FAQID = readID("faq_id")
		snapshot.FAQPageID = readText("page_id")
		snapshot.FAQLocale = readText("locale")
	case "wheelset_selection_request":
		version := readID("version")
		if raw := object["version"]; len(raw) > 0 && string(raw) != "null" && version != "1" {
			parseError = errors.New("unregistered wheelset snapshot version")
		}
		if kind := readText("kind"); kind != "" && kind != "wheelset_selection_request" {
			parseError = errors.New("unregistered wheelset snapshot kind")
		}
		snapshot.FlowSlug = readText("flow_slug")
		snapshot.FlowVersion = readID("flow_version")
		if (snapshot.FlowSlug == "") != (snapshot.FlowVersion == "") {
			parseError = errors.New("wheelset flow identity is incomplete")
		}
		// Historical payloads can also contain a product card; current producers
		// use flow_slug + version_number, not a flow-version database ID.
		snapshot.ProductID = readID("product.id")
		snapshot.VariantID = readID("product.variant_id")
	}
	if parseError != nil {
		graph.markIncomplete("business_snapshot_reference_shape_unsupported", table, "业务卡片引用的类型、版本或关联身份不受支持；保留原始内容，不能猜测关联对象。")
		graph.blockCustomerServiceCandidateRowForContractIssue(table, row, "business_snapshot_candidate_reference_invalid", "该行的业务卡片引用尚未通过恢复契约检查。")
		return nil
	}
	if len(graph.BusinessSnapshots)+len(snapshot.Lines)+1 > graph.Limits.Rows {
		graph.markIncomplete("business_snapshot_reference_resource_limit", table, "业务快照引用超过逐侧上限，检查未完成。")
		return nil
	}
	graph.BusinessSnapshotReferenceCount += len(snapshot.Lines) + 1
	if graph.BusinessSnapshotReferenceCount > graph.Limits.Rows {
		graph.markIncomplete("business_snapshot_reference_resource_limit", table, "业务快照引用总量超过逐侧上限，检查未完成。")
		return nil
	}
	graph.BusinessSnapshots = append(graph.BusinessSnapshots, snapshot)
	return media
}

type customerServiceBusinessSnapshotLookupCache struct {
	Rows map[string][]*moduleRecoveryDependencyRow
}

func (cache *customerServiceBusinessSnapshotLookupCache) lookupCustomerServiceSnapshotBusinessRows(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, table string, key map[string]string, requiredColumns []string) ([]*moduleRecoveryDependencyRow, error) {
	if graph == nil {
		return nil, nil
	}
	var columns []string
	keys := map[string]json.RawMessage{}
	for column, value := range key {
		columns = append(columns, column)
		encoded, _ := json.Marshal(value)
		keys[column] = encoded
	}
	sort.Strings(columns)
	for _, column := range append(append([]string{}, columns...), requiredColumns...) {
		if column == "_snapshot_configuration_sha256" {
			column = "config"
		}
		if !graph.Observation.Columns[table][column] {
			graph.markIncomplete("business_snapshot_reference_schema_missing", table, "缺少业务快照引用所需字段，不能核对共享对象身份。")
			return nil, nil
		}
	}
	encoded, _ := json.Marshal(key)
	cacheKey := graph.Side + ":" + table + ":" + string(encoded)
	if rows, exists := cache.Rows[cacheKey]; exists {
		return rows, nil
	}
	rows, err := graph.readRowsMatchingReferenceKeys(ctx, table, columns, []map[string]json.RawMessage{keys}, 2)
	if err != nil {
		return nil, err
	}
	cache.Rows[cacheKey] = rows
	return rows, nil
}

func (cache *customerServiceBusinessSnapshotLookupCache) inspectCustomerServiceSnapshotSharedReference(ctx context.Context, graph, other *moduleRecoveryRowDependencyGraph, snapshot customerServiceStructuredBusinessSnapshot, table string, key map[string]string, identityColumns []string, expected map[string]string) (*moduleRecoveryDependencyRow, error) {
	report := graph.Report.ConversationReferenceChecks
	report.SnapshotReferenceChecks++
	required := append([]string{"id"}, identityColumns...)
	for column := range expected {
		required = append(required, column)
	}
	rows, err := cache.lookupCustomerServiceSnapshotBusinessRows(ctx, graph, table, key, required)
	if err != nil {
		return nil, err
	}
	var oppositeRows []*moduleRecoveryDependencyRow
	if other != nil {
		oppositeRows, err = cache.lookupCustomerServiceSnapshotBusinessRows(ctx, other, table, key, required)
		if err != nil {
			return nil, err
		}
	}
	block := func(code, message string) {
		graph.blockCustomerServiceCandidateRowForContractIssue(snapshot.OwnerTable, snapshot.Owner, code, message)
	}
	if len(rows) > 1 || len(oppositeRows) > 1 {
		block("business_snapshot_natural_identity_ambiguous", "业务卡片的共享自然身份匹配多行，不能选择其中一行或自动改号。")
		return nil, nil
	}
	if len(rows) == 0 {
		if len(oppositeRows) > 0 {
			block("business_snapshot_source_identity_unverified", "包内共享对象已不存在而目标存在同键对象，无法证明属于同一历史身份。")
		} else {
			graph.Report.recordIssue("historical_business_snapshot_reference_unavailable", "review", graph.Side, snapshot.OwnerTable, moduleRecoveryDependencyRowPublicIdentity(snapshot.Owner, graph.Observation.PrimaryKeys[snapshot.OwnerTable]), "历史卡片所引用对象已不可用，保留消息中的原始展示快照；不自动重建共享业务对象。")
		}
		return nil, nil
	}
	current := rows[0]
	for column, value := range expected {
		if value != "" && current.referenceColumnText(column) != value {
			block("business_snapshot_relationship_identity_conflict", "业务卡片中的订单行、商品规格或客户归属与共享对象不一致，必须先核对身份。")
		}
	}
	if other != nil {
		if len(oppositeRows) == 0 {
			block("business_snapshot_target_dependency_missing", "目标缺少包内卡片所需共享业务对象，不能随客服恢复补入订单或商品。")
		} else {
			opposite := oppositeRows[0]
			for _, column := range append([]string{"id"}, identityColumns...) {
				if current.referenceColumnText(column) != opposite.referenceColumnText(column) {
					block("business_snapshot_shared_identity_changed", "两侧卡片引用的共享业务身份不同；保留目标，禁止将旧会话关联到其他客户、订单或商品。 ")
					break
				}
			}
			if current.Digest != opposite.Digest {
				graph.Report.recordIssue("business_snapshot_current_content_preserved", "review", graph.Side, table, "", "共享业务行内容已有变化；历史卡片保留原文，当前价格、库存、状态及其他内容保持目标值。")
			}
		}
	}
	report.SharedBusinessRowsPreserved++
	return current, nil
}

func inspectCustomerServiceStructuredBusinessSnapshotDependencies(ctx context.Context, source, target *moduleRecoveryRowDependencyGraph) error {
	cache := &customerServiceBusinessSnapshotLookupCache{Rows: map[string][]*moduleRecoveryDependencyRow{}}
	for _, graph := range []*moduleRecoveryRowDependencyGraph{source, target} {
		if graph == nil {
			continue
		}
		var other *moduleRecoveryRowDependencyGraph
		if graph == source {
			other = target
		}
		for _, snapshot := range graph.BusinessSnapshots {
			if err := ctx.Err(); err != nil {
				return err
			}
			if graph.Exhausted {
				break
			}
			inspect := func(table string, key map[string]string, identities []string, expected map[string]string) (*moduleRecoveryDependencyRow, error) {
				return cache.inspectCustomerServiceSnapshotSharedReference(ctx, graph, other, snapshot, table, key, identities, expected)
			}
			inspectProductAndVariant := func(productID, variantID string) error {
				if productID != "" {
					if _, err := inspect("products", map[string]string{"id": productID}, []string{"sku", "slug"}, nil); err != nil {
						return err
					}
				}
				if variantID != "" {
					if _, err := inspect("product_variants", map[string]string{"id": variantID}, []string{"product_id", "sku"}, map[string]string{"product_id": productID}); err != nil {
						return err
					}
				}
				return nil
			}
			if err := inspectProductAndVariant(snapshot.ProductID, snapshot.VariantID); err != nil {
				return err
			}
			if snapshot.OrderNumber != "" {
				var customer string
				if snapshot.OwnerTable == "ticket_messages" {
					tickets, err := cache.lookupCustomerServiceSnapshotBusinessRows(ctx, graph, "tickets", map[string]string{"id": snapshot.Owner.referenceColumnText("ticket_id")}, []string{"customer_user_id", "user_id", "visitor_session_hash"})
					if err != nil {
						return err
					}
					if len(tickets) == 1 {
						customer = tickets[0].referenceColumnText("customer_user_id")
						if customer == "" || customer == "0" {
							customer = ""
							if tickets[0].referenceColumnText("visitor_session_hash") == "" {
								customer = tickets[0].referenceColumnText("user_id")
							}
						}
					}
				}
				if customer == "" || customer == "0" {
					graph.blockCustomerServiceCandidateRowForContractIssue(snapshot.OwnerTable, snapshot.Owner, "business_snapshot_order_customer_unverified", "订单卡片缺少可核对的会话客户身份，不能推断订单访问或恢复归属。")
				}
				order, err := inspect("orders", map[string]string{"order_number": snapshot.OrderNumber}, []string{"order_number", "user_id"}, map[string]string{"user_id": customer})
				if err != nil {
					return err
				}
				for _, line := range snapshot.Lines {
					if line.ID != "" {
						expected := map[string]string{"product_id": line.ProductID, "variant_id": line.VariantID}
						if order != nil {
							expected["order_id"] = order.referenceColumnText("id")
						}
						if _, err := inspect("order_items", map[string]string{"id": line.ID}, []string{"order_id", "product_id", "variant_id"}, expected); err != nil {
							return err
						}
					}
					if err := inspectProductAndVariant(line.ProductID, line.VariantID); err != nil {
						return err
					}
				}
			} else if len(snapshot.Lines) > 0 {
				graph.blockCustomerServiceCandidateRowForContractIssue(snapshot.OwnerTable, snapshot.Owner, "business_snapshot_order_identity_missing", "订单卡片含明细但缺少订单号，无法确认明细属于哪笔订单。")
			}
			if snapshot.FAQID != "" {
				if _, err := inspect("faqs", map[string]string{"id": snapshot.FAQID}, []string{"page_id", "locale"}, map[string]string{"page_id": snapshot.FAQPageID, "locale": snapshot.FAQLocale}); err != nil {
					return err
				}
			}
			if snapshot.FlowSlug != "" {
				flow, err := inspect("selection_assistant_flows", map[string]string{"slug": snapshot.FlowSlug}, []string{"slug"}, nil)
				if err != nil {
					return err
				}
				if flow != nil {
					if _, err := inspect("selection_assistant_flow_versions", map[string]string{"flow_id": flow.referenceColumnText("id"), "version_number": snapshot.FlowVersion}, []string{"flow_id", "version_number", "_snapshot_configuration_sha256"}, nil); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
