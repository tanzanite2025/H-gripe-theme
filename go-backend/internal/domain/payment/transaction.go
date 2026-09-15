package payment

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"

	"gorm.io/gorm"
)

const (
	TransactionStatusDuplicatePaid = "duplicate_paid"
)

// Transaction 支付交易记录
type Transaction struct {
	ID                 uint   `gorm:"primarykey" json:"id"`
	OrderID            uint   `gorm:"not null;index" json:"order_id"`
	TransactionID      string `gorm:"uniqueIndex" json:"transaction_id"` // 第三方交易ID
	AttemptKey         string `gorm:"size:255;index" json:"attempt_key,omitempty"`
	ProviderRequestKey string `gorm:"size:255;index" json:"-"`
	PaymentMethod      string `gorm:"not null" json:"payment_method"`
	AmountMinor        int64  `gorm:"column:amount_minor;not null;default:0" json:"amount_minor"`
	// Amount is retained as a persistence read field while historical rows are
	// backfilled. New payment logic must use AmountMoney/AmountMinor.
	Amount           float64        `gorm:"not null" json:"-"`
	Currency         string         `gorm:"not null" json:"currency"`
	Status           string         `gorm:"index" json:"status"`               // pending, processing, requires_action, completed, duplicate_paid, failed, expired, refunded
	GatewayResponse  string         `gorm:"type:text" json:"gateway_response"` // JSON格式
	LiabilityShifted *bool          `gorm:"column:liability_shifted" json:"liability_shifted,omitempty"`
	ErrorMessage     string         `gorm:"type:text" json:"error_message"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	CompletedAt      *time.Time     `json:"completed_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Transaction) TableName() string {
	return "transactions"
}

func (t *Transaction) BeforeSave(tx *gorm.DB) error {
	t.Currency = currency.NormalizeCode(t.Currency)
	if !currency.IsValidCode(t.Currency) || !currency.IsCatalogCode(t.Currency) {
		return errors.New("payment transaction currency must be a supported ISO 4217 code")
	}
	if t.AmountMinor == 0 && t.Amount != 0 {
		value, err := domainmoney.FromMajorFloat(t.Amount, t.Currency)
		if err != nil {
			return err
		}
		t.AmountMinor = value.AmountMinor()
	}
	if t.AmountMinor < 0 {
		return errors.New("payment transaction amount cannot be negative")
	}
	return nil
}

// AmountMoney returns the transaction's exact settlement amount. The major
// field fallback exists only for rows created before amount_minor migration.
func (t Transaction) AmountMoney() (domainmoney.Money, error) {
	if t.AmountMinor == 0 && t.Amount != 0 {
		return domainmoney.FromMajorFloat(t.Amount, t.Currency)
	}
	return domainmoney.New(t.AmountMinor, t.Currency)
}
