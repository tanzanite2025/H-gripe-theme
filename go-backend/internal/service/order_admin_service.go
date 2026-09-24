package service

import (
	"commerce-platform/internal/domain/order"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *OrderService) GetAdminOrder(id uint) (*order.Order, error) {
	return s.findOrder(id)
}

func (s *OrderService) GetAdminOrderTrackingShipments(id uint) ([]shippingdomain.TrackingShipment, error) {
	if _, err := s.findOrder(id); err != nil {
		return nil, err
	}
	if s.shipping == nil {
		return nil, ErrOrderShippingNotConfigured
	}
	return s.shipping.GetTrackingShipmentsByOrderID(id)
}

func (s *OrderService) GetOrderTrackingShipments(id uint) ([]shippingdomain.TrackingShipment, error) {
	if s.shipping == nil {
		return nil, ErrOrderShippingNotConfigured
	}
	return s.shipping.GetTrackingShipmentsByOrderID(id)
}

func (s *OrderService) GetAllOrders(page, pageSize int, status string) ([]order.Order, int64, error) {
	return s.orderRepo.FindAll(page, pageSize, status)
}

func (s *OrderService) ListAdminOrders(page, pageSize int, status, paymentStatus, shippingStatus, search, startDate, endDate string) ([]order.Order, int64, error) {
	return s.orderRepo.FindAllWithFilters(page, pageSize, status, paymentStatus, shippingStatus, search, startDate, endDate)
}

func (s *OrderService) UpdateOrderStatus(id uint, status string) error {
	if isSystemManagedOrderStatus(status) {
		return fmt.Errorf("%w: %s", ErrSystemManagedOrderStatus, status)
	}
	if status == "shipped" {
		return ErrOrderFulfillmentStatusManaged
	}

	if status == "completed" {
		return s.completeOrderWithLoyaltyReward(id)
	}

	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return normalizeOrderError(err)
	}

	if !o.CanTransitionTo(status) {
		return fmt.Errorf("invalid status transition from %s to %s", o.Status, status)
	}

	if status == "cancelled" {
		return s.cancelOrderWithRollback(o)
	}

	return s.orderRepo.UpdateStatus(id, o.Status, status)
}

func isSystemManagedOrderStatus(status string) bool {
	return status == "paid" || status == "refunded" || status == "payment_expired" || status == "disputed" || status == "needs_review"
}

func (s *OrderService) UpdateShippingStatus(id uint, shippingStatus string) error {
	if shippingStatus == "shipped" {
		return ErrOrderFulfillmentStatusManaged
	}
	if shippingStatus == "delivered" && s.shipping == nil {
		return ErrOrderShippingNotConfigured
	}

	if s.txManager != nil {
		return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
			orderRecord, err := repos.Order.FindByIDForUpdate(id)
			if err != nil {
				return normalizeOrderError(err)
			}
			if shippingStatus == "delivered" {
				if repos.Shipping == nil {
					return ErrOrderShippingNotConfigured
				}
				allDelivered, err := repos.Shipping.AreAllTrackingShipmentsDelivered(id)
				if err != nil {
					return err
				}
				if !allDelivered {
					return ErrOrderShippingDeliveryNotReady
				}
			}
			if shippingStatus == "delivered" {
				deliveredAt := time.Now().UTC()
				updated, err := repos.Order.MarkDeliveredAtIfNeeded(id, deliveredAt)
				if err != nil || !updated {
					return err
				}
				return enqueueOrderDeliveredDomainEvent(
					repos.Outbox,
					orderRecord,
					"",
					orderRecord.ShippingStatus,
					deliveredAt,
					"admin_manual_update",
				)
			}
			if err := repos.Order.UpdateShippingStatus(id, shippingStatus); err != nil {
				return err
			}
			return nil
		})
	}

	_, err := s.findOrder(id)
	if err != nil {
		return err
	}
	if shippingStatus == "delivered" {
		if s.shipping == nil || s.shipping.shippingRepo == nil {
			return ErrOrderShippingNotConfigured
		}
		allDelivered, err := s.shipping.shippingRepo.AreAllTrackingShipmentsDelivered(id)
		if err != nil {
			return err
		}
		if !allDelivered {
			return ErrOrderShippingDeliveryNotReady
		}
		deliveredAt := time.Now().UTC()
		updated, err := s.orderRepo.MarkDeliveredAtIfNeeded(id, deliveredAt)
		if err != nil || !updated {
			return err
		}
		return nil
	}
	return s.orderRepo.UpdateShippingStatus(id, shippingStatus)
}

