package repository

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
)

// This is a read projection, not a classification or visibility update. Deleted
// tickets and orphaned messages still contain sensitive historical attachments.
func (r *MediaRepository) HistoricalUnclassifiedMediaRequiresConversationAccess(asset *media.MediaAsset) (bool, error) {
	if !media.IsUnclassifiedHistoricalMediaAssetWithUnknownUploadProvenance(asset) {
		return false, nil
	}
	query, err := r.includeRegisteredDerivativeURLsAndObjectKeysInMediaReferenceQuery(newMediaAssetReferenceQuery(asset), asset)
	if err != nil {
		return false, err
	}
	query.IncludeRetainedDeletedPlacements = true
	// Authoritative managed owners remain conversation-scoped even without a
	// message projection. Legacy URL/publication heuristics cannot override them.
	for _, scanner := range []mediaAssetReferenceScanner{r.customerServiceMessageMediaAssetBindingReferences, r.customerServicePendingUploadSessionAssetReferences} {
		references, err := scanner(query)
		if err != nil {
			return false, err
		}
		if len(references) > 0 {
			return true, nil
		}
	}
	if !r.hasTable("ticket_messages") {
		return false, nil
	}
	hasTicketAttachment, err := r.historicalTicketAttachmentsContainExactRegisteredMediaReference(query)
	if err == nil && !hasTicketAttachment {
		hasTicketAttachment, err = r.historicalTicketStructuredMetadataContainsPotentialRegisteredMediaReference(query)
	}
	if err != nil || !hasTicketAttachment {
		return false, err
	}
	for _, scan := range []mediaAssetReferenceScanner{
		r.productCategoryImageReferences, r.productBrandLogoReferences, r.productMediaReferences,
		r.productRichTextMediaAttributeReferences, r.productAndTemplateOptionSwatchReferences,
		r.galleryReferences, r.faqReferences, r.postReferences, r.websiteProfileSettingMediaReferences, r.workbenchFeedImageReferences,
	} {
		references, err := scan(query)
		if err != nil {
			return true, err
		}
		published, err := r.dedicatedPublicObjectReferencesIncludePublishedConsumer(references)
		if err != nil || published {
			return !published, err
		}
	}
	return true, nil
}

func (r *MediaRepository) historicalTicketAttachmentsContainExactRegisteredMediaReference(query mediaAssetReferenceQuery) (bool, error) {
	if !r.db.Migrator().HasColumn("ticket_messages", "attachments") {
		return false, nil
	}
	for afterID := uint(0); ; {
		var rows []struct {
			ID          uint
			Attachments string
		}
		if err := r.db.Table("ticket_messages").Select("id", "attachments").Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return false, err
		}
		for _, row := range rows {
			if row.Attachments == "" {
				continue
			}
			var value any
			if err := json.Unmarshal([]byte(row.Attachments), &value); err != nil {
				return false, fmt.Errorf("decode historical ticket attachment references for message %d: %w", row.ID, err)
			}
			if r.historicalTicketAttachmentJSONContainsPotentialRegisteredMediaReference(query, value) {
				return true, nil
			}
		}
		if len(rows) < 256 {
			return false, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}

func (r *MediaRepository) historicalTicketStructuredMetadataContainsPotentialRegisteredMediaReference(query mediaAssetReferenceQuery) (bool, error) {
	if !r.db.Migrator().HasColumn("ticket_messages", "metadata") {
		return false, nil
	}
	columns := "id, metadata"
	if r.db.Migrator().HasColumn("ticket_messages", "message_type") {
		columns += ", message_type"
	}
	for afterID := uint(0); ; {
		var rows []struct {
			ID                    uint
			Metadata, MessageType string
		}
		if err := r.db.Table("ticket_messages").Select(columns).Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return false, err
		}
		for _, row := range rows {
			document, err := ticket.ParseStructuredMessageMetadataKnownMediaFields(row.MessageType, row.Metadata)
			if err != nil {
				return false, fmt.Errorf("inspect retained ticket_messages.metadata #%d: invalid structured media fields", row.ID)
			}
			for _, field := range document.Fields {
				if r.historicalTicketAttachmentJSONContainsPotentialRegisteredMediaReference(query, field.URL) {
					return true, nil
				}
			}
		}
		if len(rows) < 256 {
			return false, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}

func (r *MediaRepository) historicalTicketAttachmentJSONContainsPotentialRegisteredMediaReference(query mediaAssetReferenceQuery, value any) bool {
	switch value := value.(type) {
	case string:
		if r.matchesExactRegisteredOrTrustedMediaReference(query, value) {
			return true
		}
		// An unrecognized historical origin is uncertainty about privacy, never
		// permission to read local bytes. A complete root /uploads key can only
		// withhold anonymous access; authorized reads still require an exact
		// registration or an explicitly reviewed origin.
		parsed, err := url.Parse(strings.TrimSpace(value))
		if err != nil || parsed.User != nil || parsed.Opaque != "" ||
			(parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https") || !strings.HasPrefix(parsed.Path, "/uploads/") {
			return false
		}
		key, valid := storage.NormalizeObjectKey(strings.TrimPrefix(parsed.Path, "/uploads/"))
		if !valid {
			return false
		}
		for _, registeredKey := range query.ObjectKeys {
			if key == registeredKey {
				return true
			}
		}
		return false
	case []any:
		for _, item := range value {
			if r.historicalTicketAttachmentJSONContainsPotentialRegisteredMediaReference(query, item) {
				return true
			}
		}
	case map[string]any:
		for _, item := range value {
			if r.historicalTicketAttachmentJSONContainsPotentialRegisteredMediaReference(query, item) {
				return true
			}
		}
	}
	return false
}
