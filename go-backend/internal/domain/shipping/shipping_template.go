package shipping

import (
	"commerce-platform/internal/domain/currency"
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
	ID               uint           `gorm:"primarykey" json:"id"`
	Name             string         `gorm:"not null" json:"name"`
	Type             string         `gorm:"not null" json:"type"` // weight, quantity, price
	FreeShipping     bool           `gorm:"default:false" json:"free_shipping"`
	FreeThreshold    float64        `gorm:"default:0" json:"free_threshold"` // 免邮门槛
	DefaultFee       float64        `gorm:"default:0" json:"default_fee"`
	DisplayPriceData datatypes.JSON `gorm:"column:display_price_snapshots;type:json;not null;default:'{}'" json:"display_price_snapshots,omitempty"`
	Description      string         `gorm:"type:text" json:"description"`
	Enabled          bool           `gorm:"default:true" json:"enabled"`
	Rules            []ShippingRule `gorm:"foreignKey:TemplateID" json:"rules"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ShippingTemplate) TableName() string {
	return "shipping_templates"
}

func (t *ShippingTemplate) BeforeCreate(tx *gorm.DB) error {
	t.ensureDisplayPriceData()
	return nil
}

func (t *ShippingTemplate) BeforeSave(tx *gorm.DB) error {
	t.ensureDisplayPriceData()
	return nil
}

func (t *ShippingTemplate) ensureDisplayPriceData() {
	if len(t.DisplayPriceData) == 0 {
		t.DisplayPriceData = datatypes.JSON([]byte("{}"))
	}
}

// ShippingRule 运费规则
type ShippingRule struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	TemplateID       uint           `gorm:"not null;index" json:"template_id"`
	Region           string         `json:"region"` // 地区代码，如 US, CN, EU
	MinValue         float64        `gorm:"default:0" json:"min_value"`
	MaxValue         float64        `gorm:"default:0" json:"max_value"`
	Fee              float64        `gorm:"not null" json:"fee"`
	Additional       float64        `gorm:"default:0" json:"additional"` // 续重/续件费用
	DisplayPriceData datatypes.JSON `gorm:"column:display_price_snapshots;type:json;not null;default:'{}'" json:"display_price_snapshots,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}

// TableName 指定表名
func (ShippingRule) TableName() string {
	return "shipping_rules"
}

func (r *ShippingRule) BeforeCreate(tx *gorm.DB) error {
	r.ensureDisplayPriceData()
	return nil
}

func (r *ShippingRule) BeforeSave(tx *gorm.DB) error {
	r.ensureDisplayPriceData()
	return nil
}

func (r *ShippingRule) ensureDisplayPriceData() {
	if len(r.DisplayPriceData) == 0 {
		r.DisplayPriceData = datatypes.JSON([]byte("{}"))
	}
}

func TemplateDisplayPriceSnapshotsJSON(values map[string][]currency.DisplayPriceSnapshot) datatypes.JSON {
	return currency.DisplayPriceSnapshotMapJSON(values, "", ShippingTemplateDisplayPriceFields...)
}

func RuleDisplayPriceSnapshotsJSON(values map[string][]currency.DisplayPriceSnapshot) datatypes.JSON {
	return currency.DisplayPriceSnapshotMapJSON(values, "", ShippingRuleDisplayPriceFields...)
}
