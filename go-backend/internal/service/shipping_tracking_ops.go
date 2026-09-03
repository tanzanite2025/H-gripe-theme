package service

import (
	"context"
	"strings"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/tracking"
	"commerce-platform/internal/repository"
)

func (s *ShippingService) NewTrackingClientForProvider(providerID uint) (tracking.TrackingService, error) {
	if providerID == 0 {
		return nil, ErrTrackingProviderRequired
	}

	provider, err := s.GetTrackingProviderConfig(providerID)
	if err != nil {
		return nil, err
	}
	return s.newTrackingClientFromProvider(provider)
}

func (s *ShippingService) GetTrackingShipmentByOrderID(orderID uint) (*shipping.TrackingShipment, error) {
	if orderID == 0 {
		return nil, ErrTrackingOrderRequired
	}
	return s.shippingRepo.FindTrackingShipmentByOrderID(orderID)
}

func (s *ShippingService) ListTrackingShipments(filter TrackingShipmentListFilter) ([]shipping.TrackingShipment, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	if filter.Limit > 500 {
		filter.Limit = 500
	}

	return s.shippingRepo.FindAllTrackingShipments(repository.TrackingShipmentFilter{
		SyncStatus:          strings.ToLower(strings.TrimSpace(filter.SyncStatus)),
		RegistrationStatus:  strings.ToLower(strings.TrimSpace(filter.RegistrationStatus)),
		TrackingNumber:      strings.TrimSpace(filter.TrackingNumber),
		ProviderCarrierCode: strings.TrimSpace(filter.ProviderCarrierCode),
		Keyword:             strings.TrimSpace(filter.Keyword),
		OrderID:             filter.OrderID,
		ProviderID:          filter.ProviderID,
		CarrierID:           filter.CarrierID,
		CarrierServiceID:    filter.CarrierServiceID,
		Enabled:             filter.Enabled,
		DueOnly:             filter.DueOnly,
		Limit:               filter.Limit,
	})
}

func (s *ShippingService) SyncDueTrackingShipments(ctx context.Context, limit int) (*TrackingShipmentSyncBatchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	shipments, err := s.shippingRepo.FindDueTrackingShipments(limit, time.Now())
	if err != nil {
		return nil, err
	}

	batch := &TrackingShipmentSyncBatchResult{
		Matched: len(shipments),
		Results: make([]TrackingSyncResult, 0, len(shipments)),
		Errors:  make([]TrackingShipmentSyncFailure, 0),
	}
	for _, shipment := range shipments {
		result, err := s.SyncTracking(ctx, TrackingSyncInput{
			OrderID:                  shipment.OrderID,
			ProviderID:               shipment.TrackingProviderID,
			TrackingNumber:           shipment.TrackingNumber,
			ProviderCarrierCode:      shipment.ProviderCarrierCode,
			CarrierID:                shipment.CarrierID,
			CarrierServiceID:         shipment.CarrierServiceID,
			TrackingCarrierMappingID: shipment.TrackingCarrierMappingID,
		})
		if err != nil {
			batch.Failed++
			batch.Errors = append(batch.Errors, TrackingShipmentSyncFailure{
				OrderID:        shipment.OrderID,
				TrackingNumber: shipment.TrackingNumber,
				Error:          err.Error(),
			})
			continue
		}

		batch.Synced++
		batch.Results = append(batch.Results, *result)
	}

	return batch, nil
}

