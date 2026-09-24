package payment

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PaymentRiskLevel string

const (
	PaymentRiskLevelNormal   PaymentRiskLevel = "normal"
	PaymentRiskLevelWarning  PaymentRiskLevel = "warning"
	PaymentRiskLevelCritical PaymentRiskLevel = "critical"
)

// PaymentRiskAmountMinorByCurrency is an exact, non-additive amount summary.
// Values are expressed in the smallest unit of the currency named by each
// key. Keeping currencies separate is essential: USD minor must never be
// added to JPY whole units in a portfolio report.
type PaymentRiskAmountMinorByCurrency map[string]int64

func normalizePaymentRiskAmountTotals(values PaymentRiskAmountMinorByCurrency) PaymentRiskAmountMinorByCurrency {
	if len(values) == 0 {
		return PaymentRiskAmountMinorByCurrency{}
	}
	result := make(PaymentRiskAmountMinorByCurrency, len(values))
	for code, amount := range values {
		code = currency.NormalizeCode(strings.TrimSpace(code))
		if code == "" || !currency.IsCatalogCode(code) || amount == 0 {
			continue
		}
		result[code] += amount
	}
	return result
}

func encodePaymentRiskAmountTotals(values PaymentRiskAmountMinorByCurrency) datatypes.JSON {
	values = normalizePaymentRiskAmountTotals(values)
	if len(values) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(encoded)
}

func decodePaymentRiskAmountTotals(raw datatypes.JSON) PaymentRiskAmountMinorByCurrency {
	result := PaymentRiskAmountMinorByCurrency{}
	if len(raw) == 0 {
		return result
	}
	var values map[string]int64
	if err := json.Unmarshal(raw, &values); err != nil {
		return result
	}
	return normalizePaymentRiskAmountTotals(values)
}

// PaymentRiskSnapshot is an internal rolling-window view. The rate fields are
// operational indicators, not claims about a provider's official program.
type PaymentRiskSnapshot struct {
	ID                     uint      `gorm:"primarykey" json:"id"`
	Provider               string    `gorm:"index;not null" json:"provider"`
	WindowDays             int       `gorm:"not null" json:"window_days"`
	WindowStart            time.Time `gorm:"index;not null" json:"window_start"`
	WindowEnd              time.Time `gorm:"index;not null" json:"window_end"`
	SuccessfulPaymentCount int64     `gorm:"not null;default:0" json:"successful_payment_count"`
	// SuccessfulPaymentAmountMinorByCurrency is persisted as JSON so every
	// currency remains an independent exact total. The legacy major-unit field
	// below is a read-only persistence projection and is not exposed or used by
	// the risk engine.
	SuccessfulPaymentAmountMinorByCurrency     PaymentRiskAmountMinorByCurrency `gorm:"-" json:"successful_payment_amount_minor_by_currency"`
	SuccessfulPaymentAmountMinorByCurrencyJSON datatypes.JSON                   `gorm:"column:successful_payment_amount_minor_by_currency;type:jsonb;not null;default:'{}'" json:"-"`
	DisputeCount                               int64                            `gorm:"not null;default:0" json:"dispute_count"`
	DisputeAmountMinorByCurrency               PaymentRiskAmountMinorByCurrency `gorm:"-" json:"dispute_amount_minor_by_currency"`
	DisputeAmountMinorByCurrencyJSON           datatypes.JSON                   `gorm:"column:dispute_amount_minor_by_currency;type:jsonb;not null;default:'{}'" json:"-"`
	EarlyFraudWarningCount                     int64                            `gorm:"not null;default:0" json:"early_fraud_warning_count"`
	RefundCount                                int64                            `gorm:"not null;default:0" json:"refund_count"`
	RefundAmountMinorByCurrency                PaymentRiskAmountMinorByCurrency `gorm:"-" json:"refund_amount_minor_by_currency"`
	RefundAmountMinorByCurrencyJSON            datatypes.JSON                   `gorm:"column:refund_amount_minor_by_currency;type:jsonb;not null;default:'{}'" json:"-"`
	CheckoutAttemptCount                       int64                            `gorm:"not null;default:0" json:"checkout_attempt_count"`
	ThreeDSUpgradeCount                        int64                            `gorm:"not null;default:0" json:"three_ds_upgrade_count"`
	ThreeDSChallengeCount                      int64                            `gorm:"not null;default:0" json:"three_ds_challenge_count"`
	ThreeDSExemptionCount                      int64                            `gorm:"not null;default:0" json:"three_ds_exemption_count"`
	DisputeActivityRate                        float64                          `gorm:"not null;default:0" json:"dispute_activity_rate"`
	EarlyFraudWarningRate                      float64                          `gorm:"not null;default:0" json:"early_fraud_warning_rate"`
	RefundRate                                 float64                          `gorm:"not null;default:0" json:"refund_rate"`
	ThreeDSUpgradeRate                         float64                          `gorm:"not null;default:0" json:"three_ds_upgrade_rate"`
	Level                                      PaymentRiskLevel                 `gorm:"index;not null" json:"level"`
	RecommendedAction                          string                           `gorm:"not null;default:''" json:"recommended_action"`
	ReasonsJSON                                string                           `gorm:"type:text" json:"-"`
	ComputedAt                                 time.Time                        `gorm:"index;not null" json:"computed_at"`
	CreatedAt                                  time.Time                        `json:"created_at"`
}