type resolvedOrderTrackingUpdate struct {
	trackingShipment    TrackingShipmentInput
	carrierName         string
	trackingURLTemplate string
	autoRegister        bool
}

func resolveOrderTrackingUpdate(
	shippingService *ShippingService,
	input OrderTrackingUpdateInput,
) (*resolvedOrderTrackingUpdate, error) {
	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if trackingNumber == "" {
		return nil, ErrTrackingNumberRequired
	}
	if shippingService == nil {
		return nil, ErrOrderShippingNotConfigured
	}

	carrierIDInput := input.CarrierID
	carrierServiceIDInput := input.CarrierServiceID

	resolution, err := shippingService.ResolveTrackingCarrier(TrackingCarrierResolutionInput{
		ProviderID:       input.TrackingProviderID,
		CarrierID:        carrierIDInput,
		CarrierServiceID: carrierServiceIDInput,
	})
	if err != nil {
		return nil, err
	}

	var carrierID *uint
	if resolution.Carrier != nil {
		carrierID = uintPtr(resolution.Carrier.ID)
	} else if hasPositiveID(carrierIDInput) {
		carrierID = carrierIDInput
	}

	var carrierServiceID *uint
	if resolution.CarrierService != nil {
		carrierServiceID = uintPtr(resolution.CarrierService.ID)
	} else if hasPositiveID(carrierServiceIDInput) {
		carrierServiceID = carrierServiceIDInput
	}

	carrierName := ""
	trackingURLTemplate := ""
	if resolution.Carrier != nil {
		carrierName = strings.TrimSpace(resolution.Carrier.Name)
		trackingURLTemplate = strings.TrimSpace(resolution.Carrier.TrackingURL)
	}
	if carrierName == "" {
		carrierName = strings.TrimSpace(resolution.ProviderCarrierName)
	}
	if carrierName == "" {
		carrierName = strings.TrimSpace(resolution.ProviderCarrierCode)
	}

	return &resolvedOrderTrackingUpdate{
		trackingShipment: TrackingShipmentInput{
			TrackingProviderID:       resolution.Provider.ID,
			TrackingNumber:           trackingNumber,
			ProviderCarrierCode:      resolution.ProviderCarrierCode,
			CarrierID:                carrierID,
			CarrierServiceID:         carrierServiceID,
			TrackingCarrierMappingID: uintPtr(resolution.Mapping.ID),
		},
		carrierName:         carrierName,
		trackingURLTemplate: trackingURLTemplate,
		autoRegister:        resolution.Provider.AutoRegister,
	}, nil
}

