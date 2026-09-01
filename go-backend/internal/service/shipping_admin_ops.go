package service

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

func (s *ShippingService) ListCarriers(enabledOnly bool) ([]shipping.Carrier, error) {
	return s.shippingRepo.FindAllCarriers(enabledOnly)
}

func (s *ShippingService) GetCarrier(id uint) (*shipping.Carrier, error) {
	return s.shippingRepo.FindCarrierByID(id)
}

func (s *ShippingService) CreateCarrier(carrier *shipping.Carrier) error {
	return s.shippingRepo.CreateCarrier(carrier)
}

func (s *ShippingService) UpdateCarrier(carrier *shipping.Carrier) error {
	return s.shippingRepo.UpdateCarrier(carrier)
}

func (s *ShippingService) DeleteCarrier(id uint) error {
	return s.shippingRepo.DeleteCarrier(id)
}

func (s *ShippingService) ListTrackingProviderConfigs(enabledOnly bool) ([]shipping.TrackingProviderConfig, error) {
	return s.shippingRepo.FindAllTrackingProviderConfigs(enabledOnly)
}

func (s *ShippingService) GetTrackingProviderConfig(id uint) (*shipping.TrackingProviderConfig, error) {
	return s.shippingRepo.FindTrackingProviderConfigByID(id)
}

func (s *ShippingService) GetTrackingProviderConfigByCode(providerCode string) (*shipping.TrackingProviderConfig, error) {
	return s.shippingRepo.FindTrackingProviderConfigByCode(providerCode)
}

func (s *ShippingService) CreateTrackingProviderConfig(provider *shipping.TrackingProviderConfig) error {
	return s.shippingRepo.CreateTrackingProviderConfig(provider)
}

func (s *ShippingService) UpdateTrackingProviderConfig(provider *shipping.TrackingProviderConfig) error {
	return s.shippingRepo.UpdateTrackingProviderConfig(provider)
}

func (s *ShippingService) DeleteTrackingProviderConfig(id uint) error {
	return s.shippingRepo.DeleteTrackingProviderConfig(id)
}

func (s *ShippingService) ListTrackingCarrierMappings(enabledOnly bool) ([]shipping.TrackingCarrierMapping, error) {
	return s.shippingRepo.FindAllTrackingCarrierMappings(enabledOnly)
}

func (s *ShippingService) GetTrackingCarrierMapping(id uint) (*shipping.TrackingCarrierMapping, error) {
	return s.shippingRepo.FindTrackingCarrierMappingByID(id)
}

func (s *ShippingService) CreateTrackingCarrierMapping(mapping *shipping.TrackingCarrierMapping) error {
	return s.shippingRepo.CreateTrackingCarrierMapping(mapping)
}

func (s *ShippingService) UpdateTrackingCarrierMapping(mapping *shipping.TrackingCarrierMapping) error {
	return s.shippingRepo.UpdateTrackingCarrierMapping(mapping)
}

func (s *ShippingService) DeleteTrackingCarrierMapping(id uint) error {
	return s.shippingRepo.DeleteTrackingCarrierMapping(id)
}

