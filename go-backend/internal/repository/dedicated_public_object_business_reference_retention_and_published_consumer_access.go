package repository

import (
	"fmt"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

// These are read projections of authoritative business rows, not another asset
// registration or reference ledger. Retention and permission are independent.
type DedicatedPublicObjectBusinessReferenceState struct {
	RetainedReferences   []media.AssetReference
	HasPublishedConsumer bool
}

func (r *MediaRepository) FindDedicatedPublicObjectBusinessReferenceState(key string) (DedicatedPublicObjectBusinessReferenceState, error) {
	return r.findDedicatedPublicObjectBusinessReferenceStateForRetentionOrPublicAccess(key, false)
}

func (r *MediaRepository) findDedicatedPublicObjectBusinessReferenceStateForRetentionOrPublicAccess(key string, publishedConsumersOnly bool) (DedicatedPublicObjectBusinessReferenceState, error) {
	state := DedicatedPublicObjectBusinessReferenceState{}
	key, valid := storage.NormalizeObjectKey(key)
	if !valid || dedicatedSiteIdentityOrVisualShowcaseNamespaceForObjectKey(key) == "" {
		return state, fmt.Errorf("invalid dedicated public object key")
	}
	query := mediaAssetReferenceQuery{URLs: []string{"/uploads/" + key, "uploads/" + key}, ObjectKeys: []string{key}, IncludeRetainedDeletedPlacements: true}
	scanners := []mediaAssetReferenceScanner{
		r.productCategoryImageReferences, r.productBrandLogoReferences, r.productMediaReferences, r.productRichTextMediaAttributeReferences,
		r.productAndTemplateOptionSwatchReferences, r.galleryReferences, r.faqReferences,
		r.postReferences, r.websiteProfileSettingMediaReferences, r.workbenchFeedImageReferences,
	}
	if !publishedConsumersOnly {
		scanners = append(scanners, r.siteIdentitySettingMediaReferences, r.dedicatedPublicObjectLegacyCommunityAndPrivateAttachmentReferences, r.orderEvidenceImmutableAttachmentAndSnapshotReferences)
	}
	for _, scan := range scanners {
		items, err := scan(query)
		if err != nil {
			return state, err
		}
		state.RetainedReferences = append(state.RetainedReferences, items...)
		if publishedConsumersOnly {
			allowed, err := r.dedicatedPublicObjectReferencesIncludePublishedConsumer(items)
			if err != nil || allowed {
				state.HasPublishedConsumer = allowed
				return state, err
			}
		}
	}
	// An ID-only placement must resolve through the URL of its registration in
	// this owner's provider. A matching stored key alone is only a collision:
	// Logo may live in a different physical store from the shared library.
	if r.hasTable("media_assets") {
		for afterID := uint(0); ; {
			var rows []media.MediaAsset
			if err := r.db.Unscoped().Select("id", "url", "storage_key").Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
				return state, err
			}
			for _, row := range rows {
				trustedOriginal := r.matchesExactRegisteredOrTrustedMediaReference(query, row.URL)
				if !publishedConsumersOnly && (trustedOriginal || registeredMediaReferenceQueryContainsExactObjectKey(query, row.StorageKey)) {
					state.RetainedReferences = append(state.RetainedReferences, newMediaReference(media.ReferenceCategoryContent, "dedicated_object_registration_conflict", row.ID, 0, "通用登记与专用对象身份冲突，需核对存储", "storage_key, url"))
				}
				if !trustedOriginal {
					continue
				}
				idQuery := mediaAssetReferenceQuery{AssetID: row.ID, IncludeRetainedDeletedPlacements: true}
				idScanners := []mediaAssetReferenceScanner{r.productCategoryImageReferences, r.productMediaReferences, r.productAndTemplateOptionSwatchReferences, r.galleryReferences}
				if !publishedConsumersOnly {
					idScanners = append(idScanners, r.customerServiceMessageMediaAssetBindingReferences, r.customerServicePendingUploadSessionAssetReferences)
				}
				for _, scan := range idScanners {
					items, err := scan(idQuery)
					if err != nil {
						return state, err
					}
					state.RetainedReferences = append(state.RetainedReferences, items...)
					if publishedConsumersOnly {
						allowed, err := r.dedicatedPublicObjectReferencesIncludePublishedConsumer(items)
						if err != nil || allowed {
							state.HasPublishedConsumer = allowed
							return state, err
						}
					}
				}
			}
			if len(rows) < 256 {
				break
			}
			afterID = rows[len(rows)-1].ID
		}
	}
	if !publishedConsumersOnly && r.hasAssetDerivativesTable() {
		for afterID := uint(0); ; {
			var rows []media.MediaAssetDerivative
			if err := r.db.Unscoped().Select("id", "media_asset_id", "url", "storage_key").Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
				return state, err
			}
			for _, row := range rows {
				if r.matchesExactRegisteredOrTrustedMediaReference(query, row.URL) || registeredMediaReferenceQueryContainsExactObjectKey(query, row.StorageKey) {
					state.RetainedReferences = append(state.RetainedReferences, newMediaReference(media.ReferenceCategoryContent, "dedicated_object_registration_conflict", row.ID, row.MediaAssetID, "通用衍生登记与专用对象身份冲突，需核对存储", "derivative.storage_key, derivative.url"))
				}
			}
			if len(rows) < 256 {
				break
			}
			afterID = rows[len(rows)-1].ID
		}
	}
	seen := make(map[string]bool)
	unique := make([]media.AssetReference, 0, len(state.RetainedReferences))
	for _, reference := range state.RetainedReferences {
		identity := fmt.Sprintf("%s/%d/%d/%s", reference.ResourceType, reference.ResourceID, reference.ParentResourceID, reference.Field)
		if seen[identity] {
			continue
		}
		seen[identity] = true
		unique = append(unique, reference)
		if state.HasPublishedConsumer {
			continue
		}
		allowed, err := r.dedicatedPublicObjectReferenceHasPublishedConsumer(reference)
		if err != nil {
			return state, err
		}
		state.HasPublishedConsumer = allowed
	}
	state.RetainedReferences = unique
	return state, nil
}

