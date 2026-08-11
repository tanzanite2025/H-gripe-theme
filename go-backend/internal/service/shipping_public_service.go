package service

import "commerce-platform/internal/domain/shipping"

func (s *ShippingService) ListPublicTemplates() ([]shipping.ShippingTemplate, error) {
	templates, err := s.ListTemplates()
	if err != nil {
		return nil, err
	}

	publicTemplates := make([]shipping.ShippingTemplate, 0, len(templates))
	for _, template := range templates {
		if template.Enabled {
			publicTemplates = append(publicTemplates, template)
		}
	}
	return publicTemplates, nil
}

func (s *ShippingService) GetPublicTemplate(id uint) (*shipping.ShippingTemplate, error) {
	template, err := s.GetTemplate(id)
	if err != nil {
		return nil, err
	}
	if !template.Enabled {
		return nil, ErrShippingNotFound
	}
	return template, nil
}

func (s *ShippingService) ListPublicCarriers() ([]shipping.Carrier, error) {
	return s.ListCarriers(true)
}

func (s *ShippingService) GetPublicCarrier(id uint) (*shipping.Carrier, error) {
	carrier, err := s.GetCarrier(id)
	if err != nil {
		return nil, err
	}
	if !carrier.Enabled {
		return nil, ErrShippingNotFound
	}
	return carrier, nil
}

func (s *ShippingService) ListPublicCarrierServices() ([]shipping.CarrierService, error) {
	services, err := s.ListCarrierServices(true)
	if err != nil {
		return nil, err
	}

	publicServices := make([]shipping.CarrierService, 0, len(services))
	for _, service := range services {
		if publicCarrierServiceVisible(&service) {
			publicServices = append(publicServices, service)
		}
	}
	return publicServices, nil
}

func (s *ShippingService) GetPublicCarrierService(id uint) (*shipping.CarrierService, error) {
	service, err := s.GetCarrierService(id)
	if err != nil {
		return nil, err
	}
	if !publicCarrierServiceVisible(service) {
		return nil, ErrShippingNotFound
	}
	return service, nil
}

func publicCarrierServiceVisible(service *shipping.CarrierService) bool {
	if service == nil || !service.Enabled {
		return false
	}
	if service.Carrier == nil || !service.Carrier.Enabled {
		return false
	}
	if service.Template != nil && !service.Template.Enabled {
		return false
	}
	return true
}

func (s *ShippingService) ListPublicZones() ([]shipping.ShippingZone, error) {
	zones, err := s.ListZones()
	if err != nil {
		return nil, err
	}

	publicZones := make([]shipping.ShippingZone, 0, len(zones))
	for _, zone := range zones {
		if zone.Enabled {
			publicZones = append(publicZones, zone)
		}
	}
	return publicZones, nil
}

func (s *ShippingService) GetPublicZone(id uint) (*shipping.ShippingZone, error) {
	zone, err := s.GetZone(id)
	if err != nil {
		return nil, err
	}
	if !zone.Enabled {
		return nil, ErrShippingNotFound
	}
	return zone, nil
}

func (s *ShippingService) ListPublicPackagingRules() ([]shipping.PackagingRule, error) {
	rules, err := s.ListPackagingRules()
	if err != nil {
		return nil, err
	}

	publicRules := make([]shipping.PackagingRule, 0, len(rules))
	for _, rule := range rules {
		if rule.IsActive {
			publicRules = append(publicRules, rule)
		}
	}
	return publicRules, nil
}

func (s *ShippingService) GetPublicPackagingRule(id uint) (*shipping.PackagingRule, error) {
	rule, err := s.GetPackagingRule(id)
	if err != nil {
		return nil, err
	}
	if !rule.IsActive {
		return nil, ErrShippingNotFound
	}
	return rule, nil
}

func (s *ShippingService) GetPublicProductPackagingRules(productID uint) ([]shipping.PackagingRule, error) {
	if s.productRepo != nil {
		product, err := s.productRepo.FindByID(productID)
		if err != nil {
			return nil, err
		}
		if product.Status != "active" {
			return nil, ErrProductNotFound
		}
	}
	return s.GetProductPackagingRules(productID)
}
