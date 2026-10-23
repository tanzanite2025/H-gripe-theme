package localbackup

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"commerce-platform/internal/domain/ticket"
)

// Reuse the rendering contract: navigation URLs and arbitrary nested strings
// are not media. Bounded metadata is parsed privately and never serialized in
// a report, unlike full message content which is never selected at all.
func (cfg Configuration) customerServiceStructuredMetadataLocalFileReferences(graph *moduleRecoveryRowDependencyGraph, table string, row *moduleRecoveryDependencyRow, includeLegacyAttachments bool) []moduleRecoveryFileReference {
	messageType := strings.ToLower(strings.TrimSpace(row.referenceColumnText("_structured_message_type")))
	metadata := row.referenceColumnText("_structured_metadata")
	known := false
	for _, registered := range []string{"", "text", "system", "link", "product", "order", "faq", "config_confirm", "wheelset_selection_request", "image", "video"} {
		known = known || messageType == registered
	}
	if !known {
		graph.markIncomplete("structured_message_type_reference_contract_missing", table, "消息类型没有内容及媒体引用恢复契约，不能猜测其正文或元数据语义。")
		return nil
	}
	document, err := ticket.ParseStructuredMessageMetadataKnownMediaFields(messageType, metadata)
	if err != nil {
		graph.markIncomplete("structured_message_metadata_media_fields_invalid", table, "结构化消息媒体字段无法解析；未输出元数据或扩大恢复范围。")
		return nil
	}
	var references []string
	references = append(references, cfg.collectCustomerServiceStructuredBusinessSnapshotAndOrderLineMediaReferences(graph, table, row)...)
	for _, field := range document.Fields {
		references = append(references, field.URL)
	}
	if includeLegacyAttachments {
		raw := strings.TrimSpace(row.referenceColumnText("attachments"))
		if raw != "" && raw != "null" {
			var attachments []string
			if json.Unmarshal([]byte(raw), &attachments) != nil {
				graph.markIncomplete("preserved_message_attachment_references_invalid", table, "保留消息或自动回复的附件字段无法解析，不能确认清理对象范围。")
			} else {
				references = append(references, attachments...)
			}
		}
	}
	var files []moduleRecoveryFileReference
	seen := map[string]bool{}
	for _, reference := range references {
		root, key := cfg.resolveLocalMediaReference(reference)
		if key == "" {
			graph.markIncomplete("structured_message_media_reference_unresolved", table, "结构化媒体引用不能映射到商城本地文件，须核对历史对象。")
			continue
		}
		file := moduleRecoveryFileReference{Root: root, Key: key, Table: table, RowIdentity: moduleRecoveryDependencyRowPublicIdentity(row, graph.Observation.PrimaryKeys[table])}
		if !seen[file.packagePath()] {
			files = append(files, file)
			seen[file.packagePath()] = true
		}
	}
	return files
}

