package payment

import (
	paymentdomain "commerce-platform/internal/domain/payment"
	"time"
)

type paymentMethodResponse struct {
	ID                uint      `json:"id"`
	Name              string    `json:"name"`
	Code              string    `json:"code"`
	Provider          string    `json:"provider,omitempty"`
	Icon              string    `json:"icon"`
	Description       string    `json:"description"`
	FeeType           string    `json:"fee_type"`
	FeeValue          float64   `json:"fee_value"`
	MinAmount         float64   `json:"min_amount"`
	MaxAmount         float64   `json:"max_amount"`
	Enabled           bool      `json:"enabled"`
	Available         bool      `json:"available"`
	UnavailableReason string    `json:"unavailable_reason,omitempty"`
	SortOrder         int       `json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type transactionResponse struct {
	ID               uint       `json:"id"`
	OrderID          uint       `json:"order_id"`
	TransactionID    string     `json:"transaction_id"`
	PaymentMethod    string     `json:"payment_method"`
	AmountMinor      int64      `json:"amount_minor"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	LiabilityShifted *bool      `json:"liability_shifted,omitempty"`
	ErrorMessage     string     `json:"error_message,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CompletedAt      *time.Time `json:"completed_at"`
}

type refundResponse struct {
	ID                    uint                     `json:"id"`
	OrderID               uint                     `json:"order_id"`
	TransactionID         uint                     `json:"transaction_id"`
	RefundID              *string                  `json:"refund_id,omitempty"`
	AmountMinor           int64                    `json:"amount_minor"`
	GiftCardAmountMinor   int64                    `json:"gift_card_refund_amount_minor"`
	RequestedAmountMinor  int64                    `json:"requested_amount_minor"`
	DiscountClawbackMinor int64                    `json:"discount_clawback_amount_minor"`
	Currency              string                   `json:"currency"`
	LineItems             []refundLineItemResponse `json:"line_items,omitempty"`
	Reason                string                   `json:"reason"`
	Status                string                   `json:"status"`
	CreatedAt             time.Time                `json:"created_at"`
	UpdatedAt             time.Time                `json:"updated_at"`
	CompletedAt           *time.Time               `json:"completed_at"`
}

type refundLineItemResponse struct {
	ID                uint   `json:"id"`
	OrderItemID       uint   `json:"order_item_id"`
	ProductID         uint   `json:"product_id"`
	VariantID         *uint  `json:"variant_id,omitempty"`
	ProductName       string `json:"product_name"`
	SKU               string `json:"sku"`
	Quantity          int    `json:"quantity"`
	Currency          string `json:"currency"`
	UnitPriceMinor    int64  `json:"unit_price_minor"`
	LineSubtotalMinor int64  `json:"line_subtotal_minor"`
	LineTaxMinor      int64  `json:"line_tax_minor"`
	LineDiscountMinor int64  `json:"line_discount_minor"`
	LineTotalMinor    int64  `json:"line_total_minor"`
	Restock           bool   `json:"restock"`
}

func paymentMethodToResponse(method paymentdomain.PaymentMethod) paymentMethodResponse {
	return paymentMethodResponse{
		ID:          method.ID,
		Name:        method.Name,
		Code:        method.Code,
		Provider:    paymentMethodProvider(method.Code),
		Icon:        method.Icon,
		Description: method.Description,
		FeeType:     method.FeeType,
		FeeValue:    method.FeeValue,
		MinAmount:   method.MinAmount,
		MaxAmount:   method.MaxAmount,
		Enabled:     method.Enabled,
		Available:   method.Enabled,
		SortOrder:   method.SortOrder,
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

func transactionToResponse(transaction paymentdomain.Transaction) transactionResponse {
	amountMinor := transaction.AmountMinor
	if amountMinor == 0 && transaction.Amount != 0 {
		if amount, err := transaction.AmountMoney(); err == nil {
			amountMinor = amount.AmountMinor()
		}
	}
	return transactionResponse{
		ID:               transaction.ID,
		OrderID:          transaction.OrderID,
		TransactionID:    transaction.TransactionID,
		PaymentMethod:    transaction.PaymentMethod,
		AmountMinor:      amountMinor,
		Currency:         transaction.Currency,
		Status:           transaction.Status,
		LiabilityShifted: transaction.LiabilityShifted,
		ErrorMessage:     transaction.ErrorMessage,
		CreatedAt:        transaction.CreatedAt,
		UpdatedAt:        transaction.UpdatedAt,
		CompletedAt:      transaction.CompletedAt,
	}
}

func transactionsToResponse(transactions []paymentdomain.Transaction) []transactionResponse {
	items := make([]transactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		items = append(items, transactionToResponse(transaction))
	}
	return items
}

func refundToResponse(refund paymentdomain.Refund) refundResponse {
	amountMinor := refund.AmountMinor
	requestedMinor := refund.RequestedAmountMinor
	discountMinor := refund.DiscountClawbackAmountMinor
	if amountMinor == 0 && refund.Amount != 0 {
		if amount, err := refund.AmountMoney(); err == nil {
			amountMinor = amount.AmountMinor()
		}
	}
	if requestedMinor == 0 && refund.RequestedAmount != 0 {
		if amount, err := refund.RequestedAmountMoney(); err == nil {
			requestedMinor = amount.AmountMinor()
		}
	}
	return refundResponse{
		ID:                    refund.ID,
		OrderID:               refund.OrderID,
		TransactionID:         refund.TransactionID,
		RefundID:              refund.RefundID,
		AmountMinor:           amountMinor,
		GiftCardAmountMinor:   refund.GiftCardRefundAmountMinor,
		RequestedAmountMinor:  requestedMinor,
		DiscountClawbackMinor: discountMinor,
		Currency:              refund.Currency,
		LineItems:             refundLineItemsToResponse(refund.LineItems),
		Reason:                refund.Reason,
		Status:                refund.Status,
		CreatedAt:             refund.CreatedAt,
		UpdatedAt:             refund.UpdatedAt,
		CompletedAt:           refund.CompletedAt,
	}
}

func refundLineItemsToResponse(lineItems []paymentdomain.RefundLineItem) []refundLineItemResponse {
	if len(lineItems) == 0 {
		return nil
	}
	items := make([]refundLineItemResponse, 0, len(lineItems))
	for _, item := range lineItems {
		unitPriceMinor := item.UnitPriceMinor
		if unitPriceMinor == 0 && item.UnitPrice != 0 {
			if value, err := item.UnitPriceMoney(); err == nil {
				unitPriceMinor = value.AmountMinor()
			}
		}
		subtotalMinor := item.LineSubtotalMinor
		if subtotalMinor == 0 && item.LineSubtotalAmount != 0 {
			if value, err := item.LineSubtotalMoney(); err == nil {
				subtotalMinor = value.AmountMinor()
			}
		}
		taxMinor := item.LineTaxMinor
		if taxMinor == 0 && item.LineTaxAmount != 0 {
			if value, err := item.LineTaxMoney(); err == nil {
				taxMinor = value.AmountMinor()
			}
		}
		discountMinor := item.LineDiscountMinor
		if discountMinor == 0 && item.LineDiscountAmount != 0 {
			if value, err := item.LineDiscountMoney(); err == nil {
				discountMinor = value.AmountMinor()
			}
		}
		totalMinor := item.LineTotalMinor
		if totalMinor == 0 && item.LineTotalAmount != 0 {
			if value, err := item.LineTotalMoney(); err == nil {
				totalMinor = value.AmountMinor()
			}
		}
		items = append(items, refundLineItemResponse{
			ID:                item.ID,
			OrderItemID:       item.OrderItemID,
			ProductID:         item.ProductID,
			VariantID:         item.VariantID,
			ProductName:       item.ProductName,
			SKU:               item.SKU,
			Quantity:          item.Quantity,
			Currency:          item.Currency,
			UnitPriceMinor:    unitPriceMinor,
			LineSubtotalMinor: subtotalMinor,
			LineTaxMinor:      taxMinor,
			LineDiscountMinor: discountMinor,
			LineTotalMinor:    totalMinor,
			Restock:           item.Restock,
		})
	}
	return items
}

func refundsToResponse(refunds []paymentdomain.Refund) []refundResponse {
	items := make([]refundResponse, 0, len(refunds))
	for _, refund := range refunds {
		items = append(items, refundToResponse(refund))
	}
	return items
}
