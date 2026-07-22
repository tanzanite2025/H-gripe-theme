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

// CarrierService 承运商线路服务
type CarrierService struct {
	ID                    uint              `gorm:"primarykey" json:"id"`
	CarrierID             uint              `gorm:"not null;index;uniqueIndex:idx_shipping_carrier_service_code" json:"carrier_id"`
	TemplateID            *uint             `gorm:"index" json:"template_id"`
	ServiceCode           string            `gorm:"type:varchar(80);not null;uniqueIndex:idx_shipping_carrier_service_code" json:"service_code"`
	ServiceName           string            `gorm:"type:varchar(160);not null" json:"service_name"`
	RouteName             string            `gorm:"type:varchar(160)" json:"route_name"`
	Countries             string            `gorm:"type:text;default:'[]';not null" json:"countries"`
	Currency              string            `gorm:"type:varchar(10);default:'USD';not null" json:"currency"`
	BillingMode           string            `gorm:"type:varchar(40);default:'actual_weight';not null" json:"billing_mode"`
	FirstWeightGrams      int               `gorm:"default:0;not null" json:"first_weight_grams"`
	AdditionalWeightGrams int               `gorm:"default:0;not null" json:"additional_weight_grams"`
	MinChargeWeightGrams  int               `gorm:"default:0;not null" json:"min_charge_weight_grams"`
	VolumetricDivisor     int               `gorm:"default:6000;not null" json:"volumetric_divisor"`
	FuelSurchargePercent  float64           `gorm:"type:decimal(8,3);default:0;not null" json:"fuel_surcharge_percent"`
	RemoteSurcharge       float64           `gorm:"type:decimal(12,2);default:0;not null" json:"remote_surcharge"`
	EtaMinDays            int               `gorm:"default:0;not null" json:"eta_min_days"`
	EtaMaxDays            int               `gorm:"default:0;not null" json:"eta_max_days"`
	Enabled               bool              `gorm:"default:true;not null;index" json:"enabled"`
	SortOrder             int               `gorm:"default:0;not null;index" json:"sort_order"`
	Description           string            `gorm:"type:text" json:"description"`
	Carrier               *Carrier          `gorm:"foreignKey:CarrierID" json:"carrier,omitempty"`
	Template              *ShippingTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
	DeletedAt             gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (CarrierService) TableName() string {
	return "shipping_carrier_services"
}

// TrackingProviderConfig 物流追踪服务商配置，例如 17TRACK、AfterShip 等
type TrackingProviderConfig struct {
	ID                     uint           `gorm:"primarykey" json:"id"`
	ProviderCode           string         `gorm:"type:varchar(80);not null;index" json:"provider_code"`
	ProviderName           string         `gorm:"type:varchar(160);not null" json:"provider_name"`
	Environment            string         `gorm:"type:varchar(40);default:'production';not null" json:"environment"`
	BaseURL                string         `gorm:"type:text" json:"base_url"`
	APIKey                 string         `gorm:"type:text" json:"api_key"`
	WebhookSecret          string         `gorm:"type:text" json:"webhook_secret"`
	WebhookEnabled         bool           `gorm:"default:false;not null" json:"webhook_enabled"`
	AutoRegister           bool           `gorm:"default:false;not null" json:"auto_register"`
	PollingEnabled         bool           `gorm:"default:false;not null" json:"polling_enabled"`
	PollingIntervalMinutes int            `gorm:"default:60;not null" json:"polling_interval_minutes"`
	RequestTimeoutSeconds  int            `gorm:"default:15;not null" json:"request_timeout_seconds"`
	Enabled                bool           `gorm:"default:true;not null;index" json:"enabled"`
	SortOrder              int            `gorm:"default:0;not null;index" json:"sort_order"`
	Description            string         `gorm:"type:text" json:"description"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TrackingProviderConfig) TableName() string {
	return "shipping_tracking_providers"
}

// TrackingCarrierMapping maps local carriers or carrier services to provider-specific carrier codes.
type TrackingCarrierMapping struct {
	ID                  uint                    `gorm:"primarykey" json:"id"`
	ProviderID          uint                    `gorm:"not null;index" json:"provider_id"`
	Scope               string                  `gorm:"type:varchar(40);default:'carrier';not null;index" json:"scope"` // carrier, carrier_service
	CarrierID           *uint                   `gorm:"index" json:"carrier_id"`
	CarrierServiceID    *uint                   `gorm:"index" json:"carrier_service_id"`
	ProviderCarrierCode string                  `gorm:"type:varchar(120);not null;index" json:"provider_carrier_code"`
	ProviderCarrierName string                  `gorm:"type:varchar(160)" json:"provider_carrier_name"`
	Enabled             bool                    `gorm:"default:true;not null;index" json:"enabled"`
	Priority            int                     `gorm:"default:0;not null;index" json:"priority"`
	Description         string                  `gorm:"type:text" json:"description"`
	Provider            *TrackingProviderConfig `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	Carrier             *Carrier                `gorm:"foreignKey:CarrierID" json:"carrier,omitempty"`
	CarrierService      *CarrierService         `gorm:"foreignKey:CarrierServiceID" json:"carrier_service,omitempty"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
	DeletedAt           gorm.DeletedAt          `gorm:"index" json:"-"`
}

func (TrackingCarrierMapping) TableName() string {
	return "shipping_tracking_carrier_mappings"
}

// TrackingShipment tracks the sync lifecycle for an order tracking number.
// Order fields remain the source for provider/carrier selection; this table stores operational status only.
type TrackingShipment struct {
	ID                       uint                    `gorm:"primarykey" json:"id"`
	OrderID                  uint                    `gorm:"not null;index;uniqueIndex:idx_shipping_tracking_shipments_order" json:"order_id"`
	TrackingProviderID       uint                    `gorm:"not null;index" json:"tracking_provider_id"`
	TrackingNumber           string                  `gorm:"type:varchar(120);not null;index" json:"tracking_number"`
	ProviderCarrierCode      string                  `gorm:"type:varchar(120);not null;index" json:"provider_carrier_code"`
	CarrierID                *uint                   `gorm:"index" json:"carrier_id"`
	CarrierServiceID         *uint                   `gorm:"index" json:"carrier_service_id"`
	TrackingCarrierMappingID *uint                   `gorm:"index" json:"tracking_carrier_mapping_id"`
	RegistrationStatus       string                  `gorm:"type:varchar(40);default:'pending';not null;index" json:"registration_status"` // pending, registered, failed
	SyncStatus               string                  `gorm:"type:varchar(40);default:'pending';not null;index" json:"sync_status"`         // pending, syncing, synced, failed
	EventCount               int                     `gorm:"default:0;not null" json:"event_count"`
	LastEventAt              *time.Time              `gorm:"index" json:"last_event_at"`
	LastSyncedAt             *time.Time              `gorm:"index" json:"last_synced_at"`
	NextSyncAt               *time.Time              `gorm:"index" json:"next_sync_at"`
	LastError                string                  `gorm:"type:text" json:"last_error"`
	Enabled                  bool                    `gorm:"default:true;not null;index" json:"enabled"`
	Provider                 *TrackingProviderConfig `gorm:"foreignKey:TrackingProviderID" json:"provider,omitempty"`
	Carrier                  *Carrier                `gorm:"foreignKey:CarrierID" json:"carrier,omitempty"`
	CarrierService           *CarrierService         `gorm:"foreignKey:CarrierServiceID" json:"carrier_service,omitempty"`
	Mapping                  *TrackingCarrierMapping `gorm:"foreignKey:TrackingCarrierMappingID" json:"mapping,omitempty"`
	CreatedAt                time.Time               `json:"created_at"`
	UpdatedAt                time.Time               `json:"updated_at"`
	DeletedAt                gorm.DeletedAt          `gorm:"index" json:"-"`
}

func (TrackingShipment) TableName() string {
	return "shipping_tracking_shipments"
}

// TrackingEvent 物流追踪事件
type TrackingEvent struct {
	ID                  uint      `gorm:"primarykey" json:"id"`
	OrderID             uint      `gorm:"not null;index" json:"order_id"`
	TrackingNumber      string    `gorm:"index" json:"tracking_number"`
	ProviderCarrierCode string    `json:"provider_carrier_code"`
	Status              string    `json:"status"`
	Location            string    `json:"location"`
	Description         string    `gorm:"type:text" json:"description"`
	EventTime           time.Time `json:"event_time"`
	CreatedAt           time.Time `json:"created_at"`
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
	ID        uint           `gorm:"primarykey" json:"id"`
	RuleID    uint           `gorm:"not null;index;uniqueIndex:idx_shipping_packaging_rule_apply_rule_product" json:"rule_id"`
	ProductID uint           `gorm:"not null;index;uniqueIndex:idx_shipping_packaging_rule_apply_rule_product;uniqueIndex:idx_shipping_packaging_rule_apply_product" json:"product_id"`
	Rule      *PackagingRule `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
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
