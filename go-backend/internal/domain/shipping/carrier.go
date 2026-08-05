package shipping

import (
	"time"

	"gorm.io/gorm"
)

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
	Currency              string            `gorm:"type:varchar(10);not null" json:"currency"`
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
