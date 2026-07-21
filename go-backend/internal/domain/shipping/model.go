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
	Countries   string         `gorm:"type:text" json:"countries"`    // JSON数组
	States      string         `gorm:"type:text" json:"states"`       // JSON数组
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

// PackagingRule 包装规格规则
type PackagingRule struct {
	ID          uint                 `gorm:"primarykey" json:"id"`
	RuleName    string               `gorm:"type:varchar(100);not null" json:"rule_name"`
	Description string               `gorm:"type:text" json:"description"`
	BoxWeight   float64              `gorm:"type:decimal(10,3);default:0;not null" json:"box_weight"`
	BoxLength   float64              `gorm:"type:decimal(10,2);default:0;not null" json:"box_length"`
	BoxWidth    float64              `gorm:"type:decimal(10,2);default:0;not null" json:"box_width"`
	BoxHeight   float64              `gorm:"type:decimal(10,2);default:0;not null" json:"box_height"`
	MaxWeight   float64              `gorm:"type:decimal(10,3);default:0;not null" json:"max_weight"`
	IsActive    bool                 `gorm:"default:true;index" json:"is_active"`
	Applies     []PackagingRuleApply `gorm:"foreignKey:RuleID" json:"applies"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// TableName 指定表名
func (PackagingRule) TableName() string {
	return "shipping_packaging_rules"
}

// PackagingRuleApply 包装规则应用的产品
type PackagingRuleApply struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	RuleID    uint      `gorm:"not null;index" json:"rule_id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (PackagingRuleApply) TableName() string {
	return "shipping_packaging_rule_applies"
}

// ShippingTemplateBinding maps a shipping template to a default, product type, product, or variant scope.
type ShippingTemplateBinding struct {
	ID            uint              `gorm:"primarykey" json:"id"`
	TemplateID    uint              `gorm:"not null;index" json:"template_id"`
	Scope         string            `gorm:"type:varchar(30);not null;index" json:"scope"` // default, product_type, product, variant
	ProductTypeID *uint             `gorm:"index" json:"product_type_id"`
	ProductID     *uint             `gorm:"index" json:"product_id"`
	VariantID     *uint             `gorm:"index" json:"variant_id"`
	Priority      int               `gorm:"default:0;not null;index" json:"priority"`
	Enabled       bool              `gorm:"default:true;not null;index" json:"enabled"`
	Template      *ShippingTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (ShippingTemplateBinding) TableName() string {
	return "shipping_template_bindings"
}
