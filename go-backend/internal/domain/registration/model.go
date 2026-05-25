package registration

import (
	"time"

	"gorm.io/gorm"
)

// ProductRegistration 产品注册
type ProductRegistration struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	UserID          uint           `gorm:"not null;index" json:"user_id"`
	ProductID       uint           `gorm:"not null;index" json:"product_id"`
	SerialNumber    string         `gorm:"uniqueIndex;not null" json:"serial_number"`
	PurchaseDate    time.Time      `gorm:"not null" json:"purchase_date"`
	PurchaseProof   string         `json:"purchase_proof"` // 购买凭证图片URL
	Retailer        string         `json:"retailer"` // 购买商家
	WarrantyPeriod  int            `gorm:"not null" json:"warranty_period"` // 保修期（月）
	WarrantyExpires time.Time      `gorm:"not null" json:"warranty_expires"`
	Status          string         `gorm:"index;default:'active'" json:"status"` // active, expired, claimed
	Notes           string         `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ProductRegistration) TableName() string {
	return "product_registrations"
}

// WarrantyClaim 保修申请
type WarrantyClaim struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	RegistrationID uint           `gorm:"not null;index" json:"registration_id"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	IssueType      string         `gorm:"not null" json:"issue_type"` // defect, damage, malfunction
	Description    string         `gorm:"type:text;not null" json:"description"`
	Images         string         `gorm:"type:text" json:"images"` // JSON数组
	Status         string         `gorm:"index;default:'submitted'" json:"status"` // submitted, reviewing, approved, rejected, completed
	Resolution     string         `gorm:"type:text" json:"resolution"`
	ProcessedBy    uint           `json:"processed_by"`
	ProcessedAt    *time.Time     `json:"processed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (WarrantyClaim) TableName() string {
	return "warranty_claims"
}
