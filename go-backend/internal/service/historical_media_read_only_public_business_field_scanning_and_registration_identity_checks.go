package service

import (
	"fmt"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/faqcontent"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/websitecontent"
)

type historicalMediaReadOnlyBusinessFieldSource struct {
	Table, Field, AssetIDField, ExpectedType, Markup, SettingKey, DedicatedNamespace, StorageKeyField string
}

func historicalMediaRecognizedPublicBusinessFieldSourcesForReadOnlyInventory() []historicalMediaReadOnlyBusinessFieldSource {
	sources := []historicalMediaReadOnlyBusinessFieldSource{
		{Table: "product_media", Field: "url", AssetIDField: "media_asset_id"},
		{Table: "product_media", Field: "thumbnail_url", ExpectedType: "image"},
		{Table: "product_media", Field: "poster_url", ExpectedType: "image"},
		{Table: "products", Field: "description", Markup: "product"},
		{Table: "products", Field: "short_description", Markup: "product"},
		{Table: "product_information_templates", Field: "content", Markup: "product"},
		{Table: "product_categories", Field: "seo_intro", Markup: "product"},
		{Table: "product_category_translations", Field: "seo_intro", Markup: "product"},
		{Table: "product_categories", Field: "image_url", AssetIDField: "image_media_asset_id", ExpectedType: "image"},
		{Table: "product_brands", Field: "logo_url", ExpectedType: "image"},
		{Table: "product_variant_option_values", Field: "swatch_url", AssetIDField: "swatch_media_asset_id", ExpectedType: "image"},
		{Table: "product_spec_option_items", Field: "swatch_url", AssetIDField: "swatch_media_asset_id", ExpectedType: "image"},
		{Table: "posts", Field: "featured_image", ExpectedType: "image"},
		{Table: "posts", Field: "featured_img", ExpectedType: "image"},
		{Table: "posts", Field: "content", Markup: "post"},
		{Table: "faqs", Field: "answer_image_url", ExpectedType: "image"},
		{Table: "faqs", Field: "answer", Markup: "faq"},
		{Table: "galleries", Field: "cover_image", ExpectedType: "image"},
		{Table: "gallery_images", Field: "url", AssetIDField: "media_asset_id", ExpectedType: "image"},
		{Table: "gallery_images", Field: "thumbnail", ExpectedType: "image"},
		{Table: "workbench_feed_media", Field: "file_path", ExpectedType: "image"},
		{Table: "ticket_messages", Field: "metadata", Markup: "structured_message"},
		{Table: "ticket_auto_replies", Field: "metadata", Markup: "structured_message"},
		{Table: "site_logo_assets", Field: "url", DedicatedNamespace: "site-logo", StorageKeyField: "storage_key", ExpectedType: "image"},
		{Table: "site_favicon_assets", Field: "url", DedicatedNamespace: "site-favicon", StorageKeyField: "storage_key", ExpectedType: "image"},
		{Table: "visual_showcase_items", Field: "image_url", DedicatedNamespace: "visual-showcase", StorageKeyField: "storage_key", ExpectedType: "image"},
		{Table: "visual_showcase_items", Field: "thumbnail_url", DedicatedNamespace: "visual-showcase", ExpectedType: "image"},
	}
	for _, key := range []string{setting.SiteLogoSettingKey, setting.SiteFaviconSettingKey, setting.WebsiteProfileKeyAvatarURL, setting.WebsiteProfileKeyFactoryImageURL, setting.WebsiteProfileKeyStatementBody, setting.WebsiteProfileKeyStatementParagraph1, setting.WebsiteProfileKeyStatementParagraph2} {
		source := historicalMediaReadOnlyBusinessFieldSource{Table: "settings", Field: "value", SettingKey: key, ExpectedType: "image"}
		if setting.IsWebsiteProfileSharedMediaRichTextSettingKey(key) {
			source.Markup, source.ExpectedType = "website", ""
		}
		sources = append(sources, source)
	}
	return sources
}

