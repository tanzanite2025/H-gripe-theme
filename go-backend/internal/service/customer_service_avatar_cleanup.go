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
}

func NewCustomerServiceAvatarCleanupHandler(
	userRepo *repository.UserRepository,
	storageSvc storage.StorageService,
) *CustomerServiceAvatarCleanupHandler {
	return &CustomerServiceAvatarCleanupHandler{
		userRepo: userRepo,
		storage:  storageSvc,
	}
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
	key, err := h.storage.ObjectKey(payload.URL)
	if err != nil || !IsCustomerServiceAvatarStorageKey(key) {
		// A malformed or foreign URL must never be passed to storage.Delete.
		// Mark it processed so a corrupted event cannot retry forever.
		return nil
	}

	current, err := h.userRepo.IsCurrentCustomerServiceAgentAvatarURL(h.storage.GetURL(key))
	if err != nil {
		return fmt.Errorf("check customer-service avatar ownership: %w", err)
	}
	if current {
		// The object became current again or an event was stale. Preserve the
		// active avatar and let this cleanup event finish harmlessly.
		return nil
	}

	if err := h.storage.Delete(ctx, payload.URL); err != nil {
		return fmt.Errorf("delete customer-service avatar: %w", err)
	}
	return nil
}
