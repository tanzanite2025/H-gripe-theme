package service

import (
	"commerce-platform/internal/domain/shipping"
	"encoding/json"
	"sort"
)

// hydrateShippingRouteAggregatePublishedDestinations refreshes identity and
// country read models from published collections. It never writes provider
// prices into customer templates or persists quote calculations.
func (s *ShippingService) hydrateShippingRouteAggregatePublishedDestinations(groups []shipping.ShippingRouteGroup) error {
	services := make([]shipping.CarrierService, 0)
	for _, group := range groups {
		services = append(services, group.CarrierServices...)
	}
	projected, _, err := s.projectCarrierServicesFromPublishedCollections(services)
	if err != nil {
		return err
	}
	byID := make(map[uint]shipping.CarrierService, len(projected))
	for _, service := range projected {
		byID[service.ID] = service
	}
	for index := range groups {
		countries := make(map[string]struct{})
		for memberIndex, original := range groups[index].CarrierServices {
			current, exists := byID[original.ID]
			if !exists {
				groups[index].CarrierServices[memberIndex].Enabled = false
				continue
			}
			groups[index].CarrierServices[memberIndex] = current
			if !current.Enabled || current.Carrier == nil || !current.Carrier.Enabled {
				continue
			}
			var codes []string
			if err := json.Unmarshal([]byte(shipping.NormalizeShippingServiceCollectionCountryCodes(current.Countries)), &codes); err != nil {
				return err
			}
			for _, code := range codes {
				countries[code] = struct{}{}
			}
		}
		codes := make([]string, 0, len(countries))
		for code := range countries {
			codes = append(codes, code)
		}
		sort.Strings(codes)
		encoded, err := json.Marshal(codes)
		if err != nil {
			return err
		}
		groups[index].Countries = string(encoded)
	}
	return nil
}
