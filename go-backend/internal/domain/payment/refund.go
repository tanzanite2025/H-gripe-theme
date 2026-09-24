package payment

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Refund 退款记录
type Refund struct {
	ID                          uint           `gorm:"primarykey" json:"id"`
	OrderID                     uint           `gorm:"not null;index" json:"order_id"`
	TransactionID               uint           `gorm:"index" json:"transaction_id"`
	RefundID                    *string        `gorm:"uniqueIndex" json:"refund_id,omitempty"`
	Currency                    string         `gorm:"size:3;not null;default:'USD'" json:"currency"`
	AmountMinor                 int64          `gorm:"column:amount_minor;not null;default:0" json:"amount_minor"`
	RequestedAmountMinor        int64          `gorm:"column:requested_amount_minor;not null;default:0" json:"requested_amount_minor"`
	DiscountClawbackAmountMinor int64          `gorm:"column:discount_clawback_amount_minor;not null;default:0" json:"discount_clawback_amount_minor"`
	LoyaltySettlementPrepared   bool           `gorm:"not null;default:false" json:"-"`
	LoyaltyPointsClawback       int            `gorm:"not null;default:0" json:"-"`
	LoyaltyPointsDebt           int            `gorm:"column:loyalty_points_debt;not null;default:0" json:"-"`
	CalculationSnapshot         string         `gorm:"type:text" json:"calculation_snapshot"`
	FXSnapshotData              datatypes.JSON `gorm:"column:fx_snapshot;type:jsonb;not null;default:'{}'" json:"-"`
	// Provider settlement facts are separate from the customer-facing refund
	// amount because a gateway may convert the deduction at refund time.
	SettlementAmountMinor          int64  `gorm:"column:settlement_amount_minor;not null;default:0" json:"settlement_amount_minor"`
	SettlementCurrency             string `gorm:"column:settlement_currency;size:3;not null;default:''" json:"settlement_currency"`
	SettlementBalanceTransactionID string `gorm:"column:settlement_balance_transaction_id;size:255;not null;default:''" json:"settlement_balance_transaction_id"`
	// Signed convention: positive means a loss (actual settlement deduction
	// exceeds the historical base-currency value), negative means a gain.
	FXGainLossMinor    int64            `gorm:"column:fx_gain_loss_minor;not null;default:0" json:"fx_gain_loss_minor"`
	FXGainLossCurrency string           `gorm:"column:fx_gain_loss_currency;size:3;not null;default:''" json:"fx_gain_loss_currency"`
	LineItems          []RefundLineItem `gorm:"foreignKey:RefundID" json:"line_items,omitempty"`
	Reason             string           `gorm:"type:text" json:"reason"`
	Status             string           `gorm:"index" json:"status"` // pending, completed, failed
	RefundedBy         uint             `json:"refunded_by"`         // 操作人ID
	GatewayResponse    string           `gorm:"type:text" json:"gateway_response"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	CompletedAt        *time.Time       `json:"completed_at"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (r *Refund) BeforeSave(tx *gorm.DB) error {
	r.Currency = currency.NormalizeCode(r.Currency)
	r.SettlementCurrency = currency.NormalizeCode(r.SettlementCurrency)
	r.FXGainLossCurrency = currency.NormalizeCode(r.FXGainLossCurrency)
	if r.Currency == "" {
		// Historical fixtures may not carry a currency snapshot yet. Leave the
		// field unset rather than guessing a currency; production workflows set
		// it from the captured payment transaction before persistence.
		return nil
	}
	if !currency.IsCatalogCode(r.Currency) {
		return errors.New("refund currency must be a supported ISO 4217 code")
	}
	if r.SettlementAmountMinor < 0 {
		return errors.New("refund settlement amount cannot be negative")
	}
	if r.SettlementAmountMinor > 0 && !currency.IsCatalogCode(r.SettlementCurrency) {
		return errors.New("refund settlement currency must be a supported ISO 4217 code")
	}
	if r.FXGainLossMinor != 0 && !currency.IsCatalogCode(r.FXGainLossCurrency) {
		return errors.New("refund FX gain/loss currency must be a supported ISO 4217 code")
	}
	for _, amount := range []int64{r.AmountMinor, r.RequestedAmountMinor, r.DiscountClawbackAmountMinor} {
		if amount < 0 {
			return errors.New("refund monetary amounts cannot be negative")
		}
	}
	return nil
}

func (r Refund) AmountMoney() (domainmoney.Money, error) {
	return domainmoney.New(r.AmountMinor, r.Currency)
}

func (r Refund) RequestedAmountMoney() (domainmoney.Money, error) {
	return domainmoney.New(r.RequestedAmountMinor, r.Currency)
}

func (r Refund) DiscountClawbackMoney() (domainmoney.Money, error) {
	return domainmoney.New(r.DiscountClawbackAmountMinor, r.Currency)
}

// TableName 指定表名
func (Refund) TableName() string {
	return "refunds"
}

// RefundLineItem records the exact order item facts that produced a refund.
// Product snapshots are copied from order_items so later catalog edits cannot
// alter historical refund accounting.
type RefundLineItem struct {
	ID                uint       `gorm:"primarykey" json:"id"`
	RefundID          uint       `gorm:"not null;index" json:"refund_id"`
	OrderID           uint       `gorm:"not null;index" json:"order_id"`
	OrderItemID       uint       `gorm:"not null;index" json:"order_item_id"`
	ProductID         uint       `gorm:"not null;index" json:"product_id"`
	VariantID         *uint      `gorm:"index" json:"variant_id,omitempty"`
	ProductName       string     `gorm:"not null" json:"product_name"`
	SKU               string     `json:"sku"`
	Quantity          int        `gorm:"not null" json:"quantity"`
	Currency          string     `gorm:"size:3;not null;default:'USD'" json:"currency"`
	UnitPriceMinor    int64      `gorm:"column:unit_price_minor;not null;default:0" json:"unit_price_minor"`
	LineSubtotalMinor int64      `gorm:"column:line_subtotal_minor;not null;default:0" json:"line_subtotal_minor"`
	LineTaxMinor      int64      `gorm:"column:line_tax_minor;not null;default:0" json:"line_tax_minor"`
	LineDiscountMinor int64      `gorm:"column:line_discount_minor;not null;default:0" json:"line_discount_minor"`
	LineTotalMinor    int64      `gorm:"column:line_total_minor;not null;default:0" json:"line_total_minor"`
	Restock           bool       `gorm:"not null;default:false" json:"restock"`
	RestockedAt       *time.Time `json:"restocked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (i *RefundLineItem) BeforeSave(tx *gorm.DB) error {
	i.Currency = currency.NormalizeCode(i.Currency)
	if i.Currency == "" {
		i.Currency = currency.DefaultPrimaryCurrency
	}
	if !currency.IsCatalogCode(i.Currency) {
		return errors.New("refund line item currency must be a supported ISO 4217 code")
	}
	for _, amount := range []int64{i.UnitPriceMinor, i.LineSubtotalMinor, i.LineTaxMinor, i.LineDiscountMinor, i.LineTotalMinor} {
		if amount < 0 {
			return errors.New("refund line item monetary amounts cannot be negative")
		}
	}
	return nil
}

func (i RefundLineItem) UnitPriceMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.UnitPriceMinor, i.Currency)
}

func (i RefundLineItem) LineSubtotalMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.LineSubtotalMinor, i.Currency)
}

func (i RefundLineItem) LineTaxMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.LineTaxMinor, i.Currency)
}

func (i RefundLineItem) LineDiscountMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.LineDiscountMinor, i.Currency)
}

func (i RefundLineItem) LineTotalMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.LineTotalMinor, i.Currency)
}

func (RefundLineItem) TableName() string {
	return "refund_line_items"
}
