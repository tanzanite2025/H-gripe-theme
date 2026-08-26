package service

import (
	"commerce-platform/internal/domain/order"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

func (s *OrderService) GetAdminOrder(id uint) (*order.Order, error) {
	return s.findOrder(id)
}

func (s *OrderService) GetAdminOrderTrackingShipment(id uint) (*shippingdomain.TrackingShipment, error) {
	if _, err := s.findOrder(id); err != nil {
		return nil, err
	}
	if s.shipping == nil {
		return nil, ErrOrderShippingNotConfigured
	}

	shipment, err := s.shipping.GetTrackingShipmentByOrderID(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return shipment, nil
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

	if status == "shipped" && s.txManager != nil {
		return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
			lockedOrder, err := repos.Order.FindByIDForUpdateWithItems(id)
			if err != nil {
				return normalizeOrderError(err)
			}
			if !lockedOrder.CanTransitionTo(status) {
				return fmt.Errorf("invalid status transition from %s to %s", lockedOrder.Status, status)
			}
			if err := repos.Order.UpdateStatus(id, status); err != nil {
				return err
			}
			return nil
		})
	}

	return s.orderRepo.UpdateStatus(id, status)
}

func isSystemManagedOrderStatus(status string) bool {
	return status == "paid" || status == "refunded" || status == "payment_expired"
}

func (s *OrderService) UpdateShippingStatus(id uint, shippingStatus string) error {
	if s.txManager != nil {
		return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
			_, err := repos.Order.FindByIDForUpdate(id)
			if err != nil {
				return normalizeOrderError(err)
			}
			if err := repos.Order.UpdateShippingStatus(id, shippingStatus); err != nil {
				return err
			}
			return nil
		})
	}

	if _, err := s.findOrder(id); err != nil {
		return err
	}
	return s.orderRepo.UpdateShippingStatus(id, shippingStatus)
}

type resolvedOrderTrackingUpdate struct {
	trackingInfo     order.TrackingInfoUpdate
	trackingShipment TrackingShipmentInput
	autoRegister     bool
}

func resolveOrderTrackingUpdate(
	shippingService *ShippingService,
	existingOrder *order.Order,
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
	if !hasPositiveID(carrierIDInput) && !hasPositiveID(carrierServiceIDInput) && existingOrder != nil {
		carrierIDInput = existingOrder.CarrierID
		carrierServiceIDInput = existingOrder.CarrierServiceID
	}

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

	return &resolvedOrderTrackingUpdate{
		trackingInfo: order.TrackingInfoUpdate{
			TrackingNumber:           trackingNumber,
			TrackingProviderID:       uintPtr(resolution.Provider.ID),
			CarrierID:                carrierID,
			CarrierServiceID:         carrierServiceID,
			TrackingCarrierMappingID: uintPtr(resolution.Mapping.ID),
			ProviderCarrierCode:      resolution.ProviderCarrierCode,
			ProviderCarrierName:      resolution.ProviderCarrierName,
		},
		trackingShipment: TrackingShipmentInput{
			TrackingProviderID:       resolution.Provider.ID,
			TrackingNumber:           trackingNumber,
			ProviderCarrierCode:      resolution.ProviderCarrierCode,
			CarrierID:                carrierID,
			CarrierServiceID:         carrierServiceID,
			TrackingCarrierMappingID: uintPtr(resolution.Mapping.ID),
		},
		autoRegister: resolution.Provider.AutoRegister,
	}, nil
}

func (s *OrderService) UpdateTrackingInfo(ctx context.Context, id uint, input OrderTrackingUpdateInput) error {
	o, err := s.findOrder(id)
	if err != nil {
		return err
	}

	resolvedTracking, err := resolveOrderTrackingUpdate(s.shipping, o, input)
	if err != nil {
		return err
	}
	resolvedTracking.trackingShipment.OrderID = id

	if s.txManager != nil {
		if err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
			if repos.Shipping == nil {
				return ErrOrderShippingNotConfigured
			}
			if err := repos.Order.UpdateTrackingInfo(id, resolvedTracking.trackingInfo); err != nil {
				return err
			}

			shippingService := NewShippingService(repos.Shipping)
			_, err := shippingService.UpsertTrackingShipment(resolvedTracking.trackingShipment)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	} else {
		if err := s.orderRepo.UpdateTrackingInfo(id, resolvedTracking.trackingInfo); err != nil {
			return err
		}
		trackingShipment, err := s.shipping.UpsertTrackingShipment(resolvedTracking.trackingShipment)
		if err != nil {
			return err
		}
		_ = trackingShipment
	}

	if resolvedTracking.autoRegister {
		return s.shipping.RegisterTrackingShipment(ctx, TrackingSyncInput{
			OrderID:                  id,
			ProviderID:               resolvedTracking.trackingShipment.TrackingProviderID,
			TrackingNumber:           resolvedTracking.trackingShipment.TrackingNumber,
			ProviderCarrierCode:      resolvedTracking.trackingShipment.ProviderCarrierCode,
			CarrierID:                resolvedTracking.trackingShipment.CarrierID,
			CarrierServiceID:         resolvedTracking.trackingShipment.CarrierServiceID,
			TrackingCarrierMappingID: resolvedTracking.trackingShipment.TrackingCarrierMappingID,
		})
	}

	return nil
}

