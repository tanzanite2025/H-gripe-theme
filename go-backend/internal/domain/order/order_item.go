package order

import "time"

// OrderItem 订单商品项
type OrderItem struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	OrderID     uint      `gorm:"not null;index" json:"order_id"`
	ProductID   uint      `gorm:"not null;index" json:"product_id"`
	VariantID   *uint     `gorm:"not null;index" json:"variant_id"`
	ProductName string    `gorm:"not null" json:"product_name"`
	SKU         string    `json:"sku"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	Price       float64   `gorm:"not null" json:"price"`
	Subtotal    float64   `gorm:"not null" json:"subtotal"`
	TaxAmount   float64   `gorm:"default:0" json:"tax_amount"`
	Discount    float64   `gorm:"default:0" json:"discount"`
	Total       float64   `gorm:"not null" json:"total"`
	Attributes  string    `gorm:"type:text" json:"attributes"` // JSON格式的商品属性
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}
