package shipping

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"fmt"
	"math/big"
	"strings"
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
	ID                    uint   `gorm:"primarykey" json:"id"`
	CarrierID             uint   `gorm:"not null;index;uniqueIndex:idx_shipping_carrier_service_code" json:"carrier_id"`
	TemplateID            *uint  `gorm:"index" json:"template_id"`
	ServiceCode           string `gorm:"type:varchar(80);not null;uniqueIndex:idx_shipping_carrier_service_code" json:"service_code"`
	ServiceName           string `gorm:"type:varchar(160);not null" json:"service_name"`
	RouteName             string `gorm:"type:varchar(160)" json:"route_name"`
	Countries             string `gorm:"type:text;default:'[]';not null" json:"countries"`
	Currency              string `gorm:"type:varchar(10);not null" json:"currency"`
	BillingMode           string `gorm:"type:varchar(40);default:'actual_weight';not null" json:"billing_mode"`
	FirstWeightGrams      int    `gorm:"default:0;not null" json:"first_weight_grams"`
	AdditionalWeightGrams int    `gorm:"default:0;not null" json:"additional_weight_grams"`
	MinChargeWeightGrams  int    `gorm:"default:0;not null" json:"min_charge_weight_grams"`
	VolumetricDivisor     int    `gorm:"default:6000;not null" json:"volumetric_divisor"`
	// FuelSurchargePercentDecimal is the exact percentage applied to the base
	// shipping fee (for example, "7.5" means 7.5%). It is the sole persisted
	// source of truth; display layers may project it to a number at their
	// boundary, but transaction calculations must use this decimal value.
	FuelSurchargePercentDecimal string `gorm:"column:fuel_surcharge_percent_decimal;type:numeric(30,15);default:0;not null" json:"fuel_surcharge_percent_decimal"`
	RemoteSurchargeMinor        int64  `gorm:"column:remote_surcharge_minor;not null;default:0" json:"remote_surcharge_minor"`
	// RemotePostalCodes is a JSON array of exact codes, prefixes ("123*"),
	// or inclusive ranges ("10000-10099") that qualify for RemoteSurchargeMinor.
	// An empty list preserves the legacy behavior of applying the surcharge to
	// every shipment served by this carrier service.
	RemotePostalCodes string            `gorm:"type:text;default:'[]';not null" json:"remote_postal_codes"`
	EtaMinDays        int               `gorm:"default:0;not null" json:"eta_min_days"`
	EtaMaxDays        int               `gorm:"default:0;not null" json:"eta_max_days"`
	Enabled           bool              `gorm:"default:true;not null;index" json:"enabled"`
	SortOrder         int               `gorm:"default:0;not null;index" json:"sort_order"`
	Description       string            `gorm:"type:text" json:"description"`
	Carrier           *Carrier          `gorm:"foreignKey:CarrierID" json:"carrier,omitempty"`
	Template          *ShippingTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (CarrierService) TableName() string {
	return "shipping_carrier_services"
}

func (s *CarrierService) BeforeCreate(tx *gorm.DB) error { return s.validateMoneyFields() }
func (s *CarrierService) BeforeSave(tx *gorm.DB) error   { return s.validateMoneyFields() }

func (s *CarrierService) validateMoneyFields() error {
	if strings.TrimSpace(s.FuelSurchargePercentDecimal) == "" {
		s.FuelSurchargePercentDecimal = "0"
	}
	if s.RemoteSurchargeMinor < 0 {
		return fmt.Errorf("remote surcharge cannot be negative")
	}
	if _, err := s.FuelSurchargeRate(); err != nil {
		return err
	}
	return nil
}

// FuelSurchargeRate parses the configured surcharge percentage as an exact
// rational. Empty values are normalized to zero for zero-value construction
// and partial updates; persisted/API values cannot contain fractional syntax
// and are limited to NUMERIC(30,15)'s scale.
func (s CarrierService) FuelSurchargeRate() (*big.Rat, error) {
	value := strings.TrimSpace(s.FuelSurchargePercentDecimal)
	if value == "" {
		value = "0"
	}
	if strings.Contains(value, "/") {
		return nil, fmt.Errorf("fuel surcharge percent must be a decimal")
	}
	rate, ok := new(big.Rat).SetString(value)
	if !ok || rate.Sign() < 0 || rate.Cmp(big.NewRat(100, 1)) > 0 {
		return nil, fmt.Errorf("fuel surcharge percent must be between 0 and 100")
	}
	scaled := new(big.Rat).Mul(rate, new(big.Rat).SetInt64(1_000_000_000_000_000))
	if scaled.Denom().Cmp(big.NewInt(1)) != 0 {
		return nil, fmt.Errorf("fuel surcharge percent supports at most 15 fractional digits")
	}
	return rate, nil
}

func (s CarrierService) RemoteSurchargeMoney() (domainmoney.Money, error) {
	return domainmoney.New(s.RemoteSurchargeMinor, currency.NormalizeCode(s.Currency))
}