func (s *ShippingService) ResolveTrackingCarrier(input TrackingCarrierResolutionInput) (*TrackingCarrierResolution, error) {
	if input.ProviderID == 0 {
		return nil, ErrTrackingProviderRequired
	}
	if !hasPositiveID(input.CarrierID) && !hasPositiveID(input.CarrierServiceID) {
		return nil, ErrTrackingLocalTargetRequired
	}

	provider, err := s.GetTrackingProviderConfig(input.ProviderID)
	if err != nil {
		return nil, err
	}
	if !provider.Enabled {
		return nil, fmt.Errorf("%w: %s", ErrTrackingProviderDisabled, provider.ProviderName)
	}

	var carrier *shipping.Carrier
	var carrierID *uint
	if hasPositiveID(input.CarrierID) {
		carrierID = input.CarrierID
	}

	var carrierService *shipping.CarrierService
	if hasPositiveID(input.CarrierServiceID) {
		carrierService, err = s.GetCarrierService(*input.CarrierServiceID)
		if err != nil {
			return nil, err
		}
		if !carrierService.Enabled {
			return nil, fmt.Errorf("%w: %s", ErrTrackingCarrierServiceDisabled, carrierService.ServiceName)
		}
		if carrierService.Carrier != nil {
			carrier = carrierService.Carrier
		}
		serviceCarrierID := carrierService.CarrierID
		carrierID = &serviceCarrierID

		mapping, err := s.shippingRepo.FindEnabledTrackingCarrierMappingByCarrierService(provider.ID, carrierService.ID)
		if err == nil {
			return trackingCarrierResolution(provider, carrier, carrierService, mapping), nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	if hasPositiveID(carrierID) {
		if carrier == nil || carrier.ID != *carrierID {
			carrier, err = s.GetCarrier(*carrierID)
			if err != nil {
				return nil, err
			}
		}
		if !carrier.Enabled {
			return nil, fmt.Errorf("%w: %s", ErrTrackingCarrierDisabled, carrier.Name)
		}

		mapping, err := s.shippingRepo.FindEnabledTrackingCarrierMappingByCarrier(provider.ID, carrier.ID)
		if err == nil {
			return trackingCarrierResolution(provider, carrier, carrierService, mapping), nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	return nil, ErrTrackingCarrierMappingMissing
}

func (s *ShippingService) ListCarrierServices(enabledOnly bool) ([]shipping.CarrierService, error) {
	return s.shippingRepo.FindAllCarrierServices(enabledOnly)
}

func (s *ShippingService) GetCarrierService(id uint) (*shipping.CarrierService, error) {
	return s.shippingRepo.FindCarrierServiceByID(id)
}

func (s *ShippingService) CreateCarrierService(service *shipping.CarrierService) error {
	if err := s.prepareCarrierServiceCurrency(service); err != nil {
		return err
	}
	return s.shippingRepo.CreateCarrierService(service)
}

func (s *ShippingService) UpdateCarrierService(service *shipping.CarrierService) error {
	if err := s.prepareCarrierServiceCurrency(service); err != nil {
		return err
	}
	return s.shippingRepo.UpdateCarrierService(service)
}

func (s *ShippingService) prepareCarrierServiceCurrency(carrierService *shipping.CarrierService) error {
	if carrierService == nil {
		return errors.New("carrier service is required")
	}
	if strings.TrimSpace(carrierService.Currency) == "" && carrierService.ID > 0 {
		existing, err := s.GetCarrierService(carrierService.ID)
		if err != nil {
			return err
		}
		carrierService.Currency = existing.Currency
	}
	normalized, err := s.normalizeShippingSourceCurrency(carrierService.Currency)
	if err != nil {
		return err
	}
	carrierService.Currency = normalized
	return nil
}

func (s *ShippingService) DeleteCarrierService(id uint) error {
	return s.shippingRepo.DeleteCarrierService(id)
}

func (s *ShippingService) ListZones() ([]shipping.ShippingZone, error) {
	return s.shippingRepo.FindAllZones()
}

func (s *ShippingService) GetZone(id uint) (*shipping.ShippingZone, error) {
	return s.shippingRepo.FindZoneByID(id)
}

func (s *ShippingService) CreateZone(zone *shipping.ShippingZone) error {
	return s.shippingRepo.CreateZone(zone)
}

func (s *ShippingService) UpdateZone(zone *shipping.ShippingZone) error {
	return s.shippingRepo.UpdateZone(zone)
}

func (s *ShippingService) DeleteZone(id uint) error {
	return s.shippingRepo.DeleteZone(id)
}

func (s *ShippingService) GetTrackingEventsByTrackingNumber(trackingNumber string) ([]shipping.TrackingEvent, error) {
	return s.shippingRepo.FindTrackingEventsByTrackingNumber(trackingNumber)
}

func (s *ShippingService) GetTrackingShipmentByTrackingNumber(trackingNumber string) (*shipping.TrackingShipment, error) {
	return s.shippingRepo.FindTrackingShipmentByTrackingNumber(trackingNumber)
}

func (s *ShippingService) GetTrackingEventsByOrderID(orderID uint) ([]shipping.TrackingEvent, error) {
	return s.shippingRepo.FindTrackingEventsByOrderID(orderID)
}

func (s *ShippingService) ListPackagingRules() ([]shipping.PackagingRule, error) {
	return s.shippingRepo.FindAllPackagingRules()
}

func (s *ShippingService) GetPackagingRule(id uint) (*shipping.PackagingRule, error) {
	return s.shippingRepo.FindPackagingRuleByID(id)
}

func (s *ShippingService) CreatePackagingRule(rule *shipping.PackagingRule) error {
	return s.shippingRepo.CreatePackagingRule(rule)
}

func (s *ShippingService) UpdatePackagingRule(rule *shipping.PackagingRule) error {
	return s.shippingRepo.UpdatePackagingRule(rule)
}

func (s *ShippingService) DeletePackagingRule(id uint) error {
	return s.shippingRepo.DeletePackagingRule(id)
}

func (s *ShippingService) CreatePackagingRuleApply(apply *shipping.PackagingRuleApply) error {
	if apply == nil {
		return errors.New("packaging rule apply is required")
	}
	if apply.RuleID == 0 {
		return errors.New("packaging rule id is required")
	}
	if apply.ProductID == 0 {
		return errors.New("product id is required")
	}

	if _, err := s.shippingRepo.FindPackagingRuleByID(apply.RuleID); err != nil {
		if repository.IsRecordNotFound(err) {
			return fmt.Errorf("packaging rule ID %d does not exist", apply.RuleID)
		}
		return err
	}

	if s.productRepo != nil {
		if _, err := s.productRepo.FindByID(apply.ProductID); err != nil {
			if repository.IsRecordNotFound(err) {
				return fmt.Errorf("product ID %d does not exist", apply.ProductID)
			}
			return err
		}
	}

	existing, err := s.shippingRepo.FindPackagingRuleApplyByProductID(apply.ProductID)
	if err == nil && existing != nil && existing.ID > 0 {
		if existing.RuleID == apply.RuleID {
			return errors.New("packaging rule already applies to this product")
		}
		return errors.New("product already has a packaging rule")
	}
	if err != nil && !repository.IsRecordNotFound(err) {
		return err
	}

	return s.shippingRepo.CreatePackagingRuleApply(apply)
}

func (s *ShippingService) DeletePackagingRuleApply(id uint) error {
	return s.shippingRepo.DeletePackagingRuleApply(id)
}
