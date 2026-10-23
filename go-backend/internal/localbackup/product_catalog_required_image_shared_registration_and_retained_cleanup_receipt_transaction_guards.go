package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
)

const productCatalogRequiredImageCleanupProtectionVersion = 1

// This binding records a read-only decision. No receipt is imported, cancelled,
// reset or consumed; the complete database binding also covers its original row.
type productCatalogRequiredImageCleanupProtection struct {
	Version                                 int    `json:"version"`
	RequiredImages                          int    `json:"required_images"`
	InspectedRegistrations                  int    `json:"inspected_registrations"`
	RetainedCleanupReceipts                 int    `json:"retained_cleanup_receipts"`
	RequiredPathsSHA256                     string `json:"required_paths_sha256"`
	RegistrationRowsSHA256                  string `json:"registration_rows_sha256"`
	CleanupRowsSHA256                       string `json:"cleanup_rows_sha256"`
	StandardWriteGuardVersion               int    `json:"standard_write_guard_version,omitempty"`
	MediaPlacementGuardVersion              int    `json:"media_placement_guard_version,omitempty"`
	RichTextMediaGuardVersion               int    `json:"rich_text_media_guard_version,omitempty"`
	TemplateAndCategoryRichTextGuardVersion int    `json:"template_and_category_rich_text_guard_version,omitempty"`
}

type productCatalogMediaObjectRegistration struct {
	ID, ParentID, Key, URL, DeletedAt, Visibility, Partition, Status, WriteLease, WriteLeaseExpiry string
	PublicBusinessReuseAllowed                                                                     bool
}

