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
	ID                          uint             `gorm:"primarykey" json:"id"`
	OrderID                     uint             `gorm:"not null;index" json:"order_id"`
	TransactionID               uint             `gorm:"index" json:"transaction_id"`
	RefundID                    *string          `gorm:"uniqueIndex" json:"refund_id,omitempty"`
	Currency                    string           `gorm:"size:3;not null;default:'USD'" json:"currency"`
	AmountMinor                 int64            `gorm:"column:amount_minor;not null;default:0" json:"amount_minor"`
	GiftCardRefundAmountMinor   int64            `gorm:"column:gift_card_refund_amount_minor;not null;default:0" json:"gift_card_refund_amount_minor"`
	RequestedAmountMinor        int64            `gorm:"column:requested_amount_minor;not null;default:0" json:"requested_amount_minor"`
	DiscountClawbackAmountMinor int64            `gorm:"column:discount_clawback_amount_minor;not null;default:0" json:"discount_clawback_amount_minor"`
	Amount                      float64          `gorm:"not null" json:"-"` // legacy persistence projection; use AmountMinor
	GiftCardRefundAmount        float64          `gorm:"not null;default:0" json:"-"`
	RequestedAmount             float64          `gorm:"not null;default:0" json:"-"`
	DiscountClawbackAmount      float64          `gorm:"not null;default:0" json:"-"`
	LoyaltySettlementPrepared   bool             `gorm:"not null;default:false" json:"-"`
	LoyaltyPointsClawback       int              `gorm:"not null;default:0" json:"-"`
	LoyaltyPointsReturned       int              `gorm:"not null;default:0" json:"-"`
	LoyaltyCashDeductionAmount  float64          `gorm:"not null;default:0" json:"-"`
	CalculationSnapshot         string           `gorm:"type:text" json:"calculation_snapshot"`
	FXSnapshotData              datatypes.JSON   `gorm:"column:fx_snapshot;type:jsonb;not null;default:'{}'" json:"-"`
	LineItems                   []RefundLineItem `gorm:"foreignKey:RefundID" json:"line_items,omitempty"`
	Reason                      string           `gorm:"type:text" json:"reason"`
	Status                      string           `gorm:"index" json:"status"` // pending, completed, failed
	RefundedBy                  uint             `json:"refunded_by"`         // 操作人ID
	GatewayResponse             string           `gorm:"type:text" json:"gateway_response"`
	CreatedAt                   time.Time        `json:"created_at"`
	UpdatedAt                   time.Time        `json:"updated_at"`
	CompletedAt                 *time.Time       `json:"completed_at"`
	DeletedAt                   gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (r *Refund) BeforeSave(tx *gorm.DB) error {
	r.Currency = currency.NormalizeCode(r.Currency)
	if r.Currency == "" {
		// Historical fixtures may not carry a currency snapshot yet. Leave the
		// field unset rather than guessing a currency; production workflows set
		// it from the captured payment transaction before persistence.
		return nil
	}
	if !currency.IsCatalogCode(r.Currency) {
		return errors.New("refund currency must be a supported ISO 4217 code")
	}
	fields := []struct {
		major float64
		minor *int64
	}{
		{r.Amount, &r.AmountMinor},
		{r.GiftCardRefundAmount, &r.GiftCardRefundAmountMinor},
		{r.RequestedAmount, &r.RequestedAmountMinor},
		{r.DiscountClawbackAmount, &r.DiscountClawbackAmountMinor},
	}
	for _, field := range fields {
		// Keep the persisted exact snapshot in sync when a legacy major-unit
		// projection is updated by an existing workflow.
		if field.major != 0 {
			value, err := domainmoney.FromMajorFloat(field.major, r.Currency)
			if err != nil {
				return err
			}
			*field.minor = value.AmountMinor()
		}
		if *field.minor < 0 {
			return errors.New("refund monetary amounts cannot be negative")
		}
	}
	return nil
}

func (r Refund) AmountMoney() (domainmoney.Money, error) {
	if r.AmountMinor == 0 && r.Amount != 0 {
		return domainmoney.FromMajorFloat(r.Amount, r.Currency)
	}
	return domainmoney.New(r.AmountMinor, r.Currency)
}

func (r Refund) RequestedAmountMoney() (domainmoney.Money, error) {
	if r.RequestedAmountMinor == 0 && r.RequestedAmount != 0 {
		return domainmoney.FromMajorFloat(r.RequestedAmount, r.Currency)
	}
	return domainmoney.New(r.RequestedAmountMinor, r.Currency)
}

func (r Refund) GiftCardRefundMoney() (domainmoney.Money, error) {
	if r.GiftCardRefundAmountMinor == 0 && r.GiftCardRefundAmount != 0 {
		return domainmoney.FromMajorFloat(r.GiftCardRefundAmount, r.Currency)
	}
	return domainmoney.New(r.GiftCardRefundAmountMinor, r.Currency)
}

func (r Refund) DiscountClawbackMoney() (domainmoney.Money, error) {
	if r.DiscountClawbackAmountMinor == 0 && r.DiscountClawbackAmount != 0 {
		return domainmoney.FromMajorFloat(r.DiscountClawbackAmount, r.Currency)
	}
	return domainmoney.New(r.DiscountClawbackAmountMinor, r.Currency)
}

