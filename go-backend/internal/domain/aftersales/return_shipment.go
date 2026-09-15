package aftersales

import "time"

// AfterSalesReturnShipment is a physical return package attached to an
// after-sales case. Keeping it separate from the case supports a replacement
// label or a later multi-package return without losing the original warehouse
// and tracking audit trail.
type AfterSalesReturnShipment struct {
	ID               uint       `gorm:"primarykey" json:"id"`
	CaseID           uint       `gorm:"not null;index" json:"case_id"`
	WarehouseName    string     `gorm:"size:200;not null;default:''" json:"warehouse_name"`
	WarehouseAddress string     `gorm:"type:text;not null;default:''" json:"warehouse_address"`
	Carrier          string     `gorm:"size:100;not null;default:''" json:"carrier"`
	TrackingNumber   string     `gorm:"size:200;not null;default:''" json:"tracking_number"`
	TrackingURL      string     `gorm:"type:text;not null;default:''" json:"tracking_url"`
	LabelURL         string     `gorm:"type:text;not null;default:''" json:"label_url"`
	ShippedAt        *time.Time `gorm:"index" json:"shipped_at,omitempty"`
	ReceivedAt       *time.Time `gorm:"index" json:"received_at,omitempty"`
	ReceivedBy       *uint      `gorm:"index" json:"received_by,omitempty"`
	CreatedBy        uint       `gorm:"not null;default:0" json:"created_by"`
	UpdatedBy        uint       `gorm:"not null;default:0" json:"updated_by"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (AfterSalesReturnShipment) TableName() string {
	return "after_sales_return_shipments"
}
