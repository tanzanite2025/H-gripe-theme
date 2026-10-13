package shipping

import (
	shippingdomain "commerce-platform/internal/domain/shipping"

	"gorm.io/datatypes"
)

// PublicShippingCustomerBillingTemplateResponse exposes customer pricing only.
// Provider settlement fields must never enter this contract, including through
// nested route aggregates or carrier identities.
type PublicShippingCustomerBillingTemplateResponse struct {
	ID                                  uint                                          `json:"id"`
	Name                                string                                        `json:"name"`
	Type                                string                                        `json:"type"`
	TemplateKind                        string                                        `json:"template_kind"`
	IsSystemManaged                     bool                                          `json:"is_system_managed"`
	RouteGroupID                        *uint                                         `json:"route_group_id,omitempty"`
	Currency                            string                                        `json:"currency"`
	FreeShipping                        bool                                          `json:"free_shipping"`
	FreeThresholdMinor                  int64                                         `json:"free_threshold_minor"`
	DefaultFeeMinor                     int64                                         `json:"default_fee_minor"`
	CustomerBillingMode                 string                                        `json:"customer_billing_mode"`
	CustomerFirstWeightGrams            int                                           `json:"customer_first_weight_grams"`
	CustomerAdditionalWeightGrams       int                                           `json:"customer_additional_weight_grams"`
	CustomerMinChargeWeightGrams        int                                           `json:"customer_min_charge_weight_grams"`
	CustomerVolumetricDivisor           int                                           `json:"customer_volumetric_divisor"`
	CustomerFuelSurchargePercentDecimal string                                        `json:"customer_fuel_surcharge_percent_decimal"`
	CustomerRemoteSurchargeMinor        int64                                         `json:"customer_remote_surcharge_minor"`
	CustomerRemotePostalCodes           string                                        `json:"customer_remote_postal_codes"`
	CustomerEtaMinDays                  int                                           `json:"customer_eta_min_days"`
	CustomerEtaMaxDays                  int                                           `json:"customer_eta_max_days"`
	FreeShippingCountries               string                                        `json:"free_shipping_countries"`
	DisplayPriceSnapshots               datatypes.JSON                                `json:"display_price_snapshots,omitempty"`
	Description                         string                                        `json:"description"`
	Enabled                             bool                                          `json:"enabled"`
	RouteGroup                          *PublicShippingRouteAggregateIdentityResponse `json:"route_group,omitempty"`
	Rules                               []PublicShippingCustomerBillingRuleResponse   `json:"rules"`
}

type PublicShippingCustomerBillingRuleResponse struct {
	ID                    uint           `json:"id"`
	TemplateID            uint           `json:"template_id"`
	Region                string         `json:"region"`
	Currency              string         `json:"currency"`
	MinValueMinor         int64          `json:"min_value_minor"`
	MaxValueMinor         int64          `json:"max_value_minor"`
	FeeMinor              int64          `json:"fee_minor"`
	AdditionalMinor       int64          `json:"additional_minor"`
	DisplayPriceSnapshots datatypes.JSON `json:"display_price_snapshots,omitempty"`
}

type PublicShippingRouteAggregateIdentityResponse struct {
	ID              uint                                           `json:"id"`
	Name            string                                         `json:"name"`
	Countries       string                                         `json:"countries"`
	Enabled         bool                                           `json:"enabled"`
	CarrierServices []PublicShippingCarrierServiceIdentityResponse `json:"carrier_services"`
}

// PublicShippingCarrierServiceIdentityResponse identifies a route. Its billing
// mode, weight steps, surcharges, currency and ETA are deliberately absent:
// buyer-facing values are owned by the customer template and quoted plan.
type PublicShippingCarrierServiceIdentityResponse struct {
	ID          uint                                   `json:"id"`
	CarrierID   uint                                   `json:"carrier_id"`
	ServiceCode string                                 `json:"service_code"`
	ServiceName string                                 `json:"service_name"`
	RouteName   string                                 `json:"route_name,omitempty"`
	Countries   string                                 `json:"countries"`
	Enabled     bool                                   `json:"enabled"`
	SortOrder   int                                    `json:"sort_order"`
	Carrier     *PublicShippingCarrierIdentityResponse `json:"carrier,omitempty"`
}

type PublicShippingCarrierIdentityResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	TrackingURL string `json:"tracking_url"`
}

