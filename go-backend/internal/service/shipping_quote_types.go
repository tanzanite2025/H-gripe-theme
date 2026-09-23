package service

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/shipping"
	"time"

	"gorm.io/datatypes"
)

type ShippingCalculationInput struct {
	TemplateID  uint
	Weight      float64
	Quantity    int
	AmountMinor int64
	Country     string
}

type ShippingQuoteItemInput struct {
	ProductID                      uint           `json:"product_id"`
	VariantID                      *uint          `json:"variant_id,omitempty"`
	ProductSpecificationTemplateID *uint          `json:"product_specification_template_id,omitempty"`
	ShippingTemplateID             *uint          `json:"shipping_template_id,omitempty"`
	Quantity                       int            `json:"quantity"`
	UnitPriceMinor                 int64          `json:"unit_price_minor"`
	WeightGrams                    int            `json:"weight_grams"`
	PackagingWeightDeltaGrams      int            `json:"packaging_weight_delta_grams,omitempty"`
	ConfigurationData              datatypes.JSON `json:"configuration,omitempty"`
}

type ShippingQuoteInput struct {
	Country             string                   `json:"country"`
	PostalCode          string                   `json:"postal_code,omitempty"`
	Currency            string                   `json:"currency,omitempty"`
	DisplayCurrency     string                   `json:"display_currency,omitempty"`
	ShippingQuoteID     string                   `json:"shipping_quote_id,omitempty"`
	SelectedQuotePlanID string                   `json:"selected_quote_plan_id,omitempty"`
	Items               []ShippingQuoteItemInput `json:"items"`
}

type ShippingQuoteItem struct {
	ProductID                      uint   `json:"product_id"`
	VariantID                      *uint  `json:"variant_id,omitempty"`
	ProductSpecificationTemplateID *uint  `json:"product_specification_template_id,omitempty"`
	TemplateID                     uint   `json:"template_id"`
	TemplateName                   string `json:"template_name"`
	PackagingRuleID                *uint  `json:"packaging_rule_id,omitempty"`
	PackagingRuleName              string `json:"packaging_rule_name,omitempty"`
	Quantity                       int    `json:"quantity"`
	UnitPriceDecimal               string `json:"-"` // display-only projection
	UnitPriceMinor                 int64  `json:"unit_price_minor"`
	AmountDecimal                  string `json:"-"` // display-only projection
	AmountMinor                    int64  `json:"amount_minor"`
	WeightGrams                    int    `json:"weight_grams"`
	PackagingWeightGrams           int    `json:"packaging_weight_grams"`
	ChargeWeightGrams              int    `json:"charge_weight_grams"`
	ShippingFeeDecimal             string `json:"-"` // display-only projection
	ShippingFeeMinor               int64  `json:"shipping_fee_minor"`
	FreeShipping                   bool   `json:"free_shipping"`
}

type ShippingQuote struct {
	ID                 string                          `json:"id"`
	RateVersion        string                          `json:"rate_version"`
	ExpiresAt          time.Time                       `json:"expires_at"`
	ShippingFeeDecimal string                          `json:"-"` // display-only projection
	ShippingFeeMinor   int64                           `json:"shipping_fee_minor"`
	FreeShipping       bool                            `json:"free_shipping"`
	Currency           string                          `json:"currency,omitempty"`
	DisplayPrice       *currency.DisplayPriceSnapshot  `json:"display_price,omitempty"`
	DisplayPrices      []currency.DisplayPriceSnapshot `json:"display_prices,omitempty"`
	DisplayCurrency    string                          `json:"display_currency,omitempty"`
	Items              []ShippingQuoteItem             `json:"items,omitempty"`
	Plans              []ShippingQuotePlan             `json:"plans"`
	SelectedPlan       *ShippingQuotePlan              `json:"selected_plan"`
}

type ShippingQuotePlan struct {
	ID                 string                          `json:"id"`
	Currency           string                          `json:"currency"`
	ShippingFeeDecimal string                          `json:"-"` // display-only projection
	ShippingFeeMinor   int64                           `json:"shipping_fee_minor"`
	DisplayPrice       *currency.DisplayPriceSnapshot  `json:"display_price,omitempty"`
	DisplayPrices      []currency.DisplayPriceSnapshot `json:"display_prices,omitempty"`
	FreeShipping       bool                            `json:"free_shipping"`
	EtaMinDays         int                             `json:"eta_min_days"`
	EtaMaxDays         int                             `json:"eta_max_days"`
	Legs               []ShippingQuoteLeg              `json:"legs"`
}

type ShippingQuoteLeg struct {
	GroupKey               string                          `json:"group_key"`
	ItemIndexes            []int                           `json:"item_indexes"`
	AllocationBasis        string                          `json:"allocation_basis"`
	CarrierID              uint                            `json:"carrier_id"`
	CarrierName            string                          `json:"carrier_name"`
	CarrierCode            string                          `json:"carrier_code"`
	CarrierServiceID       uint                            `json:"carrier_service_id"`
	ServiceCode            string                          `json:"service_code"`
	ServiceName            string                          `json:"service_name"`
	RouteName              string                          `json:"route_name,omitempty"`
	TemplateID             uint                            `json:"template_id"`
	TemplateName           string                          `json:"template_name"`
	Currency               string                          `json:"currency,omitempty"`
	BillingMode            string                          `json:"billing_mode"`
	ActualWeightGrams      int                             `json:"actual_weight_grams"`
	VolumetricWeightGrams  int                             `json:"volumetric_weight_grams"`
	ChargeWeightGrams      int                             `json:"charge_weight_grams"`
	BillableWeightGrams    int                             `json:"billable_weight_grams"`
	BaseFeeDecimal         string                          `json:"-"` // display-only projection
	FuelSurchargeDecimal   string                          `json:"-"` // display-only projection
	RemoteSurchargeDecimal string                          `json:"-"` // display-only projection
	ShippingFeeDecimal     string                          `json:"-"` // display-only projection
	BaseFeeMinor           int64                           `json:"base_fee_minor"`
	FuelSurchargeMinor     int64                           `json:"fuel_surcharge_minor"`
	RemoteSurchargeMinor   int64                           `json:"remote_surcharge_minor"`
	ShippingFeeMinor       int64                           `json:"shipping_fee_minor"`
	DisplayPrice           *currency.DisplayPriceSnapshot  `json:"display_price,omitempty"`
	DisplayPrices          []currency.DisplayPriceSnapshot `json:"display_prices,omitempty"`
	FreeShipping           bool                            `json:"free_shipping"`
	EtaMinDays             int                             `json:"eta_min_days"`
	EtaMaxDays             int                             `json:"eta_max_days"`
	SortOrder              int                             `json:"sort_order"`
}

type resolvedShippingItem struct {
	ShippingQuoteItemInput
	Amount               domainmoney.Money
	Template             *shipping.ShippingTemplate
	PackagingRule        *shipping.PackagingRule
	PackagingWeightGrams int
	ChargeWeightGrams    int
}

type shippingQuoteGroup struct {
	Template         *shipping.ShippingTemplate
	ItemIndexes      []int
	Amount           domainmoney.Money
	Quantity         int
	TotalWeightGrams int
}