func (s *HistoricalMediaReadOnlyInventoryService) scanHistoricalMediaBusinessFieldsWithoutMutatingRecords(tx *gorm.DB, index *historicalMediaReadOnlyRegistrationIndex, report *HistoricalMediaReadOnlyInventoryReport) error {
	for _, source := range historicalMediaRecognizedPublicBusinessFieldSourcesForReadOnlyInventory() {
		if source.Table == "posts" && (source.Field == "featured_image" || source.Field == "featured_img") && tx.Migrator().HasTable("posts") && !tx.Migrator().HasColumn("posts", source.Field) {
			otherColumn := "featured_image"
			if source.Field == otherColumn {
				otherColumn = "featured_img"
			}
			if tx.Migrator().HasColumn("posts", otherColumn) {
				continue // An absent compatibility alias is not a missing source.
			}
		}
		coverageField := source.Field
		if source.SettingKey != "" {
			coverageField += "(" + source.SettingKey + ")"
		}
		coverage := HistoricalMediaReadOnlyInventorySourceCoverage{Table: source.Table, Field: coverageField, Complete: true}
		if !tx.Migrator().HasTable(source.Table) {
			coverage.Unavailable, coverage.Complete = true, false
			report.Sources = append(report.Sources, coverage)
			continue
		}
		required := []string{"id", source.Field}
		if source.AssetIDField != "" {
			required = append(required, source.AssetIDField)
		}
		if source.StorageKeyField != "" {
			required = append(required, source.StorageKeyField)
		}
		if source.SettingKey != "" {
			required = append(required, "key")
		}
		if source.Markup == "structured_message" {
			required = append(required, "message_type")
		}
		for _, column := range required {
			if !tx.Migrator().HasColumn(source.Table, column) {
				coverage.Unavailable, coverage.Complete = true, false
				report.recordHistoricalMediaInventoryIssue(HistoricalMediaReadOnlyInventoryIssue{Code: "business_source_column_unavailable", Table: source.Table, Field: column}, "")
			}
		}
		if coverage.Unavailable {
			report.Sources = append(report.Sources, coverage)
			continue
		}
		assetIDSelect, storageKeySelect := "NULL AS media_asset_id", "'' AS storage_key"
		retainedRecordSelect := "FALSE AS retained_deleted_record"
		if tx.Migrator().HasColumn(source.Table, "deleted_at") {
			retainedRecordSelect = "(deleted_at IS NOT NULL) AS retained_deleted_record"
		}
		condition := "COALESCE(" + source.Field + ", '') <> ''"
		if source.AssetIDField != "" {
			assetIDSelect = source.AssetIDField + " AS media_asset_id"
			condition = "(" + condition + " OR COALESCE(" + source.AssetIDField + ", 0) <> 0)"
		}
		if source.StorageKeyField != "" {
			storageKeySelect = source.StorageKeyField + " AS storage_key"
			condition = "(" + condition + " OR COALESCE(" + source.StorageKeyField + ", '') <> '')"
		}
		for {
			query := tx.Table(source.Table).Where("id > ?", coverage.LastRecordID).Where(condition)
			if source.SettingKey != "" {
				query = query.Where("key = ?", source.SettingKey)
			}
			limit := min(report.Options.BatchSize, report.Options.MaxRecordsPerSource-coverage.RowsScanned)
			if limit == 0 {
				if err := query.Count(&coverage.RemainingRows).Error; err != nil {
					return err
				}
				coverage.Complete = coverage.RemainingRows == 0
				break
			}
			var rows []struct {
				ID                    uint
				Reference             string
				MediaAssetID          *uint
				StorageKey            string
				RetainedDeletedRecord bool
				MessageType           string
			}
			columns := "id, " + source.Field + " AS reference, " + assetIDSelect + ", " + storageKeySelect + ", " + retainedRecordSelect
			if source.Markup == "structured_message" {
				columns += ", message_type"
			}
			if err := query.Select(columns).Order("id ASC").Limit(limit).Find(&rows).Error; err != nil {
				return fmt.Errorf("read historical media source %s.%s: %w", source.Table, source.Field, err)
			}
			for _, row := range rows {
				issue := HistoricalMediaReadOnlyInventoryIssue{Table: source.Table, RecordID: row.ID, Field: coverageField, RetainedDeletedRecord: row.RetainedDeletedRecord}
				if row.MediaAssetID != nil {
					issue.AssetID = *row.MediaAssetID
				}
				switch source.Markup {
				case "":
					s.inspectHistoricalMediaBusinessReferenceIdentity(index, report, source, issue, row.Reference, source.ExpectedType, row.StorageKey)
				case "structured_message":
					document, err := ticket.ParseStructuredMessageMetadataKnownMediaFields(row.MessageType, row.Reference)
					if err != nil {
						issue.Code = "business_structured_media_could_not_be_parsed"
						report.recordHistoricalMediaInventoryIssue(issue, "")
					} else {
						for _, field := range document.Fields {
							fieldIssue := issue
							fieldIssue.Field += "." + field.Path
							s.inspectHistoricalMediaBusinessReferenceIdentity(index, report, source, fieldIssue, field.URL, field.ExpectedMediaType, "")
						}
					}
				default:
					markup := row.Reference
					var err error
					switch source.Markup {
					case "faq":
						markup, err = faqcontent.SanitizeAnswer(markup)
					case "website":
						markup, err = websitecontent.NormalizeWebsiteBody(markup)
					case "post":
						markup, err = safehtml.Sanitize(markup)
					}
					if err != nil {
						issue.Code = "business_markup_could_not_be_parsed"
						report.recordHistoricalMediaInventoryIssue(issue, "")
						continue
					}
					document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(markup)
					if err != nil {
						issue.Code = "business_markup_could_not_be_parsed"
						report.recordHistoricalMediaInventoryIssue(issue, "")
						continue
					}
					for _, attribute := range document.References {
						attributeIssue := issue
						attributeIssue.Field += "." + attribute.ElementAndAttribute
						s.inspectHistoricalMediaBusinessReferenceIdentity(index, report, source, attributeIssue, *attribute.URL, attribute.ExpectedMediaType, "")
					}
				}
				coverage.LastRecordID = row.ID
			}
			// Advance even if a row's malformed markup was recorded and skipped.
			if len(rows) > 0 {
				coverage.LastRecordID = rows[len(rows)-1].ID
			}
			coverage.RowsScanned += len(rows)
			if len(rows) < limit {
				break
			}
		}
		report.Sources = append(report.Sources, coverage)
	}
	return nil
}

