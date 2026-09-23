package shipping

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
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
	ID                 uint   `gorm:"primarykey" json:"id"`
	Name               string `gorm:"not null" json:"name"`
	Type               string `gorm:"not null" json:"type"` // weight, quantity, price
	Currency           string `gorm:"size:3;not null;default:'USD';index" json:"currency"`
	FreeShipping       bool   `gorm:"default:false" json:"free_shipping"`
	FreeThresholdMinor int64  `gorm:"column:free_threshold_minor;not null;default:0" json:"free_threshold_minor"`
	DefaultFeeMinor    int64  `gorm:"column:default_fee_minor;not null;default:0" json:"default_fee_minor"`
	// DisplayPriceData is hydrated from ShippingDisplayPriceSnapshot. It is a
	// storefront/read-model projection and is never persisted on the
	// transactional shipping_templates row.
	DisplayPriceData     datatypes.JSON                `gorm:"-" json:"display_price_snapshots,omitempty"`
	DisplayPriceSnapshot *ShippingDisplayPriceSnapshot `gorm:"-" json:"-"`
	Description          string                        `gorm:"type:text" json:"description"`
	Enabled              bool                          `gorm:"default:true" json:"enabled"`
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
	return t.validateMoneyFields()
}

func (t *ShippingTemplate) BeforeSave(tx *gorm.DB) error {
	return t.validateMoneyFields()
}

func (t *ShippingTemplate) validateMoneyFields() error {
	t.ensureSourceCurrency()
	if t.FreeThresholdMinor < 0 || t.DefaultFeeMinor < 0 {
		return errors.New("shipping template monetary fields cannot be negative")
	}
	return nil
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
