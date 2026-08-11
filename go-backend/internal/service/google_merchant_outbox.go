package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"commerce-platform/internal/domain/merchant"
	"commerce-platform/internal/domain/outbox"
)

type GoogleMerchantOutboxHandler struct {
	service *GoogleMerchantService
}

func NewGoogleMerchantOutboxHandler(service *GoogleMerchantService) *GoogleMerchantOutboxHandler {
	return &GoogleMerchantOutboxHandler{service: service}
}

func (h *GoogleMerchantOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.service == nil {
		return errors.New("Google Merchant outbox handler is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	switch event.EventType {
	case outbox.EventTypeMerchantProductUpsert:
		var payload outbox.MerchantProductSyncPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode Merchant product upsert event: %w", err)
		}
		return h.syncProductOffers(ctx, payload.ProductID)
	case outbox.EventTypeMerchantProductWithdraw:
		var payload outbox.MerchantProductSyncPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode Merchant product withdraw event: %w", err)
		}
		return h.withdrawProductOffers(ctx, payload.ProductID)
	case outbox.EventTypeMerchantOfferRevalidate:
		var payload outbox.MerchantOfferRevalidatePayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return fmt.Errorf("decode Merchant offer revalidate event: %w", err)
		}
		return h.revalidateOffer(ctx, payload.OfferID)
	default:
		return fmt.Errorf("unsupported Merchant outbox event type %s", event.EventType)
	}
}

func (h *GoogleMerchantOutboxHandler) revalidateOffer(ctx context.Context, offerID uint) error {
	if offerID == 0 {
		return errors.New("Merchant offer revalidate event is missing offer_id")
	}
	offer, err := h.service.offers.FindOfferByID(offerID)
	if err != nil {
		return normalizeGoogleMerchantOfferError(err)
	}
	if merchantOfferShouldBeWithdrawn(offer) || offer.PublicationStatus != "ready" {
		if !googleMerchantOfferHasRemoteSubmission(offer) {
			return nil
		}
		_, err := h.service.RemoveRemoteOffer(ctx, offer.ID)
		return err
	}
	_, err = h.service.ValidateOffer(offer.ID)
	return err
}

type GoogleMerchantReconcileResult struct {
	Considered int      `json:"considered"`
	Synced     int      `json:"synced"`
	Withdrawn  int      `json:"withdrawn"`
	Skipped    int      `json:"skipped"`
	Failed     int      `json:"failed"`
	Errors     []string `json:"errors,omitempty"`
}

func (s *GoogleMerchantService) ReconcileOffers(ctx context.Context) (GoogleMerchantReconcileResult, error) {
	result := GoogleMerchantReconcileResult{Errors: []string{}}
	if s == nil || s.offers == nil {
		return result, errors.New("Google Merchant service is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	offers, err := s.offers.ListOffers()
	if err != nil {
		return result, err
	}

	for index := range offers {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		offer := &offers[index]
		result.Considered++
		if merchantOfferShouldBeWithdrawn(offer) || offer.PublicationStatus != "ready" {
			if !googleMerchantOfferHasRemoteSubmission(offer) {
				result.Skipped++
				continue
			}
			if _, err := s.RemoveRemoteOffer(ctx, offer.ID); err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("offer %d: %v", offer.ID, err))
				continue
			}
			result.Withdrawn++
			continue
		}
		if _, err := s.SyncOffer(ctx, offer.ID); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("offer %d: %v", offer.ID, err))
			continue
		}
		result.Synced++
	}

	if result.Failed > 0 {
		return result, fmt.Errorf("Google Merchant reconciliation failed for %d offers", result.Failed)
	}
	return result, nil
}

func (h *GoogleMerchantOutboxHandler) syncProductOffers(ctx context.Context, productID uint) error {
	if productID == 0 {
		return errors.New("Merchant product event is missing product_id")
	}
	offers, err := h.service.offers.ListOffersByProductID(productID)
	if err != nil {
		return err
	}
	for index := range offers {
		offer := &offers[index]
		if merchantOfferShouldBeWithdrawn(offer) {
			if !googleMerchantOfferHasRemoteSubmission(offer) {
				continue
			}
			if _, err := h.service.RemoveRemoteOffer(ctx, offer.ID); err != nil {
				return err
			}
			continue
		}
		if offer.PublicationStatus != "ready" {
			if googleMerchantOfferHasRemoteSubmission(offer) {
				if _, err := h.service.RemoveRemoteOffer(ctx, offer.ID); err != nil {
					return err
				}
			}
			continue
		}
		if _, err := h.service.SyncOffer(ctx, offer.ID); err != nil {
			return err
		}
	}
	return nil
}

func (h *GoogleMerchantOutboxHandler) withdrawProductOffers(ctx context.Context, productID uint) error {
	if productID == 0 {
		return errors.New("Merchant product event is missing product_id")
	}
	offers, err := h.service.offers.ListOffersByProductID(productID)
	if err != nil {
		return err
	}
	for index := range offers {
		offer := &offers[index]
		if !googleMerchantOfferHasRemoteSubmission(offer) {
			continue
		}
		if _, err := h.service.RemoveRemoteOffer(ctx, offer.ID); err != nil {
			return err
		}
	}
	return nil
}

func merchantOfferShouldBeWithdrawn(offer *merchant.GoogleMerchantOffer) bool {
	if offer == nil || offer.Product == nil || offer.Variant == nil {
		return true
	}
	return offer.Product.Status != "active" || !offer.Variant.IsActive
}