func (cfg Configuration) inspectProductCatalogRequiredImagesAndRetainedCleanupReceipts(ctx context.Context, tx *sql.Tx) (*productCatalogRequiredImageCleanupProtection, error) {
	// Include target-only products and their images, not just uploaded rows.
	snapshot, err := cfg.captureProductCatalogModuleSnapshot(ctx, tx)
	if err != nil {
		return nil, err
	}
	paths, err := cfg.collectProductCatalogPackagePathsWithoutReadingLiveFiles(snapshot)
	if err != nil {
		return nil, err
	}
	if len(paths) > cfg.MaxFiles {
		return nil, errors.New("product required image inspection exceeds file capacity")
	}
	keys := map[string]bool{}
	for filename := range paths {
		if strings.HasPrefix(filename, "media/uploads/") {
			key := strings.TrimPrefix(filename, "media/uploads/")
			if normalized, valid := storage.NormalizeObjectKey(key); !valid || normalized != key {
				return nil, errors.New("product required image has a noncanonical object key")
			}
			keys[key] = true
		}
	}
	assets, err := readProductCatalogRows(ctx, tx, "media_assets", "true")
	if err != nil {
		return nil, err
	}
	derivatives, err := readProductCatalogRows(ctx, tx, "media_asset_derivatives", "true")
	if err != nil {
		return nil, err
	}
	if len(assets)+len(derivatives) > productCatalogMaximumRows {
		return nil, errors.New("product shared media inspection exceeds registration capacity")
	}
	registrations := map[string]productCatalogMediaObjectRegistration{}
	parents := map[string]productCatalogMediaObjectRegistration{}
	derivativeIDs := map[string]productCatalogMediaObjectRegistration{}
	requiredAssetIDs := map[string]bool{}
	for _, table := range []struct {
		name string
		rows []json.RawMessage
	}{{"media_assets", assets}, {"media_asset_derivatives", derivatives}} {
		for _, row := range table.rows {
			field := func(column string) string { value, _ := productCatalogRowString(row, column); return value }
			registration := productCatalogMediaObjectRegistration{ID: field("id"), ParentID: field("media_asset_id"), Key: field("storage_key"), URL: field("url"), DeletedAt: field("deleted_at"), Visibility: field("visibility"), Partition: field("library_partition"), Status: field("status"), WriteLease: field("customer_service_derivative_write_token"), WriteLeaseExpiry: field("customer_service_derivative_write_expires_at")}
			if table.name == "media_assets" {
				registration.PublicBusinessReuseAllowed = media.CanMediaAssetBeReusedByPublicBusinessContent(&media.MediaAsset{
					StorageKey: registration.Key, URL: registration.URL, Status: registration.Status, Visibility: registration.Visibility,
					LibraryPartition: media.MediaAssetLibraryPartition(registration.Partition), UploaderActorType: media.MediaUploadActorType(field("uploader_actor_type")),
					UploadOrigin: media.MediaAssetUploadOrigin(field("upload_origin")), CustomerServiceMediaPurpose: field("customer_service_media_purpose"),
				})
			}
			_, urlKey := cfg.resolveLocalMediaReference(registration.URL)
			key := registration.Key
			if key == "" {
				key = urlKey
			}
			normalized, valid := storage.NormalizeObjectKey(key)
			// A malformed or legacy spelling must not hide a collision. Foreign
			// URLs never acquire local ownership through a filename suffix.
			if keys[key] || keys[urlKey] || valid && keys[normalized] {
				if !valid || key != normalized || registration.URL != "" && urlKey != key {
					return nil, errors.New("product shared media registration has conflicting object identity")
				}
				if registration.DeletedAt != "" || registration.WriteLease != "" || registration.WriteLeaseExpiry != "" {
					return nil, errors.New("product required image has a retired registration or active write lease")
				}
				if table.name == "media_assets" && (!registration.PublicBusinessReuseAllowed || registration.Partition == "customer_service") {
					return nil, errors.New("product required image registration is not eligible for public business reuse")
				}
				if _, exists := registrations[key]; exists {
					return nil, errors.New("product required image has multiple shared object registrations")
				}
				registrations[key] = registration
				if table.name == "media_assets" {
					requiredAssetIDs[registration.ID] = true
				} else {
					requiredAssetIDs[registration.ParentID] = true
				}
			}
			registration.Key = key
			if table.name == "media_assets" {
				parents[registration.ID] = registration
			} else {
				derivativeIDs[registration.ID] = registration
			}
		}
	}
	for _, derivative := range derivativeIDs {
		if !keys[derivative.Key] {
			continue
		}
		parent, exists := parents[derivative.ParentID]
		if !exists || parent.DeletedAt != "" || !parent.PublicBusinessReuseAllowed || parent.Partition == "customer_service" || parent.WriteLease != "" || parent.WriteLeaseExpiry != "" {
			return nil, errors.New("product required derivative has no active public parent registration")
		}
		if strings.HasPrefix(derivative.Key, "media-derivatives/") && !strings.HasPrefix(derivative.Key, "media-derivatives/"+derivative.ParentID+"/") {
			return nil, errors.New("product required derivative namespace disagrees with its parent registration")
		}
	}
	for key := range keys {
		prefix := storage.CustomerServiceConversationAttachmentDerivativesPrefix + "/"
		if strings.HasPrefix(key, prefix) {
			parentID, _, _ := strings.Cut(strings.TrimPrefix(key, prefix), "/")
			if canonicalModuleRecoveryDurableEventPositiveID(parentID) {
				requiredAssetIDs[parentID] = true
			}
		}
	}
	// Unknown cleanup types are blocked too, rather than silently omitted by a
	// fixed IN list when the business dispatcher acquires a new handler.
	events, err := readProductCatalogRows(ctx, tx, "outbox_events", "event_type ILIKE '%cleanup%'")
	if err != nil {
		return nil, err
	}
	if len(events) > 20000 {
		return nil, errors.New("product retained cleanup receipt inspection exceeds capacity")
	}
	for _, row := range events {
		if err := cfg.validateProductCatalogRetainedCleanupReceipt(row, keys, requiredAssetIDs, parents, derivativeIDs); err != nil {
			id, _ := productCatalogRowString(row, "id")
			return nil, fmt.Errorf("product cleanup receipt %s blocks restore: %w", id, err)
		}
	}
	orderedPaths := []string{}
	for filename := range paths {
		orderedPaths = append(orderedPaths, filename)
	}
	sort.Strings(orderedPaths)
	result := &productCatalogRequiredImageCleanupProtection{Version: productCatalogRequiredImageCleanupProtectionVersion, RequiredImages: len(paths), InspectedRegistrations: len(assets) + len(derivatives), RetainedCleanupReceipts: len(events)}
	result.StandardWriteGuardVersion = productCatalogStandardWriteGuardVersion
	result.MediaPlacementGuardVersion = productCatalogMediaPlacementGuardVersion
	result.RichTextMediaGuardVersion = productCatalogRichTextMediaGuardVersion
	result.TemplateAndCategoryRichTextGuardVersion = productCatalogTemplateAndCategoryRichTextGuardVersion
	result.RequiredPathsSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(orderedPaths)
	if err != nil {
		return nil, err
	}
	result.RegistrationRowsSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256([][]json.RawMessage{assets, derivatives})
	if err != nil {
		return nil, err
	}
	result.CleanupRowsSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(events)
	return result, err
}