func buildPublicShippingCustomerBillingTemplateResponse(template shippingdomain.ShippingTemplate) PublicShippingCustomerBillingTemplateResponse {
	result := PublicShippingCustomerBillingTemplateResponse{
		ID: template.ID, Name: template.Name, Type: template.Type,
		TemplateKind: template.TemplateKind, IsSystemManaged: template.IsSystemManaged,
		RouteGroupID: template.RouteGroupID, Currency: template.Currency,
		FreeShipping: template.FreeShipping, FreeThresholdMinor: template.FreeThresholdMinor,
		DefaultFeeMinor: template.DefaultFeeMinor, CustomerBillingMode: template.CustomerBillingMode,
		CustomerFirstWeightGrams:            template.CustomerFirstWeightGrams,
		CustomerAdditionalWeightGrams:       template.CustomerAdditionalWeightGrams,
		CustomerMinChargeWeightGrams:        template.CustomerMinChargeWeightGrams,
		CustomerVolumetricDivisor:           template.CustomerVolumetricDivisor,
		CustomerFuelSurchargePercentDecimal: template.CustomerFuelSurchargePercentDecimal,
		CustomerRemoteSurchargeMinor:        template.CustomerRemoteSurchargeMinor,
		CustomerRemotePostalCodes:           template.CustomerRemotePostalCodes,
		CustomerEtaMinDays:                  template.CustomerEtaMinDays, CustomerEtaMaxDays: template.CustomerEtaMaxDays,
		FreeShippingCountries: template.FreeShippingCountries,
		DisplayPriceSnapshots: template.DisplayPriceData,
		Description:           template.Description, Enabled: template.Enabled,
		Rules: make([]PublicShippingCustomerBillingRuleResponse, 0, len(template.Rules)),
	}
	for _, rule := range template.Rules {
		result.Rules = append(result.Rules, PublicShippingCustomerBillingRuleResponse{
			ID: rule.ID, TemplateID: rule.TemplateID, Region: rule.Region, Currency: rule.Currency,
			MinValueMinor: rule.MinValueMinor, MaxValueMinor: rule.MaxValueMinor,
			FeeMinor: rule.FeeMinor, AdditionalMinor: rule.AdditionalMinor,
			DisplayPriceSnapshots: rule.DisplayPriceData,
		})
	}
	if template.RouteGroup != nil {
		result.RouteGroup = &PublicShippingRouteAggregateIdentityResponse{
			ID: template.RouteGroup.ID, Name: template.RouteGroup.Name,
			Countries: template.RouteGroup.Countries, Enabled: template.RouteGroup.Enabled,
			CarrierServices: buildPublicShippingCarrierServiceIdentityResponses(template.RouteGroup.CarrierServices),
		}
	}
	return result
}

func buildPublicShippingCustomerBillingTemplateResponses(templates []shippingdomain.ShippingTemplate) []PublicShippingCustomerBillingTemplateResponse {
	results := make([]PublicShippingCustomerBillingTemplateResponse, 0, len(templates))
	for _, template := range templates {
		results = append(results, buildPublicShippingCustomerBillingTemplateResponse(template))
	}
	return results
}

func buildPublicShippingCarrierServiceIdentityResponse(route shippingdomain.CarrierService) PublicShippingCarrierServiceIdentityResponse {
	result := PublicShippingCarrierServiceIdentityResponse{
		ID: route.ID, CarrierID: route.CarrierID, ServiceCode: route.ServiceCode,
		ServiceName: route.ServiceName, RouteName: route.RouteName,
		Countries: route.Countries, Enabled: route.Enabled, SortOrder: route.SortOrder,
	}
	if route.Carrier != nil {
		result.Carrier = &PublicShippingCarrierIdentityResponse{
			ID: route.Carrier.ID, Name: route.Carrier.Name, Code: route.Carrier.Code,
			TrackingURL: route.Carrier.TrackingURL,
		}
	}
	return result
}

func buildPublicShippingCarrierServiceIdentityResponses(routes []shippingdomain.CarrierService) []PublicShippingCarrierServiceIdentityResponse {
	results := make([]PublicShippingCarrierServiceIdentityResponse, 0, len(routes))
	for _, route := range routes {
		results = append(results, buildPublicShippingCarrierServiceIdentityResponse(route))
	}
	return results
}
