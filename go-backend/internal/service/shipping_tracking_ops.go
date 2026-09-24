package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/pkg/resilience"
	"commerce-platform/internal/pkg/tracking"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
)

const trackingSyncLeaseDuration = 10 * time.Minute

// RequestTrackingShipmentRegistration records an idempotent registration
// command in the same database transaction that reads the current shipment.
// The external provider is contacted only by the outbox worker.
func (s *ShippingService) RequestTrackingShipmentRegistration(ctx context.Context, orderID uint, trackingNumber string) (*shipping.TrackingShipment, error) {
	if orderID == 0 {
		return nil, ErrTrackingOrderRequired
	}
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return nil, ErrTrackingNumberRequired
	}
	if s == nil || s.txManager == nil {
		return nil, ErrTrackingRegistrationTransactionNeeded
	}
	var shipment *shipping.TrackingShipment
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Shipping == nil || repos.Outbox == nil {
			return ErrOrderShippingNotConfigured
		}
		current, err := repos.Shipping.FindTrackingShipmentByOrderIDAndTrackingNumber(orderID, trackingNumber)
		if err != nil {
			return err
		}
		shipment = current
		if current.RegistrationStatus == trackingRegistrationSynced {
			return nil
		}
		if current.RegistrationStatus == trackingRegistrationUnknown {
			return fmt.Errorf("%w: tracking shipment registration requires reconciliation", resilience.ErrExternalOutcomeUnknown)
		}
		return enqueueTrackingShipmentRegistrationOutboxEvent(repos.Outbox, TrackingShipmentInput{
			OrderID:                  current.OrderID,
			TrackingProviderID:       current.TrackingProviderID,
			TrackingNumber:           current.TrackingNumber,
			ProviderCarrierCode:      current.ProviderCarrierCode,
			CarrierID:                current.CarrierID,
			CarrierServiceID:         current.CarrierServiceID,
			TrackingCarrierMappingID: current.TrackingCarrierMappingID,
		}, time.Now().UTC())
	})
	return shipment, err
}

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

func (s *ShippingService) GetTrackingShipmentsByOrderID(orderID uint) ([]shipping.TrackingShipment, error) {
	if orderID == 0 {
		return nil, ErrTrackingOrderRequired
	}
	return s.shippingRepo.FindTrackingShipmentsByOrderID(orderID)
}

