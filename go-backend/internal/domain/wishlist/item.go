package wishlist

import (
	"tanzanite/internal/domain/product"
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

type ItemResponse struct {
	ID        uint                         `json:"id"`
	ProductID uint                         `json:"product_id"`
	CreatedAt time.Time                    `json:"created_at"`
	Product   *product.ProductListResponse `json:"product,omitempty"`
}

func (i *Item) ToResponse() *ItemResponse {
	resp := &ItemResponse{
		ID:        i.ID,
		ProductID: i.ProductID,
		CreatedAt: i.CreatedAt,
	}
	if i.Product != nil {
		resp.Product = i.Product.ToListResponse()
	}
	return resp
}
