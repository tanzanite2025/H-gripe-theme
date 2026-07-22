package product

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID            uint               `gorm:"primarykey" json:"id"`
	ProductTypeID *uint              `gorm:"index" json:"product_type_id"`
	SKU           string             `gorm:"uniqueIndex;not null" json:"sku"`
	Name          string             `gorm:"not null;index" json:"name"`
	Slug          string             `gorm:"uniqueIndex:idx_product_slug_locale;not null" json:"slug"`
	Description   string             `gorm:"type:text" json:"description"`
	ShortDesc     string             `gorm:"type:text" json:"short_description"`
	Price         float64            `gorm:"not null" json:"price"`
	SalePrice     *float64           `json:"sale_price"`
	Stock         int                `gorm:"default:0" json:"stock"`
	Weight        int                `gorm:"column:weight_grams" json:"weight_grams"` // 克
	Status        string             `gorm:"default:'active';index" json:"status"`    // active, inactive, out_of_stock
	Locale        string             `gorm:"uniqueIndex:idx_product_slug_locale;default:'en';index" json:"locale"`
	ParentID      *uint              `gorm:"index" json:"parent_id"` // 翻译关联
	Featured      bool               `gorm:"default:false" json:"featured"`
	ViewCount     int                `gorm:"default:0" json:"view_count"`
	MetaTitle     string             `json:"meta_title"`
	MetaDesc      string             `gorm:"type:text" json:"meta_description"`
	Media         []ProductMedia     `gorm:"foreignKey:ProductID" json:"media,omitempty"`
	ProductType   *ProductType       `gorm:"foreignKey:ProductTypeID" json:"product_type,omitempty"`
	SpecValues    []ProductSpecValue `gorm:"foreignKey:ProductID" json:"spec_values,omitempty"`
	Variants      []ProductVariant   `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	DeletedAt     gorm.DeletedAt     `gorm:"index" json:"-"`
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

// ProductMedia stores product-owned media placement.
// Galleries are intentionally separate showcase collections and should not be
// used as the source of truth for product images or videos.
type ProductMedia struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ProductID    uint           `gorm:"not null;index" json:"product_id"`
	VariantID    *uint          `gorm:"index" json:"variant_id,omitempty"`
	MediaAssetID *uint          `gorm:"index" json:"media_asset_id,omitempty"`
	MediaType    string         `gorm:"default:'image';not null;index" json:"media_type"`
	Role         string         `gorm:"default:'gallery';not null;index" json:"role"`
	URL          string         `gorm:"not null" json:"url"`
	ThumbnailURL string         `json:"thumbnail_url"`
	PosterURL    string         `json:"poster_url"`
	Alt          string         `json:"alt"`
	Title        string         `json:"title"`
	Locale       string         `gorm:"index" json:"locale"`
	SortOrder    int            `gorm:"default:0;not null" json:"sort_order"`
	IsPrimary    bool           `gorm:"default:false;not null" json:"is_primary"`
	IsVisible    bool           `gorm:"default:true;not null" json:"is_visible"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductMedia) TableName() string {
	return "product_media"
}

