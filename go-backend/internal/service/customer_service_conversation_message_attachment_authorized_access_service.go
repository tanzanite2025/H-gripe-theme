package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"

	"gorm.io/gorm"
)

var ErrCustomerServiceMessageAttachmentUnavailable = errors.New("customer-service message attachment is unavailable")

type CustomerServiceConversationMessageAttachmentAccessService struct {
	tickets *TicketService
	media   *MediaService
}

func NewCustomerServiceConversationMessageAttachmentAccessService(tickets *TicketService, mediaService *MediaService) *CustomerServiceConversationMessageAttachmentAccessService {
	return &CustomerServiceConversationMessageAttachmentAccessService{tickets: tickets, media: mediaService}
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) OpenPublicCustomerServiceMessageAttachment(
	ctx context.Context, conversationID string, owner CustomerServiceOwner, messageID uint, attachmentIndex int, preset string,
) (*MediaAssetFile, error) {
	if s == nil || s.tickets == nil || s.media == nil {
		return nil, ErrMediaStorageUnavailable
	}
	conversation, err := s.tickets.GetPublicCustomerServiceConversation(conversationID, owner)
	if err != nil {
		return nil, err
	}
	message, err := s.tickets.ticketRepo.FindCustomerServiceAttachmentMessageByConversationAndID(ctx, conversation.ID, messageID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerServiceMessageAttachmentUnavailable
		}
		return nil, err
	}
	return s.openRegisteredCustomerServiceMessageAttachment(ctx, *message, attachmentIndex, preset)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) OpenAgentCustomerServiceMessageAttachment(
	ctx context.Context, ticketID, agentUserID uint, canViewAll bool, messageID uint, attachmentIndex int, preset string,
) (*MediaAssetFile, error) {
	if s == nil || s.tickets == nil || s.media == nil {
		return nil, ErrMediaStorageUnavailable
	}
	if agentUserID == 0 {
		return nil, ErrCustomerServiceAgentAccessDenied
	}
	if _, err := s.tickets.GetCustomerServiceConversationForAgent(ticketID, agentUserID, canViewAll); err != nil {
		return nil, err
	}
	message, err := s.tickets.ticketRepo.FindCustomerServiceAttachmentMessageByConversationAndID(ctx, ticketID, messageID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerServiceMessageAttachmentUnavailable
		}
		return nil, err
	}
	return s.openRegisteredCustomerServiceMessageAttachment(ctx, *message, attachmentIndex, preset)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) openRegisteredCustomerServiceMessageAttachment(ctx context.Context, message ticket.TicketMessage, attachmentIndex int, preset string) (*MediaAssetFile, error) {
	if file, handled, err := s.openStructuredMetadataMediaAtMembershipIndex(ctx, message, attachmentIndex, preset); handled || err != nil {
		return file, err
	}
	if message.AttachmentReferenceFormat == ticket.TicketMessageMediaAssetBindings {
		asset, key, err := s.resolveBoundCustomerServiceMessageAttachment(ctx, message, attachmentIndex)
		if err != nil {
			return nil, err
		}
		return s.openCustomerServiceAttachmentAssetFile(ctx, asset, key, preset)
	}
	var references []string
	if json.Unmarshal([]byte(message.Attachments), &references) != nil || attachmentIndex < 0 || attachmentIndex >= len(references) {
		return nil, ErrCustomerServiceMessageAttachmentUnavailable
	}
	asset, originalKey, err := s.resolveRegisteredCustomerServiceMessageAttachment(references[attachmentIndex])
	if err != nil {
		return nil, err
	}
	return s.openCustomerServiceAttachmentAssetFile(ctx, asset, originalKey, preset)
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) openCustomerServiceAttachmentAssetFile(ctx context.Context, asset *media.MediaAsset, originalKey, preset string) (*MediaAssetFile, error) {
	var err error
	key, mimeType, filename := originalKey, asset.MimeType, asset.OriginalFilename
	if filename == "" {
		filename = asset.Filename
	}
	if preset != "" {
		asset.Derivatives, err = s.media.repo.FindAssetDerivatives(asset.ID)
		if err != nil {
			return nil, err
		}
		var selected *media.MediaAssetDerivative
		for index := range asset.Derivatives {
			candidate := &asset.Derivatives[index]
			if candidate.MediaAssetID == asset.ID && candidate.Preset == preset && !candidate.DeletedAt.Valid &&
				(selected == nil || candidate.PresetVersion > selected.PresetVersion || (candidate.PresetVersion == selected.PresetVersion && candidate.ID > selected.ID)) {
				selected = candidate
			}
		}
		if selected == nil {
			return nil, ErrCustomerServiceMessageAttachmentUnavailable
		}
		key, err = s.validCustomerServiceAttachmentDerivativeStorageKey(asset, *selected)
		if err != nil {
			return nil, err
		}
		mimeType, filename = selected.MimeType, path.Base(key)
	}
	// Every supported production provider has ObjectOpener. Require a stream:
	// a public redirect or bearer URL must not bypass conversation authorization.
	if asset.MediaType == "video" && preset == "" {
		if rangeOpener, ok := s.media.storage.(storage.ObjectByteRangeSourceOpener); ok {
			source, err := rangeOpener.OpenObjectByteRangeSource(ctx, key)
			if err != nil {
				return nil, err
			}
			if source == nil {
				return nil, ErrMediaStorageUnavailable
			}
			metadata := source.Metadata()
			return &MediaAssetFile{ReadCloser: source, ByteRangeSource: source, Filename: filename, MimeType: mimeType, Size: metadata.Size, ModTime: metadata.ModTime}, nil
		}
	}
	opener, ok := s.media.storage.(storage.ObjectOpener)
	if !ok {
		return nil, ErrMediaStorageUnavailable
	}
	object, err := opener.Open(ctx, key)
	if err != nil {
		return nil, err
	}
	if object == nil || object.ReadCloser == nil {
		return nil, ErrMediaStorageUnavailable
	}
	return &MediaAssetFile{ReadCloser: object.ReadCloser, Filename: filename, MimeType: mimeType, Size: object.Size, ModTime: object.ModTime}, nil
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) resolveRegisteredCustomerServiceMessageAttachment(reference string) (*media.MediaAsset, string, error) {
	if s == nil || s.media == nil || s.media.repo == nil || s.media.storage == nil {
		return nil, "", ErrMediaStorageUnavailable
	}
	value := strings.TrimSpace(reference)
	parsed, err := url.Parse(value)
	if err != nil || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	if parsed.Host != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	key, err := s.media.storage.ObjectKey(value)
	if err != nil {
		key, err = s.media.ResolveTrustedMediaObjectKeyForPublicBusinessReference(value)
	}
	if err != nil {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	usedReviewedHistoricalOrigin := false
	if key == "" {
		key = s.media.reviewedHistoricalCustomerServiceAttachmentObjectKey(value)
		usedReviewedHistoricalOrigin = key != ""
	}
	asset, derivative, err := s.media.repo.FindUniquelyRegisteredMediaObjectByExactURLOrTrustedKey(value, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
		}
		return nil, "", err
	}
	if derivative != nil {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	if usedReviewedHistoricalOrigin && !media.IsUnclassifiedHistoricalMediaAssetWithUnknownUploadProvenance(asset) {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	if key == "" {
		// An exact registered URL can predate the configured origin. This
		// grants access only through this already authorized message endpoint.
		key = assetStorageKey(asset)
		// The old URL identifies the registration, but its physical shared key
		// must also have one owner, including retained registrations.
		asset, derivative, err = s.media.repo.FindUniquelyRegisteredMediaObjectByExactURLOrTrustedKey(value, key)
		if err != nil {
			return nil, "", err
		}
		if derivative != nil {
			return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
		}
	}
	return validateRegisteredCustomerServiceAttachmentAsset(asset, key)
}

func validateRegisteredCustomerServiceAttachmentAsset(asset *media.MediaAsset, key string) (*media.MediaAsset, string, error) {
	normalizedKey, valid := storage.NormalizeObjectKey(key)
	if asset == nil || !valid || normalizedKey != key || asset.DeletedAt.Valid || asset.Status != "active" || assetStorageKey(asset) != key {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	if (asset.MediaType != "image" || !allowedPublicImageAttachmentMimeType(asset.MimeType)) &&
		(asset.MediaType != "video" || !allowedPublicVideoAttachmentMimeType(asset.MimeType)) {
		return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	if asset.Visibility == "public" && !storage.IsPrivateObjectKey(key) {
		return asset, key, nil
	}
	if asset.Visibility == "private" && storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(key) {
		return asset, key, nil
	}
	// A chat URL cannot confer access to warranty, order evidence, or other
	// private domains. Each of those retains its own business authorization.
	return nil, "", ErrCustomerServiceMessageAttachmentUnavailable
}

func (s *CustomerServiceConversationMessageAttachmentAccessService) validCustomerServiceAttachmentDerivativeStorageKey(asset *media.MediaAsset, derivative media.MediaAssetDerivative) (string, error) {
	key, ok := storage.NormalizeObjectKey(derivative.StorageKey)
	if !ok || derivative.MediaAssetID != asset.ID || !allowedPublicImageAttachmentMimeType(derivative.MimeType) {
		return "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	if asset.Visibility == "private" {
		expectedPrefix := fmt.Sprintf("%s/%d/%s/v%d/", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID, derivative.Preset, normalizeMediaDerivativeVersion(derivative.PresetVersion))
		if !storage.IsCustomerServiceConversationAttachmentDerivativeObjectKey(key) || !strings.HasPrefix(key, expectedPrefix) {
			return "", ErrCustomerServiceMessageAttachmentUnavailable
		}
	} else if storage.IsPrivateObjectKey(key) {
		return "", ErrCustomerServiceMessageAttachmentUnavailable
	}
	return key, nil
}
