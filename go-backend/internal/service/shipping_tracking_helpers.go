package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/resilience"
	"commerce-platform/internal/pkg/tracking"
	"commerce-platform/internal/repository"
)

func (s *ShippingService) ensureTrackingShipmentForSync(input TrackingSyncInput, trackingNumber string, providerCarrierCode string) (*shipping.TrackingShipment, error) {
	existing, err := s.GetTrackingShipmentByOrderID(input.OrderID)
	if err == nil && existing.TrackingProviderID == input.ProviderID && existing.TrackingNumber == trackingNumber && existing.ProviderCarrierCode == providerCarrierCode {
		return existing, nil
	}
	if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}

	carrierID := input.CarrierID
	carrierServiceID := input.CarrierServiceID
	trackingCarrierMappingID := input.TrackingCarrierMappingID
	if err == nil {
		if !hasPositiveID(carrierID) {
			carrierID = existing.CarrierID
		}
		if !hasPositiveID(carrierServiceID) {
			carrierServiceID = existing.CarrierServiceID
		}
		if !hasPositiveID(trackingCarrierMappingID) {
			trackingCarrierMappingID = existing.TrackingCarrierMappingID
		}
	}

	return s.UpsertTrackingShipment(TrackingShipmentInput{
		OrderID:                  input.OrderID,
		TrackingProviderID:       input.ProviderID,
		TrackingNumber:           trackingNumber,
		ProviderCarrierCode:      providerCarrierCode,
		CarrierID:                carrierID,
		CarrierServiceID:         carrierServiceID,
		TrackingCarrierMappingID: trackingCarrierMappingID,
	})
}

func latestTrackingEventTime(events []shipping.TrackingEvent) *time.Time {
	var latest time.Time
	for _, event := range events {
		if event.EventTime.After(latest) {
			latest = event.EventTime
		}
	}
	if latest.IsZero() {
		return nil
	}
	return &latest
}

func (s *ShippingService) updateOrderShippingStatusIfDelivered(
	orderID uint,
	trackingNumber string,
	providerCarrierCode string,
	status string,
	statusCode int,
	events []shipping.TrackingEvent,
	source string,
) error {
	if s == nil || s.orderRepo == nil || !trackingStatusIndicatesDelivery(status, statusCode, events) {
		return nil
	}

	deliveredAt := referralDeliveryTime(events)
	var currentOrder *order.Order
	var updated bool
	var err error
	if s.txManager != nil {
		err = s.txManager.WithinTx(func(repos repository.TxRepositories) error {
			if repos.Order == nil {
				return errors.New("order repository is not configured")
			}
			currentOrder, err = repos.Order.FindByIDForUpdate(orderID)
			if err != nil {
				return err
			}
			updated, err = repos.Order.MarkDeliveredAtIfNeeded(orderID, deliveredAt)
			if err != nil || !updated {
				return err
			}
			return enqueueReferralOrderDeliveredOutboxEvent(repos.Outbox, orderID, deliveredAt, source)
		})
	} else {
		currentOrder, err = s.orderRepo.FindByIDBasic(orderID)
		if err == nil {
			updated, err = s.orderRepo.MarkDeliveredAtIfNeeded(orderID, deliveredAt)
		}
	}
	if err != nil {
		return err
	}
	if !updated {
		return nil
	}
	oldShippingStatus := strings.TrimSpace(currentOrder.ShippingStatus)

	recordServiceAudit(s.auditRecorder, serviceAuditEvent{
		Action:     "execute",
		Resource:   "order_delivery",
		ResourceID: orderID,
		Status:     "success",
		Changes: map[string]interface{}{
			"tracking_number":       strings.TrimSpace(trackingNumber),
			"provider_carrier_code": strings.TrimSpace(providerCarrierCode),
			"tracking_status":       deliveryTrackingStatusForAudit(status, statusCode, events),
			"status_code":           statusCode,
			"source":                strings.TrimSpace(source),
		},
		OldValue: map[string]interface{}{
			"order_number":    strings.TrimSpace(currentOrder.OrderNumber),
			"shipping_status": oldShippingStatus,
		},
		NewValue: map[string]interface{}{
			"order_number":    strings.TrimSpace(currentOrder.OrderNumber),
			"shipping_status": "delivered",
		},
	})
	return nil
}

