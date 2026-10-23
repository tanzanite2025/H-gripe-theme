package localbackup

import (
	"context"
	"encoding/json"
	"strconv"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
)

// Whole-asset cleanup owns a media asset ID. Failed derivative cleanup owns an
// exact key under a parent ID. Retired derivative cleanup instead names a
// derivative registration, while its aggregate names the parent media asset.
// Neither derivative receipt acquires the parent's other objects by ID alone.
func (cfg Configuration) inspectCustomerServiceMediaCleanupResourceAndRetiredDerivativeRegistrationIdentity(ctx context.Context, graph *moduleRecoveryRowDependencyGraph, event *moduleRecoveryDependencyRow, scope *customerServiceDurableEventScope, resourceType, resourceID string, cleanupKeys []string, related bool) (bool, error) {
	if resourceType != "media_asset" && resourceType != "unregistered_media_derivative" && resourceType != "retired_media_derivative" {
		return false, nil
	}
	overlaps := false
	if resourceType == "media_asset" {
		// The worker parses decimal IDs. Detect a retained target even when a
		// noncanonical receipt uses a leading zero, then reject that spelling.
		if parsedID, err := strconv.ParseUint(resourceID, 10, 64); err == nil {
			overlaps = scope.MediaAssets[strconv.FormatUint(parsedID, 10)]
		}
	}
	if !related && !overlaps {
		return false, nil
	}
	aggregateID := event.referenceColumnText("aggregate_id")
	if !canonicalModuleRecoveryDurableEventPositiveID(resourceID) || event.referenceColumnText("aggregate_type") != outbox.AggregateTypeMediaAsset || !canonicalModuleRecoveryDurableEventPositiveID(aggregateID) || resourceType != "retired_media_derivative" && aggregateID != resourceID {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_cleanup_resource_identity_conflict", "blocking", "媒体清理的资源身份或父媒体聚合身份不符合该资源类型的契约。", false)
		return overlaps, nil
	}
	if resourceType == "media_asset" {
		return overlaps, nil
	}
	if len(cleanupKeys) != 1 {
		recordModuleRecoveryDurableEventIssue(graph, event, "durable_derivative_cleanup_exact_key_required", "blocking", "派生文件清理必须保留唯一的准确对象键，不能解释为空范围或整组父媒体对象。", true)
		return false, nil
	}
	if resourceType != "retired_media_derivative" {
		return false, nil
	}
	for _, column := range []string{"id", "media_asset_id", "storage_key", "url", "deleted_at"} {
		if !graph.Observation.Columns["media_asset_derivatives"][column] {
			graph.markIncomplete("retired_derivative_cleanup_registration_schema_missing", "media_asset_derivatives", "退役派生文件清理缺少登记身份、父媒体、准确路径或退役状态字段，不能完成收据核对。")
			return false, nil
		}
	}
	registrations, err := graph.readRowsMatchingReferenceKeys(ctx, "media_asset_derivatives", []string{"id"}, []map[string]json.RawMessage{{"id": json.RawMessage(resourceID)}}, 1)
	if err != nil || graph.Exhausted {
		return false, err
	}
	// The actual handler returns without deleting when whole-asset cleanup has
	// already removed this registration. Do not reinterpret its ID as an asset.
	if len(registrations) == 0 {
		return false, nil
	}
	registration := registrations[0]
	parentID := registration.referenceColumnText("media_asset_id")
	if parentID != aggregateID || registration.referenceColumnText("deleted_at") == "" {
		recordModuleRecoveryDurableEventIssue(graph, event, "retired_derivative_cleanup_registration_identity_conflict", "blocking", "退役派生文件收据与实际父媒体或退役状态不一致，须核对原始登记；未修改收据。", false)
	}
	parents, err := graph.readRowsMatchingReferenceKeys(ctx, "media_assets", []string{"id"}, []map[string]json.RawMessage{{"id": registration.Values["media_asset_id"]}}, 1)
	if err != nil || graph.Exhausted {
		return false, err
	}
	if len(parents) == 0 {
		graph.markIncomplete("retired_derivative_cleanup_registered_parent_missing", "media_asset_derivatives", "退役派生登记仍存在但父媒体缺失，实际清理处理器要求核对归属，不能判为安全保留。")
	}
	key := registration.referenceColumnText("storage_key")
	url := registration.referenceColumnText("url")
	_, urlKey := cfg.resolveLocalMediaReference(url)
	if key == "" {
		key = urlKey
	}
	normalized, valid := storage.NormalizeObjectKey(key)
	if !valid || normalized != key || key != cleanupKeys[0] || url != "" && (urlKey == "" || urlKey != key) {
		recordModuleRecoveryDurableEventIssue(graph, event, "retired_derivative_cleanup_registered_object_identity_conflict", "blocking", "退役派生文件收据与登记键或本地 URL 的准确对象身份不一致，不能只凭文件名或父媒体 ID 判定。", true)
	}
	return false, nil
}
