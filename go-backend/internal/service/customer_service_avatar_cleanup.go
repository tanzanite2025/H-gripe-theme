package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

type CustomerServiceAvatarCleanupHandler struct {
	userRepo *repository.UserRepository
	storage  storage.StorageService
	siteURL  string
}

func NewCustomerServiceAvatarCleanupHandler(
	userRepo *repository.UserRepository,
	storageSvc storage.StorageService,
	siteURLs ...string,
) *CustomerServiceAvatarCleanupHandler {
	handler := &CustomerServiceAvatarCleanupHandler{
		userRepo: userRepo,
		storage:  storageSvc,
	}
	if len(siteURLs) > 0 {
		handler.siteURL = strings.TrimSpace(siteURLs[0])
	}
	return handler
}

func (h *CustomerServiceAvatarCleanupHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.userRepo == nil || h.storage == nil {
		return errors.New("customer-service avatar cleanup handler is not configured")
	}
	if event.EventType != outbox.EventTypeCustomerServiceAvatarCleanup {
		return fmt.Errorf("unsupported customer-service avatar cleanup event %s", event.EventType)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	var payload outbox.CustomerServiceAvatarCleanupPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode customer-service avatar cleanup event: %w", err)
	}
	payload.URL = strings.TrimSpace(payload.URL)
	key, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(h.storage, CustomerServiceAvatarStoragePrefix, payload.URL)
	if err != nil {
		// A malformed or foreign URL must never be passed to storage.Delete.
		// Mark it processed so a corrupted event cannot retry forever.
		return nil
	}

	return h.userRepo.WithLockedCustomerServiceAvatarNamespaceForPhysicalCleanup(ctx, func(repo *repository.MediaRepository) error {
		repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(&MediaService{storage: h.storage, siteURL: h.siteURL})
		references, err := repo.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, h.storage.GetURL(key), 0)
		if err != nil {
			return fmt.Errorf("check retained customer-service avatar references: %w", err)
		}
		if len(references) > 0 {
			return &MediaAssetInUseError{References: references}
		}
		if err := h.storage.Delete(ctx, payload.URL); err != nil {
			return fmt.Errorf("delete customer-service avatar: %w", err)
		}
		return nil
	})
}