func referralDeliveryTime(events []shipping.TrackingEvent) time.Time {
	var deliveredAt time.Time
	for _, event := range events {
		if !trackingStatusTextIndicatesDelivery(event.Status) || event.EventTime.IsZero() {
			continue
		}
		if deliveredAt.IsZero() || event.EventTime.Before(deliveredAt) {
			deliveredAt = event.EventTime
		}
	}
	if deliveredAt.IsZero() {
		if latest := latestTrackingEventTime(events); latest != nil {
			deliveredAt = *latest
		}
	}
	if deliveredAt.IsZero() {
		deliveredAt = time.Now().UTC()
	}
	return deliveredAt.UTC()
}

func deliveryTrackingStatusForAudit(status string, statusCode int, events []shipping.TrackingEvent) string {
	if normalized := strings.TrimSpace(status); normalized != "" {
		return normalized
	}

	for _, event := range events {
		if trackingStatusTextIndicatesDelivery(event.Status) {
			return strings.TrimSpace(event.Status)
		}
	}

	if statusCode == 4 {
		return "delivered"
	}

	for _, event := range events {
		if normalized := strings.TrimSpace(event.Status); normalized != "" {
			return normalized
		}
	}
	return ""
}

func trackingStatusIndicatesDelivery(status string, statusCode int, events []shipping.TrackingEvent) bool {
	if statusCode == 4 || trackingStatusTextIndicatesDelivery(status) {
		return true
	}

	for _, event := range events {
		if trackingStatusTextIndicatesDelivery(event.Status) {
			return true
		}
	}
	return false
}

func trackingStatusTextIndicatesDelivery(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	return strings.Contains(normalized, "delivered") ||
		strings.Contains(normalized, "signed") ||
		strings.Contains(normalized, "签收") ||
		strings.Contains(normalized, "妥投")
}

func nextTrackingSyncAt(provider *shipping.TrackingProviderConfig, now time.Time) *time.Time {
	if provider == nil || !provider.PollingEnabled {
		return nil
	}

	intervalMinutes := provider.PollingIntervalMinutes
	if intervalMinutes <= 0 {
		intervalMinutes = defaultTrackingPollingIntervalMinutes
	}

	next := now.Add(time.Duration(intervalMinutes) * time.Minute)
	return &next
}

func newTrackingClientFromProvider(provider *shipping.TrackingProviderConfig) (tracking.TrackingService, error) {
	return newTrackingClientFromProviderWithResilience(provider, resilience.HTTPRetryPolicy{}, nil)
}

func (s *ShippingService) newTrackingClientFromProvider(provider *shipping.TrackingProviderConfig) (tracking.TrackingService, error) {
	if s == nil {
		return newTrackingClientFromProvider(provider)
	}
	s.trackingMu.RLock()
	retry := s.trackingRetry
	breaker := s.trackingBreaker
	s.trackingMu.RUnlock()
	return newTrackingClientFromProviderWithResilience(provider, retry, breaker)
}

func newTrackingClientFromProviderWithResilience(
	provider *shipping.TrackingProviderConfig,
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) (tracking.TrackingService, error) {
	if provider == nil || provider.ID == 0 {
		return nil, ErrTrackingProviderRequired
	}
	if !provider.Enabled {
		return nil, fmt.Errorf("%w: %s", ErrTrackingProviderDisabled, provider.ProviderName)
	}

	providerCode := strings.ToLower(strings.TrimSpace(provider.ProviderCode))
	if providerCode == "mock" {
		return tracking.NewMockTrackingService(), nil
	}

	apiKey := strings.TrimSpace(provider.APIKey)
	if apiKey == "" {
		return nil, ErrTrackingProviderAPIKeyMissing
	}

	baseURL := strings.TrimRight(strings.TrimSpace(provider.BaseURL), "/")
	if baseURL == "" {
		return nil, ErrTrackingProviderBaseURLMissing
	}

	timeoutSeconds := provider.RequestTimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 15
	}

	switch providerCode {
	case "17track", "track17":
		return tracking.NewTrackingService(&tracking.Config{
			Provider: provider.ProviderCode,
			APIKey:   apiKey,
			BaseURL:  baseURL,
			Timeout:  time.Duration(timeoutSeconds) * time.Second,
			Retry:    retry,
			Breaker:  breaker,
		}), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrTrackingProviderUnsupported, provider.ProviderCode)
	}
}

