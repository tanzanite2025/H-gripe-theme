package admin

import (
	"strings"
	shippingdomain "tanzanite/internal/domain/shipping"
	"time"
)

type shippingTemplateRequest struct {
	Name          string                `json:"name" binding:"required"`
	Type          string                `json:"type" binding:"required"`
	FreeShipping  bool                  `json:"free_shipping"`
	FreeThreshold float64               `json:"free_threshold"`
	DefaultFee    float64               `json:"default_fee"`
	Description   string                `json:"description"`
	Enabled       *bool                 `json:"enabled"`
	Rules         []shippingRuleRequest `json:"rules"`
}

type shippingRuleRequest struct {
	ID         uint    `json:"id"`
	Region     string  `json:"region"`
	MinValue   float64 `json:"min_value"`
	MaxValue   float64 `json:"max_value"`
	Fee        float64 `json:"fee"`
	Additional float64 `json:"additional"`
}

type shippingZoneRequest struct {
	Name        string `json:"name" binding:"required"`
	Countries   string `json:"countries"`
	States      string `json:"states"`
	PostalCodes string `json:"postal_codes"`
	Enabled     *bool  `json:"enabled"`
}

type shippingTemplateBindingRequest struct {
	TemplateID    uint   `json:"template_id" binding:"required"`
	Scope         string `json:"scope" binding:"required"`
	ProductTypeID *uint  `json:"product_type_id"`
	ProductID     *uint  `json:"product_id"`
	VariantID     *uint  `json:"variant_id"`
	Priority      int    `json:"priority"`
	Enabled       *bool  `json:"enabled"`
}

type shippingCarrierServiceRequest struct {
	CarrierID             uint    `json:"carrier_id" binding:"required"`
	TemplateID            *uint   `json:"template_id"`
	ServiceCode           string  `json:"service_code" binding:"required"`
	ServiceName           string  `json:"service_name" binding:"required"`
	RouteName             string  `json:"route_name"`
	Countries             string  `json:"countries"`
	Currency              string  `json:"currency"`
	BillingMode           string  `json:"billing_mode"`
	FirstWeightGrams      int     `json:"first_weight_grams"`
	AdditionalWeightGrams int     `json:"additional_weight_grams"`
	MinChargeWeightGrams  int     `json:"min_charge_weight_grams"`
	VolumetricDivisor     int     `json:"volumetric_divisor"`
	FuelSurchargePercent  float64 `json:"fuel_surcharge_percent"`
	RemoteSurcharge       float64 `json:"remote_surcharge"`
	EtaMinDays            int     `json:"eta_min_days"`
	EtaMaxDays            int     `json:"eta_max_days"`
	Enabled               *bool   `json:"enabled"`
	SortOrder             int     `json:"sort_order"`
	Description           string  `json:"description"`
}

type shippingTrackingProviderRequest struct {
	ProviderCode           string `json:"provider_code" binding:"required"`
	ProviderName           string `json:"provider_name" binding:"required"`
	Environment            string `json:"environment"`
	BaseURL                string `json:"base_url"`
	APIKey                 string `json:"api_key"`
	WebhookSecret          string `json:"webhook_secret"`
	WebhookEnabled         *bool  `json:"webhook_enabled"`
	AutoRegister           *bool  `json:"auto_register"`
	PollingEnabled         *bool  `json:"polling_enabled"`
	PollingIntervalMinutes int    `json:"polling_interval_minutes"`
	RequestTimeoutSeconds  int    `json:"request_timeout_seconds"`
	Enabled                *bool  `json:"enabled"`
	SortOrder              int    `json:"sort_order"`
	Description            string `json:"description"`
}

type shippingTrackingCarrierMappingRequest struct {
	ProviderID          uint   `json:"provider_id" binding:"required"`
	Scope               string `json:"scope" binding:"required"`
	CarrierID           *uint  `json:"carrier_id"`
	CarrierServiceID    *uint  `json:"carrier_service_id"`
	ProviderCarrierCode string `json:"provider_carrier_code" binding:"required"`
	ProviderCarrierName string `json:"provider_carrier_name"`
	Enabled             *bool  `json:"enabled"`
	Priority            int    `json:"priority"`
	Description         string `json:"description"`
}

