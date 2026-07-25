package registration

import (
	"tanzanite/internal/domain/order"
	"time"

	"gorm.io/gorm"
)

// WarrantyClaim 保修申请
type WarrantyClaim struct {
	ID             uint                    `gorm:"primarykey" json:"id"`
	RegistrationID *uint                   `gorm:"index" json:"registration_id"`
	OrderItemID    *uint                   `gorm:"index" json:"order_item_id"`
	UserID         uint                    `gorm:"not null;index" json:"user_id"`
	IssueType      string                  `gorm:"not null" json:"issue_type"` // defect, damage, malfunction
	Description    string                  `gorm:"type:text;not null" json:"description"`
	Images         string                  `gorm:"type:text" json:"images"` // JSON数组
	OrderNumber    string                  `gorm:"size:80" json:"order_number"`
	Email          string                  `gorm:"size:190" json:"email"`
	TirePressure   string                  `gorm:"size:40" json:"tire_pressure"`
	IsTubeless     bool                    `gorm:"default:false" json:"is_tubeless"`
	VideoURL       string                  `json:"video_url"`
	Status         string                  `gorm:"index;default:'submitted'" json:"status"` // submitted, reviewing, approved, rejected, completed
	Resolution     string                  `gorm:"type:text" json:"resolution"`
	ProcessedBy    uint                    `json:"processed_by"`
	ProcessedAt    *time.Time              `json:"processed_at"`
	Registration   *ProductRegistration    `gorm:"foreignKey:RegistrationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"registration,omitempty"`
	OrderItem      *order.OrderItem        `gorm:"foreignKey:OrderItemID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"order_item,omitempty"`
	ServiceRecords []WarrantyServiceRecord `gorm:"foreignKey:ClaimID" json:"service_records,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	DeletedAt      gorm.DeletedAt          `gorm:"index" json:"-"`
}

// TableName 指定表名
func (WarrantyClaim) TableName() string {
	return "warranty_claims"
}
