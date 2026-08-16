package order

import "time"

// OrderItem 订单商品项
type OrderItem struct {
	ID                     uint      `gorm:"primarykey" json:"id"`
	OrderID                uint      `gorm:"not null;index" json:"order_id"`
	ProductID              uint      `gorm:"not null;index" json:"product_id"`
	VariantID              *uint     `gorm:"not null;index" json:"variant_id"`
	ProductName            string    `gorm:"not null" json:"product_name"`
	SKU                    string    `json:"sku"`
	Quantity               int       `gorm:"not null" json:"quantity"`
	Price                  float64   `gorm:"not null" json:"price"`
	Subtotal               float64   `gorm:"not null" json:"subtotal"`
	TaxAmount              float64   `gorm:"default:0" json:"tax_amount"`
	Discount               float64   `gorm:"default:0" json:"discount"`
	Total                  float64   `gorm:"not null" json:"total"`
	Attributes             string    `gorm:"type:text" json:"attributes"` // JSON格式的商品属性
	HSCode                 string    `gorm:"column:hs_code;size:12" json:"hs_code"`
	CNCode                 string    `gorm:"column:cn_code;size:12" json:"cn_code"`
	CountryOfOrigin        string    `gorm:"column:country_of_origin;size:2" json:"country_of_origin"`
	CustomsDescription     string    `gorm:"column:customs_description;size:255" json:"customs_description"`
	DeclaredValue          *float64  `gorm:"column:declared_value;type:numeric(12,2)" json:"declared_value"`
	DeclaredValueConfirmed bool      `gorm:"column:declared_value_confirmed;not null;default:false" json:"declared_value_confirmed"`
	CreatedAt              time.Time `json:"created_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}
