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
	ugcshowcasedomain "commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/pkg/storage"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	objectCleanupResourceMediaAsset      = "media_asset"
	objectCleanupResourceSiteLogo        = "site_logo"
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
) (*outbox.Event, error) {
	resourceType = strings.TrimSpace(resourceType)
	resourceID = strings.TrimSpace(resourceID)
	aggregateType = strings.TrimSpace(aggregateType)
	aggregateID = strings.TrimSpace(aggregateID)
	if resourceType == "" || resourceID == "" || aggregateType == "" || aggregateID == "" {
		return nil, fmt.Errorf("%w: cleanup identity is incomplete", ErrObjectStorageCleanupUnavailable)
	}

	keys := normalizeObjectCleanupKeys(objectKeys)
	now := time.Now().UTC()
	payload, err := json.Marshal(outbox.ObjectStorageCleanupPayload{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ObjectKeys:   keys,
		RequestedAt:  now,
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
	media      *MediaService
	siteLogo   *SiteLogoService
	homeVisual *HomeVisualTileService
	showcase   *UGCShowcaseService
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
	case objectCleanupResourceMediaAsset:
		return h.cleanupMediaAsset(ctx, payload.ResourceID)
	case objectCleanupResourceSiteLogo:
		return h.cleanupSiteLogo(ctx, payload.ObjectKeys)
	case objectCleanupResourceHomeVisualTile:
		return h.cleanupHomeVisualTiles(ctx, payload.ObjectKeys)
	case objectCleanupResourceUGCShowcase:
		return h.cleanupShowcase(ctx, payload.ResourceID, payload.ObjectKeys)
	default:
		return fmt.Errorf("unsupported object storage cleanup resource %q", payload.ResourceType)
	}
}

func (h *ObjectStorageCleanupOutboxHandler) cleanupMediaAsset(ctx context.Context, resourceID string) error {
	if h == nil || h.media == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	assetID, err := strconv.ParseUint(resourceID, 10, 64)
	if err != nil || assetID == 0 || uint64(uint(assetID)) != assetID {
		return fmt.Errorf("invalid media asset cleanup id %q", resourceID)
	}
	return h.media.cleanupDeletedAsset(ctx, uint(assetID))
}

func (h *ObjectStorageCleanupOutboxHandler) cleanupSiteLogo(ctx context.Context, keys []string) error {
	if h == nil || h.siteLogo == nil || h.siteLogo.repo == nil || h.siteLogo.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	current, err := h.siteLogo.repo.Current()
	if err != nil {
		return fmt.Errorf("load current site logo before cleanup: %w", err)
	}
	for _, key := range keys {
		if !IsSiteLogoStorageKey(key) {
			continue
		}
		if current != nil && strings.TrimSpace(current.StorageKey) == key {
			continue
		}
		if err := h.siteLogo.storage.Delete(ctx, key); err != nil {
			return fmt.Errorf("delete retired site logo %s: %w", key, err)
		}
	}
	return nil
}

func (h *ObjectStorageCleanupOutboxHandler) cleanupHomeVisualTiles(ctx context.Context, keys []string) error {
	if h == nil || h.homeVisual == nil || h.homeVisual.repo == nil || h.homeVisual.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	for _, key := range keys {
		if !IsHomeVisualTileStorageKey(key) {
			continue
		}
		count, err := h.homeVisual.repo.CountItemsByStorageKey(key)
		if err != nil {
			return fmt.Errorf("check home visual tile references for %s: %w", key, err)
		}
		if count > 0 {
			continue
		}
		if err := h.homeVisual.storage.Delete(ctx, key); err != nil {
			return fmt.Errorf("delete retired home visual tile %s: %w", key, err)
		}
	}
	return nil
}

func (h *ObjectStorageCleanupOutboxHandler) cleanupShowcase(ctx context.Context, resourceID string, keys []string) error {
	if h == nil || h.showcase == nil || h.showcase.repo == nil || h.showcase.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	ownerID, err := strconv.ParseUint(resourceID, 10, 64)
	if err != nil || ownerID == 0 || uint64(uint(ownerID)) != ownerID {
		return fmt.Errorf("invalid showcase cleanup id %q", resourceID)
	}
	owner, ownerErr := h.showcase.repo.GetByID(uint(ownerID))
	if ownerErr != nil && !isRecordNotFound(ownerErr) {
		return fmt.Errorf("load showcase owner before cleanup: %w", ownerErr)
	}
	for _, key := range keys {
		if !showcaseStorageKeyIsPending(key) {
			continue
		}
		items, err := h.showcase.repo.ListByImageStorageKeyCandidate(key)
		if err != nil {
			return fmt.Errorf("check showcase image references for %s: %w", key, err)
		}
		stillReferenced := false
		for _, item := range items {
			matches, matchErr := h.showcase.showcaseImageMatchesStorageKey(item, key)
			if matchErr != nil {
				return matchErr
			}
			if !matches {
				continue
			}
			if item.ID == uint(ownerID) && item.Status == ugcshowcasedomain.StatusRejected {
				continue
			}
			stillReferenced = true
			break
		}
		if stillReferenced {
			continue
		}
		if err := h.showcase.storage.Delete(ctx, key); err != nil {
			return fmt.Errorf("delete retired showcase image %s: %w", key, err)
		}
		if owner != nil && owner.Status == ugcshowcasedomain.StatusRejected {
			if err := h.showcase.pruneShowcaseImageReference(owner, key); err != nil {
				return fmt.Errorf("prune retired showcase image reference %s: %w", key, err)
			}
		}
	}
	return nil
}

// pruneShowcaseImageReference removes a successfully cleaned pending object
// from the rejected owner's JSON image list. It is intentionally idempotent:
// duplicate deliveries and already-pruned references are harmless.
func (s *UGCShowcaseService) pruneShowcaseImageReference(item *ugcshowcasedomain.UGCShowcase, key string) error {
	if s == nil || s.repo == nil || item == nil {
		return nil
	}
	imageReferences, err := decodeShowcaseImageURLs(item.Images)
	if err != nil {
		return err
	}
	remaining := make([]string, 0, len(imageReferences))
	removed := false
	for _, reference := range imageReferences {
		imageKey, keyErr := s.storage.ObjectKey(reference)
		if keyErr == nil && imageKey == key {
			removed = true
			continue
		}
		remaining = append(remaining, reference)
	}
	if !removed {
		return nil
	}
	imagesJSON, err := json.Marshal(remaining)
	if err != nil {
		return fmt.Errorf("encode remaining showcase images: %w", err)
	}
	updated, err := s.repo.UpdateImagesByStatus(item.ID, ugcshowcasedomain.StatusRejected, datatypes.JSON(imagesJSON))
	if err != nil {
		return err
	}
	if updated {
		item.Images = datatypes.JSON(imagesJSON)
	}
	return nil
}

func isRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