// BeforeSave and AfterFind keep the transport-facing maps and their compact
// database representation synchronized. This avoids any implicit float or
// cross-currency coercion in GORM.
func (s *PaymentRiskSnapshot) BeforeSave(tx *gorm.DB) error {
	for name, values := range map[string]PaymentRiskAmountMinorByCurrency{
		"successful payment": s.SuccessfulPaymentAmountMinorByCurrency,
		"dispute":            s.DisputeAmountMinorByCurrency,
		"refund":             s.RefundAmountMinorByCurrency,
	} {
		for code, amount := range values {
			if amount < 0 {
				return fmt.Errorf("payment risk %s amount cannot be negative", name)
			}
			if strings.TrimSpace(code) != "" && !currency.IsCatalogCode(code) {
				return fmt.Errorf("payment risk %s currency %q is unsupported", name, code)
			}
		}
	}
	s.SuccessfulPaymentAmountMinorByCurrency = normalizePaymentRiskAmountTotals(s.SuccessfulPaymentAmountMinorByCurrency)
	s.DisputeAmountMinorByCurrency = normalizePaymentRiskAmountTotals(s.DisputeAmountMinorByCurrency)
	s.RefundAmountMinorByCurrency = normalizePaymentRiskAmountTotals(s.RefundAmountMinorByCurrency)
	s.SuccessfulPaymentAmountMinorByCurrencyJSON = encodePaymentRiskAmountTotals(s.SuccessfulPaymentAmountMinorByCurrency)
	s.DisputeAmountMinorByCurrencyJSON = encodePaymentRiskAmountTotals(s.DisputeAmountMinorByCurrency)
	s.RefundAmountMinorByCurrencyJSON = encodePaymentRiskAmountTotals(s.RefundAmountMinorByCurrency)
	return nil
}

func (s *PaymentRiskSnapshot) AfterFind(tx *gorm.DB) error {
	s.SuccessfulPaymentAmountMinorByCurrency = decodePaymentRiskAmountTotals(s.SuccessfulPaymentAmountMinorByCurrencyJSON)
	s.DisputeAmountMinorByCurrency = decodePaymentRiskAmountTotals(s.DisputeAmountMinorByCurrencyJSON)
	s.RefundAmountMinorByCurrency = decodePaymentRiskAmountTotals(s.RefundAmountMinorByCurrencyJSON)
	return nil
}

func (PaymentRiskSnapshot) TableName() string {
	return "payment_risk_snapshots"
}