func (s *OrderService) UpdateTrackingInfo(ctx context.Context, id uint, input OrderTrackingUpdateInput) error {
	if _, err := s.findOrder(id); err != nil {
		return err
	}

	resolvedTracking, err := resolveOrderTrackingUpdate(s.shipping, input)
	if err != nil {
		return err
	}
	resolvedTracking.trackingShipment.OrderID = id
	// Tracking facts and any registration command must commit atomically. Every
	// update therefore goes through the transaction manager.
	if s.txManager == nil {
		return ErrOrderTrackingTransactionNeeded
	}

	if s.txManager != nil {
		if err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
			if repos.Shipping == nil || repos.Order == nil {
				return ErrOrderShippingNotConfigured
			}
			if resolvedTracking.autoRegister && repos.Outbox == nil {
				return errors.New("tracking registration outbox is not configured")
			}
			orderRecord, err := repos.Order.FindByIDForUpdate(id)
			if err != nil {
				return err
			}

			shippingService := NewShippingService(repos.Shipping)
			_, existingShipmentErr := repos.Shipping.FindTrackingShipmentByOrderIDAndTrackingNumber(id, resolvedTracking.trackingShipment.TrackingNumber)
			if existingShipmentErr != nil && !repository.IsRecordNotFound(existingShipmentErr) {
				return existingShipmentErr
			}
			shipment, err := shippingService.UpsertTrackingShipment(resolvedTracking.trackingShipment)
			if err != nil {
				return err
			}
			resolvedTracking.trackingShipment.ID = shipment.ID
			if repository.IsRecordNotFound(existingShipmentErr) && orderRecord.ShippingStatus == "delivered" {
				if err := repos.Order.ReopenShippingStatusForAdditionalShipment(id); err != nil {
					return err
				}
			}
			if resolvedTracking.autoRegister {
				if err := enqueueTrackingShipmentRegistrationOutboxEvent(
					repos.Outbox,
					resolvedTracking.trackingShipment,
					time.Now().UTC(),
				); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// FulfillOrder records a dispatch as one transaction: the order, its shipping
// status, and its 17TRACK task either persist together or do not change.
// Registration happens after commit so a temporary provider outage never blocks
// a parcel that has already been physically handed to the carrier.
func (s *OrderService) FulfillOrder(ctx context.Context, id uint, input OrderTrackingUpdateInput) (*OrderFulfillmentResult, error) {
	return s.FulfillOrderWithIdempotency(ctx, id, input, 0, "", "")
}

func (s *OrderService) FulfillOrderWithIdempotency(
	ctx context.Context,
	id uint,
	input OrderTrackingUpdateInput,
	adminUserID uint,
	idempotencyKey string,
	requestHash string,
) (*OrderFulfillmentResult, error) {
	if s.txManager == nil {
		return nil, ErrOrderFulfillmentTransactionNeeded
	}
	if s.shipping == nil {
		return nil, ErrOrderShippingNotConfigured
	}
	if s.orderEvidence == nil {
		return nil, ErrOrderFulfillmentEvidenceNotConfigured
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	requestHash = strings.TrimSpace(requestHash)
	if idempotencyKey != "" {
		if adminUserID == 0 {
			return nil, ErrOrderIdempotencyUnavailable
		}
		if requestHash == "" {
			return nil, ErrOrderIdempotencyHashRequired
		}
	}

	var resolvedTracking *resolvedOrderTrackingUpdate
	var idempotencyRecord *order.OrderIdempotency
	if err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Order == nil || repos.Shipping == nil || repos.OrderEvidence == nil {
			return ErrOrderShippingNotConfigured
		}
		if idempotencyKey != "" {
			if repos.OrderIdempotency == nil {
				return ErrOrderFulfillmentIdempotencyUnavailable
			}
			const fulfillmentScope = "admin_order_fulfillment"
			record := &order.OrderIdempotency{
				UserID:         adminUserID,
				Scope:          fulfillmentScope,
				IdempotencyKey: idempotencyKey,
				RequestHash:    requestHash,
			}
			claimed, err := repos.OrderIdempotency.TryCreate(record)
			if err != nil {
				return err
			}
			if !claimed {
				existing, err := repos.OrderIdempotency.FindByUserScopeKey(adminUserID, fulfillmentScope, idempotencyKey)
				if err != nil {
					return err
				}
				if existing.RequestHash != requestHash {
					return ErrOrderFulfillmentIdempotencyConflict
				}
				if existing.OrderID == nil || *existing.OrderID == 0 {
					return ErrOrderFulfillmentIdempotencyInProgress
				}
				if *existing.OrderID != id {
					return ErrOrderFulfillmentIdempotencyConflict
				}
				return nil
			}
			idempotencyRecord = record
		}

		o, err := repos.Order.FindByIDForUpdateWithItems(id)
		if err != nil {
			return normalizeOrderError(err)
		}
		if o.FulfillmentHold {
			return ErrOrderFulfillmentOnHold
		}
		if repos.Payment == nil {
			return ErrOrderFulfillmentOnHold
		}
		pendingRefund, err := repos.Payment.HasPendingRefundByOrderID(id)
		if err != nil {
			return err
		}
		if pendingRefund {
			return ErrOrderFulfillmentBlockedByPendingRefund
		}
		activeStripeDispute, err := repos.Payment.HasActiveStripeDisputeByOrderID(id)
		if err != nil {
			return err
		}
		activePayPalDispute, err := repos.Payment.HasActivePayPalDisputeByOrderID(id)
		if err != nil {
			return err
		}
		activePaymentReview, err := repos.Payment.HasActivePaymentReviewByOrderID(id)
		if err != nil {
			return err
		}
		if activeStripeDispute || activePayPalDispute || activePaymentReview {
			return ErrOrderFulfillmentOnHold
		}
		if o.Status == "shipped" {
			if o.PaymentStatus != "paid" {
				return ErrOrderFulfillmentPaymentRequired
			}
			txShippingService := NewShippingService(repos.Shipping)
			resolvedTracking, err = resolveOrderTrackingUpdate(txShippingService, input)
			if err != nil {
				return err
			}
			resolvedTracking.trackingShipment.OrderID = id
			_, existingShipmentErr := repos.Shipping.FindTrackingShipmentByOrderIDAndTrackingNumber(id, resolvedTracking.trackingShipment.TrackingNumber)
			if existingShipmentErr != nil && !repository.IsRecordNotFound(existingShipmentErr) {
				return existingShipmentErr
			}
			shipment, err := txShippingService.UpsertTrackingShipment(resolvedTracking.trackingShipment)
			if err != nil {
				return err
			}
			resolvedTracking.trackingShipment.ID = shipment.ID
			if repository.IsRecordNotFound(existingShipmentErr) && o.ShippingStatus == "delivered" {
				if err := repos.Order.ReopenShippingStatusForAdditionalShipment(id); err != nil {
					return err
				}
			}
			shippedAt := time.Now().UTC()
			if o.ShippedAt != nil && !o.ShippedAt.IsZero() {
				shippedAt = o.ShippedAt.UTC()
			}
			if err := enqueueOrderShippedDomainEvent(
				repos.Outbox,
				o,
				resolvedTracking,
				o.Status,
				o.ShippingStatus,
				shippedAt,
			); err != nil {
				return err
			}
			if resolvedTracking.autoRegister {
				if err := enqueueTrackingShipmentRegistrationOutboxEvent(repos.Outbox, resolvedTracking.trackingShipment, shippedAt); err != nil {
					return err
				}
			}
			if idempotencyRecord != nil {
				if err := repos.OrderIdempotency.BindOrderID(idempotencyRecord.ID, id); err != nil {
					return err
				}
			}
			return nil
		}
		if o.PaymentStatus != "paid" {
			return ErrOrderFulfillmentPaymentRequired
		}
		if o.Status != "paid" && o.Status != "processing" && o.Status != "shipped" {
			return fmt.Errorf("%w: %s", ErrOrderFulfillmentNotAllowed, o.Status)
		}
		if orderRequiresProduction(o) &&
			order.NormalizeProductionStatus(o.ProductionStatus) != order.ProductionStatusCompleted {
			return ErrOrderProductionNotCompleted
		}
		if o.SignatureRequired && !input.SignatureConfirmed {
			return ErrOrderFulfillmentSignatureConfirmationRequired
		}
		if err := ValidateOrderCustomsDeclaredValuesAreConfirmed(o); err != nil {
			return err
		}
		if _, err := s.orderEvidence.CheckFulfillmentReadiness(repos, o); err != nil {
			return err
		}

		txShippingService := NewShippingService(repos.Shipping)
		resolvedTracking, err = resolveOrderTrackingUpdate(txShippingService, input)
		if err != nil {
			return err
		}
		resolvedTracking.trackingShipment.OrderID = id

		shipment, err := txShippingService.UpsertTrackingShipment(resolvedTracking.trackingShipment)
		if err != nil {
			return err
		}
		resolvedTracking.trackingShipment.ID = shipment.ID

		shippedAt := time.Now().UTC()
		if o.ShippedAt != nil && !o.ShippedAt.IsZero() {
			shippedAt = o.ShippedAt.UTC()
		} else {
			o.ShippedAt = &shippedAt
		}
		previousOrderStatus := o.Status
		previousShippingStatus := o.ShippingStatus
		if o.Status != "shipped" {
			if err := repos.Order.UpdateStatus(id, o.Status, "shipped"); err != nil {
				return err
			}
		}
		if o.ShippingStatus != "shipped" {
			if err := repos.Order.UpdateShippingStatus(id, "shipped"); err != nil {
				return err
			}
		}
		if err := enqueueOrderShippedDomainEvent(
			repos.Outbox,
			o,
			resolvedTracking,
			previousOrderStatus,
			previousShippingStatus,
			shippedAt,
		); err != nil {
			return err
		}
		if resolvedTracking.autoRegister {
			if err := enqueueTrackingShipmentRegistrationOutboxEvent(repos.Outbox, resolvedTracking.trackingShipment, shippedAt); err != nil {
				return err
			}
		}
		if idempotencyRecord != nil {
			if err := repos.OrderIdempotency.BindOrderID(idempotencyRecord.ID, id); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	result := &OrderFulfillmentResult{}

	fulfilledOrder, err := s.GetAdminOrder(id)
	if err != nil {
		return nil, err
	}
	trackingShipments, err := s.GetAdminOrderTrackingShipments(id)
	if err != nil {
		return nil, err
	}
	result.Order = fulfilledOrder
	result.TrackingShipments = trackingShipments
	return result, nil
}

// ValidateOrderCustomsDeclaredValuesAreConfirmed ensures every order line has
// a positive, explicitly confirmed customs declaration before external dispatch.
func ValidateOrderCustomsDeclaredValuesAreConfirmed(orderRecord *order.Order) error {
	if orderRecord == nil {
		return ErrOrderCustomsDeclarationIncomplete
	}
	for _, item := range orderRecord.Items {
		if item.DeclaredValueMinor == nil ||
			!item.DeclaredValueConfirmed ||
			*item.DeclaredValueMinor <= 0 {
			itemReference := strings.TrimSpace(item.SKU)
			if itemReference == "" {
				itemReference = fmt.Sprintf("order_item_id=%d", item.ID)
			}
			return fmt.Errorf(
				"%w: order %s item %s",
				ErrOrderCustomsDeclarationIncomplete,
				orderRecord.OrderNumber,
				itemReference,
			)
		}
	}
	return nil
}

func (s *OrderService) SyncOrderTracking(ctx context.Context, id uint, trackingNumber string) (*TrackingSyncResult, error) {
	o, err := s.findOrder(id)
	if err != nil {
		return nil, err
	}
	if s.shipping == nil {
		return nil, ErrOrderShippingNotConfigured
	}
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return nil, ErrTrackingNumberRequired
	}
	shipment, err := s.shipping.GetTrackingShipmentByOrderIDAndTrackingNumber(o.ID, trackingNumber)
	if err != nil {
		return nil, err
	}

	return s.shipping.SyncTracking(ctx, TrackingSyncInput{
		OrderID:                  o.ID,
		ProviderID:               shipment.TrackingProviderID,
		TrackingNumber:           shipment.TrackingNumber,
		ProviderCarrierCode:      shipment.ProviderCarrierCode,
		CarrierID:                shipment.CarrierID,
		CarrierServiceID:         shipment.CarrierServiceID,
		TrackingCarrierMappingID: shipment.TrackingCarrierMappingID,
	})
}

func (s *OrderService) UpdateAdminNote(id uint, adminNote string) error {
	o, err := s.findOrder(id)
	if err != nil {
		return err
	}

	o.AdminNote = adminNote
	return s.orderRepo.Update(o)
}

// UpdateOrderItemCustoms updates the customs declaration in the order item's
// currency smallest unit. Major-unit values are not accepted at this boundary.
func (s *OrderService) UpdateOrderItemCustoms(orderID, orderItemID uint, declaredValue *int64, confirmed bool) error {
	if orderID == 0 || orderItemID == 0 {
		return ErrOrderItemNotFound
	}
	if declaredValue != nil && *declaredValue < 0 {
		return ErrDeclaredValueInvalid
	}
	if confirmed && declaredValue == nil {
		return ErrDeclaredValueConfirmationRequired
	}
	if s == nil || s.txManager == nil {
		return ErrOrderCustomsUpdateTransactionNeeded
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		o, err := repos.Order.FindByIDForUpdateWithItems(orderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		if orderCustomsDeclarationLocked(o) {
			return ErrOrderCustomsUpdateLocked
		}

		found := false
		for _, item := range o.Items {
			if item.ID == orderItemID {
				found = true
				break
			}
		}
		if !found {
			return ErrOrderItemNotFound
		}

		if declaredValue == nil {
			confirmed = false
		}
		return repos.Order.UpdateOrderItemCustoms(orderID, orderItemID, declaredValue, confirmed)
	})
}

func orderCustomsDeclarationLocked(o *order.Order) bool {
	if o == nil {
		return false
	}
	return o.Status == "shipped" ||
		o.Status == "completed" ||
		o.ShippingStatus == "shipped" ||
		o.ShippingStatus == "delivered" ||
		(o.ShippedAt != nil && !o.ShippedAt.IsZero())
}

// HideUnpaidCancelledOrPaymentExpiredOrderFromDefaultQueries keeps the
// existing admin action limited to orders that never reached a paid state.
// The order row is locked before the financial-state check so a payment
// callback cannot win a race after the check and before the soft delete.
func (s *OrderService) HideUnpaidCancelledOrPaymentExpiredOrderFromDefaultQueries(id uint) error {
	if s == nil || s.txManager == nil {
		return ErrOrderHideTransactionRequired
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		lockedOrder, err := repos.Order.FindByIDForUpdate(id)
		if err != nil {
			return normalizeOrderError(err)
		}
		if !isOrderEligibleForDefaultQueryHideWithoutPaymentActivity(lockedOrder) {
			return ErrOrderHideNotAllowed
		}

		hidden, err := repos.Order.SoftDeleteUnpaidCancelledOrPaymentExpiredOrderRecord(id)
		if err != nil {
			return err
		}
		if !hidden {
			return ErrOrderHideNotAllowed
		}
		return nil
	})
}

func isOrderEligibleForDefaultQueryHideWithoutPaymentActivity(o *order.Order) bool {
	if o == nil || o.Status == "refunded" || o.PaymentStatus == "paid" || o.PaymentStatus == "refunded" {
		return false
	}
	return (o.Status == "cancelled" && o.PaymentStatus == "unpaid") ||
		(o.Status == "payment_expired" && o.PaymentStatus == "expired")
}

func (s *OrderService) GetAdminStats() (map[string]interface{}, error) {
	return s.orderRepo.GetStats()
}

func (s *OrderService) GetSalesByDateRange(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	return s.orderRepo.GetSalesByDateRange(startDate, endDate)
}
