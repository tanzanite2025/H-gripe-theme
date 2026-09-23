package shipping

import (
	"time"

	"gorm.io/datatypes"
)

// ShipmentRecord is an optional after-sales attachment for an already shipped
// order. Order status and package tracking are sourced from their respective
// order and tracking-shipment records; this table stores warranty evidence.
type ShipmentRecord struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	OrderID         uint           `gorm:"not null;uniqueIndex:idx_shipment_records_order" json:"order_id"`
	OrderNumber     string         `gorm:"type:varchar(255);not null;index" json:"order_number"`
	UserID          uint           `gorm:"index" json:"user_id"`
	CustomerName    string         `gorm:"type:varchar(255)" json:"customer_name"`
	CustomerEmail   string         `gorm:"type:varchar(190);index" json:"customer_email"`
	ShippedAt       time.Time      `gorm:"not null;index" json:"shipped_at"`
	ItemsSnapshot   datatypes.JSON `gorm:"column:items_snapshot;type:jsonb;not null;default:'[]'" json:"items_snapshot"`
	ProductCodes    datatypes.JSON `gorm:"column:product_codes;type:jsonb;not null;default:'[]'" json:"product_codes"`
	DetailsBound    bool           `gorm:"column:details_bound;not null;default:false" json:"-"`
	ShippingNote    string         `gorm:"type:text" json:"shipping_note"`
	ShippingImages  datatypes.JSON `gorm:"column:shipping_images;type:jsonb;not null;default:'[]'" json:"shipping_images"`
	WarrantyMonths  int            `gorm:"not null;default:12" json:"warranty_months"`
	WarrantyStartAt time.Time      `gorm:"not null;index" json:"warranty_start_at"`
	WarrantyExpires time.Time      `gorm:"not null;index" json:"warranty_expires"`
	Status          string         `gorm:"type:varchar(40);not null;default:'active';index" json:"status"` // active, expired, cancelled
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// These fields are populated only for the order-backed admin/public view.
	RecordBound       bool                    `gorm:"-" json:"record_bound"`
	OrderStatus       string                  `gorm:"-" json:"order_status"`
	ShippingState     string                  `gorm:"-" json:"shipping_status"`
	TrackingShipments []ShipmentRecordPackage `gorm:"-" json:"tracking_shipments"`
}

// ShipmentRecordPackage exposes package facts without provider credentials.
type ShipmentRecordPackage struct {
	ID                  uint   `json:"id"`
	TrackingNumber      string `json:"tracking_number"`
	ProviderCarrierCode string `json:"provider_carrier_code"`
	Enabled             bool   `json:"enabled"`
}

func (ShipmentRecord) TableName() string {
	return "shipment_records"
}
