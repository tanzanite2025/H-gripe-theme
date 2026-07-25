package product

import (
	"time"

	"gorm.io/gorm"
)

// Cart 购物车
type Cart struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    *uint          `gorm:"index" json:"user_id"`    // 可为空（游客购物车）
	SessionID string         `gorm:"index" json:"session_id"` // 游客会话ID
	Items     []CartItem     `gorm:"foreignKey:CartID" json:"items"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Cart) TableName() string {
	return "carts"
}

// CartItem 购物车项目
type CartItem struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	CartID    uint            `gorm:"not null;index;uniqueIndex:idx_cart_product_variant" json:"cart_id"`
	ProductID uint            `gorm:"not null;index;uniqueIndex:idx_cart_product_variant" json:"product_id"`
	VariantID *uint           `gorm:"not null;index;uniqueIndex:idx_cart_product_variant" json:"variant_id"`
	Quantity  int             `gorm:"not null" json:"quantity"`
	Price     float64         `gorm:"not null" json:"price"` // 快照价格
	Product   *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Variant   *ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (CartItem) TableName() string {
	return "cart_items"
}

// CartSummary 购物车摘要
type CartSummary struct {
	ItemCount int        `json:"item_count"`
	Total     float64    `json:"total"`
	Items     []CartItem `json:"items"`
}
