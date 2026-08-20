package productiontest

import "time"

const (
	OrderMarkerStatusActive     = "active"
	OrderMarkerStatusReconciled = "reconciled"
	OrderMarkerStatusVoided     = "voided"
)

// OrderMarker is written immediately after a production-test order is created.
type OrderMarker struct {
	ID                   uint      `gorm:"primarykey" json:"id"`
	OrderID              uint      `gorm:"not null;uniqueIndex" json:"order_id"`
	OrderNumber          string    `gorm:"size:80;not null;index" json:"order_number"`
	UserID               uint      `gorm:"not null;index" json:"user_id"`
	TestAccountID        uint      `gorm:"not null;index" json:"test_account_id"`
	Status               string    `gorm:"size:24;not null;default:'active';index" json:"status"`
	HoldFulfillment      bool      `gorm:"not null;default:true" json:"hold_fulfillment"`
	ExcludeFromRevenue   bool      `gorm:"not null;default:true" json:"exclude_from_revenue"`
	ExcludeFromAnalytics bool      `gorm:"not null;default:true" json:"exclude_from_analytics"`
	Reason               string    `gorm:"type:text;not null;default:''" json:"reason"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (OrderMarker) TableName() string {
	return "production_test_order_markers"
}

func NewOrderMarker(orderID uint, orderNumber string, decision PurchaseDecision) (OrderMarker, bool) {
	if !decision.Allowed || !decision.MarkerRequired || orderID == 0 || decision.UserID == 0 || decision.TestAccountID == 0 {
		return OrderMarker{}, false
	}

	return OrderMarker{
		OrderID:              orderID,
		OrderNumber:          orderNumber,
		UserID:               decision.UserID,
		TestAccountID:        decision.TestAccountID,
		Status:               OrderMarkerStatusActive,
		HoldFulfillment:      decision.HoldFulfillment,
		ExcludeFromRevenue:   true,
		ExcludeFromAnalytics: true,
		Reason:               "production test checkout",
	}, true
}

