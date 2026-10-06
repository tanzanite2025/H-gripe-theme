package service

import (
	"fmt"
	"strings"

	"commerce-platform/internal/domain/shipping"
)

// projectCarrierServicesFromPublishedCollections refreshes collection-backed
// routes from the current 4PX and Yanwen publication state. Collection-backed
// routes are resolved by their stable collection record ID. Unpublished,
// disabled, removed, or legacy-unlinked routes are omitted so public reads and
// new quotes fail closed while generic carrier routes retain their stored
// configuration.
func (s *ShippingService) projectCarrierServicesFromPublishedCollections(
	services []shipping.CarrierService,
) ([]shipping.CarrierService, map[uint]struct{}, error) {
	projectedServices := make([]shipping.CarrierService, 0, len(services))
	templatesWithPublishedCollectionRoutes := make(map[uint]struct{})
	needsFpxCollection := false
	needsYanwenCollection := false

	for i := range services {
		if services[i].Template != nil && services[i].Template.IsSystemFreeShippingTemplate() {
			continue
		}
		provider := publishedCarrierServiceProvider(services[i])
		serviceCode := strings.ToUpper(strings.TrimSpace(services[i].ServiceCode))
		isYanwenService := provider == "YANWEN" || strings.HasPrefix(serviceCode, "YANWEN:")
		if isYanwenService {
			if services[i].TemplateID != nil {
				templatesWithPublishedCollectionRoutes[*services[i].TemplateID] = struct{}{}
			}
			needsYanwenCollection = true
			continue
		}
		if provider == "4PX" {
			if services[i].TemplateID != nil {
				templatesWithPublishedCollectionRoutes[*services[i].TemplateID] = struct{}{}
			}
			needsFpxCollection = true
		}
	}

	fpxChannelsByID := make(map[uint]shipping.FpxChannel)
	if needsFpxCollection {
		channels, err := s.shippingRepo.FindAllFpxChannelsForEnvironment(shipping.FpxChannelEnvironmentProduction, true)
		if err != nil {
			return nil, nil, fmt.Errorf("load enabled 4PX service collection: %w", err)
		}
		for _, channel := range channels {
			fpxChannelsByID[channel.ID] = channel
		}
	}

	yanwenChannelsByID := make(map[uint]YanwenPublishedCollectionReference)
	if needsYanwenCollection {
		if s.yanwenPublishedCollection == nil {
			return nil, nil, fmt.Errorf("load enabled Yanwen service collection: service is not configured")
		}
		channels, err := s.yanwenPublishedCollection.ListProductionYanwenCollectionReferences()
		if err != nil {
			return nil, nil, fmt.Errorf("load enabled Yanwen service collection: %w", err)
		}
		for _, channel := range channels {
			yanwenChannelsByID[channel.ID] = channel
		}
	}

	for i := range services {
		carrierService := services[i]
		if carrierService.Template != nil && carrierService.Template.IsSystemFreeShippingTemplate() {
			continue
		}
		provider := publishedCarrierServiceProvider(carrierService)
		serviceCode := strings.ToUpper(strings.TrimSpace(carrierService.ServiceCode))
		isYanwenCode := strings.HasPrefix(serviceCode, "YANWEN:")

		if provider == "YANWEN" || isYanwenCode {
			if isYanwenCode && provider != "" && provider != "YANWEN" {
				continue
			}
			if carrierService.FpxChannelID != nil || carrierService.YanwenPublishedChannelID == nil {
				continue
			}
			channel, ok := yanwenChannelsByID[*carrierService.YanwenPublishedChannelID]
			if !ok {
				continue
			}
			carrierService.ProviderCode = "YANWEN"
			carrierService.FpxChannelID = nil
			carrierService.YanwenPublishedChannelID = &channel.ID
			carrierService.ServiceCode = "YANWEN:" + normalizeYanwenProductCode(channel.ProductCode)
			carrierService.ServiceName = channel.DisplayName
			carrierService.Countries = shipping.NormalizeShippingServiceCollectionCountryCodes(channel.Countries)
			projectedServices = append(projectedServices, carrierService)
			continue
		}

		if provider == "4PX" {
			if carrierService.YanwenPublishedChannelID != nil || carrierService.FpxChannelID == nil {
				continue
			}
			channel, ok := fpxChannelsByID[*carrierService.FpxChannelID]
			if !ok {
				continue
			}
			carrierService.ProviderCode = "4PX"
			carrierService.FpxChannelID = &channel.ID
			carrierService.YanwenPublishedChannelID = nil
			carrierService.ServiceCode = channel.ServiceCode
			carrierService.ServiceName = channel.DisplayName
			carrierService.Countries = shipping.NormalizeShippingServiceCollectionCountryCodes(channel.Countries)
			projectedServices = append(projectedServices, carrierService)
			continue
		}

		if carrierService.FpxChannelID != nil || carrierService.YanwenPublishedChannelID != nil {
			if carrierService.TemplateID != nil {
				templatesWithPublishedCollectionRoutes[*carrierService.TemplateID] = struct{}{}
			}
			continue
		}
		projectedServices = append(projectedServices, carrierService)
	}

	return projectedServices, templatesWithPublishedCollectionRoutes, nil
}

func publishedCarrierServiceProvider(carrierService shipping.CarrierService) string {
	if carrierService.Carrier == nil {
		return normalizePublishedCarrierProviderCode(carrierService.ProviderCode)
	}
	return normalizePublishedCarrierProviderCode(carrierService.Carrier.Code)
}

func normalizeYanwenProductCode(value string) string {
	productCode := strings.ToUpper(strings.TrimSpace(value))
	return strings.TrimPrefix(productCode, "YANWEN:")
}
