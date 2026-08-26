package warranty

import (
	"time"

	"gorm.io/gorm"
)

// WarrantyServiceRecord 保修服务记录
type WarrantyServiceRecord struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	ClaimID        uint           `gorm:"not null;index" json:"claim_id"`
	ServiceType    string         `gorm:"size:80;not null;default:'inspection'" json:"service_type"`
	Status         string         `gorm:"size:50;not null;default:'open';index" json:"status"`
	Summary        string         `gorm:"type:text;not null" json:"summary"`
	CostAmount     float64        `gorm:"type:numeric(12,2);default:0" json:"cost_amount"`
	Currency       string         `gorm:"size:8;not null" json:"currency"`
	PerformedBy    uint           `gorm:"index" json:"performed_by"`
	CreatedBy      uint           `gorm:"index" json:"created_by"`
	PerformedAt    *time.Time     `json:"performed_at"`
	Claim          *WarrantyClaim `gorm:"foreignKey:ClaimID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"claim,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (WarrantyServiceRecord) TableName() string {
	return "warranty_service_records"
}
