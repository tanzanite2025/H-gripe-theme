package admin

import (
	paymentdomain "commerce-platform/internal/domain/payment"
	"errors"
	"strings"
	"time"
)

type paymentMethodRequest struct {
	Name        string  `json:"name" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Icon        string  `json:"icon"`
	Description string  `json:"description"`
	FeeType     string  `json:"fee_type"`
	FeeValue    float64 `json:"fee_value"`
	MinAmount   float64 `json:"min_amount"`
	MaxAmount   float64 `json:"max_amount"`
	Enabled     *bool   `json:"enabled"`
	SortOrder   int     `json:"sort_order"`
	Settings    string  `json:"settings"`
}

type paymentMethodResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	FeeType     string    `json:"fee_type"`
	FeeValue    float64   `json:"fee_value"`
	MinAmount   float64   `json:"min_amount"`
	MaxAmount   float64   `json:"max_amount"`
	Enabled     bool      `json:"enabled"`
	SortOrder   int       `json:"sort_order"`
	Settings    string    `json:"settings"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
		Name:        strings.TrimSpace(r.Name),
		Code:        strings.ToLower(strings.TrimSpace(r.Code)),
		Icon:        strings.TrimSpace(r.Icon),
		Description: strings.TrimSpace(r.Description),
		FeeType:     feeType,
		FeeValue:    r.FeeValue,
		MinAmount:   r.MinAmount,
		MaxAmount:   r.MaxAmount,
		Enabled:     enabled,
		SortOrder:   r.SortOrder,
		Settings:    strings.TrimSpace(r.Settings),
	}, nil
}

func paymentMethodToResponse(method paymentdomain.PaymentMethod) paymentMethodResponse {
	return paymentMethodResponse{
		ID:          method.ID,
		Name:        method.Name,
		Code:        method.Code,
		Icon:        method.Icon,
		Description: method.Description,
		FeeType:     method.FeeType,
		FeeValue:    method.FeeValue,
		MinAmount:   method.MinAmount,
		MaxAmount:   method.MaxAmount,
		Enabled:     method.Enabled,
		SortOrder:   method.SortOrder,
		Settings:    method.Settings,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
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
	if method.FeeValue < 0 || method.MinAmount < 0 || method.MaxAmount < 0 {
		return errors.New("payment method numeric fields cannot be negative")
	}
	if method.MaxAmount > 0 && method.MinAmount > method.MaxAmount {
		return errors.New("minimum amount cannot be greater than maximum amount")
	}

	return nil
}