type shippingTrackingProviderResponse struct {
	ID                      uint      `json:"id"`
	ProviderCode            string    `json:"provider_code"`
	ProviderName            string    `json:"provider_name"`
	Environment             string    `json:"environment"`
	BaseURL                 string    `json:"base_url"`
	APIKeyConfigured        bool      `json:"api_key_configured"`
	WebhookSecretConfigured bool      `json:"webhook_secret_configured"`
	WebhookEnabled          bool      `json:"webhook_enabled"`
	AutoRegister            bool      `json:"auto_register"`
	PollingEnabled          bool      `json:"polling_enabled"`
	PollingIntervalMinutes  int       `json:"polling_interval_minutes"`
	RequestTimeoutSeconds   int       `json:"request_timeout_seconds"`
	Enabled                 bool      `json:"enabled"`
	SortOrder               int       `json:"sort_order"`
	Description             string    `json:"description"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func (r shippingTemplateRequest) toDomain() shippingdomain.ShippingTemplate {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	template := shippingdomain.ShippingTemplate{
		Name:          strings.TrimSpace(r.Name),
		Type:          strings.TrimSpace(r.Type),
		FreeShipping:  r.FreeShipping,
		FreeThreshold: r.FreeThreshold,
		DefaultFee:    r.DefaultFee,
		Description:   strings.TrimSpace(r.Description),
		Enabled:       enabled,
	}

	for _, rule := range r.Rules {
		template.Rules = append(template.Rules, rule.toDomain())
	}

	return template
}

func (r shippingRuleRequest) toDomain() shippingdomain.ShippingRule {
	return shippingdomain.ShippingRule{
		ID:         r.ID,
		Region:     strings.ToUpper(strings.TrimSpace(r.Region)),
		MinValue:   r.MinValue,
		MaxValue:   r.MaxValue,
		Fee:        r.Fee,
		Additional: r.Additional,
	}
}

func (r shippingCarrierServiceRequest) toDomain() shippingdomain.CarrierService {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	currency := strings.ToUpper(strings.TrimSpace(r.Currency))
	if currency == "" {
		currency = "USD"
	}

	billingMode := strings.ToLower(strings.TrimSpace(r.BillingMode))
	if billingMode == "" {
		billingMode = "actual_weight"
	}

	volumetricDivisor := r.VolumetricDivisor
	if volumetricDivisor <= 0 {
		volumetricDivisor = 6000
	}

	return shippingdomain.CarrierService{
		CarrierID:             r.CarrierID,
		TemplateID:            r.TemplateID,
		ServiceCode:           strings.ToUpper(strings.TrimSpace(r.ServiceCode)),
		ServiceName:           strings.TrimSpace(r.ServiceName),
		RouteName:             strings.TrimSpace(r.RouteName),
		Countries:             strings.TrimSpace(r.Countries),
		Currency:              currency,
		BillingMode:           billingMode,
		FirstWeightGrams:      r.FirstWeightGrams,
		AdditionalWeightGrams: r.AdditionalWeightGrams,
		MinChargeWeightGrams:  r.MinChargeWeightGrams,
		VolumetricDivisor:     volumetricDivisor,
		FuelSurchargePercent:  r.FuelSurchargePercent,
		RemoteSurcharge:       r.RemoteSurcharge,
		EtaMinDays:            r.EtaMinDays,
		EtaMaxDays:            r.EtaMaxDays,
		Enabled:               enabled,
		SortOrder:             r.SortOrder,
		Description:           strings.TrimSpace(r.Description),
	}
}

func (r shippingTrackingProviderRequest) toDomain() shippingdomain.TrackingProviderConfig {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	webhookEnabled := false
	if r.WebhookEnabled != nil {
		webhookEnabled = *r.WebhookEnabled
	}

	autoRegister := false
	if r.AutoRegister != nil {
		autoRegister = *r.AutoRegister
	}

	pollingEnabled := false
	if r.PollingEnabled != nil {
		pollingEnabled = *r.PollingEnabled
	}

	environment := strings.ToLower(strings.TrimSpace(r.Environment))
	if environment == "" {
		environment = "production"
	}

	pollingInterval := r.PollingIntervalMinutes
	if pollingInterval <= 0 {
		pollingInterval = 60
	}

	requestTimeout := r.RequestTimeoutSeconds
	if requestTimeout <= 0 {
		requestTimeout = 15
	}

	return shippingdomain.TrackingProviderConfig{
		ProviderCode:           strings.ToUpper(strings.TrimSpace(r.ProviderCode)),
		ProviderName:           strings.TrimSpace(r.ProviderName),
		Environment:            environment,
		BaseURL:                strings.TrimSpace(r.BaseURL),
		APIKey:                 strings.TrimSpace(r.APIKey),
		WebhookSecret:          strings.TrimSpace(r.WebhookSecret),
		WebhookEnabled:         webhookEnabled,
		AutoRegister:           autoRegister,
		PollingEnabled:         pollingEnabled,
		PollingIntervalMinutes: pollingInterval,
		RequestTimeoutSeconds:  requestTimeout,
		Enabled:                enabled,
		SortOrder:              r.SortOrder,
		Description:            strings.TrimSpace(r.Description),
	}
}

func (r shippingTrackingCarrierMappingRequest) toDomain() shippingdomain.TrackingCarrierMapping {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	scope := strings.ToLower(strings.TrimSpace(r.Scope))
	if scope == "" {
		scope = "carrier"
	}

	return shippingdomain.TrackingCarrierMapping{
		ProviderID:          r.ProviderID,
		Scope:               scope,
		CarrierID:           r.CarrierID,
		CarrierServiceID:    r.CarrierServiceID,
		ProviderCarrierCode: strings.TrimSpace(r.ProviderCarrierCode),
		ProviderCarrierName: strings.TrimSpace(r.ProviderCarrierName),
		Enabled:             enabled,
		Priority:            r.Priority,
		Description:         strings.TrimSpace(r.Description),
	}
}

func (r shippingZoneRequest) toDomain() shippingdomain.ShippingZone {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	return shippingdomain.ShippingZone{
		Name:        strings.TrimSpace(r.Name),
		Countries:   strings.TrimSpace(r.Countries),
		States:      strings.TrimSpace(r.States),
		PostalCodes: strings.TrimSpace(r.PostalCodes),
		Enabled:     enabled,
	}
}

func (r shippingTemplateBindingRequest) toDomain() shippingdomain.ShippingTemplateBinding {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	return shippingdomain.ShippingTemplateBinding{
		TemplateID:    r.TemplateID,
		Scope:         strings.ToLower(strings.TrimSpace(r.Scope)),
		ProductTypeID: r.ProductTypeID,
		ProductID:     r.ProductID,
		VariantID:     r.VariantID,
		Priority:      r.Priority,
		Enabled:       enabled,
	}
}