func (r Refund) LoyaltyCashDeductionMoney() (domainmoney.Money, error) {
	return domainmoney.FromMajorFloat(r.LoyaltyCashDeductionAmount, r.Currency)
}

// TableName 指定表名
func (Refund) TableName() string {
	return "refunds"
}

// RefundLineItem records the exact order item facts that produced a refund.
// Product snapshots are copied from order_items so later catalog edits cannot
// alter historical refund accounting.
type RefundLineItem struct {
	ID                uint   `gorm:"primarykey" json:"id"`
	RefundID          uint   `gorm:"not null;index" json:"refund_id"`
	OrderID           uint   `gorm:"not null;index" json:"order_id"`
	OrderItemID       uint   `gorm:"not null;index" json:"order_item_id"`
	ProductID         uint   `gorm:"not null;index" json:"product_id"`
	VariantID         *uint  `gorm:"index" json:"variant_id,omitempty"`
	ProductName       string `gorm:"not null" json:"product_name"`
	SKU               string `json:"sku"`
	Quantity          int    `gorm:"not null" json:"quantity"`
	Currency          string `gorm:"size:3;not null;default:'USD'" json:"currency"`
	UnitPriceMinor    int64  `gorm:"column:unit_price_minor;not null;default:0" json:"unit_price_minor"`
	LineSubtotalMinor int64  `gorm:"column:line_subtotal_minor;not null;default:0" json:"line_subtotal_minor"`
	LineTaxMinor      int64  `gorm:"column:line_tax_minor;not null;default:0" json:"line_tax_minor"`
	LineDiscountMinor int64  `gorm:"column:line_discount_minor;not null;default:0" json:"line_discount_minor"`
	LineTotalMinor    int64  `gorm:"column:line_total_minor;not null;default:0" json:"line_total_minor"`
	// Legacy major-unit projections are retained for historical rows and
	// internal migration tooling only; transport contracts expose minor units.
	UnitPrice          float64    `gorm:"not null;default:0" json:"-"`
	LineSubtotalAmount float64    `gorm:"not null;default:0" json:"-"`
	LineTaxAmount      float64    `gorm:"not null;default:0" json:"-"`
	LineDiscountAmount float64    `gorm:"not null;default:0" json:"-"`
	LineTotalAmount    float64    `gorm:"not null;default:0" json:"-"`
	Restock            bool       `gorm:"not null;default:false" json:"restock"`
	RestockedAt        *time.Time `json:"restocked_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (i *RefundLineItem) BeforeSave(tx *gorm.DB) error {
	i.Currency = currency.NormalizeCode(i.Currency)
	if i.Currency == "" {
		i.Currency = currency.DefaultPrimaryCurrency
	}
	if !currency.IsCatalogCode(i.Currency) {
		return errors.New("refund line item currency must be a supported ISO 4217 code")
	}
	fields := []struct {
		major float64
		minor *int64
	}{
		{i.UnitPrice, &i.UnitPriceMinor},
		{i.LineSubtotalAmount, &i.LineSubtotalMinor},
		{i.LineTaxAmount, &i.LineTaxMinor},
		{i.LineDiscountAmount, &i.LineDiscountMinor},
		{i.LineTotalAmount, &i.LineTotalMinor},
	}
	for _, field := range fields {
		if *field.minor == 0 && field.major != 0 {
			value, err := domainmoney.FromMajorFloat(field.major, i.Currency)
			if err != nil {
				return err
			}
			*field.minor = value.AmountMinor()
		}
		if *field.minor < 0 {
			return errors.New("refund line item monetary amounts cannot be negative")
		}
	}
	return nil
}

func (i RefundLineItem) UnitPriceMoney() (domainmoney.Money, error) {
	if i.UnitPriceMinor == 0 && i.UnitPrice != 0 {
		return domainmoney.FromMajorFloat(i.UnitPrice, i.Currency)
	}
	return domainmoney.New(i.UnitPriceMinor, i.Currency)
}

func (i RefundLineItem) LineSubtotalMoney() (domainmoney.Money, error) {
	if i.LineSubtotalMinor == 0 && i.LineSubtotalAmount != 0 {
		return domainmoney.FromMajorFloat(i.LineSubtotalAmount, i.Currency)
	}
	return domainmoney.New(i.LineSubtotalMinor, i.Currency)
}

func (i RefundLineItem) LineTaxMoney() (domainmoney.Money, error) {
	if i.LineTaxMinor == 0 && i.LineTaxAmount != 0 {
		return domainmoney.FromMajorFloat(i.LineTaxAmount, i.Currency)
	}
	return domainmoney.New(i.LineTaxMinor, i.Currency)
}

func (i RefundLineItem) LineDiscountMoney() (domainmoney.Money, error) {
	if i.LineDiscountMinor == 0 && i.LineDiscountAmount != 0 {
		return domainmoney.FromMajorFloat(i.LineDiscountAmount, i.Currency)
	}
	return domainmoney.New(i.LineDiscountMinor, i.Currency)
}

func (i RefundLineItem) LineTotalMoney() (domainmoney.Money, error) {
	if i.LineTotalMinor == 0 && i.LineTotalAmount != 0 {
		return domainmoney.FromMajorFloat(i.LineTotalAmount, i.Currency)
	}
	return domainmoney.New(i.LineTotalMinor, i.Currency)
}

func (RefundLineItem) TableName() string {
	return "refund_line_items"
}
