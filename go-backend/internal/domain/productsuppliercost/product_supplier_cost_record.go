package productsuppliercost

import (
	"commerce-platform/internal/domain/currency"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const DefaultCurrency = "USD"

// ProductSupplierCostRecord stores SKU-level unit cost and supplier reference
// data for one product snapshot. It is not a supplier order, receiving record,
// payment record, inventory record, or supplier-side transaction.
type ProductSupplierCostRecord struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	ProductCode string `gorm:"size:160;not null;uniqueIndex" json:"product_code"`
	ProductName string `gorm:"size:255;not null;index" json:"product_name"`
	// Monetary values are persisted as exact minor units. UnitCost maps to the
	// historical purchase_price concept, but the physical column is now
	// purchase_price_minor.
	UnitCostMinor int64 `gorm:"column:purchase_price_minor;not null;default:0" json:"unit_cost_minor"`
	Currency                     string    `gorm:"size:3;not null" json:"currency"`
	SupplierName                 string    `gorm:"size:255;not null" json:"supplier_name"`
	SupplierContactName          string    `gorm:"size:255" json:"supplier_contact_name"`
	SupplierPhone                string    `gorm:"size:80" json:"supplier_phone"`
	SupplierEmail                string    `gorm:"size:190" json:"supplier_email"`
	LeadTimeDays                 int       `gorm:"not null;default:0" json:"lead_time_days"`
	MinimumOrderQuantity         int       `gorm:"not null;default:1" json:"minimum_order_quantity"`
	InboundShippingUnitCostMinor int64     `gorm:"column:inbound_shipping_unit_cost_minor;not null;default:0" json:"inbound_shipping_unit_cost_minor"`
	PackagingUnitCostMinor       int64     `gorm:"column:packaging_unit_cost_minor;not null;default:0" json:"packaging_unit_cost_minor"`
	OtherUnitCostMinor           int64     `gorm:"column:other_unit_cost_minor;not null;default:0" json:"other_unit_cost_minor"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

func (ProductSupplierCostRecord) TableName() string {
	// Keep the historical physical table name for migration and deployment
	// compatibility. The table stores supplier-cost data only; it does not
	// represent a supplier-side workflow.
	return "product_procurement_records"
}

func (p *ProductSupplierCostRecord) BeforeCreate(tx *gorm.DB) error {
	return p.normalize()
}

func (p *ProductSupplierCostRecord) BeforeSave(tx *gorm.DB) error {
	return p.normalize()
}

func (p *ProductSupplierCostRecord) normalize() error {
	p.Currency = currency.NormalizeCode(strings.TrimSpace(p.Currency))
	if p.Currency == "" {
		p.Currency = DefaultCurrency
	}
	if !currency.IsCatalogCode(p.Currency) {
		return fmt.Errorf("unsupported currency %q", p.Currency)
	}
	if p.UnitCostMinor < 0 || p.InboundShippingUnitCostMinor < 0 || p.PackagingUnitCostMinor < 0 || p.OtherUnitCostMinor < 0 {
		return fmt.Errorf("supplier cost amounts cannot be negative")
	}
	p.ProductCode = strings.TrimSpace(p.ProductCode)
	p.ProductName = strings.TrimSpace(p.ProductName)
	p.SupplierName = strings.TrimSpace(p.SupplierName)
	p.SupplierContactName = strings.TrimSpace(p.SupplierContactName)
	p.SupplierPhone = strings.TrimSpace(p.SupplierPhone)
	p.SupplierEmail = strings.TrimSpace(p.SupplierEmail)
	return nil
}