func (r *MediaRepository) dedicatedPublicObjectReferencesIncludePublishedConsumer(references []media.AssetReference) (bool, error) {
	for _, reference := range references {
		allowed, err := r.dedicatedPublicObjectReferenceHasPublishedConsumer(reference)
		if err != nil || allowed {
			return allowed, err
		}
	}
	return false, nil
}

func (r *MediaRepository) dedicatedPublicObjectReferenceHasPublishedConsumer(reference media.AssetReference) (bool, error) {
	var count int64
	query := r.db
	switch reference.ResourceType {
	case "product_rich_text":
		query = query.Table("products").Where("id = ? AND deleted_at IS NULL AND status = ?", reference.ResourceID, "active")
	case "product_media":
		if !r.hasTable("products") {
			return false, nil
		}
		query = query.Table("product_media AS placements").Joins("JOIN products AS owners ON owners.id = placements.product_id").Where("placements.id = ? AND placements.deleted_at IS NULL AND placements.is_visible = ? AND COALESCE(placements.url, '') <> '' AND owners.deleted_at IS NULL AND owners.status = ?", reference.ResourceID, true, "active")
	case "product_option_swatch":
		if !r.hasTable("products") || !r.hasTable("product_spec_definitions") || !r.hasTable("product_specification_templates") {
			return false, nil
		}
		query = query.Table("product_variant_option_values AS placements").
			Joins("JOIN products AS owners ON owners.id = placements.product_id").
			Joins("JOIN product_specification_templates AS templates ON templates.id = owners.product_specification_template_id").
			Joins("JOIN product_spec_definitions AS definitions ON definitions.id = placements.spec_definition_id AND definitions.product_specification_template_id = templates.id").
			Where("placements.id = ? AND placements.is_enabled = ? AND definitions.is_visible = ? AND TRIM(definitions.role) IN ? AND owners.deleted_at IS NULL AND owners.status = ?", reference.ResourceID, true, true, []string{"variant", "custom_option"}, "active")
	case "product_category":
		query = query.Table("product_categories").Where("id = ? AND is_enabled = ?", reference.ResourceID, true)
	case "product_brand_logo":
		query = query.Table("product_brands").Where("id = ? AND is_enabled = ?", reference.ResourceID, true)
	case "post", "faq", "gallery_cover":
		table := map[string]string{"post": "posts", "faq": "faqs", "gallery_cover": "galleries"}[reference.ResourceType]
		query = query.Table(table).Where("id = ? AND deleted_at IS NULL AND status = ?", reference.ResourceID, "published")
	case "gallery_image":
		if !r.hasTable("galleries") {
			return false, nil
		}
		query = query.Table("gallery_images AS placements").Joins("JOIN galleries AS owners ON owners.id = placements.gallery_id").Where("placements.id = ? AND placements.deleted_at IS NULL AND owners.deleted_at IS NULL AND owners.status = ?", reference.ResourceID, "published")
	case "website_profile_setting", "refund_cancellation_policy_setting":
		query = query.Table("settings").Where("id = ? AND deleted_at IS NULL AND is_public = ?", reference.ResourceID, true)
	case "workbench_feed_media":
		if !r.hasTable("workbench_feed_entries") || !r.db.Migrator().HasColumn("workbench_feed_media", "deleted_at") {
			return false, nil
		}
		query = query.Table("workbench_feed_media AS placements").Joins("JOIN workbench_feed_entries AS owners ON owners.id = placements.entry_id").Where("placements.id = ? AND placements.deleted_at IS NULL AND owners.deleted_at IS NULL AND owners.status = ?", reference.ResourceID, "published")
	default:
		// Identity mirrors, templates, private conversations and colliding
		// registrations retain bytes but do not grant anonymous access.
		return false, nil
	}
	err := query.Count(&count).Error
	return count > 0, err
}
