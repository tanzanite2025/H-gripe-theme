package order

import "time"

// OrderSummary 订单摘要（用于列表）
type OrderSummary struct {
	ID               uint       `json:"id"`
	OrderNumber      string     `json:"order_number"`
	UserID           uint       `json:"user_id"`
	Status           string     `json:"status"`
	PaymentStatus    string     `json:"payment_status"`
	ShippingStatus   string     `json:"shipping_status"`
	TotalAmountMinor int64      `json:"total_amount_minor"`
	Currency         string     `json:"currency"`
	ItemCount        int        `json:"item_count"`
	CustomerName     string     `json:"customer_name"`
	CreatedAt        time.Time  `json:"created_at"`
	PaidAt           *time.Time `json:"paid_at"`
}

// ToSummary 转换为摘要
func (o *Order) ToSummary() *OrderSummary {
	customerName := o.ShippingAddress.FirstName + " " + o.ShippingAddress.LastName
	totalAmountMinor := o.TotalAmountMinor
	return &OrderSummary{
		ID:               o.ID,
		OrderNumber:      o.OrderNumber,
		UserID:           o.UserID,
		Status:           o.Status,
		PaymentStatus:    o.PaymentStatus,
		ShippingStatus:   o.ShippingStatus,
		TotalAmountMinor: totalAmountMinor,
		Currency:         o.Currency,
		ItemCount:        len(o.Items),
		CustomerName:     customerName,
		CreatedAt:        o.CreatedAt,
		PaidAt:           o.PaidAt,
	}
}