func (cfg Configuration) includeCustomerServiceStructuredMetadataReferences(ctx context.Context, graph *moduleRecoveryRowDependencyGraph) error {
	for _, table := range []string{"ticket_messages", "ticket_auto_replies"} {
		if !graph.Observation.Columns[table]["message_type"] || !graph.Observation.Columns[table]["metadata"] {
			graph.markIncomplete("structured_message_reference_schema_missing", table, "缺少消息类型或结构化元数据字段，不能完成媒体及清理引用检查。")
		}
	}
	var messages []*moduleRecoveryDependencyRow
	for identity, message := range graph.Rows["ticket_messages"] {
		if graph.Scoped["ticket_messages"][identity] {
			messages = append(messages, message)
		}
	}
	sort.Slice(messages, func(i, j int) bool { return messages[i].Identity < messages[j].Identity })
	for _, message := range messages {
		files := cfg.customerServiceStructuredMetadataLocalFileReferences(graph, "ticket_messages", message, false)
		if len(graph.LegacyFiles)+len(files) > graph.Limits.Rows {
			graph.markIncomplete("structured_message_media_reference_resource_limit", "ticket_messages", "结构化媒体引用超过逐侧文件上限。")
			break
		}
		graph.LegacyFiles = append(graph.LegacyFiles, files...)
	}
	// Auto replies (including disabled rules) and messages of other categories
	// remain target data. Their references protect cleanup without joining the
	// graph of candidate replacement rows.
	for _, table := range []string{"ticket_auto_replies", "ticket_messages"} {
		if !graph.Observation.Columns[table]["metadata"] || !graph.Observation.Columns[table]["message_type"] {
			continue
		}
		predicate := "TRUE"
		var arguments []any
		if table == "ticket_messages" {
			predicate = `NOT EXISTS(SELECT 1 FROM public.tickets ticket WHERE ticket.id=record.ticket_id AND ticket.category='customer_service')`
			if !graph.Observation.Columns["tickets"]["category"] || !graph.Observation.Columns[table]["ticket_id"] {
				continue
			}
			if graph.SelectedCustomerServiceRows != nil {
				var identities []json.RawMessage
				for identity := range graph.Scoped[table] {
					identities = append(identities, json.RawMessage(identity))
				}
				if identities == nil {
					identities = []json.RawMessage{}
				}
				encoded, err := json.Marshal(identities)
				if err != nil {
					return err
				}
				predicate = `NOT EXISTS(SELECT 1 FROM jsonb_array_elements($1::jsonb) scoped(key) WHERE to_jsonb(record) @> scoped.key)`
				arguments = []any{string(encoded)}
			}
		}
		rows, err := graph.readBoundedRelationshipRows(ctx, table, predicate, arguments, graph.Limits.Rows)
		if err != nil {
			return err
		}
		for _, row := range rows {
			files := cfg.customerServiceStructuredMetadataLocalFileReferences(graph, table, row, true)
			if len(graph.RetainedStructuredFiles)+len(files) > graph.Limits.Rows {
				graph.markIncomplete("structured_message_media_reference_resource_limit", table, "保留消息与自动回复的媒体引用超过上限。")
				break
			}
			graph.RetainedStructuredFiles = append(graph.RetainedStructuredFiles, files...)
		}
	}
	keysByPath := map[string]map[string]json.RawMessage{}
	for _, file := range graph.RetainedStructuredFiles {
		encoded, _ := json.Marshal(file.Key)
		keysByPath[file.packagePath()] = map[string]json.RawMessage{"storage_key": encoded}
	}
	var paths []string
	for path := range keysByPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for start := 0; start < len(paths) && !graph.Exhausted; start += 200 {
		var keys []map[string]json.RawMessage
		for _, path := range paths[start:min(start+200, len(paths))] {
			keys = append(keys, keysByPath[path])
		}
		assets, err := graph.readRowsMatchingReferenceKeys(ctx, "media_assets", []string{"storage_key"}, keys, graph.Limits.Rows)
		if err != nil {
			return err
		}
		for _, asset := range assets {
			root, key := cfg.customerServiceMediaAssetLocalIdentity(asset)
			if keysByPath["media/"+root+"/"+key] != nil {
				graph.RetainedStructuredAssetIDs[asset.referenceColumnText("id")] = true
			}
		}
		if !graph.Observation.Columns["media_asset_derivatives"]["storage_key"] {
			continue
		}
		derivatives, err := graph.readRowsMatchingReferenceKeys(ctx, "media_asset_derivatives", []string{"storage_key"}, keys, graph.Limits.Rows)
		if err != nil {
			return err
		}
		for derivativeStart := 0; derivativeStart < len(derivatives) && !graph.Exhausted; derivativeStart += 200 {
			batch := derivatives[derivativeStart:min(derivativeStart+200, len(derivatives))]
			var parentKeys []map[string]json.RawMessage
			for _, derivative := range batch {
				parentKeys = append(parentKeys, map[string]json.RawMessage{"id": derivative.Values["media_asset_id"]})
			}
			parents, err := graph.readRowsMatchingReferenceKeys(ctx, "media_assets", []string{"id"}, parentKeys, graph.Limits.Rows)
			if err != nil {
				return err
			}
			parentsByID := map[string]*moduleRecoveryDependencyRow{}
			for _, parent := range parents {
				parentsByID[parent.referenceColumnText("id")] = parent
			}
			for _, derivative := range batch {
				parent := parentsByID[derivative.referenceColumnText("media_asset_id")]
				if parent == nil {
					graph.markIncomplete("preserved_structured_media_parent_missing", "media_asset_derivatives", "保留消息引用的派生对象缺少父媒体，无法确认清理对象范围。")
					continue
				}
				root, _ := cfg.customerServiceMediaAssetLocalIdentity(parent)
				key := derivative.referenceColumnText("storage_key")
				if keysByPath["media/"+root+"/"+key] != nil {
					graph.RetainedStructuredAssetIDs[parent.referenceColumnText("id")] = true
				}
			}
		}
	}
	return nil
}
