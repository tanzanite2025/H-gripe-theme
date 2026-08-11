package registration

import (
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/user"
	"time"

	"gorm.io/gorm"
)

// ProductRegistration 产品注册
type ProductRegistration struct {
	ID              uint             `gorm:"primarykey" json:"id"`
	UserID          uint             `gorm:"not null;index" json:"user_id"`
	ProductID       uint             `gorm:"not null;index" json:"product_id"`
	SerialNumber    string           `gorm:"uniqueIndex;not null" json:"serial_number"`
	PurchaseDate    time.Time        `gorm:"not null" json:"purchase_date"`
	PurchaseProof   string           `json:"purchase_proof"`                  // 购买凭证图片URL
	Retailer        string           `json:"retailer"`                        // 购买商家
	WarrantyPeriod  int              `gorm:"not null" json:"warranty_period"` // 保修期（月）
	WarrantyExpires time.Time        `gorm:"not null" json:"warranty_expires"`
	Status          string           `gorm:"index;default:'active'" json:"status"` // active, expired, claimed
	Notes           string           `gorm:"type:text" json:"notes"`
	User            *user.User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product         *product.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ProductRegistration) TableName() string {
	return "product_registrations"
}
