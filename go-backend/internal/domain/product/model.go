package product

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	SKU         string         `gorm:"uniqueIndex;not null" json:"sku"`
	Name        string         `gorm:"not null;index" json:"name"`
	Slug        string         `gorm:"uniqueIndex:idx_product_slug_locale;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	ShortDesc   string         `gorm:"type:text" json:"short_description"`
	Price       float64        `gorm:"not null" json:"price"`
	SalePrice   *float64       `json:"sale_price"`
	Stock       int            `gorm:"default:0" json:"stock"`
	Weight      int            `json:"weight_grams"` // 克
	Status      string         `gorm:"default:'active';index" json:"status"` // active, inactive, out_of_stock
	Locale      string         `gorm:"uniqueIndex:idx_product_slug_locale;default:'en';index" json:"locale"`
	ParentID    *uint          `gorm:"index" json:"parent_id"` // 翻译关联
	Featured    bool           `gorm:"default:false" json:"featured"`
	ViewCount   int            `gorm:"default:0" json:"view_count"`
	MetaTitle   string         `json:"meta_title"`
	MetaDesc    string         `gorm:"type:text" json:"meta_description"`
	Images      []ProductImage `gorm:"foreignKey:ProductID" json:"images"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

// BeforeCreate GORM钩子：创建前
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.Locale == "" {
		p.Locale = "en"
	}
	if p.Status == "" {
		p.Status = "active"
	}
	return nil
}

// ProductImage 产品图片
type ProductImage struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	URL       string    `gorm:"not null" json:"url"`
	Alt       string    `json:"alt"`
	Order     int       `gorm:"default:0" json:"order"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (ProductImage) TableName() string {
	return "product_images"
}

// ProductListResponse 产品列表响应
type ProductListResponse struct {
	ID          uint     `json:"id"`
	SKU         string   `json:"sku"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	ShortDesc   string   `json:"short_description"`
	Price       float64  `json:"price"`
	SalePrice   *float64 `json:"sale_price"`
	Stock       int      `json:"stock"`
	Status      string   `json:"status"`
	Locale      string   `json:"locale"`
	Featured    bool     `json:"featured"`
	FeaturedImg string   `json:"featured_image"`
}

// ToListResponse 转换为列表响应
func (p *Product) ToListResponse() *ProductListResponse {
	resp := &ProductListResponse{
		ID:        p.ID,
		SKU:       p.SKU,
		Name:      p.Name,
		Slug:      p.Slug,
		ShortDesc: p.ShortDesc,
		Price:     p.Price,
		SalePrice: p.SalePrice,
		Stock:     p.Stock,
		Status:    p.Status,
		Locale:    p.Locale,
		Featured:  p.Featured,
	}
	
	// 获取第一张图片作为特色图片
	if len(p.Images) > 0 {
		resp.FeaturedImg = p.Images[0].URL
	}
	
	return resp
}

// Cart 购物车
type Cart struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    *uint          `gorm:"index" json:"user_id"` // 可为空（游客购物车）
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
	ID        uint      `gorm:"primarykey" json:"id"`
	CartID    uint      `gorm:"not null;index" json:"cart_id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"not null" json:"price"` // 快照价格
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CartItem) TableName() string {
	return "cart_items"
}

// CartSummary 购物车摘要
type CartSummary struct {
	ItemCount int     `json:"item_count"`
	Total     float64 `json:"total"`
}
