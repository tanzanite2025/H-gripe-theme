package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	objectCleanupResourceMediaAsset      = "media_asset"
	objectCleanupResourceSiteLogo        = "site_logo"
	objectCleanupResourceSiteFavicon     = "site_favicon"
	objectCleanupResourceHomeVisualTile  = "home_visual_tile"
	objectCleanupResourceUGCShowcase     = "ugc_showcase"
	objectStorageCleanupMaxAttempts      = 1_000_000
	objectStorageCleanupAggregateIDLimit = 80
)

var ErrObjectStorageCleanupUnavailable = errors.New("durable object storage cleanup is unavailable")

func newObjectStorageCleanupEvent(
	resourceType string,
	resourceID string,
	aggregateType string,
	aggregateID string,
	objectKeys []string,
	expectedUploadObjectBytes ...map[string]int64,
) (*outbox.Event, error) {
	resourceType = strings.TrimSpace(resourceType)
	resourceID = strings.TrimSpace(resourceID)
	aggregateType = strings.TrimSpace(aggregateType)
	aggregateID = strings.TrimSpace(aggregateID)
	if resourceType == "" || resourceID == "" || aggregateType == "" || aggregateID == "" {
		return nil, fmt.Errorf("%w: cleanup identity is incomplete", ErrObjectStorageCleanupUnavailable)
	}

	keys := normalizeObjectCleanupKeys(objectKeys)
	var expectedBytes map[string]int64
	if len(expectedUploadObjectBytes) > 0 {
		expectedBytes = make(map[string]int64)
		for _, key := range keys {
			if size, known := expectedUploadObjectBytes[0][key]; known && size >= 0 {
				expectedBytes[key] = size
			}
		}
	}
	now := time.Now().UTC()
	payload, err := json.Marshal(outbox.ObjectStorageCleanupPayload{
		ResourceType:              resourceType,
		ResourceID:                resourceID,
		ObjectKeys:                keys,
		RequestedAt:               now,
		ExpectedUploadObjectBytes: expectedBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("encode object storage cleanup event: %w", err)
	}

	identity := resourceType + "\x00" + resourceID + "\x00" + strings.Join(keys, "\x00")
	digest := sha256.Sum256([]byte(identity))
	if len(aggregateID) > objectStorageCleanupAggregateIDLimit {
		aggregateID = fmt.Sprintf("%x", sha256.Sum256([]byte(aggregateID)))
	}
	return &outbox.Event{
		EventKey:      fmt.Sprintf("%s:%x", outbox.EventTypeObjectStorageCleanup, digest[:]),
		EventType:     outbox.EventTypeObjectStorageCleanup,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Payload:       datatypes.JSON(payload),
		MaxAttempts:   objectStorageCleanupMaxAttempts,
		AvailableAt:   now,
	}, nil
}

func normalizeObjectCleanupKeys(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		key, ok := storage.NormalizeObjectKey(value)
		if !ok {
			continue
		}
		unique[key] = struct{}{}
	}
	keys := make([]string, 0, len(unique))
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// ObjectStorageCleanupOutboxHandler re-checks domain ownership before every
// delete. This makes delayed events safe when an object key becomes current or
// referenced again before the worker receives it.
type ObjectStorageCleanupOutboxHandler struct {
	suggestionAndPayPalAttachments  *SuggestionAndPayPalAttachmentUploadService
	privateOrderBusinessAttachments *PrivateOrderBusinessAttachmentUploadService
	media                           *MediaService
	siteLogo                        *SiteLogoService
	siteFavicon                     *SiteFaviconService
	homeVisual                      *HomeVisualTileService
	showcase                        *UGCShowcaseService
	orderEvidenceAttachment         *OrderEvidenceAttachmentService
}

func NewObjectStorageCleanupOutboxHandler(
	mediaService *MediaService,
	siteLogoService *SiteLogoService,
	homeVisualService *HomeVisualTileService,
	showcaseService *UGCShowcaseService,
) *ObjectStorageCleanupOutboxHandler {
	return &ObjectStorageCleanupOutboxHandler{
		media:      mediaService,
		siteLogo:   siteLogoService,
		homeVisual: homeVisualService,
		showcase:   showcaseService,
	}
}

func (h *ObjectStorageCleanupOutboxHandler) ConfigureSiteFaviconService(service *SiteFaviconService) {
	if h != nil {
		h.siteFavicon = service
	}
}

func (h *ObjectStorageCleanupOutboxHandler) ConfigureOrderEvidenceAttachmentService(service *OrderEvidenceAttachmentService) {
	if h != nil {
		h.orderEvidenceAttachment = service
	}
}

func (h *ObjectStorageCleanupOutboxHandler) ConfigurePrivateOrderBusinessAttachmentUploads(uploads *PrivateOrderBusinessAttachmentUploadService) {
	if h != nil {
		h.privateOrderBusinessAttachments = uploads
	}
}

func (h *ObjectStorageCleanupOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if event.EventType != outbox.EventTypeObjectStorageCleanup {
		return fmt.Errorf("unsupported object storage cleanup event %s", event.EventType)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	var payload outbox.ObjectStorageCleanupPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode object storage cleanup event: %w", err)
	}
	payload.ResourceType = strings.TrimSpace(payload.ResourceType)
	payload.ResourceID = strings.TrimSpace(payload.ResourceID)
	payload.ObjectKeys = normalizeObjectCleanupKeys(payload.ObjectKeys)

	switch payload.ResourceType {
	case outbox.SuggestionAndPayPalAttachmentCleanupResourceType:
		return h.suggestionAndPayPalAttachments.cleanupSuggestionOrPayPalAttachmentFromOriginalReceipt(ctx, event)
	case repository.PrivateOrderBusinessAttachmentFailedUploadResourceType:
		return h.privateOrderBusinessAttachments.cleanupPrivateOrderBusinessAttachmentFailedBatchWithRetainedReferences(ctx, payload.ResourceID, payload.ObjectKeys)
	case objectCleanupResourceExpiredBusinessDraftMediaUpload:
		if h.media == nil {
			return ErrObjectStorageCleanupUnavailable
		}
		return h.media.cleanupExpiredBusinessDraftMediaObjectsFromDurableReceipt(ctx, payload.ResourceID, payload.ObjectKeys)
	case objectCleanupResourceFailedOrderEvidenceUpload:
		return h.orderEvidenceAttachment.cleanupFailedOrderEvidenceUploadWithImmutableHistoryRetention(ctx, payload.ResourceID, payload.ObjectKeys)
	case objectCleanupResourceMediaAsset:
		return h.cleanupMediaAssetFromOriginalDurableReceipt(ctx, payload.ResourceID, payload.ObjectKeys)
	case objectCleanupResourceUnregisteredMediaUpload, objectCleanupResourceUnregisteredMediaDerivative, objectCleanupResourceRetiredMediaDerivative:
		return h.media.cleanupUnregisteredOrRetiredMediaObjectFromDurableEvent(ctx, payload.ResourceType, payload.ResourceID, payload.ObjectKeys)
	case objectCleanupResourceCustomerServiceUnregisteredUpload:
		return h.media.cleanupRetiredUnregisteredCustomerServiceUploadOriginal(ctx, payload.ObjectKeys)
	case objectCleanupResourceSiteLogo:
		return h.cleanupSiteLogo(ctx, payload.ObjectKeys)
	case objectCleanupResourceSiteFavicon:
		return h.cleanupSiteFavicon(ctx, payload.ObjectKeys)
	case objectCleanupResourceHomeVisualTile:
		return h.cleanupHomeVisualTiles(ctx, payload.ObjectKeys)
	case objectCleanupResourceFailedUGCShowcaseUpload:
		if h.showcase == nil {
			return ErrObjectStorageCleanupUnavailable
		}
		_, err := h.showcase.cleanupUGCShowcaseObjectsWithRetainedReferencesAndLockedRejectedOwner(ctx, 0, payload.ObjectKeys, true)
		return err
	case objectCleanupResourceUGCShowcase:
		return h.cleanupShowcase(ctx, payload.ResourceID, payload.ObjectKeys)
	default:
		return fmt.Errorf("unsupported object storage cleanup resource %q", payload.ResourceType)
	}
}

func (h *ObjectStorageCleanupOutboxHandler) cleanupSiteFavicon(ctx context.Context, keys []string) error {
	if h == nil || h.siteFavicon == nil || h.siteFavicon.repo == nil || h.siteFavicon.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	return h.siteFavicon.repo.WithLockedCurrentSiteFaviconForObjectCleanup(ctx, func(current *sitefavicondomain.Asset, references *repository.MediaRepository) error {
		for _, key := range keys {
			if !IsSiteFaviconStorageKey(key) || (current != nil && current.StorageKey == key) {
				continue
			}
			if err := checkDedicatedPublicObjectBusinessReferencesBeforePhysicalCleanup(references, h.siteFavicon.storage, h.siteFavicon.siteURL, key); err != nil {
				return err
			}
			if err := h.siteFavicon.storage.Delete(ctx, key); err != nil {
				return fmt.Errorf("delete retired site favicon %s: %w", key, err)
			}
		}
		return nil
	})
}

func (h *ObjectStorageCleanupOutboxHandler) cleanupMediaAssetFromOriginalDurableReceipt(ctx context.Context, resourceID string, keys []string) error {
	if h == nil || h.media == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	assetID, err := strconv.ParseUint(resourceID, 10, 64)
	if err != nil || assetID == 0 || uint64(uint(assetID)) != assetID {
		return fmt.Errorf("invalid media asset cleanup id %q", resourceID)
	}
	return h.media.cleanupDeletedAsset(ctx, uint(assetID), keys)
}

func (h *ObjectStorageCleanupOutboxHandler) cleanupSiteLogo(ctx context.Context, keys []string) error {
	if h == nil || h.siteLogo == nil || h.siteLogo.repo == nil || h.siteLogo.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	return h.siteLogo.repo.WithLockedCurrentSiteLogoForObjectCleanup(ctx, func(current *sitelogodomain.Asset, references *repository.MediaRepository) error {
		for _, key := range keys {
			if !IsSiteLogoStorageKey(key) || (current != nil && strings.TrimSpace(current.StorageKey) == key) {
				continue
			}
			if err := checkDedicatedPublicObjectBusinessReferencesBeforePhysicalCleanup(references, h.siteLogo.storage, h.siteLogo.siteURL, key); err != nil {
				return err
			}
			if err := h.siteLogo.storage.Delete(ctx, key); err != nil {
				return fmt.Errorf("delete retired site logo %s: %w", key, err)
			}
		}
		return nil
	})
}

func (h *ObjectStorageCleanupOutboxHandler) cleanupHomeVisualTiles(ctx context.Context, keys []string) error {
	if h == nil || h.homeVisual == nil || h.homeVisual.repo == nil || h.homeVisual.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	return h.homeVisual.cleanupUnreferencedDedicatedVisualShowcaseObjectsUnderTransactionLock(ctx, keys)
}

func (h *ObjectStorageCleanupOutboxHandler) cleanupShowcase(ctx context.Context, resourceID string, keys []string) error {
	if h == nil || h.showcase == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	ownerID, err := strconv.ParseUint(resourceID, 10, 64)
	if err != nil || ownerID == 0 || uint64(uint(ownerID)) != ownerID {
		return fmt.Errorf("invalid showcase cleanup owner identity")
	}
	_, err = h.showcase.cleanupUGCShowcaseObjectsWithRetainedReferencesAndLockedRejectedOwner(ctx, uint(ownerID), keys, false)
	return err
}

func isRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
