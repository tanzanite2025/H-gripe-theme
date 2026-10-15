package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
)

type CustomerServiceMessageAttachmentDerivativeAccessDescription struct {
	Preset    string `json:"preset"`
	AccessURL string `json:"access_url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type CustomerServiceMessageAttachmentAccessDescription struct {
	AttachmentIndex int                                                           `json:"attachment_index"`
	AssetID         uint                                                          `json:"asset_id"`
	MediaType       string                                                        `json:"media_type"`
	AccessURL       string                                                        `json:"access_url"`
	Derivatives     []CustomerServiceMessageAttachmentDerivativeAccessDescription `json:"derivatives,omitempty"`
}

// Descriptions are generated only for messages already authorized by their
// history/send handler. The file endpoints independently recheck authorization
// on every request, including attachment membership and the parent asset.
func (s *CustomerServiceConversationMessageAttachmentAccessService) DescribePublicCustomerServiceMessageAttachments(message ticket.TicketMessage, conversationID string) ([]string, []CustomerServiceMessageAttachmentAccessDescription) {
	if s == nil || s.media == nil {
		if message.AttachmentReferenceFormat == ticket.TicketMessageMediaAssetBindings {
			return []string{}, nil
		}
		return parseCustomerServiceMessageAttachmentReferences(message.Attachments), nil
	}
	prefix := fmt.Sprintf("%s/api/v1/customer-service/conversations/%s/messages/%d/attachments",
		strings.TrimRight(s.media.siteURL, "/"), url.PathEscape(conversationID), message.ID)
	return s.describeAuthorizedCustomerServiceMessageAttachments(message, prefix, false)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) DescribeAgentCustomerServiceMessageAttachments(message ticket.TicketMessage) ([]string, []CustomerServiceMessageAttachmentAccessDescription) {
	if s == nil || s.media == nil {
		if message.AttachmentReferenceFormat == ticket.TicketMessageMediaAssetBindings {
			return []string{}, nil
		}
		return parseCustomerServiceMessageAttachmentReferences(message.Attachments), nil
	}
	prefix := fmt.Sprintf("/api/admin/customer-service/conversations/%d/messages/%d/attachments", message.TicketID, message.ID)
	return s.describeAuthorizedCustomerServiceMessageAttachments(message, prefix, true)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) describeAuthorizedCustomerServiceMessageAttachments(message ticket.TicketMessage, routePrefix string, preserveHistoricalPublicReferences bool) ([]string, []CustomerServiceMessageAttachmentAccessDescription) {
	references := parseCustomerServiceMessageAttachmentReferences(message.Attachments)
	boundAssets := make(map[int]*media.MediaAsset)
	if message.AttachmentReferenceFormat == ticket.TicketMessageMediaAssetBindings {
		bindings, err := s.tickets.ticketRepo.FindCustomerServiceMessageMediaAssetBindings(message.ID)
		if err != nil {
			return []string{}, nil
		}
		references = make([]string, len(bindings))
		for _, binding := range bindings {
			if binding.AttachmentIndex < 0 || binding.AttachmentIndex >= len(references) {
				continue
			}
			if asset, _, lookupErr := s.resolveCustomerServiceMessageMediaAssetBinding(context.Background(), message, binding); lookupErr == nil {
				references[binding.AttachmentIndex] = asset.URL
				boundAssets[binding.AttachmentIndex] = asset
			}
		}
	}
	compatibleURLs := make([]string, 0, len(references))
	descriptions := make([]CustomerServiceMessageAttachmentAccessDescription, 0, len(references))
	for attachmentIndex, reference := range references {
		compatiblePublicReference := s.media.CanonicalPublicMediaURL(reference)
		if preserveHistoricalPublicReferences && compatiblePublicReference != "" {
			// Backoffice historically returns stored public references as-is.
			// Still suppress private references using the common public resolver.
			compatiblePublicReference = reference
		}
		asset := boundAssets[attachmentIndex]
		if message.AttachmentReferenceFormat != ticket.TicketMessageMediaAssetBindings {
			asset, _, _ = s.resolveRegisteredCustomerServiceMessageAttachment(reference)
		}
		if asset == nil {
			// Unregistered historical public URLs retain their existing response.
			// Private namespaces are suppressed by the canonical public resolver.
			compatibleURLs = append(compatibleURLs, compatiblePublicReference)
			continue
		}
		accessURL := fmt.Sprintf("%s/%d/file", routePrefix, attachmentIndex)
		description := CustomerServiceMessageAttachmentAccessDescription{
			AttachmentIndex: attachmentIndex, AssetID: asset.ID, MediaType: asset.MediaType, AccessURL: accessURL,
		}
		asset.Derivatives, _ = s.media.repo.FindAssetDerivatives(asset.ID)
		for _, derivative := range asset.Derivatives {
			if _, err := s.validCustomerServiceAttachmentDerivativeStorageKey(asset, derivative); err != nil || derivative.DeletedAt.Valid {
				continue
			}
			description.Derivatives = append(description.Derivatives, CustomerServiceMessageAttachmentDerivativeAccessDescription{
				Preset: derivative.Preset, AccessURL: accessURL + "?preset=" + url.QueryEscape(derivative.Preset), Width: derivative.Width, Height: derivative.Height,
			})
		}
		descriptions = append(descriptions, description)
		if asset.Visibility == "private" {
			compatibleURLs = append(compatibleURLs, accessURL)
		} else {
			compatibleURLs = append(compatibleURLs, compatiblePublicReference)
		}
	}
	return compatibleURLs, descriptions
}

func parseCustomerServiceMessageAttachmentReferences(value string) []string {
	var references []string
	if json.Unmarshal([]byte(value), &references) != nil || references == nil {
		return []string{}
	}
	return references
}