// ProductListResponse 产品列表响应
type ProductListResponse struct {
	ID           uint     `json:"id"`
	SKU          string   `json:"sku"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	ShortDesc    string   `json:"short_description"`
	Price        float64  `json:"price"`
	SalePrice    *float64 `json:"sale_price"`
	Stock        int      `json:"stock"`
	Status       string   `json:"status"`
	Locale       string   `json:"locale"`
	Featured     bool     `json:"featured"`
	FeaturedImg  string   `json:"featured_image"`
	VariantCount int      `json:"variant_count"`
}

// ToListResponse 转换为列表响应
func (p *Product) ToListResponse() *ProductListResponse {
	price, salePrice := p.DisplayPrices()
	resp := &ProductListResponse{
		ID:           p.ID,
		SKU:          p.DisplaySKU(),
		Name:         p.Name,
		Slug:         p.Slug,
		ShortDesc:    p.ShortDesc,
		Price:        price,
		SalePrice:    salePrice,
		Stock:        p.TotalVariantStock(),
		Status:       p.Status,
		Locale:       p.Locale,
		Featured:     p.Featured,
		VariantCount: len(p.Variants),
	}

	if len(p.Media) > 0 {
		for _, item := range p.Media {
			if item.MediaType == "image" && item.IsVisible && item.URL != "" {
				resp.FeaturedImg = item.URL
				break
			}
		}
	}

	return resp
}

func (p *Product) ActiveVariants() []ProductVariant {
	var variants []ProductVariant
	for _, variant := range p.Variants {
		if variant.IsActive {
			variants = append(variants, variant)
		}
	}
	return variants
}

func (p *Product) DefaultVariant() *ProductVariant {
	activeVariants := p.ActiveVariants()
	if len(activeVariants) == 0 {
		return nil
	}
	for i := range activeVariants {
		if activeVariants[i].IsDefault {
			return &activeVariants[i]
		}
	}
	return &activeVariants[0]
}

func (p *Product) DisplaySKU() string {
	if variant := p.DefaultVariant(); variant != nil {
		return variant.SKU
	}
	return p.SKU
}

func (p *Product) DisplayPrices() (float64, *float64) {
	if variant := p.DefaultVariant(); variant != nil {
		return variant.Price, variant.SalePrice
	}
	return p.Price, p.SalePrice
}

func (p *Product) TotalVariantStock() int {
	activeVariants := p.ActiveVariants()
	if len(activeVariants) == 0 {
		return p.Stock
	}

	total := 0
	for _, variant := range activeVariants {
		total += variant.Stock
	}
	return total
}

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

// ProductAttribute 商品属性
type ProductAttribute struct {
	ID           uint             `gorm:"primarykey" json:"id"`
	Name         string           `gorm:"type:varchar(120);not null" json:"name"`
	Slug         string           `gorm:"type:varchar(120);uniqueIndex;not null" json:"slug"`
	Type         string           `gorm:"type:varchar(32);default:'select';not null" json:"type"` // select, text
	SortOrder    int              `gorm:"default:0;not null" json:"sort_order"`
	IsFilterable bool             `gorm:"default:true;not null" json:"is_filterable"`
	AffectsSKU   bool             `gorm:"default:true;not null" json:"affects_sku"`
	AffectsStock bool             `gorm:"default:false;not null" json:"affects_stock"`
	IsEnabled    bool             `gorm:"default:true;not null" json:"is_enabled"`
	Meta         string           `gorm:"type:text" json:"meta"`
	Values       []AttributeValue `gorm:"foreignKey:AttributeID" json:"values"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// TableName 指定表名
func (ProductAttribute) TableName() string {
	return "product_attributes"
}

// AttributeValue 属性值
type AttributeValue struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	AttributeID uint      `gorm:"not null;index" json:"attribute_id"`
	Name        string    `gorm:"type:varchar(120);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(120);not null" json:"slug"`
	Value       string    `gorm:"type:varchar(255)" json:"value"`
	SortOrder   int       `gorm:"default:0;not null" json:"sort_order"`
	IsEnabled   bool      `gorm:"default:true;not null" json:"is_enabled"`
	Meta        string    `gorm:"type:text" json:"meta"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AttributeValue) TableName() string {
	return "product_attribute_values"
}

type ProductType struct {
	ID              uint             `gorm:"primarykey" json:"id"`
	Name            string           `gorm:"type:varchar(120);not null" json:"name"`
	Slug            string           `gorm:"type:varchar(120);uniqueIndex;not null" json:"slug"`
	Description     string           `gorm:"type:text" json:"description"`
	SortOrder       int              `gorm:"default:0;not null" json:"sort_order"`
	IsEnabled       bool             `gorm:"default:true;not null" json:"is_enabled"`
	SpecDefinitions []SpecDefinition `gorm:"foreignKey:ProductTypeID" json:"spec_definitions,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

func (ProductType) TableName() string {
	return "product_types"
}

type SpecDefinition struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	ProductTypeID   uint      `gorm:"not null;index;uniqueIndex:idx_product_type_spec_slug" json:"product_type_id"`
	Group           string    `gorm:"type:varchar(80);default:'specs';not null" json:"group"`
	Name            string    `gorm:"type:varchar(120);not null" json:"name"`
	Slug            string    `gorm:"type:varchar(120);not null;uniqueIndex:idx_product_type_spec_slug" json:"slug"`
	FieldType       string    `gorm:"type:varchar(32);default:'text';not null" json:"field_type"`
	Unit            string    `gorm:"type:varchar(32)" json:"unit"`
	IsRequired      bool      `gorm:"default:false;not null" json:"is_required"`
	IsFilterable    bool      `gorm:"default:false;not null" json:"is_filterable"`
	IsVisible       bool      `gorm:"default:true;not null" json:"is_visible"`
	IsVariantOption bool      `gorm:"default:false;not null" json:"is_variant_option"`
	SortOrder       int       `gorm:"default:0;not null" json:"sort_order"`
	Options         string    `gorm:"type:text" json:"options"`
	Validation      string    `gorm:"type:text" json:"validation"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (SpecDefinition) TableName() string {
	return "product_spec_definitions"
}

type ProductSpecValue struct {
	ID               uint            `gorm:"primarykey" json:"id"`
	ProductID        uint            `gorm:"not null;index;uniqueIndex:idx_product_spec_value" json:"product_id"`
	SpecDefinitionID uint            `gorm:"not null;index;uniqueIndex:idx_product_spec_value" json:"spec_definition_id"`
	Value            string          `gorm:"type:text;not null" json:"value"`
	SpecDefinition   *SpecDefinition `gorm:"foreignKey:SpecDefinitionID" json:"definition,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (ProductSpecValue) TableName() string {
	return "product_spec_values"
}

type ProductVariant struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ProductID    uint           `gorm:"not null;index;uniqueIndex:idx_product_variant_options" json:"product_id"`
	SKU          string         `gorm:"type:varchar(120);uniqueIndex;not null" json:"sku"`
	Title        string         `gorm:"type:varchar(160)" json:"title"`
	OptionValues string         `gorm:"type:text;not null;uniqueIndex:idx_product_variant_options" json:"option_values"`
	Price        float64        `gorm:"not null" json:"price"`
	SalePrice    *float64       `json:"sale_price"`
	Stock        int            `gorm:"default:0;not null" json:"stock"`
	Weight       int            `gorm:"column:weight_grams" json:"weight_grams"`
	IsDefault    bool           `gorm:"default:false;not null" json:"is_default"`
	IsActive     bool           `gorm:"default:true;not null" json:"is_active"`
	SortOrder    int            `gorm:"default:0;not null" json:"sort_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductVariant) TableName() string {
	return "product_variants"
}

func (v *ProductVariant) EffectivePrice() float64 {
	if v.SalePrice != nil {
		return *v.SalePrice
	}
	return v.Price
}

func (v *ProductVariant) BeforeCreate(tx *gorm.DB) error {
	if v.OptionValues == "" {
		v.OptionValues = "{}"
	}
	return nil
}
