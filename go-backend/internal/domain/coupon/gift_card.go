package coupon

import (
	"errors"
	"math"
	"time"

	"commerce-platform/internal/domain/currency"

	"gorm.io/gorm"
)

// GiftCard 礼品卡
type GiftCard struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	Code           string         `gorm:"uniqueIndex;not null" json:"code"`
	InitialValue   float64        `gorm:"-" json:"initial_value"`
	Balance        float64        `gorm:"-" json:"balance"`
	InitialCents   int64          `gorm:"column:initial_value_cents;not null;check:initial_value_cents_non_negative,initial_value_cents >= 0" json:"initial_value_cents"`
	BalanceCents   int64          `gorm:"column:balance_cents;not null;check:balance_cents_non_negative,balance_cents >= 0" json:"balance_cents"`
	Currency       string         `gorm:"not null" json:"currency"`
	Status         string         `gorm:"index" json:"status"` // active, used, expired, cancelled
	OwnerUserID    *uint          `gorm:"index" json:"owner_user_id,omitempty"`
	Origin         string         `gorm:"index;not null;default:'admin'" json:"origin"` // admin, loyalty_redemption
	RecipientEmail string         `json:"recipient_email"`
	RecipientName  string         `json:"recipient_name"`
	SenderName     string         `json:"sender_name"`
	Message        string         `gorm:"type:text" json:"message"`
	CoverImage     string         `json:"cover_image"`
	ExpiresAt      *time.Time     `json:"expires_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (GiftCard) TableName() string {
	return "gift_cards"
}

func (gc *GiftCard) BeforeSave(tx *gorm.DB) error {
	gc.Currency = currency.NormalizeCode(gc.Currency)
	if !currency.IsValidCode(gc.Currency) || !currency.IsCatalogCode(gc.Currency) {
		return errors.New("gift card currency must be a supported ISO 4217 code")
	}
	if gc.InitialValue > 0 || gc.InitialCents == 0 {
		gc.InitialCents = AmountToCents(gc.InitialValue)
	}
	if gc.Balance > 0 || gc.BalanceCents == 0 {
		gc.BalanceCents = AmountToCents(gc.Balance)
	}
	gc.syncAmountsFromCents()
	return nil
}

func (gc *GiftCard) AfterFind(tx *gorm.DB) error {
	gc.syncAmountsFromCents()
	return nil
}

func (gc *GiftCard) syncAmountsFromCents() {
	gc.InitialValue = CentsToAmount(gc.InitialCents)
	gc.Balance = CentsToAmount(gc.BalanceCents)
}

// IsValid 检查礼品卡是否有效
func (gc *GiftCard) IsValid() bool {
	if gc.Status != "active" {
		return false
	}
	if gc.BalanceCents <= 0 {
		return false
	}
	if gc.ExpiresAt != nil && time.Now().After(*gc.ExpiresAt) {
		return false
	}
	return true
}

// GiftCardTransaction 礼品卡交易记录
type GiftCardTransaction struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	GiftCardID   uint      `gorm:"not null;index" json:"gift_card_id"`
	RedemptionID *uint     `gorm:"index" json:"redemption_id,omitempty"`
	OrderID      uint      `gorm:"index" json:"order_id"`
	Type         string    `gorm:"not null" json:"type"` // issue, use, refund
	Amount       float64   `gorm:"-" json:"amount"`
	Balance      float64   `gorm:"-" json:"balance"` // 交易后余额
	AmountCents  int64     `gorm:"column:amount_cents;not null" json:"-"`
	BalanceCents int64     `gorm:"column:balance_cents;not null" json:"-"`
	Note         string    `gorm:"type:text" json:"note"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (GiftCardTransaction) TableName() string {
	return "gift_card_transactions"
}

func (t *GiftCardTransaction) BeforeSave(tx *gorm.DB) error {
	if t.Amount != 0 || t.AmountCents == 0 {
		t.AmountCents = AmountToCents(t.Amount)
	}
	if t.Balance != 0 || t.BalanceCents == 0 {
		t.BalanceCents = AmountToCents(t.Balance)
	}
	t.syncAmountsFromCents()
	return nil
}

func (t *GiftCardTransaction) AfterFind(tx *gorm.DB) error {
	t.syncAmountsFromCents()
	return nil
}

func (t *GiftCardTransaction) syncAmountsFromCents() {
	t.Amount = CentsToAmount(t.AmountCents)
	t.Balance = CentsToAmount(t.BalanceCents)
}

func AmountToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func CentsToAmount(cents int64) float64 {
	return float64(cents) / 100
}
