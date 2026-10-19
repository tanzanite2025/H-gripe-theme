package service

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
)

// Metadata media follows existing attachments in the file endpoint's membership
// index. It stays in its card field; it is not added to the displayed attachments.
func (s *CustomerServiceConversationMessageAttachmentAccessService) structuredMetadataMediaFieldsAndAttachmentOffset(message ticket.TicketMessage) ([]ticket.StructuredMessageMediaField, int, error) {
	offset := len(parseCustomerServiceMessageAttachmentReferences(message.Attachments))
	if message.AttachmentReferenceFormat == ticket.TicketMessageMediaAssetBindings {
		bindings, err := s.tickets.ticketRepo.FindCustomerServiceMessageMediaAssetBindings(message.ID)
		if err != nil {
			return nil, 0, err
		}
		offset = 0
		for _, binding := range bindings {
			if binding.AttachmentIndex >= offset {
				offset = binding.AttachmentIndex + 1
			}
		}
	}
	document, err := ticket.ParseStructuredMessageMetadataKnownMediaFields(message.MessageType, message.Metadata)
	if err != nil {
		return nil, offset, ErrCustomerServiceMessageAttachmentUnavailable
	}
	return document.Fields, offset, nil
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) resolveRegisteredStructuredMetadataMediaReference(reference string) (*media.MediaAsset, string, string, error) {
	parsed, err := url.Parse(reference)
	if err != nil || parsed.User != nil {
		return nil, "", "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	parsed.RawQuery, parsed.Fragment, parsed.RawFragment, parsed.ForceQuery = "", "", "", false
	key, err := s.media.ResolveRegisteredMediaObjectKeyForCleanupOwnership(parsed.String())
	if err != nil {
		return nil, "", "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	if key == "" {
		key = s.media.reviewedHistoricalCustomerServiceAttachmentObjectKey(reference)
	}
	asset, derivative, err := s.media.repo.FindUniquelyRegisteredMediaObjectForExistingConsumerReadIncludingRetiredDerivative(reference, key)
	if err != nil || asset == nil {
		return nil, "", "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	resolvedAsset, originalKey, err := s.resolveRegisteredCustomerServiceMessageAttachment(asset.URL)
	if err != nil {
		return nil, "", "", err
	}
	preset := ""
	if derivative != nil {
		derivativeKey, err := s.validCustomerServiceAttachmentDerivativeStorageKey(resolvedAsset, *derivative)
		if err != nil {
			return nil, "", "", err
		}
		// This membership identifies an exact historical generation. It must
		// still work if that preset was disabled or its current version changed.
		consumerAsset := *resolvedAsset
		consumerAsset.MimeType = derivative.MimeType
		consumerAsset.OriginalFilename = path.Base(derivativeKey)
		return &consumerAsset, derivativeKey, "", nil
	}
	return resolvedAsset, originalKey, preset, nil
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) MetadataWithAuthorizedProtectedMediaURLsForPublicMessage(message ticket.TicketMessage, descriptions []CustomerServiceMessageAttachmentAccessDescription, conversationID string) string {
	if s == nil || s.media == nil {
		return message.Metadata
	}
	prefix := fmt.Sprintf("%s/api/v1/customer-service/conversations/%s/messages/%d/attachments", strings.TrimRight(s.media.siteURL, "/"), url.PathEscape(conversationID), message.ID)
	return s.metadataWithAuthorizedProtectedStructuredMediaURLs(message, descriptions, prefix)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) MetadataWithAuthorizedProtectedMediaURLsForAgentMessage(message ticket.TicketMessage, descriptions []CustomerServiceMessageAttachmentAccessDescription) string {
	if s == nil || s.media == nil {
		return message.Metadata
	}
	prefix := fmt.Sprintf("/api/admin/customer-service/conversations/%d/messages/%d/attachments", message.TicketID, message.ID)
	return s.metadataWithAuthorizedProtectedStructuredMediaURLs(message, descriptions, prefix)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) metadataWithAuthorizedProtectedStructuredMediaURLs(message ticket.TicketMessage, descriptions []CustomerServiceMessageAttachmentAccessDescription, prefix string) string {
	fields, offset, err := s.structuredMetadataMediaFieldsAndAttachmentOffset(message)
	if err != nil {
		return s.MetadataWithAuthorizedProtectedAttachmentURLs(message.Metadata, descriptions)
	}
	document, err := ticket.ParseStructuredMessageMetadataKnownMediaFields(message.MessageType, message.Metadata)
	if err != nil {
		return s.MetadataWithAuthorizedProtectedAttachmentURLs(message.Metadata, descriptions)
	}
	combined := append([]CustomerServiceMessageAttachmentAccessDescription(nil), descriptions...)
	for index, field := range fields {
		asset, _, _, err := s.resolveRegisteredStructuredMetadataMediaReference(field.URL)
		if err != nil {
			document.Fields[index].URL = s.media.CanonicalPublicMediaURL(field.URL)
			continue
		}
		if err := s.media.projectHistoricalTicketAttachmentConversationAccess(asset); err != nil {
			document.Fields[index].URL = ""
			continue
		}
		if asset.Visibility != "private" && !asset.RequiresConversationAccess {
			continue
		}
		accessURL := fmt.Sprintf("%s/%d/file", prefix, offset+index)
		document.Fields[index].URL = accessURL
		combined = append(combined, CustomerServiceMessageAttachmentAccessDescription{AssetID: asset.ID, AttachmentIndex: offset + index, AccessURL: accessURL, MediaType: asset.MediaType, RequiresConversationAccess: true})
	}
	projected, err := document.RenderMetadataWithCanonicalMediaFields()
	if err != nil {
		return ""
	}
	return s.MetadataWithAuthorizedProtectedAttachmentURLs(projected, combined)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) openStructuredMetadataMediaAtMembershipIndex(ctx context.Context, message ticket.TicketMessage, attachmentIndex int, preset string) (*MediaAssetFile, bool, error) {
	fields, offset, err := s.structuredMetadataMediaFieldsAndAttachmentOffset(message)
	if attachmentIndex < offset {
		return nil, false, nil
	}
	if err != nil {
		return nil, true, err
	}
	index := attachmentIndex - offset
	if index < 0 || index >= len(fields) {
		return nil, true, ErrCustomerServiceMessageAttachmentUnavailable
	}
	asset, key, defaultPreset, err := s.resolveRegisteredStructuredMetadataMediaReference(fields[index].URL)
	if err != nil {
		return nil, true, err
	}
	if preset == "" {
		preset = defaultPreset
	}
	file, err := s.openCustomerServiceAttachmentAssetFile(ctx, asset, key, preset)
	return file, true, err
}
