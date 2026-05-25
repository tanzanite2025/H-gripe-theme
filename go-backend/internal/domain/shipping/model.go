package shipping

import (
	"time"

	"gorm.io/gorm"
)

// ShippingTemplate 运费模板
type ShippingTemplate struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	Name            string         `gorm:"not null" json:"name"`
	Type            string         `gorm:"not null" json:"type"` // weight, quantity, price
	FreeShipping    bool           `gorm:"default:false" json:"free_shipping"`
	FreeThreshold   float64        `gorm:"default:0" json:"free_threshold"` // 免邮门槛
	DefaultFee      float64        `gorm:"default:0" json:"default_fee"`
	Description     string         `gorm:"type:text" json:"description"`
	Enabled         bool           `gorm:"default:true" json:"enabled"`
	Rules           []ShippingRule `gorm:"foreignKey:TemplateID" json:"rules"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
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

// Carrier 物流公司
type Carrier struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Code        string         `gorm:"uniqueIndex;not null" json:"code"`
	TrackingURL string         `json:"tracking_url"`
	APIEndpoint string         `json:"api_endpoint"`
	APIKey      string         `json:"api_key"`
	APISecret   string         `json:"api_secret"`
	Contact     string         `json:"contact"`
	Phone       string         `json:"phone"`
	Email       string         `json:"email"`
	ServiceArea string         `gorm:"type:text" json:"service_area"` // JSON格式的服务区域
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Carrier) TableName() string {
	return "carriers"
}

// TrackingEvent 物流追踪事件
type TrackingEvent struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	OrderID        uint      `gorm:"not null;index" json:"order_id"`
	TrackingNumber string    `gorm:"index" json:"tracking_number"`
	CarrierCode    string    `json:"carrier_code"`
	Status         string    `json:"status"`
	Location       string    `json:"location"`
	Description    string    `gorm:"type:text" json:"description"`
	EventTime      time.Time `json:"event_time"`
	CreatedAt      time.Time `json:"created_at"`
}

// TableName 指定表名
func (TrackingEvent) TableName() string {
	return "tracking_events"
}

// ShippingZone 配送区域
type ShippingZone struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Countries   string         `gorm:"type:text" json:"countries"` // JSON数组
	States      string         `gorm:"type:text" json:"states"`    // JSON数组
	PostalCodes string         `gorm:"type:text" json:"postal_codes"` // JSON数组
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ShippingZone) TableName() string {
	return "shipping_zones"
}
