package shipping

import (
	"time"

	"gorm.io/gorm"
)

// ShippingTemplate 运费模板
type ShippingTemplate struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	Type          string         `gorm:"not null" json:"type"` // weight, quantity, price
	FreeShipping  bool           `gorm:"default:false" json:"free_shipping"`
	FreeThreshold float64        `gorm:"default:0" json:"free_threshold"` // 免邮门槛
	DefaultFee    float64        `gorm:"default:0" json:"default_fee"`
	Description   string         `gorm:"type:text" json:"description"`
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	Rules         []ShippingRule `gorm:"foreignKey:TemplateID" json:"rules"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ShippingTemplate) TableName() string {
	return "shipping_templates"
}

// ShippingRule 运费规则
type ShippingRule struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	TemplateID uint      `gorm:"not null;index" json:"template_id"`
	Region     string    `json:"region"` // 地区代码，如 US, CN, EU
	MinValue   float64   `gorm:"default:0" json:"min_value"`
	MaxValue   float64   `gorm:"default:0" json:"max_value"`
	Fee        float64   `gorm:"not null" json:"fee"`
	Additional float64   `gorm:"default:0" json:"additional"` // 续重/续件费用
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (ShippingRule) TableName() string {
	return "shipping_rules"
}