func (s *ShippingService) GetTrackingShipmentByOrderIDAndTrackingNumber(orderID uint, trackingNumber string) (*shipping.TrackingShipment, error) {
	if orderID == 0 {
		return nil, ErrTrackingOrderRequired
	}
	if strings.TrimSpace(trackingNumber) == "" {
		return nil, ErrTrackingNumberRequired
	}
	return s.shippingRepo.FindTrackingShipmentByOrderIDAndTrackingNumber(orderID, trackingNumber)
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

	shipments, err := s.shippingRepo.ClaimDueTrackingShipments(
		limit,
		time.Now().UTC(),
		"tracking-poll-"+uuid.NewString(),
		trackingSyncLeaseDuration,
	)
	if err != nil {
		return nil, err
	}

	batch := &TrackingShipmentSyncBatchResult{
		Matched: len(shipments),
		Results: make([]TrackingSyncResult, 0, len(shipments)),
		Errors:  make([]TrackingShipmentSyncFailure, 0),
	}
	for index := range shipments {
		shipment := &shipments[index]
		result, err := s.syncClaimedTracking(ctx, TrackingSyncInput{
			OrderID:                  shipment.OrderID,
			ProviderID:               shipment.TrackingProviderID,
			TrackingNumber:           shipment.TrackingNumber,
			ProviderCarrierCode:      shipment.ProviderCarrierCode,
			CarrierID:                shipment.CarrierID,
			CarrierServiceID:         shipment.CarrierServiceID,
			TrackingCarrierMappingID: shipment.TrackingCarrierMappingID,
		}, shipment, shipment.Provider)
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
	persistedEvents, err := s.shippingRepo.FindTrackingEventsByOrderIDAndTrackingNumber(shipment.OrderID, trackingNumber)
	if err != nil {
		return nil, err
	}
	if err := s.shippingRepo.ApplyTrackingWebhookSyncSuccess(
		shipment.OrderID,
		trackingNumber,
		len(persistedEvents),
		latestTrackingEventTime(persistedEvents),
		time.Now().UTC(),
	); err != nil {
		return nil, err
	}
	if err := s.updateOrderShippingStatusIfDelivered(
		shipment.OrderID,
		trackingNumber,
		providerCarrierCode,
		input.Status,
		input.StatusCode,
		events,
		"tracking_webhook",
	); err != nil {
		return nil, err
	}
	if s.afterSalesService != nil && isReturnDeliveryWebhook(input.Status, input.Events) {
		if _, err := s.afterSalesService.ApplyCarrierWebhook(AfterSalesCarrierWebhookInput{
			TrackingNumber: trackingNumber,
			Status:         input.Status,
			EventTime:      latestTrackingWebhookEventTime(input.Events),
		}); err != nil && !errors.Is(err, ErrAfterSalesCaseNotFound) {
			return nil, err
		}
	}

	updatedShipment, err := s.GetTrackingShipmentByOrderIDAndTrackingNumber(shipment.OrderID, trackingNumber)
	if err != nil {
		return nil, err
	}

	return &TrackingWebhookResult{
		Shipment: updatedShipment,
		Events:   events,
	}, nil
}

func isReturnDeliveryWebhook(status string, events []TrackingWebhookEventInput) bool {
	if normalizeAfterSalesCarrierStatus(status) != "" {
		return true
	}
	for _, event := range events {
		if normalizeAfterSalesCarrierStatus(event.Status) != "" {
			return true
		}
	}
	return false
}

func latestTrackingWebhookEventTime(events []TrackingWebhookEventInput) time.Time {
	var latest time.Time
	for _, event := range events {
		if event.EventTime.After(latest) {
			latest = event.EventTime
		}
	}
	return latest
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
	return s.GetTrackingShipmentByOrderIDAndTrackingNumber(input.OrderID, trackingNumber)
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

	existing, err := s.GetTrackingShipmentByOrderIDAndTrackingNumber(input.OrderID, trackingNumber)
	if err == nil &&
		existing.TrackingProviderID == input.ProviderID &&
		existing.TrackingNumber == trackingNumber &&
		existing.ProviderCarrierCode == providerCarrierCode &&
		existing.RegistrationStatus == trackingRegistrationSynced {
		return nil
	}
	if err == nil && existing.RegistrationStatus == trackingRegistrationUnknown {
		return fmt.Errorf("%w: tracking shipment registration requires reconciliation", resilience.ErrExternalOutcomeUnknown)
	}
	if err != nil && !repository.IsRecordNotFound(err) {
		return err
	}

	client, err := s.NewTrackingClientForProvider(input.ProviderID)
	if err != nil {
		_ = s.shippingRepo.UpdateTrackingShipmentRegistrationStatusForTracking(input.OrderID, trackingNumber, trackingRegistrationFailed, err.Error())
		return err
	}

	registrar, ok := client.(tracking.TrackingRegistrar)
	if !ok {
		return s.shippingRepo.UpdateTrackingShipmentRegistrationStatusForTracking(input.OrderID, trackingNumber, trackingRegistrationSynced, "")
	}

	if err := registrar.RegisterTrackings(ctx, []tracking.TrackingRequest{{TrackingNumber: trackingNumber, Carrier: providerCarrierCode}}); err != nil {
		status := trackingRegistrationFailed
		if errors.Is(err, resilience.ErrExternalOutcomeUnknown) {
			status = trackingRegistrationUnknown
		}
		_ = s.shippingRepo.UpdateTrackingShipmentRegistrationStatusForTracking(input.OrderID, trackingNumber, status, err.Error())
		return err
	}

	return s.shippingRepo.UpdateTrackingShipmentRegistrationStatusForTracking(input.OrderID, trackingNumber, trackingRegistrationSynced, "")
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
	claim, err := s.shippingRepo.ClaimTrackingShipmentForSync(
		input.OrderID,
		trackingNumber,
		time.Now().UTC(),
		"tracking-manual-"+uuid.NewString(),
		trackingSyncLeaseDuration,
	)
	if err != nil {
		return nil, err
	}
	return s.syncClaimedTracking(ctx, input, claim, provider)
}

func (s *ShippingService) syncClaimedTracking(
	ctx context.Context,
	input TrackingSyncInput,
	claim *shipping.TrackingShipment,
	provider *shipping.TrackingProviderConfig,
) (*TrackingSyncResult, error) {
	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	providerCarrierCode := strings.TrimSpace(input.ProviderCarrierCode)
	if claim == nil || claim.SyncStatus != "syncing" || strings.TrimSpace(claim.SyncLeaseOwner) == "" {
		return nil, repository.ErrTrackingSyncLeaseLost
	}
	if provider == nil {
		var err error
		provider, err = s.GetTrackingProviderConfig(input.ProviderID)
		if err != nil {
			return nil, s.failClaimedTracking(claim, nil, err)
		}
	}
	if provider.AutoRegister && claim.RegistrationStatus != trackingRegistrationSynced {
		return nil, s.failClaimedTracking(claim, provider, errors.New("tracking registration is not complete"))
	}
	if err := s.shippingRepo.RenewTrackingShipmentSyncLease(claim, time.Now().UTC(), trackingSyncLeaseDuration); err != nil {
		return nil, err
	}

	client, err := s.newTrackingClientFromProvider(provider)
	if err != nil {
		return nil, s.failClaimedTracking(claim, provider, err)
	}

	info, err := client.Track(ctx, trackingNumber, providerCarrierCode)
	if err != nil {
		return nil, s.failClaimedTracking(claim, provider, err)
	}
	if err := s.shippingRepo.RenewTrackingShipmentSyncLease(claim, time.Now().UTC(), trackingSyncLeaseDuration); err != nil {
		return nil, err
	}

	events := trackingInfoToDomainEvents(input.OrderID, trackingNumber, providerCarrierCode, info)
	if err := s.shippingRepo.UpsertTrackingEvents(input.OrderID, trackingNumber, events); err != nil {
		return nil, s.failClaimedTracking(claim, provider, err)
	}
	persistedEvents, err := s.shippingRepo.FindTrackingEventsByOrderIDAndTrackingNumber(input.OrderID, trackingNumber)
	if err != nil {
		return nil, s.failClaimedTracking(claim, provider, err)
	}
	completedAt := time.Now().UTC()
	if err := s.shippingRepo.CompleteTrackingShipmentSync(
		claim,
		len(persistedEvents),
		latestTrackingEventTime(persistedEvents),
		nextTrackingSyncAt(provider, completedAt),
		completedAt,
	); err != nil {
		return nil, err
	}
	if info != nil {
		if err := s.updateOrderShippingStatusIfDelivered(
			input.OrderID,
			trackingNumber,
			providerCarrierCode,
			info.Status,
			info.StatusCode,
			events,
			"tracking_sync",
		); err != nil {
			return nil, err
		}
		if s.afterSalesService != nil {
			if _, err := s.afterSalesService.ApplyCarrierTrackingFact(AfterSalesCarrierWebhookInput{
				TrackingNumber: trackingNumber,
				Status:         info.Status,
				EventTime:      trackingFactEventTime(events),
			}); err != nil && !errors.Is(err, ErrAfterSalesCaseNotFound) {
				return nil, err
			}
		}
	}
	shipment, err := s.GetTrackingShipmentByOrderIDAndTrackingNumber(input.OrderID, trackingNumber)
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

func trackingFactEventTime(events []shipping.TrackingEvent) time.Time {
	if latest := latestTrackingEventTime(events); latest != nil {
		return *latest
	}
	return time.Now().UTC()
}

func (s *ShippingService) failClaimedTracking(claim *shipping.TrackingShipment, provider *shipping.TrackingProviderConfig, cause error) error {
	now := time.Now().UTC()
	if err := s.shippingRepo.FailTrackingShipmentSync(claim, cause.Error(), nextTrackingSyncAt(provider, now), now); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}
