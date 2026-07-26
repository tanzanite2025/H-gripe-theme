package admin

import (
	"errors"
	"regexp"
	"strings"
	paymentdomain "tanzanite/internal/domain/payment"
)

var paymentCurrencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)

type paymentMethodRequest struct {
	Name                string  `json:"name" binding:"required"`
	Code                string  `json:"code" binding:"required"`
	Icon                string  `json:"icon"`
	Description         string  `json:"description"`
	FeeType             string  `json:"fee_type"`
	FeeValue            float64 `json:"fee_value"`
	MinAmount           float64 `json:"min_amount"`
	MaxAmount           float64 `json:"max_amount"`
	SupportedCurrencies string  `json:"supported_currencies"`
	Enabled             *bool   `json:"enabled"`
	SortOrder           int     `json:"sort_order"`
	Settings            string  `json:"settings"`
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

	currencies, err := normalizePaymentCurrencies(r.SupportedCurrencies)
	if err != nil {
		return paymentdomain.PaymentMethod{}, err
	}

	return paymentdomain.PaymentMethod{
		Name:                strings.TrimSpace(r.Name),
		Code:                strings.ToLower(strings.TrimSpace(r.Code)),
		Icon:                strings.TrimSpace(r.Icon),
		Description:         strings.TrimSpace(r.Description),
		FeeType:             feeType,
		FeeValue:            r.FeeValue,
		MinAmount:           r.MinAmount,
		MaxAmount:           r.MaxAmount,
		SupportedCurrencies: strings.Join(currencies, ","),
		Enabled:             enabled,
		SortOrder:           r.SortOrder,
		Settings:            strings.TrimSpace(r.Settings),
	}, nil
}

func normalizePaymentCurrencies(input string) ([]string, error) {
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ';' || r == '，' || r == '；' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})

	seen := make(map[string]bool, len(parts))
	currencies := make([]string, 0, len(parts))
	for _, part := range parts {
		code := strings.ToUpper(strings.TrimSpace(part))
		if code == "" {
			continue
		}
		if !paymentCurrencyCodePattern.MatchString(code) {
			return nil, errors.New("supported currencies must be 3-letter currency codes")
		}
		if seen[code] {
			continue
		}
		seen[code] = true
		currencies = append(currencies, code)
	}

	return currencies, nil
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
	if method.FeeValue < 0 || method.MinAmount < 0 || method.MaxAmount < 0 {
		return errors.New("payment method numeric fields cannot be negative")
	}
	if method.MaxAmount > 0 && method.MinAmount > method.MaxAmount {
		return errors.New("minimum amount cannot be greater than maximum amount")
	}

	return nil
}
