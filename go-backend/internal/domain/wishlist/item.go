package wishlist

import (
	"commerce-platform/internal/domain/product"
	"time"

	"gorm.io/gorm"
)

type Item struct {
	ID        uint             `gorm:"primarykey" json:"id"`
	UserID    uint             `gorm:"not null;uniqueIndex:idx_wishlist_user_product;index" json:"user_id"`
	ProductID uint             `gorm:"not null;uniqueIndex:idx_wishlist_user_product;index" json:"product_id"`
	Product   *product.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (Item) TableName() string {
	return "wishlist_items"
}
