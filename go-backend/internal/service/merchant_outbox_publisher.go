package service

import (
	"encoding/json"
	"fmt"
	"time"

	"tanzanite/internal/domain/outbox"
	"tanzanite/internal/repository"

	"gorm.io/datatypes"
)

type MerchantProductEventPublisher interface {
	EnqueueProductUpsert(productID uint, reason string) error
	EnqueueProductWithdraw(productID uint, reason string) error
}

type MerchantOfferEventPublisher interface {
	EnqueueOfferRevalidate(offerID uint, reason string) error
}

type MerchantCatalogEventPublisher interface {
	MerchantProductEventPublisher
	MerchantOfferEventPublisher
}

type MerchantOutboxPublisher struct {
	repo *repository.OutboxRepository
}

func NewMerchantOutboxPublisher(repo *repository.OutboxRepository) *MerchantOutboxPublisher {
	return &MerchantOutboxPublisher{repo: repo}
}

func (p *MerchantOutboxPublisher) EnqueueProductUpsert(productID uint, reason string) error {
	return p.enqueueProductEvent(outbox.EventTypeMerchantProductUpsert, productID, reason)
}

func (p *MerchantOutboxPublisher) EnqueueProductWithdraw(productID uint, reason string) error {
	return p.enqueueProductEvent(outbox.EventTypeMerchantProductWithdraw, productID, reason)
}

func (p *MerchantOutboxPublisher) EnqueueOfferRevalidate(offerID uint, reason string) error {
	if p == nil || p.repo == nil || offerID == 0 {
		return nil
	}

	payload, err := json.Marshal(outbox.MerchantOfferRevalidatePayload{
		OfferID: offerID,
		Reason:  reason,
	})
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return p.repo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d:%d", outbox.EventTypeMerchantOfferRevalidate, offerID, now.UnixNano()),
		EventType:     outbox.EventTypeMerchantOfferRevalidate,
		AggregateType: outbox.AggregateTypeMerchantOffer,
		AggregateID:   fmt.Sprint(offerID),
		Payload:       datatypes.JSON(payload),
		AvailableAt:   now,
	})
}

func (p *MerchantOutboxPublisher) enqueueProductEvent(eventType string, productID uint, reason string) error {
	if p == nil || p.repo == nil || productID == 0 {
		return nil
	}

	payload, err := json.Marshal(outbox.MerchantProductSyncPayload{
		ProductID: productID,
		Reason:    reason,
	})
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return p.repo.CreateEvent(&outbox.Event{
		EventKey:      fmt.Sprintf("%s:%d:%d", eventType, productID, now.UnixNano()),
		EventType:     eventType,
		AggregateType: outbox.AggregateTypeProduct,
		AggregateID:   fmt.Sprint(productID),
		Payload:       datatypes.JSON(payload),
		AvailableAt:   now,
	})
}