func (s *ShippingService) ApplyTrackingWebhook(input TrackingWebhookInput) (*TrackingWebhookResult, error) {
	if input.ProviderID == 0 {
		return nil, ErrTrackingProviderRequired
	}

	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if trackingNumber == "" {
		return nil, ErrTrackingNumberRequired
	}

	providerCarrierCode := strings.TrimSpace(input.ProviderCarrierCode)
	shipment, err := s.shippingRepo.FindTrackingShipmentByProviderTrackingNumber(input.ProviderID, trackingNumber, providerCarrierCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(shipment.ProviderCarrierCode) != "" {
		providerCarrierCode = shipment.ProviderCarrierCode
	}

	events := trackingWebhookEventsToDomainEvents(shipment.OrderID, trackingNumber, providerCarrierCode, input)
	if err := s.shippingRepo.UpsertTrackingEvents(shipment.OrderID, trackingNumber, events); err != nil {
		return nil, err
	}
	persistedEvents, err := s.shippingRepo.FindTrackingEventsByOrderID(shipment.OrderID)
	if err != nil {
		return nil, err
	}
	if err := s.shippingRepo.UpdateTrackingShipmentSyncSuccess(shipment.OrderID, len(persistedEvents), latestTrackingEventTime(persistedEvents), nil); err != nil {
		return nil, err
	}
	if err := s.updateOrderShippingStatusIfDelivered(
		shipment.OrderID,
		input.Status,
		input.StatusCode,
		events,
	); err != nil {
		return nil, err
	}

	updatedShipment, err := s.GetTrackingShipmentByOrderID(shipment.OrderID)
	if err != nil {
		return nil, err
	}

	return &TrackingWebhookResult{
		Shipment: updatedShipment,
		Events:   events,
	}, nil
}

func (s *ShippingService) UpsertTrackingShipment(input TrackingShipmentInput) (*shipping.TrackingShipment, error) {
	if input.OrderID == 0 {
		return nil, ErrTrackingOrderRequired
	}
	if input.TrackingProviderID == 0 {
		return nil, ErrTrackingProviderRequired
	}

	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if trackingNumber == "" {
		return nil, ErrTrackingNumberRequired
	}

	providerCarrierCode := strings.TrimSpace(input.ProviderCarrierCode)
	if providerCarrierCode == "" {
		return nil, ErrTrackingCarrierCodeRequired
	}

	shipment := &shipping.TrackingShipment{
		OrderID:                  input.OrderID,
		TrackingProviderID:       input.TrackingProviderID,
		TrackingNumber:           trackingNumber,
		ProviderCarrierCode:      providerCarrierCode,
		CarrierID:                input.CarrierID,
		CarrierServiceID:         input.CarrierServiceID,
		TrackingCarrierMappingID: input.TrackingCarrierMappingID,
		RegistrationStatus:       trackingRegistrationPending,
		SyncStatus:               trackingSyncPending,
		EventCount:               0,
		LastError:                "",
		Enabled:                  true,
	}
	if err := s.shippingRepo.UpsertTrackingShipment(shipment); err != nil {
		return nil, err
	}
	return s.GetTrackingShipmentByOrderID(input.OrderID)
}

func (s *ShippingService) UpsertAndMaybeRegisterTrackingShipment(ctx context.Context, input TrackingShipmentInput) (*shipping.TrackingShipment, error) {
	shipment, err := s.UpsertTrackingShipment(input)
	if err != nil {
		return nil, err
	}

	provider, err := s.GetTrackingProviderConfig(input.TrackingProviderID)
	if err != nil {
		return nil, err
	}
	if !provider.AutoRegister {
		return shipment, nil
	}

	if err := s.RegisterTrackingShipment(ctx, TrackingSyncInput{
		OrderID:                  input.OrderID,
		ProviderID:               input.TrackingProviderID,
		TrackingNumber:           input.TrackingNumber,
		ProviderCarrierCode:      input.ProviderCarrierCode,
		CarrierID:                input.CarrierID,
		CarrierServiceID:         input.CarrierServiceID,
		TrackingCarrierMappingID: input.TrackingCarrierMappingID,
	}); err != nil {
		return nil, err
	}
	return s.GetTrackingShipmentByOrderID(input.OrderID)
}

func (s *ShippingService) RegisterTrackingShipment(ctx context.Context, input TrackingSyncInput) error {
	if input.OrderID == 0 {
		return ErrTrackingOrderRequired
	}
	if input.ProviderID == 0 {
		return ErrTrackingProviderRequired
	}

	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if trackingNumber == "" {
		return ErrTrackingNumberRequired
	}
	providerCarrierCode := strings.TrimSpace(input.ProviderCarrierCode)
	if providerCarrierCode == "" {
		return ErrTrackingCarrierCodeRequired
	}

	existing, err := s.GetTrackingShipmentByOrderID(input.OrderID)
	if err == nil &&
		existing.TrackingProviderID == input.ProviderID &&
		existing.TrackingNumber == trackingNumber &&
		existing.ProviderCarrierCode == providerCarrierCode &&
		existing.RegistrationStatus == trackingRegistrationSynced {
		return nil
	}
	if err != nil && !repository.IsRecordNotFound(err) {
		return err
	}

	client, err := s.NewTrackingClientForProvider(input.ProviderID)
	if err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentRegistrationStatus(input.OrderID, trackingRegistrationFailed, err.Error())
		return err
	}

	registrar, ok := client.(tracking.TrackingRegistrar)
	if !ok {
		return nil
	}

	if err := registrar.RegisterTrackings(ctx, []tracking.TrackingRequest{{TrackingNumber: trackingNumber, Carrier: providerCarrierCode}}); err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentRegistrationStatus(input.OrderID, trackingRegistrationFailed, err.Error())
		return err
	}

	return s.shippingRepo.UpdateTrackingShipmentRegistrationStatus(input.OrderID, trackingRegistrationSynced, "")
}

