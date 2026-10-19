package service

import (
	"encoding/json"
	"net/url"
	"strings"
)

// Legacy metadata can repeat an original or thumbnail outside attachments.
// Rewrite only registered objects belonging to protected attachments in this
// already authorized response. Never change another business URL by suffix.
func (s *CustomerServiceConversationMessageAttachmentAccessService) MetadataWithAuthorizedProtectedAttachmentURLs(value string, descriptions []CustomerServiceMessageAttachmentAccessDescription) string {
	protected := make(map[uint]string)
	for _, description := range descriptions {
		if description.RequiresConversationAccess {
			protected[description.AssetID] = description.AccessURL
		}
	}
	if len(protected) == 0 || s == nil || s.media == nil || s.media.repo == nil {
		return value
	}
	var metadata any
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.UseNumber()
	if !json.Valid([]byte(value)) || decoder.Decode(&metadata) != nil {
		return ""
	}
	metadata = s.replaceProtectedAttachmentMetadataURLsRecursively(metadata, protected)
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) replaceProtectedAttachmentMetadataURLsRecursively(value any, protected map[uint]string) any {
	switch typed := value.(type) {
	case string:
		key, err := s.media.ResolveTrustedMediaObjectKeyForPublicBusinessReference(typed)
		if s.media.storage != nil {
			if parsed, parseErr := url.Parse(typed); parseErr == nil {
				parsed.RawQuery, parsed.Fragment, parsed.RawFragment, parsed.ForceQuery = "", "", "", false
				if storedKey, storageErr := s.media.storage.ObjectKey(parsed.String()); storageErr == nil {
					key, err = storedKey, nil
				}
			}
		}
		if err != nil {
			return typed
		}
		if key == "" {
			key = s.media.reviewedHistoricalCustomerServiceAttachmentObjectKey(typed)
		}
		asset, derivative, err := s.media.repo.FindUniquelyRegisteredMediaObjectForExistingConsumerReadIncludingRetiredDerivative(typed, key)
		if err != nil || asset == nil || protected[asset.ID] == "" {
			return typed
		}
		address := protected[asset.ID]
		if derivative != nil {
			address += "?preset=" + url.QueryEscape(derivative.Preset)
		}
		return address
	case []any:
		for index, item := range typed {
			typed[index] = s.replaceProtectedAttachmentMetadataURLsRecursively(item, protected)
		}
	case map[string]any:
		for key, item := range typed {
			typed[key] = s.replaceProtectedAttachmentMetadataURLsRecursively(item, protected)
		}
	}
	return value
}
