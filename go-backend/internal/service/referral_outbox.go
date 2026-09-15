package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

func enqueueReferralOrderPaidOutboxEvent(
	repo *repository.OutboxRepository,
	orderRecord *order.Order,
	paidAt time.Time,
	transactionID string,
) error {
	if repo == nil || orderRecord == nil {
		return nil
	}
	amount, err := domainmoney.FromMajorFloat(orderRecord.TotalAmount, orderRecord.Currency)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(outbox.ReferralOrderPaidPayload{
		OrderID:     orderRecord.ID,
		UserID:      orderRecord.UserID,
		AmountMinor: amount.AmountMinor(),
		Currency:    strings.ToUpper(strings.TrimSpace(orderRecord.Currency)),
		PaidAt:      paidAt.UTC(),
	})
	if err != nil {
		return err
	}
	return repo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d:%s", outbox.EventTypeReferralOrderPaid, orderRecord.ID, strings.TrimSpace(transactionID)),
		EventType:     outbox.EventTypeReferralOrderPaid,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   fmt.Sprint(orderRecord.ID),
		Payload:       datatypes.JSON(payload),
	})
}

func enqueueReferralOrderDeliveredOutboxEvent(
	repo *repository.OutboxRepository,
	orderID uint,
	deliveredAt time.Time,
	source string,
) error {
	if repo == nil || orderID == 0 {
		return nil
	}
	payload, err := json.Marshal(outbox.ReferralOrderDeliveredPayload{
		OrderID:     orderID,
		DeliveredAt: deliveredAt.UTC(),
		Source:      strings.TrimSpace(source),
	})
	if err != nil {
		return err
	}
	return repo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d", outbox.EventTypeReferralOrderDelivered, orderID),
		EventType:     outbox.EventTypeReferralOrderDelivered,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   fmt.Sprint(orderID),
		Payload:       datatypes.JSON(payload),
	})
}

func enqueueReferralOrderInvalidatedOutboxEvent(
	repo *repository.OutboxRepository,
	orderID uint,
	occurredAt time.Time,
	reason string,
	source string,
	reference string,
) error {
	if repo == nil || orderID == 0 {
		return nil
	}
	payload, err := json.Marshal(outbox.ReferralOrderInvalidatedPayload{
		OrderID:    orderID,
		OccurredAt: occurredAt.UTC(),
		Reason:     strings.TrimSpace(reason),
		Source:     strings.TrimSpace(source),
		Reference:  strings.TrimSpace(reference),
	})
	if err != nil {
		return err
	}
	eventKey := fmt.Sprintf(
		"%s:%d:%s:%s",
		outbox.EventTypeReferralOrderInvalidated,
		orderID,
		strings.TrimSpace(source),
		strings.TrimSpace(reference),
	)
	return repo.CreateEvent(&outbox.Event{
		EventKey:      eventKey,
		EventType:     outbox.EventTypeReferralOrderInvalidated,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   fmt.Sprint(orderID),
		Payload:       datatypes.JSON(payload),
	})
}
