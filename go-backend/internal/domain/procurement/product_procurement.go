package procurement

import (
	"commerce-platform/internal/domain/currency"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const DefaultCurrency = "USD"

// ProductProcurement stores the operational sourcing data for one product
// reference. The reference is intentionally a local product code snapshot:
// this domain does not import or join the public product catalog.
type ProductProcurement struct {
	ID                      uint      `gorm:"primarykey" json:"id"`
	ProductCode             string    `gorm:"size:160;not null;uniqueIndex" json:"product_code"`
	ProductName             string    `gorm:"size:255;not null;index" json:"product_name"`
	PurchasePrice           float64   `gorm:"type:numeric(14,2);not null" json:"purchase_price"`
	Currency                string    `gorm:"size:3;not null" json:"currency"`
	SupplierName            string    `gorm:"size:255;not null" json:"supplier_name"`
	SupplierContactName     string    `gorm:"size:255" json:"supplier_contact_name"`
	SupplierPhone           string    `gorm:"size:80" json:"supplier_phone"`
	SupplierEmail           string    `gorm:"size:190" json:"supplier_email"`
	LeadTimeDays            int       `gorm:"not null;default:0" json:"lead_time_days"`
	MinimumOrderQuantity    int       `gorm:"not null;default:1" json:"minimum_order_quantity"`
	InboundShippingUnitCost float64   `gorm:"type:numeric(14,2);not null;default:0" json:"inbound_shipping_unit_cost"`
	PackagingUnitCost       float64   `gorm:"type:numeric(14,2);not null;default:0" json:"packaging_unit_cost"`
	OtherUnitCost           float64   `gorm:"type:numeric(14,2);not null;default:0" json:"other_unit_cost"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func (ProductProcurement) TableName() string {
	return "product_procurement_records"
}

func (p *ProductProcurement) BeforeCreate(tx *gorm.DB) error {
	return p.normalize()
}

func (p *ProductProcurement) BeforeSave(tx *gorm.DB) error {
	return p.normalize()
}

func (p *ProductProcurement) normalize() error {
	p.Currency = currency.NormalizeCode(strings.TrimSpace(p.Currency))
	if p.Currency == "" {
		p.Currency = DefaultCurrency
	}
	if !currency.IsCatalogCode(p.Currency) {
		return fmt.Errorf("unsupported currency %q", p.Currency)
	}
	p.ProductCode = strings.TrimSpace(p.ProductCode)
	p.ProductName = strings.TrimSpace(p.ProductName)
	p.SupplierName = strings.TrimSpace(p.SupplierName)
	p.SupplierContactName = strings.TrimSpace(p.SupplierContactName)
	p.SupplierPhone = strings.TrimSpace(p.SupplierPhone)
	p.SupplierEmail = strings.TrimSpace(p.SupplierEmail)
	return nil
}