// FulfillOrder records a dispatch as one transaction: the order, its shipping
// status, and its 17TRACK task either persist together or do not change.
// Registration happens after commit so a temporary provider outage never blocks
// a parcel that has already been physically handed to the carrier.
func (s *OrderService) FulfillOrder(ctx context.Context, id uint, input OrderTrackingUpdateInput) (*OrderFulfillmentResult, error) {
	if s.txManager == nil {
		return nil, ErrOrderFulfillmentTransactionNeeded
	}
	if s.shipping == nil {
		return nil, ErrOrderShippingNotConfigured
	}

	var resolvedTracking *resolvedOrderTrackingUpdate
	if err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.Order == nil || repos.Shipping == nil {
			return ErrOrderShippingNotConfigured
		}

		o, err := repos.Order.FindByIDForUpdateWithItems(id)
		if err != nil {
			return normalizeOrderError(err)
		}
		if o.PaymentStatus != "paid" {
			return ErrOrderFulfillmentPaymentRequired
		}
		if o.Status != "paid" && o.Status != "processing" && o.Status != "shipped" {
			return fmt.Errorf("%w: %s", ErrOrderFulfillmentNotAllowed, o.Status)
		}

		txShippingService := NewShippingService(repos.Shipping)
		resolvedTracking, err = resolveOrderTrackingUpdate(txShippingService, o, input)
		if err != nil {
			return err
		}
		resolvedTracking.trackingShipment.OrderID = id

		if err := repos.Order.UpdateTrackingInfo(id, resolvedTracking.trackingInfo); err != nil {
			return err
		}
		if _, err := txShippingService.UpsertTrackingShipment(resolvedTracking.trackingShipment); err != nil {
			return err
		}

		shippedAt := time.Now().UTC()
		if o.ShippedAt != nil && !o.ShippedAt.IsZero() {
			shippedAt = o.ShippedAt.UTC()
		} else {
			o.ShippedAt = &shippedAt
		}
		if o.Status != "shipped" {
			if err := repos.Order.UpdateStatus(id, "shipped"); err != nil {
				return err
			}
		}
		if o.ShippingStatus != "shipped" {
			if err := repos.Order.UpdateShippingStatus(id, "shipped"); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	result := &OrderFulfillmentResult{}
	if resolvedTracking != nil && resolvedTracking.autoRegister {
		if err := s.shipping.RegisterTrackingShipment(ctx, TrackingSyncInput{
			OrderID:                  id,
			ProviderID:               resolvedTracking.trackingShipment.TrackingProviderID,
			TrackingNumber:           resolvedTracking.trackingShipment.TrackingNumber,
			ProviderCarrierCode:      resolvedTracking.trackingShipment.ProviderCarrierCode,
			CarrierID:                resolvedTracking.trackingShipment.CarrierID,
			CarrierServiceID:         resolvedTracking.trackingShipment.CarrierServiceID,
			TrackingCarrierMappingID: resolvedTracking.trackingShipment.TrackingCarrierMappingID,
		}); err != nil {
			result.TrackingRegistrationError = err.Error()
		}
	}

	fulfilledOrder, err := s.GetAdminOrder(id)
	if err != nil {
		return nil, err
	}
	trackingShipment, err := s.GetAdminOrderTrackingShipment(id)
	if err != nil {
		return nil, err
	}
	result.Order = fulfilledOrder
	result.TrackingShipment = trackingShipment
	return result, nil
}

func (s *OrderService) SyncOrderTracking(ctx context.Context, id uint) (*TrackingSyncResult, error) {
	o, err := s.findOrder(id)
	if err != nil {
		return nil, err
	}
	if s.shipping == nil {
		return nil, ErrOrderShippingNotConfigured
	}
	if strings.TrimSpace(o.TrackingNumber) == "" {
		return nil, ErrTrackingNumberRequired
	}
	if !hasPositiveID(o.TrackingProviderID) {
		return nil, ErrTrackingProviderRequired
	}
	if strings.TrimSpace(o.ProviderCarrierCode) == "" {
		return nil, ErrTrackingCarrierCodeRequired
	}

	return s.shipping.SyncTracking(ctx, TrackingSyncInput{
		OrderID:                  o.ID,
		ProviderID:               *o.TrackingProviderID,
		TrackingNumber:           o.TrackingNumber,
		ProviderCarrierCode:      o.ProviderCarrierCode,
		CarrierID:                o.CarrierID,
		CarrierServiceID:         o.CarrierServiceID,
		TrackingCarrierMappingID: o.TrackingCarrierMappingID,
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

func (s *OrderService) UpdateOrderItemCustoms(orderID, orderItemID uint, declaredValue *float64, confirmed bool) error {
	if orderID == 0 || orderItemID == 0 {
		return ErrOrderItemNotFound
	}
	if declaredValue != nil && (math.IsNaN(*declaredValue) || math.IsInf(*declaredValue, 0) || *declaredValue < 0) {
		return ErrDeclaredValueInvalid
	}
	if confirmed && declaredValue == nil {
		return ErrDeclaredValueConfirmationRequired
	}

	o, err := s.findOrder(orderID)
	if err != nil {
		return err
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
	return s.orderRepo.UpdateOrderItemCustoms(orderID, orderItemID, declaredValue, confirmed)
}

func (s *OrderService) DeleteAdminOrder(id uint) error {
	o, err := s.findOrder(id)
	if err != nil {
		return err
	}

	if o.Status != "cancelled" && o.Status != "refunded" && o.Status != "payment_expired" {
		return ErrOrderDeleteNotAllowed
	}

	return s.orderRepo.Delete(id)
}

func (s *OrderService) GetAdminStats() (map[string]interface{}, error) {
	return s.orderRepo.GetStats()
}

func (s *OrderService) GetSalesByDateRange(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	return s.orderRepo.GetSalesByDateRange(startDate, endDate)
}