func (cfg Configuration) validateProductCatalogRetainedCleanupReceipt(row json.RawMessage, requiredKeys, requiredAssetIDs map[string]bool, parents, derivatives map[string]productCatalogMediaObjectRegistration) error {
	field := func(column string) string { value, _ := productCatalogRowString(row, column); return value }
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(row, &fields); err != nil {
		return err
	}
	eventType, status := field("event_type"), field("status")
	if !isCustomerServiceRecoveryAuditedCleanupEvent(eventType) || field("event_key") == "" {
		return errors.New("cleanup event type or idempotency identity is unregistered")
	}
	leased := field("locked_at") != "" || field("locked_by") != "" || status == outbox.EventStatusProcessing
	execution := customerServiceRecoveryCleanupReceiptFutureExecution(eventType, status, field("deleted_at") != "", leased)
	if leased || status == outbox.EventStatusUnknown || execution == "unsupported_execution_state" {
		return errors.New("cleanup receipt has an execution lease or an unresolved execution state")
	}
	keys := []string{}
	overlaps := false
	switch eventType {
	case outbox.EventTypeObjectStorageCleanup:
		var payload outbox.ObjectStorageCleanupPayload
		if err := decodeStrictJSON(fields["payload"], &payload); err != nil {
			return errors.New("cleanup payload has unknown fields or invalid types")
		}
		if !customerServiceRecoveryObjectCleanupResourceContractRegistered(payload.ResourceType) || strings.TrimSpace(payload.ResourceID) != payload.ResourceID || payload.ResourceID == "" || len(payload.ObjectKeys) == 0 {
			return errors.New("cleanup resource contract or exact object keys are unavailable")
		}
		keys = payload.ObjectKeys
		switch payload.ResourceType {
		case "media_asset", "unregistered_media_derivative", "retired_media_derivative":
			if !canonicalModuleRecoveryDurableEventPositiveID(payload.ResourceID) || field("aggregate_type") != outbox.AggregateTypeMediaAsset || !canonicalModuleRecoveryDurableEventPositiveID(field("aggregate_id")) {
				return errors.New("cleanup media resource or aggregate identity is invalid")
			}
			if payload.ResourceType == "retired_media_derivative" {
				if len(keys) != 1 {
					return errors.New("retired derivative cleanup requires one exact key")
				}
				if derivative, exists := derivatives[payload.ResourceID]; exists {
					_, urlKey := cfg.resolveLocalMediaReference(derivative.URL)
					if derivative.ParentID != field("aggregate_id") || derivative.DeletedAt == "" || derivative.Key != keys[0] || derivative.URL != "" && urlKey != derivative.Key || parents[derivative.ParentID].ID == "" {
						return errors.New("retired derivative cleanup disagrees with its registration or parent")
					}
				}
			} else {
				if field("aggregate_id") != payload.ResourceID {
					return errors.New("cleanup media resource and aggregate disagree")
				}
				if payload.ResourceType == "media_asset" {
					overlaps = requiredAssetIDs[payload.ResourceID]
					// The whole-asset handler also visits registrations omitted
					// from an old receipt and its private derivative namespace.
					if parent, exists := parents[payload.ResourceID]; exists {
						overlaps = overlaps || requiredKeys[parent.Key]
					}
				} else {
					if len(keys) != 1 {
						return errors.New("failed derivative cleanup requires one exact key")
					}
					parent, exists := parents[payload.ResourceID]
					draftID, draft := storage.BusinessDraftMediaReceiptIDFromObjectKey(parent.Key)
					keyDraftID, keyDraft := storage.BusinessDraftMediaReceiptIDFromObjectKey(keys[0])
					validDraft := exists && draft && keyDraft && draftID == keyDraftID && parent.Key != keys[0]
					validPublic := strings.HasPrefix(keys[0], "media-derivatives/"+payload.ResourceID+"/")
					validPrivate := strings.HasPrefix(keys[0], storage.CustomerServiceConversationAttachmentDerivativesPrefix+"/"+payload.ResourceID+"/")
					if !validDraft && !validPublic && !validPrivate {
						return errors.New("failed derivative cleanup key disagrees with its parent namespace")
					}
				}
			}
		}
	case outbox.EventTypeCustomerServiceAvatarCleanup:
		var payload outbox.CustomerServiceAvatarCleanupPayload
		if err := decodeStrictJSON(fields["payload"], &payload); err != nil {
			return errors.New("avatar cleanup payload has unknown fields or invalid types")
		}
		root, key := cfg.resolveLocalMediaReference(payload.URL)
		if root != "uploads" || !strings.HasPrefix(key, "customer-service/avatars/") || field("aggregate_type") != outbox.AggregateTypeCustomerServiceAgentProfile || !canonicalModuleRecoveryDurableEventPositiveID(field("aggregate_id")) {
			return errors.New("avatar cleanup has no trusted profile and object identity")
		}
		keys = append(keys, key)
	case outbox.EventTypeCustomerServiceRetentionCleanup:
		var payload outbox.CustomerServiceRetentionCleanupPayload
		if err := decodeStrictJSON(fields["payload"], &payload); err != nil || payload.TicketID == 0 || fmt.Sprint(payload.TicketID) != field("aggregate_id") || field("aggregate_type") != outbox.AggregateTypeCustomerServiceConversation || payload.AttachmentReferences == nil {
			return errors.New("retention cleanup has no trusted conversation and reference list")
		}
		for _, reference := range payload.AttachmentReferences {
			root, key := cfg.resolveLocalMediaReference(reference)
			if root == "" || key == "" {
				return errors.New("retention cleanup reference cannot be resolved locally")
			}
			if root == "uploads" {
				keys = append(keys, key)
			}
		}
	}
	seen := map[string]bool{}
	for _, key := range keys {
		normalized, valid := storage.NormalizeObjectKey(key)
		if !valid || normalized != key || seen[key] {
			return errors.New("cleanup object key is unsafe, noncanonical or duplicated")
		}
		seen[key] = true
		overlaps = overlaps || requiredKeys[key]
	}
	if overlaps && execution != "suppressed_original_receipt" {
		return errors.New("retained cleanup can delete a required product image; reconcile the original receipt before restoring")
	}
	return nil
}

func (cfg Configuration) inspectProductCatalogRequiredImageCleanupProtectionInDatabase(ctx context.Context, name string) (*productCatalogRequiredImageCleanupProtection, error) {
	db, err := cfg.connectDatabase(ctx, name)
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(db.Close)
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	return cfg.inspectProductCatalogRequiredImagesAndRetainedCleanupReceipts(ctx, tx)
}