func trackingInfoToDomainEvents(orderID uint, trackingNumber string, carrierCode string, info *tracking.TrackingInfo) []shipping.TrackingEvent {
	if info == nil {
		return nil
	}

	effectiveTrackingNumber := strings.TrimSpace(info.TrackingNumber)
	if effectiveTrackingNumber == "" {
		effectiveTrackingNumber = trackingNumber
	}
	effectiveCarrier := strings.TrimSpace(info.Carrier)
	if effectiveCarrier == "" {
		effectiveCarrier = carrierCode
	}

	events := make([]shipping.TrackingEvent, 0, len(info.Events))
	syncTime := time.Now()
	for _, event := range info.Events {
		eventTime := event.Time
		if eventTime.IsZero() {
			eventTime = syncTime
		}
		events = append(events, shipping.TrackingEvent{
			OrderID:                orderID,
			TrackingNumber:         effectiveTrackingNumber,
			ProviderCarrierCode:    effectiveCarrier,
			Status:                 strings.TrimSpace(event.Status),
			Location:               trackingLocationString(event.Location),
			Description:            strings.TrimSpace(event.Description),
			RecipientSignatureName: strings.TrimSpace(event.RecipientSignatureName),
			ProofOfDeliveryURL:     strings.TrimSpace(event.ProofOfDeliveryURL),
			EventTime:              eventTime,
		})
	}

	if len(events) == 0 && strings.TrimSpace(info.Status) != "" {
		eventTime := info.UpdatedAt
		if eventTime.IsZero() {
			eventTime = syncTime
		}
		events = append(events, shipping.TrackingEvent{
			OrderID:             orderID,
			TrackingNumber:      effectiveTrackingNumber,
			ProviderCarrierCode: effectiveCarrier,
			Status:              strings.TrimSpace(info.Status),
			Description:         strings.TrimSpace(info.Status),
			EventTime:           eventTime,
		})
	}

	return events
}

func trackingWebhookEventsToDomainEvents(orderID uint, trackingNumber string, carrierCode string, input TrackingWebhookInput) []shipping.TrackingEvent {
	events := make([]shipping.TrackingEvent, 0, len(input.Events))
	syncTime := time.Now()
	for _, event := range input.Events {
		eventTime := event.EventTime
		if eventTime.IsZero() {
			eventTime = syncTime
		}
		events = append(events, shipping.TrackingEvent{
			OrderID:                orderID,
			TrackingNumber:         trackingNumber,
			ProviderCarrierCode:    carrierCode,
			Status:                 strings.TrimSpace(event.Status),
			Location:               strings.TrimSpace(event.Location),
			Description:            strings.TrimSpace(event.Description),
			RecipientSignatureName: strings.TrimSpace(event.RecipientSignatureName),
			ProofOfDeliveryURL:     strings.TrimSpace(event.ProofOfDeliveryURL),
			EventTime:              eventTime,
		})
	}

	if len(events) == 0 && strings.TrimSpace(input.Status) != "" {
		events = append(events, shipping.TrackingEvent{
			OrderID:             orderID,
			TrackingNumber:      trackingNumber,
			ProviderCarrierCode: carrierCode,
			Status:              strings.TrimSpace(input.Status),
			Description:         strings.TrimSpace(input.Status),
			EventTime:           syncTime,
		})
	}

	return events
}

func trackingLocationString(location *tracking.Location) string {
	if location == nil {
		return ""
	}

	parts := []string{}
	for _, part := range []string{location.City, location.State, location.Country, location.ZipCode} {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, ", ")
}

func trackingCarrierResolution(
	provider *shipping.TrackingProviderConfig,
	carrier *shipping.Carrier,
	carrierService *shipping.CarrierService,
	mapping *shipping.TrackingCarrierMapping,
) *TrackingCarrierResolution {
	return &TrackingCarrierResolution{
		Provider:            provider,
		Carrier:             carrier,
		CarrierService:      carrierService,
		Mapping:             mapping,
		ProviderCarrierCode: mapping.ProviderCarrierCode,
		ProviderCarrierName: mapping.ProviderCarrierName,
	}
}
