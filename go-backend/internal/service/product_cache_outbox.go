package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

type ProductCacheEventPublisher interface {
	EnqueueProductCacheInvalidateByIDs(productIDs []uint, reason string) error
	EnqueueProductCacheInvalidateByProductSpecificationTemplateID(productSpecificationTemplateID uint, reason string) error
	EnqueueProductCacheInvalidateByBrandID(brandID uint, reason string) error
	EnqueueProductCacheInvalidateByInformationTemplateID(templateID uint, reason string) error
}

type ProductCacheOutboxPublisher struct {
	repo *repository.OutboxRepository
}

type ProductCacheOutboxHandler struct {
	invalidator ProductCacheInvalidationExecutor
}

func NewProductCacheOutboxPublisher(repo *repository.OutboxRepository) *ProductCacheOutboxPublisher {
	return &ProductCacheOutboxPublisher{repo: repo}
}

func (p *ProductCacheOutboxPublisher) WithRepository(repo *repository.OutboxRepository) *ProductCacheOutboxPublisher {
	return NewProductCacheOutboxPublisher(repo)
}

func NewProductCacheOutboxHandler(invalidator ProductCacheInvalidationExecutor) *ProductCacheOutboxHandler {
	return &ProductCacheOutboxHandler{invalidator: invalidator}
}

func (p *ProductCacheOutboxPublisher) EnqueueProductCacheInvalidateByIDs(productIDs []uint, reason string) error {
	ids := uniqueUintIDs(productIDs)
	if len(ids) == 0 {
		return nil
	}
	return p.enqueueProductCacheInvalidate(
		outbox.AggregateTypeProductCache,
		productCacheEventAggregateID("products", ids),
		reason,
		outbox.ProductCacheInvalidatePayload{ProductIDs: ids, Reason: reason},
	)
}

func (p *ProductCacheOutboxPublisher) EnqueueProductCacheInvalidateByProductSpecificationTemplateID(productSpecificationTemplateID uint, reason string) error {
	if productSpecificationTemplateID == 0 {
		return nil
	}
	return p.enqueueProductCacheInvalidate(
		outbox.AggregateTypeProductSpecificationTemplate,
		strconv.FormatUint(uint64(productSpecificationTemplateID), 10),
		reason,
		outbox.ProductCacheInvalidatePayload{ProductSpecificationTemplateID: productSpecificationTemplateID, Reason: reason},
	)
}

func (p *ProductCacheOutboxPublisher) EnqueueProductCacheInvalidateByBrandID(brandID uint, reason string) error {
	if brandID == 0 {
		return nil
	}
	return p.enqueueProductCacheInvalidate(
		outbox.AggregateTypeProductBrand,
		strconv.FormatUint(uint64(brandID), 10),
		reason,
		outbox.ProductCacheInvalidatePayload{ProductBrandID: brandID, Reason: reason},
	)
}

func (p *ProductCacheOutboxPublisher) EnqueueProductCacheInvalidateByInformationTemplateID(templateID uint, reason string) error {
	if templateID == 0 {
		return nil
	}
	return p.enqueueProductCacheInvalidate(
		outbox.AggregateTypeInformationTemplate,
		strconv.FormatUint(uint64(templateID), 10),
		reason,
		outbox.ProductCacheInvalidatePayload{ProductInformationTemplateID: templateID, Reason: reason},
	)
}

func (p *ProductCacheOutboxPublisher) enqueueProductCacheInvalidate(aggregateType, aggregateID, reason string, payload outbox.ProductCacheInvalidatePayload) error {
	if p == nil || p.repo == nil || aggregateID == "" {
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return p.repo.CreateEvent(&outbox.Event{
		EventKey:      productCacheInvalidateEventKey(aggregateType, aggregateID, reason, now),
		EventType:     outbox.EventTypeProductCacheInvalidate,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Payload:       datatypes.JSON(data),
		AvailableAt:   now,
	})
}

func (h *ProductCacheOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.invalidator == nil {
		return errors.New("product cache outbox handler is not configured")
	}
	if event.EventType != outbox.EventTypeProductCacheInvalidate {
		return fmt.Errorf("unsupported product cache outbox event type %s", event.EventType)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	var payload outbox.ProductCacheInvalidatePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode product cache invalidate event: %w", err)
	}

	switch {
	case len(payload.ProductIDs) > 0:
		_, err := h.invalidator.InvalidateProductCacheByIDsWithSource(payload.ProductIDs, productCacheInvalidationSourceOutbox)
		return err
	case payload.ProductSpecificationTemplateID > 0:
		_, err := h.invalidator.InvalidateProductCacheByProductSpecificationTemplateIDWithSource(payload.ProductSpecificationTemplateID, productCacheInvalidationSourceOutbox)
		return err
	case payload.ProductBrandID > 0:
		_, err := h.invalidator.InvalidateProductCacheByBrandIDWithSource(payload.ProductBrandID, productCacheInvalidationSourceOutbox)
		return err
	case payload.ProductInformationTemplateID > 0:
		_, err := h.invalidator.InvalidateProductCacheByInformationTemplateIDWithSource(payload.ProductInformationTemplateID, productCacheInvalidationSourceOutbox)
		return err
	default:
		return errors.New("product cache invalidate event has no target")
	}
}

func productCacheInvalidateEventKey(aggregateType, aggregateID, reason string, now time.Time) string {
	return fmt.Sprintf(
		"%s:%s:%s:%s:%d",
		outbox.EventTypeProductCacheInvalidate,
		strings.TrimSpace(aggregateType),
		strings.TrimSpace(aggregateID),
		normalizeProductCacheEventReason(reason),
		now.Unix(),
	)
}

func productCacheEventAggregateID(scope string, ids []uint) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		parts = append(parts, strconv.FormatUint(uint64(id), 10))
	}
	return fmt.Sprintf("%s:%s", strings.TrimSpace(scope), strings.Join(parts, ","))
}

func normalizeProductCacheEventReason(reason string) string {
	reason = strings.ToLower(strings.TrimSpace(reason))
	if reason == "" {
		return "unspecified"
	}
	replacer := strings.NewReplacer(" ", "_", ":", "_", "/", "_", "\\", "_")
	return replacer.Replace(reason)
}
