package order

import "time"

// OrderIdempotency is the durable idempotency fact for an order creation
// request. The row and the order are committed in the same transaction.
type OrderIdempotency struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	UserID         uint      `gorm:"not null;uniqueIndex:idx_order_idempotency_scope_key" json:"user_id"`
	Scope          string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_order_idempotency_scope_key" json:"scope"`
	IdempotencyKey string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_order_idempotency_scope_key" json:"idempotency_key"`
	RequestHash    string    `gorm:"type:varchar(64);not null" json:"request_hash"`
	OrderID        *uint     `gorm:"index" json:"order_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (OrderIdempotency) TableName() string {
	return "order_idempotencies"
}