func (s *HistoricalMediaReadOnlyInventoryService) inspectHistoricalMediaBusinessReferenceIdentity(index *historicalMediaReadOnlyRegistrationIndex, report *HistoricalMediaReadOnlyInventoryReport, source historicalMediaReadOnlyBusinessFieldSource, issue HistoricalMediaReadOnlyInventoryIssue, reference, expectedType, storedOwnerKey string) {
	report.ReferencesScanned++
	key := ""
	if strings.TrimSpace(reference) != "" {
		var err error
		key, err = s.resolveHistoricalMediaPublicReferenceObjectKey(reference)
		if err != nil {
			issue.Code = "business_public_media_reference_invalid"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
			return
		}
	}
	if source.DedicatedNamespace != "" {
		if key == "" || storage.DedicatedPublicObjectNamespaceForStorageKey(key) != source.DedicatedNamespace {
			issue.Code = "dedicated_owner_url_identity_requires_review"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		}
		if source.StorageKeyField != "" {
			ownerKey, valid := storage.NormalizeObjectKey(storedOwnerKey)
			if !valid || (key != "" && ownerKey != key) {
				issue.Code = "dedicated_owner_url_and_storage_key_disagree"
				report.recordHistoricalMediaInventoryIssue(issue, reference)
			}
		}
		return // Dedicated owner records do not require shared registration.
	}
	if storage.DedicatedPublicObjectNamespaceForStorageKey(key) != "" {
		issue.Code = "historical_dedicated_object_business_reference"
		report.recordHistoricalMediaInventoryIssue(issue, reference)
		return
	}
	candidates := make(map[int]bool)
	if reference != "" {
		for _, position := range index.urls[normalizedHistoricalMediaRegistrationURL(reference)] {
			candidates[position] = true
		}
	}
	if key != "" {
		for _, position := range index.keys[key] {
			candidates[position] = true
		}
	}
	if len(candidates) > 1 {
		issue.Code = "business_reference_registration_identity_ambiguous"
		report.recordHistoricalMediaInventoryIssue(issue, reference)
		return
	}
	var asset *media.MediaAsset
	if issue.AssetID != 0 {
		asset = index.assets[issue.AssetID]
		if asset == nil && index.complete {
			issue.Code = "business_asset_id_not_registered"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		}
	}
	for position := range candidates {
		registration := index.entries[position]
		if issue.AssetID != 0 && issue.AssetID != registration.AssetID {
			issue.Code, issue.RelatedRecordID = "business_asset_id_and_url_identity_disagree", registration.AssetID
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		}
		issue.AssetID = registration.AssetID
		asset = index.assets[registration.AssetID]
		if registration.DeletedAt.Valid {
			issue.Code = "business_reference_uses_retired_registration"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		}
	}
	if len(candidates) == 0 && reference != "" {
		if key != "" && index.complete {
			issue.Code = "business_managed_url_not_registered"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		} else if parsed, err := url.Parse(reference); key == "" && err == nil && parsed.Host != "" && strings.Contains(parsed.Path, "/uploads/") {
			issue.Code = "untrusted_upload_url_origin_requires_review"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		}
		if asset != nil {
			issue.Code = "business_asset_id_url_identity_requires_review"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		}
	}
	if asset != nil {
		if asset.DeletedAt.Valid {
			issue.Code = "business_reference_uses_retired_asset"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		} else if !media.CanMediaAssetBeReusedByPublicBusinessContent(asset) {
			issue.Code = "business_reference_not_eligible_for_new_public_reuse"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		}
		if expectedType != "" && asset.MediaType != expectedType {
			issue.Code = "business_field_media_type_disagrees"
			report.recordHistoricalMediaInventoryIssue(issue, reference)
		}
	}
}
