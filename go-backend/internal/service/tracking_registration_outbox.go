package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
	"gorm.io/datatypes"
)

// TrackingShipmentRegistrationOutboxHandler performs the external tracking
// provider registration after the fulfillment transaction has committed.
type TrackingShipmentRegistrationOutboxHandler struct {
	shippingService *ShippingService
}

const trackingRegistrationPayloadVersion = 1

type trackingRegistrationFingerprintInput struct {
	ProviderID               uint   `json:"provider_id"`
	TrackingNumber           string `json:"tracking_number"`
	ProviderCarrierCode      string `json:"provider_carrier_code"`
	CarrierID                *uint  `json:"carrier_id,omitempty"`
	CarrierServiceID         *uint  `json:"carrier_service_id,omitempty"`
	TrackingCarrierMappingID *uint  `json:"tracking_carrier_mapping_id,omitempty"`
}

func trackingRegistrationSourceFingerprint(input TrackingShipmentInput) string {
	canonical, _ := json.Marshal(trackingRegistrationFingerprintInput{
		ProviderID:               input.TrackingProviderID,
		TrackingNumber:           strings.TrimSpace(input.TrackingNumber),
		ProviderCarrierCode:      strings.TrimSpace(input.ProviderCarrierCode),
		CarrierID:                input.CarrierID,
		CarrierServiceID:         input.CarrierServiceID,
		TrackingCarrierMappingID: input.TrackingCarrierMappingID,
	})
	sum := sha256.Sum256(canonical)
	return fmt.Sprintf("%x", sum[:])
}

func NewTrackingShipmentRegistrationOutboxHandler(shippingService *ShippingService) *TrackingShipmentRegistrationOutboxHandler {
	return &TrackingShipmentRegistrationOutboxHandler{shippingService: shippingService}
}

func (h *TrackingShipmentRegistrationOutboxHandler) Handle(ctx context.Context, event outbox.Event) error {
	if h == nil || h.shippingService == nil {
		return errors.New("tracking registration service is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if event.EventType != outbox.EventTypeTrackingShipmentRegistration {
		return fmt.Errorf("unsupported tracking registration outbox event type %s", event.EventType)
	}
	var payload outbox.TrackingShipmentRegistrationPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode tracking registration event: %w", err)
	}
	if payload.OrderID == 0 || payload.TrackingProviderID == 0 || strings.TrimSpace(payload.TrackingNumber) == "" || strings.TrimSpace(payload.ProviderCarrierCode) == "" || payload.RequestedAt.IsZero() {
		return errors.New("tracking registration event is incomplete")
	}
	if payload.Version != trackingRegistrationPayloadVersion || strings.TrimSpace(payload.SourceFingerprint) == "" {
		return errors.New("tracking registration event version or fingerprint is invalid")
	}
	shipment, err := h.shippingService.GetTrackingShipmentByOrderIDAndTrackingNumber(payload.OrderID, strings.TrimSpace(payload.TrackingNumber))
	if err != nil {
		return err
	}
	currentInput := TrackingShipmentInput{
		OrderID:                  payload.OrderID,
		TrackingProviderID:       shipment.TrackingProviderID,
		TrackingNumber:           shipment.TrackingNumber,
		ProviderCarrierCode:      shipment.ProviderCarrierCode,
		CarrierID:                shipment.CarrierID,
		CarrierServiceID:         shipment.CarrierServiceID,
		TrackingCarrierMappingID: shipment.TrackingCarrierMappingID,
	}
	if trackingRegistrationSourceFingerprint(currentInput) != strings.TrimSpace(payload.SourceFingerprint) {
		// A corrected shipment has its own event. The old event is complete and
		// must not register stale carrier data.
		return nil
	}
	return h.shippingService.RegisterTrackingShipment(ctx, TrackingSyncInput{
		OrderID:                  payload.OrderID,
		ProviderID:               payload.TrackingProviderID,
		TrackingNumber:           strings.TrimSpace(payload.TrackingNumber),
		ProviderCarrierCode:      strings.TrimSpace(payload.ProviderCarrierCode),
		CarrierID:                payload.CarrierID,
		CarrierServiceID:         payload.CarrierServiceID,
		TrackingCarrierMappingID: payload.TrackingCarrierMappingID,
	})
}

func enqueueTrackingShipmentRegistrationOutboxEvent(
	repo *repository.OutboxRepository,
	input TrackingShipmentInput,
	requestedAt time.Time,
) error {
	if repo == nil {
		return errors.New("outbox repository is not configured")
	}
	if input.OrderID == 0 || input.TrackingProviderID == 0 || strings.TrimSpace(input.TrackingNumber) == "" || strings.TrimSpace(input.ProviderCarrierCode) == "" {
		return errors.New("tracking registration requires order, provider, tracking number, and carrier")
	}
	if requestedAt.IsZero() {
		requestedAt = time.Now().UTC()
	} else {
		requestedAt = requestedAt.UTC()
	}
	payload, err := json.Marshal(outbox.TrackingShipmentRegistrationPayload{
		Version:                  trackingRegistrationPayloadVersion,
		OrderID:                  input.OrderID,
		TrackingProviderID:       input.TrackingProviderID,
		TrackingNumber:           strings.TrimSpace(input.TrackingNumber),
		ProviderCarrierCode:      strings.TrimSpace(input.ProviderCarrierCode),
		CarrierID:                input.CarrierID,
		CarrierServiceID:         input.CarrierServiceID,
		TrackingCarrierMappingID: input.TrackingCarrierMappingID,
		SourceFingerprint:        trackingRegistrationSourceFingerprint(input),
		RequestedAt:              requestedAt,
	})
	if err != nil {
		return fmt.Errorf("encode tracking registration event: %w", err)
	}
	fingerprint := trackingRegistrationSourceFingerprint(input)
	return repo.CreateEvent(&outbox.Event{
		// Keep the unique key bounded even when a carrier emits a long tracking
		// number. The payload remains the source of truth for the command data.
		EventKey:      fmt.Sprintf("%s:%d:%d:%s", outbox.EventTypeTrackingShipmentRegistration, input.OrderID, input.TrackingProviderID, fingerprint[:32]),
		EventType:     outbox.EventTypeTrackingShipmentRegistration,
		AggregateType: outbox.AggregateTypeOrder,
		AggregateID:   fmt.Sprintf("%d", input.OrderID),
		Payload:       datatypes.JSON(payload),
		AvailableAt:   requestedAt,
	})
}
