package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

// OrderCompletionOutboxHandler runs post-completion settlement after the
// order status transaction has committed. A retry is harmless because the
// loyalty ledger has a database-level source/idempotency guard.
type OrderCompletionOutboxHandler struct {
	orderService    *OrderService
	templateService *TransactionalNotificationTemplateService
	sender          TransactionalNotificationSender
}

func NewOrderCompletionOutboxHandler(orderService *OrderService, templateServices ...*TransactionalNotificationTemplateService) *OrderCompletionOutboxHandler {
	var templateService *TransactionalNotificationTemplateService
	if len(templateServices) > 0 {
		templateService = templateServices[0]
	}
	return &OrderCompletionOutboxHandler{orderService: orderService, templateService: templateService}
}

func (h *OrderCompletionOutboxHandler) ConfigureTransactionalNotificationSender(sender TransactionalNotificationSender) {
	if h == nil {
		return
	}
	h.sender = sender
}

func (h *OrderCompletionOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.orderService == nil || h.orderService.txManager == nil {
		return errors.New("order completion service is not configured")
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if event.EventType != outbox.EventTypeOrderCompleted {
		return fmt.Errorf("unsupported order completion outbox event type %s", event.EventType)
	}
	var payload outbox.OrderCompletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode order completed payload: %w", err)
	}
	if payload.OrderID == 0 || strings.TrimSpace(payload.OrderNumber) == "" || payload.CompletedAt.IsZero() {
		return errors.New("order completed event is incomplete")
	}
	resolution, err := ResolveTransactionalNotificationTemplate(event.EventType, event.Payload)
	if err != nil {
		return err
	}
	if err := ValidateTransactionalNotificationTemplateResolution(resolution); err != nil {
		return err
	}
	if h.sender == nil && h.templateService != nil {
		if _, err := h.templateService.Resolve(resolution.TemplateCode, "en"); err != nil {
			return fmt.Errorf("resolve transactional notification template %q: %w", resolution.TemplateCode, err)
		}
	}
	if err := h.orderService.handleOrderCompletedOutbox(payload); err != nil {
		return err
	}
	if h.sender != nil {
		if err := deliverTransactionalNotification(ctx, event, h.templateService, h.sender); err != nil {
			return err
		}
	}
	return nil
}

func enqueueOrderCompletedOutboxEvent(repo *repository.OutboxRepository, orderRecord *order.Order, completedAt time.Time) error {
	if repo == nil {
		return errors.New("outbox repository is not configured")
	}
	if orderRecord == nil || orderRecord.ID == 0 || strings.TrimSpace(orderRecord.OrderNumber) == "" {
		return errors.New("order completed event requires a persisted order")
	}
	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	} else {
		completedAt = completedAt.UTC()
	}
	payload, err := json.Marshal(outbox.OrderCompletedPayload{
		CanonicalDomainEventMetadata: canonicalDomainEventMetadata(
			fmt.Sprintf("order_completion:%d", orderRecord.ID),
			completedAt,
		),
		NotificationAudienceSnapshot: orderNotificationAudience(orderRecord),
		OrderID:                      orderRecord.ID,
		OrderNumber:                  strings.TrimSpace(orderRecord.OrderNumber),
		CompletedAt:                  completedAt,
	})
	if err != nil {
		return fmt.Errorf("encode order completed event: %w", err)
	}
	return repo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d", outbox.EventTypeOrderCompleted, orderRecord.ID),
		EventType:     outbox.EventTypeOrderCompleted,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   strconv.FormatUint(uint64(orderRecord.ID), 10),
		Payload:       datatypes.JSON(payload),
		AvailableAt:   completedAt,
	})
}

func (s *OrderService) handleOrderCompletedOutbox(payload outbox.OrderCompletedPayload) error {
	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Order == nil {
			return errors.New("order repository is not configured")
		}
		o, err := repos.Order.FindByIDForUpdateWithItems(payload.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		if o.OrderNumber != strings.TrimSpace(payload.OrderNumber) {
			return errors.New("order completed event order number mismatch")
		}
		if o.Status != "completed" || o.PaymentStatus != "paid" || o.UserID == 0 {
			return nil
		}
		return s.awardOrderCompletionPoints(repos, o)
	})
}
