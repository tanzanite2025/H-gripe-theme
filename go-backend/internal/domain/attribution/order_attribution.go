package attribution

import "time"

// OrderAttribution is a signed first-party campaign snapshot bound to an order
// before checkout. It does not itself indicate that a payment succeeded.
type OrderAttribution struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	OrderID     uint      `gorm:"not null;uniqueIndex" json:"order_id"`
	Source      string    `gorm:"size:96" json:"source,omitempty"`
	Medium      string    `gorm:"size:96" json:"medium,omitempty"`
	Campaign    string    `gorm:"size:160" json:"campaign,omitempty"`
	Term        string    `gorm:"size:160" json:"term,omitempty"`
	Content     string    `gorm:"size:160" json:"content,omitempty"`
	ClickIDKind string    `gorm:"size:32" json:"click_id_kind,omitempty"`
	ClickID     string    `gorm:"size:256" json:"click_id,omitempty"`
	CapturedAt  time.Time `gorm:"not null;index" json:"captured_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func (OrderAttribution) TableName() string {
	return "order_attributions"
}
