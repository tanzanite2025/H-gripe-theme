package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"
)

// A preview URL is a transport receipt, never a stored business relation.
// Existing clients may send the URL returned by upload; new clients may send
// media_attachments directly. Ownership is checked again while binding.
func (s *CustomerServiceConversationMessageAttachmentAccessService) PrepareCustomerServiceMediaAttachmentBindingRequests(
	values []string, explicitRequests []ticket.CustomerServiceMessageMediaAttachmentBindingRequest, conversationID string, ticketID uint, maxCount int,
) ([]ticket.CustomerServiceMessageMediaAttachmentBindingRequest, error) {
	uniqueValues := make([]string, 0, len(values))
	seen := map[string]bool{}
	hasReceipt := false
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		uniqueValues = append(uniqueValues, value)
		if strings.Contains(value, "/attachment-uploads/") {
			hasReceipt = true
		}
	}
	if len(explicitRequests) == 0 && !hasReceipt {
		return nil, nil
	}
	if s == nil || s.media == nil {
		return nil, ErrMediaStorageUnavailable
	}
	if len(explicitRequests) > 0 {
		if len(uniqueValues) > 0 || len(explicitRequests) > maxCount {
			return nil, repository.ErrCustomerServiceAttachmentBindingUnavailable
		}
		return explicitRequests, nil
	}
	if len(uniqueValues) > maxCount {
		return nil, repository.ErrCustomerServiceAttachmentBindingUnavailable
	}
	prefix := fmt.Sprintf("/api/v1/customer-service/conversations/%s/attachment-uploads/", url.PathEscape(conversationID))
	if ticketID > 0 {
		prefix = fmt.Sprintf("/api/admin/customer-service/conversations/%d/attachment-uploads/", ticketID)
	}
	bindings := make([]ticket.CustomerServiceMessageMediaAttachmentBindingRequest, 0, len(uniqueValues))
	for _, reference := range uniqueValues {
		parsed, err := url.Parse(reference)
		if err != nil || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, repository.ErrCustomerServiceAttachmentBindingUnavailable
		}
		if strings.Contains(parsed.Path, "/attachment-uploads/") {
			if parsed.Host != "" && (!sameAttachmentReferenceHost(parsed.Host, s.media.siteURL) || (parsed.Scheme != "http" && parsed.Scheme != "https")) {
				return nil, repository.ErrCustomerServiceAttachmentBindingUnavailable
			}
			if !strings.HasPrefix(parsed.Path, prefix) || !strings.HasSuffix(parsed.Path, "/file") {
				return nil, repository.ErrCustomerServiceAttachmentBindingUnavailable
			}
			receiptID := strings.TrimSuffix(strings.TrimPrefix(parsed.Path, prefix), "/file")
			if _, err := uuid.Parse(receiptID); err != nil {
				return nil, repository.ErrCustomerServiceAttachmentBindingUnavailable
			}
			bindings = append(bindings, ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: receiptID})
		} else {
			asset, _, err := s.resolveRegisteredCustomerServiceMessageAttachment(reference)
			if err != nil || asset.Visibility != "public" {
				return nil, repository.ErrCustomerServiceAttachmentBindingUnavailable
			}
			bindings = append(bindings, ticket.CustomerServiceMessageMediaAttachmentBindingRequest{MediaAssetID: asset.ID})
		}
	}
	return bindings, nil
}
