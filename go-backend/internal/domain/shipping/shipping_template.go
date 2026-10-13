package shipping

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	ShippingTemplateKindCarrier            = "carrier"
	ShippingTemplateKindSystemFreeShipping = "system_free_shipping"
	ShippingTemplateTypeSystemFreeShipping = "free_shipping"

	ShippingTemplateDisplayPriceFieldDefaultFee    = "default_fee"
	ShippingTemplateDisplayPriceFieldFreeThreshold = "free_threshold"

	ShippingRuleDisplayPriceFieldMinValue   = "min_value"
	ShippingRuleDisplayPriceFieldMaxValue   = "max_value"
	ShippingRuleDisplayPriceFieldFee        = "fee"
	ShippingRuleDisplayPriceFieldAdditional = "additional"
)

var ShippingTemplateDisplayPriceFields = []string{
	ShippingTemplateDisplayPriceFieldDefaultFee,
	ShippingTemplateDisplayPriceFieldFreeThreshold,
}

var ShippingRuleDisplayPriceFields = []string{
	ShippingRuleDisplayPriceFieldMinValue,
	ShippingRuleDisplayPriceFieldMaxValue,
	ShippingRuleDisplayPriceFieldFee,
	ShippingRuleDisplayPriceFieldAdditional,
}

// ShippingTemplate 运费模板
type ShippingTemplate struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Name string `gorm:"not null" json:"name"`
	Type string `gorm:"not null" json:"type"` // weight, quantity, price, free_shipping
	// TemplateKind distinguishes the system-owned country-scoped free-shipping
	// template from ordinary carrier pricing templates. It deliberately has no
	// relation to a carrier, 4PX channel, or Yanwen channel.
	TemplateKind    string `gorm:"column:template_kind;size:40;not null;default:'carrier'" json:"template_kind"`
	IsSystemManaged bool   `gorm:"column:is_system_managed;not null;default:false" json:"is_system_managed"`
	// RouteGroupID links this customer-facing billing template to a published
	// route aggregate. Route identity and provider metadata live on that
	// aggregate; all fields below describe the amount charged to the customer.
	RouteGroupID                        *uint  `gorm:"column:route_group_id;index" json:"route_group_id,omitempty"`
	Currency                            string `gorm:"size:3;not null;default:'USD';index" json:"currency"`
	FreeShipping                        bool   `gorm:"default:false" json:"free_shipping"`
	FreeThresholdMinor                  int64  `gorm:"column:free_threshold_minor;not null;default:0" json:"free_threshold_minor"`
	DefaultFeeMinor                     int64  `gorm:"column:default_fee_minor;not null;default:0" json:"default_fee_minor"`
	CustomerBillingMode                 string `gorm:"column:customer_billing_mode;size:40;not null;default:'actual_weight'" json:"customer_billing_mode"`
	CustomerFirstWeightGrams            int    `gorm:"column:customer_first_weight_grams;not null;default:0" json:"customer_first_weight_grams"`
	CustomerAdditionalWeightGrams       int    `gorm:"column:customer_additional_weight_grams;not null;default:0" json:"customer_additional_weight_grams"`
	CustomerMinChargeWeightGrams        int    `gorm:"column:customer_min_charge_weight_grams;not null;default:0" json:"customer_min_charge_weight_grams"`
	CustomerVolumetricDivisor           int    `gorm:"column:customer_volumetric_divisor;not null;default:6000" json:"customer_volumetric_divisor"`
	CustomerFuelSurchargePercentDecimal string `gorm:"column:customer_fuel_surcharge_percent_decimal;type:numeric(30,15);not null;default:0" json:"customer_fuel_surcharge_percent_decimal"`
	CustomerRemoteSurchargeMinor        int64  `gorm:"column:customer_remote_surcharge_minor;not null;default:0" json:"customer_remote_surcharge_minor"`
	CustomerRemotePostalCodes           string `gorm:"column:customer_remote_postal_codes;type:text;not null;default:'[]'" json:"customer_remote_postal_codes"`
	CustomerEtaMinDays                  int    `gorm:"column:customer_eta_min_days;not null;default:0" json:"customer_eta_min_days"`
	CustomerEtaMaxDays                  int    `gorm:"column:customer_eta_max_days;not null;default:0" json:"customer_eta_max_days"`
	// FreeShippingCountries is a normalized JSON array of ISO alpha-2 country
	// codes. It is used only by the system free-shipping template.
	FreeShippingCountries string `gorm:"column:free_shipping_countries;type:text;not null;default:'[]'" json:"free_shipping_countries"`
	// DisplayPriceData is hydrated from ShippingDisplayPriceSnapshot. It is a
	// storefront/read-model projection and is never persisted on the
	// transactional shipping_templates row.
	DisplayPriceData     datatypes.JSON                `gorm:"-" json:"display_price_snapshots,omitempty"`
	DisplayPriceSnapshot *ShippingDisplayPriceSnapshot `gorm:"-" json:"-"`
	Description          string                        `gorm:"type:text" json:"description"`
	Enabled              bool                          `gorm:"default:true" json:"enabled"`
	RouteGroup           *ShippingRouteGroup           `gorm:"foreignKey:RouteGroupID" json:"route_group,omitempty"`
	Rules                []ShippingRule                `gorm:"foreignKey:TemplateID" json:"rules"`
	CreatedAt            time.Time                     `json:"created_at"`
	UpdatedAt            time.Time                     `json:"updated_at"`
	DeletedAt            gorm.DeletedAt                `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ShippingTemplate) TableName() string {
	return "shipping_templates"
}

func (t *ShippingTemplate) BeforeCreate(tx *gorm.DB) error {
	return t.normalizeAndValidate()
}

func (t *ShippingTemplate) BeforeSave(tx *gorm.DB) error {
	return t.normalizeAndValidate()
}

func (t *ShippingTemplate) ValidateCustomerShippingBillingConfiguration() error {
	return t.normalizeAndValidate()
}

func (t *ShippingTemplate) normalizeAndValidate() error {
	t.NormalizeSystemFreeShippingTemplateIdentity()
	t.ApplySystemFreeShippingTemplatePricingDefaults()
	t.FreeShippingCountries = NormalizeShippingCountryCodes(t.FreeShippingCountries)
	if err := ValidateShippingCountryCodes(t.FreeShippingCountries); err != nil {
		return err
	}
	t.ensureSourceCurrency()
	if t.FreeThresholdMinor < 0 || t.DefaultFeeMinor < 0 {
		return errors.New("shipping template monetary fields cannot be negative")
	}
	if t.CustomerBillingMode == "" {
		t.CustomerBillingMode = "actual_weight"
	}
	if t.CustomerVolumetricDivisor == 0 {
		t.CustomerVolumetricDivisor = 6000
	}
	if t.CustomerFirstWeightGrams < 0 || t.CustomerAdditionalWeightGrams < 0 || t.CustomerMinChargeWeightGrams < 0 || t.CustomerVolumetricDivisor < 0 || t.CustomerRemoteSurchargeMinor < 0 || t.CustomerEtaMinDays < 0 || t.CustomerEtaMaxDays < 0 {
		return errors.New("customer shipping billing fields cannot be negative")
	}
	switch t.CustomerBillingMode {
	case "actual_weight", "volumetric_weight", "greater_of_actual_and_volumetric":
	default:
		return errors.New("customer billing mode must be actual_weight, volumetric_weight or greater_of_actual_and_volumetric")
	}
	var postalRules []string
	if strings.TrimSpace(t.CustomerRemotePostalCodes) == "" {
		t.CustomerRemotePostalCodes = "[]"
	}
	if err := json.Unmarshal([]byte(t.CustomerRemotePostalCodes), &postalRules); err != nil {
		return errors.New("customer remote postal codes must be a JSON array of strings")
	}
	for _, rule := range postalRules {
		if strings.TrimSpace(rule) == "" {
			return errors.New("customer remote postal code rule cannot be empty")
		}
	}
	if t.CustomerEtaMaxDays > 0 && t.CustomerEtaMinDays > 0 && t.CustomerEtaMaxDays < t.CustomerEtaMinDays {
		return errors.New("customer eta max days cannot be less than min days")
	}
	if _, err := t.CustomerFuelSurchargeRate(); err != nil {
		return err
	}
	return nil
}

// CustomerFuelSurchargeRate parses the exact surcharge configured on the
// customer billing template. It is deliberately separate from
// CarrierService.FuelSurchargeRate, which belongs to provider-side data.
func (t ShippingTemplate) CustomerFuelSurchargeRate() (*big.Rat, error) {
	value := strings.TrimSpace(t.CustomerFuelSurchargePercentDecimal)
	if value == "" {
		value = "0"
	}
	if strings.Contains(value, "/") {
		return nil, fmt.Errorf("customer fuel surcharge percent must be a decimal")
	}
	rate, ok := new(big.Rat).SetString(value)
	if !ok || rate.Sign() < 0 || rate.Cmp(big.NewRat(100, 1)) > 0 {
		return nil, fmt.Errorf("customer fuel surcharge percent must be between 0 and 100")
	}
	scaled := new(big.Rat).Mul(rate, new(big.Rat).SetInt64(1_000_000_000_000_000))
	if scaled.Denom().Cmp(big.NewInt(1)) != 0 {
		return nil, errors.New("customer fuel surcharge percent supports at most 15 fractional digits")
	}
	return rate, nil
}

// CustomerRemoteSurchargeMoney returns the customer-facing remote-area fee.
func (t ShippingTemplate) CustomerRemoteSurchargeMoney() (domainmoney.Money, error) {
	return domainmoney.New(t.CustomerRemoteSurchargeMinor, currency.NormalizeCode(t.Currency))
}

// NormalizeSystemFreeShippingTemplateIdentity keeps the built-in template's
// discriminator fields consistent. Ownership is represented by
// IsSystemManaged, while TemplateKind and Type describe this specific
// system-free-shipping behavior. A system-managed template of another kind
// must remain usable as its own template type.
func (t *ShippingTemplate) NormalizeSystemFreeShippingTemplateIdentity() {
	if t == nil {
		return
	}

	t.TemplateKind = NormalizeShippingTemplateKind(t.TemplateKind)
	if t.TemplateKind == ShippingTemplateKindSystemFreeShipping ||
		strings.EqualFold(strings.TrimSpace(t.Type), ShippingTemplateTypeSystemFreeShipping) {
		t.TemplateKind = ShippingTemplateKindSystemFreeShipping
		t.Type = ShippingTemplateTypeSystemFreeShipping
		t.IsSystemManaged = true
	}
}

// ApplySystemFreeShippingTemplatePricingDefaults enforces the pricing
// contract for the system free-shipping template. Its fee is always zero and
// it never carries dimensional pricing rules; country scope is the only
// condition evaluated by the quote service.
func (t *ShippingTemplate) ApplySystemFreeShippingTemplatePricingDefaults() {
	if t == nil || !t.IsSystemFreeShippingTemplate() {
		return
	}
	t.FreeShipping = true
	t.FreeThresholdMinor = 0
	t.DefaultFeeMinor = 0
	t.Rules = nil
}

// NormalizeShippingTemplateKind keeps old rows compatible with the generic
// carrier template kind while allowing the system free-shipping row to be
// identified without inspecting any carrier records.
func NormalizeShippingTemplateKind(value string) string {
	kind := strings.ToLower(strings.TrimSpace(value))
	if kind == ShippingTemplateKindSystemFreeShipping {
		return ShippingTemplateKindSystemFreeShipping
	}
	return ShippingTemplateKindCarrier
}

// NormalizeShippingCountryCodes accepts the JSON and delimited formats used
// by existing shipping country fields and returns a stable JSON array.
func NormalizeShippingCountryCodes(value string) string {
	codes := parseShippingCountryCodes(value)
	if len(codes) == 0 {
		return "[]"
	}
	return marshalShippingCountryCodes(codes)
}

// ValidateShippingCountryCodes rejects malformed scope entries while keeping
// the explicit wildcard values supported by the shipping contract.
func ValidateShippingCountryCodes(value string) error {
	for _, code := range parseShippingCountryCodes(value) {
		switch code {
		case "*", "ALL", "GLOBAL", "WORLDWIDE":
			continue
		}
		if len(code) != 2 || code[0] < 'A' || code[0] > 'Z' || code[1] < 'A' || code[1] > 'Z' {
			return errors.New("shipping country scope must contain ISO alpha-2 country codes")
		}
	}
	return nil
}

// ShippingCountryCodes returns the normalized country scope of a template.
func (t ShippingTemplate) ShippingCountryCodes() []string {
	return parseShippingCountryCodes(t.FreeShippingCountries)
}

// IsSystemFreeShippingTemplate identifies the built-in country-scoped free
// shipping template independently of any carrier service collection.
func (t ShippingTemplate) IsSystemFreeShippingTemplate() bool {
	return NormalizeShippingTemplateKind(t.TemplateKind) == ShippingTemplateKindSystemFreeShipping ||
		strings.EqualFold(strings.TrimSpace(t.Type), ShippingTemplateTypeSystemFreeShipping)
}

// AllowsShippingCountry returns whether the template can be used for a
// destination. System free-shipping templates intentionally deny every
// country until an administrator explicitly selects at least one country.
func (t ShippingTemplate) AllowsShippingCountry(country string) bool {
	if !t.IsSystemFreeShippingTemplate() {
		return true
	}
	normalizedCountry := strings.ToUpper(strings.TrimSpace(country))
	if normalizedCountry == "" {
		return false
	}
	for _, code := range t.ShippingCountryCodes() {
		if code == "*" || code == "ALL" || code == "GLOBAL" || code == "WORLDWIDE" || code == normalizedCountry {
			return true
		}
	}
	return false
}

func (t *ShippingTemplate) ensureSourceCurrency() {
	if !currency.IsCatalogCode(t.Currency) {
		t.Currency = currency.DefaultPrimaryCurrency
	}
}

// FreeThresholdMoney and DefaultFeeMoney expose exact monetary values to the
// pricing engine.
func (t ShippingTemplate) FreeThresholdMoney() (domainmoney.Money, error) {
	return domainmoney.New(t.FreeThresholdMinor, t.Currency)
}

func (t ShippingTemplate) DefaultFeeMoney() (domainmoney.Money, error) {
	return domainmoney.New(t.DefaultFeeMinor, t.Currency)
}

// ShippingRule 运费规则
type ShippingRule struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	TemplateID uint   `gorm:"not null;index" json:"template_id"`
	Region     string `json:"region"` // 地区代码，如 US, CN, EU
	Currency   string `gorm:"size:3;not null;default:'USD';index" json:"currency"`
	// For weight/quantity rules these are dimensional thresholds, not money,
	// and remain ordinary scalar fields. Price rules use the minor fields.
	MinValue        float64 `gorm:"default:0" json:"-"`
	MaxValue        float64 `gorm:"default:0" json:"-"`
	MinValueMinor   int64   `gorm:"column:min_value_minor;not null;default:0" json:"min_value_minor"`
	MaxValueMinor   int64   `gorm:"column:max_value_minor;not null;default:0" json:"max_value_minor"`
	FeeMinor        int64   `gorm:"column:fee_minor;not null;default:0" json:"fee_minor"`
	AdditionalMinor int64   `gorm:"column:additional_minor;not null;default:0" json:"additional_minor"`
	// DisplayPriceData is hydrated from ShippingDisplayPriceSnapshot. It is a
	// storefront/read-model projection and is never persisted on the
	// transactional shipping_rules row.
	DisplayPriceData     datatypes.JSON                `gorm:"-" json:"display_price_snapshots,omitempty"`
	DisplayPriceSnapshot *ShippingDisplayPriceSnapshot `gorm:"-" json:"-"`
	CreatedAt            time.Time                     `json:"created_at"`
}

// TableName 指定表名
func (ShippingRule) TableName() string {
	return "shipping_rules"
}

func (r *ShippingRule) BeforeCreate(tx *gorm.DB) error {
	return r.validateMoneyFields()
}

func (r *ShippingRule) BeforeSave(tx *gorm.DB) error {
	return r.validateMoneyFields()
}

func (r *ShippingRule) validateMoneyFields() error {
	r.ensureSourceCurrency()
	if r.MinValueMinor < 0 || r.MaxValueMinor < 0 || r.FeeMinor < 0 || r.AdditionalMinor < 0 {
		return errors.New("shipping rule monetary fields cannot be negative")
	}
	return nil
}

func (r *ShippingRule) ensureSourceCurrency() {
	if !currency.IsCatalogCode(r.Currency) {
		r.Currency = currency.DefaultPrimaryCurrency
	}
}

func (r ShippingRule) MinValueMoney(code string) (domainmoney.Money, error) {
	return domainmoney.New(r.MinValueMinor, code)
}

func (r ShippingRule) MaxValueMoney(code string) (domainmoney.Money, error) {
	return domainmoney.New(r.MaxValueMinor, code)
}

func (r ShippingRule) FeeMoney(code string) (domainmoney.Money, error) {
	return domainmoney.New(r.FeeMinor, code)
}

func (r ShippingRule) AdditionalMoney(code string) (domainmoney.Money, error) {
	return domainmoney.New(r.AdditionalMinor, code)
}

func TemplateDisplayPriceSnapshotsJSON(values map[string][]currency.DisplayPriceSnapshot, baseCurrency ...string) datatypes.JSON {
	base := ""
	if len(baseCurrency) > 0 {
		base = baseCurrency[0]
	}
	return currency.DisplayPriceSnapshotMapJSON(values, base, ShippingTemplateDisplayPriceFields...)
}

func RuleDisplayPriceSnapshotsJSON(values map[string][]currency.DisplayPriceSnapshot, baseCurrency ...string) datatypes.JSON {
	base := ""
	if len(baseCurrency) > 0 {
		base = baseCurrency[0]
	}
	return currency.DisplayPriceSnapshotMapJSON(values, base, ShippingRuleDisplayPriceFields...)
}

// ShippingDisplayPriceSnapshot is the storefront display-price read model for
// a template or rule. Source minor-unit fields form the freshness fingerprint:
// a snapshot is only exposed when it still describes the current transaction
// source row. The table is deliberately separate from shipping_templates and
// shipping_rules so exchange-rate refreshes never update transactional rows.
type ShippingDisplayPriceSnapshot struct {
	ID                       uint           `gorm:"primaryKey" json:"id"`
	ScopeKey                 string         `gorm:"size:64;uniqueIndex;not null" json:"-"`
	TemplateID               uint           `gorm:"not null;index" json:"template_id"`
	RuleID                   *uint          `gorm:"index" json:"rule_id,omitempty"`
	SourceCurrency           string         `gorm:"size:3;not null" json:"source_currency"`
	SourceDefaultFeeMinor    *int64         `gorm:"column:source_default_fee_minor" json:"source_default_fee_minor,omitempty"`
	SourceFreeThresholdMinor *int64         `gorm:"column:source_free_threshold_minor" json:"source_free_threshold_minor,omitempty"`
	SourceMinValueMinor      *int64         `gorm:"column:source_min_value_minor" json:"source_min_value_minor,omitempty"`
	SourceMaxValueMinor      *int64         `gorm:"column:source_max_value_minor" json:"source_max_value_minor,omitempty"`
	SourceFeeMinor           *int64         `gorm:"column:source_fee_minor" json:"source_fee_minor,omitempty"`
	SourceAdditionalMinor    *int64         `gorm:"column:source_additional_minor" json:"source_additional_minor,omitempty"`
	DisplayPriceData         datatypes.JSON `gorm:"column:display_prices;type:json;not null;default:'{}'" json:"display_prices"`
	CreatedAt                time.Time      `json:"-"`
	UpdatedAt                time.Time      `json:"-"`
}

func (ShippingDisplayPriceSnapshot) TableName() string {
	return "shipping_display_price_snapshots"
}