func (s *ShippingService) SyncTracking(ctx context.Context, input TrackingSyncInput) (*TrackingSyncResult, error) {
	if input.OrderID == 0 {
		return nil, ErrTrackingOrderRequired
	}
	if strings.TrimSpace(input.TrackingNumber) == "" {
		return nil, ErrTrackingNumberRequired
	}
	providerCarrierCode := strings.TrimSpace(input.ProviderCarrierCode)
	if providerCarrierCode == "" {
		return nil, ErrTrackingCarrierCodeRequired
	}

	provider, err := s.GetTrackingProviderConfig(input.ProviderID)
	if err != nil {
		return nil, err
	}

	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if _, err := s.ensureTrackingShipmentForSync(input, trackingNumber, providerCarrierCode); err != nil {
		return nil, err
	}
	if provider.AutoRegister {
		if err := s.RegisterTrackingShipment(ctx, input); err != nil {
			_ = s.shippingRepo.UpdateTrackingShipmentSyncFailure(input.OrderID, err.Error(), nextTrackingSyncAt(provider, time.Now()))
			return nil, err
		}
	}
	if err := s.shippingRepo.UpdateTrackingShipmentSyncing(input.OrderID); err != nil {
		return nil, err
	}

	client, err := s.newTrackingClientFromProvider(provider)
	if err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentSyncFailure(input.OrderID, err.Error(), nextTrackingSyncAt(provider, time.Now()))
		return nil, err
	}

	info, err := client.Track(ctx, trackingNumber, providerCarrierCode)
	if err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentSyncFailure(input.OrderID, err.Error(), nextTrackingSyncAt(provider, time.Now()))
		return nil, err
	}

	events := trackingInfoToDomainEvents(input.OrderID, trackingNumber, providerCarrierCode, info)
	if err := s.shippingRepo.UpsertTrackingEvents(input.OrderID, trackingNumber, events); err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentSyncFailure(input.OrderID, err.Error(), nextTrackingSyncAt(provider, time.Now()))
		return nil, err
	}
	persistedEvents, err := s.shippingRepo.FindTrackingEventsByOrderID(input.OrderID)
	if err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentSyncFailure(input.OrderID, err.Error(), nextTrackingSyncAt(provider, time.Now()))
		return nil, err
	}
	if err := s.shippingRepo.UpdateTrackingShipmentSyncSuccess(input.OrderID, len(persistedEvents), latestTrackingEventTime(persistedEvents), nextTrackingSyncAt(provider, time.Now())); err != nil {
		return nil, err
	}
	if info != nil {
		if err := s.updateOrderShippingStatusIfDelivered(
			input.OrderID,
			info.Status,
			info.StatusCode,
			events,
		); err != nil {
			return nil, err
		}
	}
	shipment, err := s.GetTrackingShipmentByOrderID(input.OrderID)
	if err != nil {
		return nil, err
	}

	result := &TrackingSyncResult{
		TrackingNumber: trackingNumber,
		Carrier:        providerCarrierCode,
		Events:         events,
		Shipment:       shipment,
	}
	if info != nil {
		result.TrackingNumber = info.TrackingNumber
		if result.TrackingNumber == "" {
			result.TrackingNumber = trackingNumber
		}
		if strings.TrimSpace(info.Carrier) != "" {
			result.Carrier = info.Carrier
		}
		result.Status = info.Status
		result.StatusCode = info.StatusCode
		result.UpdatedAt = info.UpdatedAt
	}

	return result, nil
}
