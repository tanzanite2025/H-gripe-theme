package coupon

import "time"

// GiftCardRedemption is the immutable business record connecting a user,
// a gift card, and the points ledger entry that paid for it.
type GiftCardRedemption struct {
	ID                   uint      `gorm:"primarykey" json:"id"`
	UserID               uint      `gorm:"not null;index" json:"user_id"`
	GiftCardID           uint      `gorm:"not null;uniqueIndex" json:"gift_card_id"`
	LoyaltyTransactionID *uint     `gorm:"uniqueIndex" json:"loyalty_transaction_id,omitempty"`
	ProgramConfigID      uint      `gorm:"not null;index" json:"program_config_id"`
	IdempotencyKey       string    `gorm:"not null" json:"-"`
	Currency             string    `gorm:"not null" json:"currency"`
	GiftCardValueCents   int64     `gorm:"not null" json:"gift_card_value_cents"`
	PointsSpent          int       `gorm:"not null" json:"points_spent"`
	Status               string    `gorm:"index;not null;default:'completed'" json:"status"` // completed
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	GiftCard             *GiftCard `gorm:"foreignKey:GiftCardID" json:"gift_card,omitempty"`
}

func (GiftCardRedemption) TableName() string {
	return "gift_card_redemptions"
}
