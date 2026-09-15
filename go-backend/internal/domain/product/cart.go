package product

import (
	"errors"
	"time"

	domainmoney "commerce-platform/internal/domain/money"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	DefaultConfigurationJSON = `{"schema_version":1,"selections":[]}`
	DefaultConfigurationHash = "49aaaf6402be0bdf1a77fbed39f489bd4b6f3e6364e8aca69b5c49db4899a9a1"
)

func DefaultConfigurationJSONBytes() []byte {
	return []byte(DefaultConfigurationJSON)
}

// Cart 购物车
type Cart struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    *uint          `gorm:"index;uniqueIndex:uq_carts_active_user_id,where:user_id IS NOT NULL AND deleted_at IS NULL" json:"user_id"`                        // 可为空（游客购物车）
	SessionID string         `gorm:"index;uniqueIndex:uq_carts_active_session_id,where:user_id IS NULL AND session_id <> '' AND deleted_at IS NULL" json:"session_id"` // 游客会话ID
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
	ID                uint            `gorm:"primarykey" json:"id"`
	CartID            uint            `gorm:"not null;index;uniqueIndex:uq_cart_items_identity" json:"cart_id"`
	ProductID         uint            `gorm:"not null;index;uniqueIndex:uq_cart_items_identity" json:"product_id"`
	VariantID         *uint           `gorm:"not null;index;uniqueIndex:uq_cart_items_identity" json:"variant_id"`
	ConfigurationData datatypes.JSON  `gorm:"column:configuration;type:jsonb;not null;default:'{\"schema_version\":1,\"selections\":[]}'" json:"configuration"`
	ConfigurationHash string          `gorm:"column:configuration_hash;type:char(64);not null;default:'49aaaf6402be0bdf1a77fbed39f489bd4b6f3e6364e8aca69b5c49db4899a9a1';uniqueIndex:uq_cart_items_identity" json:"configuration_hash"`
	Quantity          int             `gorm:"not null" json:"quantity"`
	PriceMinor        int64           `gorm:"column:price_minor;not null" json:"price_minor"`
	Currency          string          `gorm:"size:3;not null;default:'USD'" json:"currency"`
	Product           *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Variant           *ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (CartItem) TableName() string {
	return "cart_items"
}

func (i CartItem) PriceMoney() (domainmoney.Money, error) {
	if i.PriceMinor < 0 {
		return domainmoney.Money{}, errors.New("cart item price cannot be negative")
	}
	return domainmoney.New(i.PriceMinor, i.Currency)
}

func (i *CartItem) SetPriceMoney(value domainmoney.Money) error {
	if err := value.Validate(); err != nil {
		return err
	}
	if value.AmountMinor() < 0 {
		return errors.New("cart item price cannot be negative")
	}
	i.PriceMinor = value.AmountMinor()
	i.Currency = value.Currency().String()
	return nil
}

// CartSummary 购物车摘要
type CartSummary struct {
	ItemCount  int               `json:"item_count"`
	Total      float64           `json:"total"`
	TotalMoney domainmoney.Money `json:"-"`
	Items      []CartItem        `json:"items"`
}
