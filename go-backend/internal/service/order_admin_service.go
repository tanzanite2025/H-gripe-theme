package service

import (
	"context"
	"fmt"
	"strings"
	"tanzanite/internal/domain/order"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/repository"
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
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return normalizeOrderError(err)
	}

	if isSystemManagedOrderStatus(status) {
		return fmt.Errorf("%w: %s", ErrSystemManagedOrderStatus, status)
	}

	if !o.CanTransitionTo(status) {
		return fmt.Errorf("invalid status transition from %s to %s", o.Status, status)
	}

	if status == "cancelled" {
		return s.cancelOrderWithRollback(o)
	}

	return s.orderRepo.UpdateStatus(id, status)
}

func isSystemManagedOrderStatus(status string) bool {
	return status == "paid" || status == "refunded"
}

func (s *OrderService) UpdateShippingStatus(id uint, shippingStatus string) error {
	if _, err := s.findOrder(id); err != nil {
		return err
	}

	return s.orderRepo.UpdateShippingStatus(id, shippingStatus)
}

func (s *OrderService) UpdateTrackingInfo(ctx context.Context, id uint, input OrderTrackingUpdateInput) error {
	o, err := s.findOrder(id)
	if err != nil {
		return err
	}

	trackingNumber := strings.TrimSpace(input.TrackingNumber)
	if trackingNumber == "" {
		return ErrTrackingNumberRequired
	}
	if s.shipping == nil {
		return ErrOrderShippingNotConfigured
	}

	carrierIDInput := input.CarrierID
	carrierServiceIDInput := input.CarrierServiceID
	if !hasPositiveID(carrierIDInput) && !hasPositiveID(carrierServiceIDInput) {
		carrierIDInput = o.CarrierID
		carrierServiceIDInput = o.CarrierServiceID
	}

	resolution, err := s.shipping.ResolveTrackingCarrier(TrackingCarrierResolutionInput{
		ProviderID:       input.TrackingProviderID,
		CarrierID:        carrierIDInput,
		CarrierServiceID: carrierServiceIDInput,
	})
	if err != nil {
		return err
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

	trackingInfo := order.TrackingInfoUpdate{
		TrackingNumber:           trackingNumber,
		TrackingProviderID:       uintPtr(resolution.Provider.ID),
		CarrierID:                carrierID,
		CarrierServiceID:         carrierServiceID,
		TrackingCarrierMappingID: uintPtr(resolution.Mapping.ID),
		ProviderCarrierCode:      resolution.ProviderCarrierCode,
		ProviderCarrierName:      resolution.ProviderCarrierName,
	}
	trackingShipment := TrackingShipmentInput{
		OrderID:                  id,
		TrackingProviderID:       resolution.Provider.ID,
		TrackingNumber:           trackingNumber,
		ProviderCarrierCode:      resolution.ProviderCarrierCode,
		CarrierID:                carrierID,
		CarrierServiceID:         carrierServiceID,
		TrackingCarrierMappingID: uintPtr(resolution.Mapping.ID),
	}

	if s.txManager != nil {
		if err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
			if repos.Shipping == nil {
				return ErrOrderShippingNotConfigured
			}
			if err := repos.Order.UpdateTrackingInfo(id, trackingInfo); err != nil {
				return err
			}

			shippingService := NewShippingService(repos.Shipping)
			_, err := shippingService.UpsertTrackingShipment(trackingShipment)
			return err
		}); err != nil {
			return err
		}
	} else {
		if err := s.orderRepo.UpdateTrackingInfo(id, trackingInfo); err != nil {
			return err
		}
		if _, err = s.shipping.UpsertTrackingShipment(trackingShipment); err != nil {
			return err
		}
	}

	if resolution.Provider.AutoRegister {
		return s.shipping.RegisterTrackingShipment(ctx, TrackingSyncInput{
			OrderID:                  id,
			ProviderID:               resolution.Provider.ID,
			TrackingNumber:           trackingNumber,
			ProviderCarrierCode:      resolution.ProviderCarrierCode,
			CarrierID:                carrierID,
			CarrierServiceID:         carrierServiceID,
			TrackingCarrierMappingID: uintPtr(resolution.Mapping.ID),
		})
	}

	return nil
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

func (s *OrderService) DeleteAdminOrder(id uint) error {
	o, err := s.findOrder(id)
	if err != nil {
		return err
	}

	if o.Status != "cancelled" && o.Status != "refunded" {
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
