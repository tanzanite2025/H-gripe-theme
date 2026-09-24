package admin

import (
	paymentdomain "commerce-platform/internal/domain/payment"
	"errors"
	"strings"
	"time"
)

type paymentMethodRequest struct {
	Name           string `json:"name" binding:"required"`
	Code           string `json:"code" binding:"required"`
	Icon           string `json:"icon"`
	Description    string `json:"description"`
	FeeType        string `json:"fee_type"`
	FeeValueMinor  int64  `json:"fee_value_minor"`
	FeeRateDecimal string `json:"fee_rate_decimal"`
	MinAmountMinor int64  `json:"min_amount_minor"`
	MaxAmountMinor int64  `json:"max_amount_minor"`
	Enabled        *bool  `json:"enabled"`
	SortOrder      int    `json:"sort_order"`
	Settings       string `json:"settings"`
}

type paymentMethodResponse struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	Icon           string    `json:"icon"`
	Description    string    `json:"description"`
	FeeType        string    `json:"fee_type"`
	FeeValueMinor  int64     `json:"fee_value_minor"`
	FeeRateDecimal string    `json:"fee_rate_decimal"`
	MinAmountMinor int64     `json:"min_amount_minor"`
	MaxAmountMinor int64     `json:"max_amount_minor"`
	Enabled        bool      `json:"enabled"`
	SortOrder      int       `json:"sort_order"`
	Settings       string    `json:"settings"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (r paymentMethodRequest) toDomain() (paymentdomain.PaymentMethod, error) {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}

	feeType := strings.ToLower(strings.TrimSpace(r.FeeType))
	if feeType == "" {
		feeType = "fixed"
	}

	return paymentdomain.PaymentMethod{
		Name:           strings.TrimSpace(r.Name),
		Code:           strings.ToLower(strings.TrimSpace(r.Code)),
		Icon:           strings.TrimSpace(r.Icon),
		Description:    strings.TrimSpace(r.Description),
		FeeType:        feeType,
		FeeValueMinor:  r.FeeValueMinor,
		FeeRateDecimal: r.FeeRateDecimal,
		MinAmountMinor: r.MinAmountMinor,
		MaxAmountMinor: r.MaxAmountMinor,
		Enabled:        enabled,
		SortOrder:      r.SortOrder,
		Settings:       strings.TrimSpace(r.Settings),
	}, nil
}

func paymentMethodToResponse(method paymentdomain.PaymentMethod) paymentMethodResponse {
	return paymentMethodResponse{
		ID:             method.ID,
		Name:           method.Name,
		Code:           method.Code,
		Icon:           method.Icon,
		Description:    method.Description,
		FeeType:        method.FeeType,
		FeeValueMinor:  method.FeeValueMinor,
		FeeRateDecimal: method.FeeRateDecimal,
		MinAmountMinor: method.MinAmountMinor,
		MaxAmountMinor: method.MaxAmountMinor,
		Enabled:        method.Enabled,
		SortOrder:      method.SortOrder,
		Settings:       method.Settings,
		CreatedAt:      method.CreatedAt,
		UpdatedAt:      method.UpdatedAt,
	}
}

func paymentMethodsToResponse(methods []paymentdomain.PaymentMethod) []paymentMethodResponse {
	items := make([]paymentMethodResponse, 0, len(methods))
	for _, method := range methods {
		items = append(items, paymentMethodToResponse(method))
	}
	return items
}

func validatePaymentMethod(method paymentdomain.PaymentMethod) error {
	if method.Name == "" {
		return errors.New("payment method name is required")
	}
	if method.Code == "" {
		return errors.New("payment method code is required")
	}
	if method.FeeType != "fixed" && method.FeeType != "percentage" {
		return errors.New("fee type must be fixed or percentage")
	}
	if method.FeeValueMinor < 0 || method.MinAmountMinor < 0 || method.MaxAmountMinor < 0 {
		return errors.New("payment method numeric fields cannot be negative")
	}
	if method.MaxAmountMinor > 0 && method.MinAmountMinor > method.MaxAmountMinor {
		return errors.New("minimum amount cannot be greater than maximum amount")
	}
	if method.FeeType == "percentage" {
		if _, err := method.FeeRate(); err != nil {
			return err
		}
	}

	return nil
}
