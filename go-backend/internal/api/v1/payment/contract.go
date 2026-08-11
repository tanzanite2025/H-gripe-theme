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
	ID            uint       `json:"id"`
	OrderID       uint       `json:"order_id"`
	TransactionID string     `json:"transaction_id"`
	PaymentMethod string     `json:"payment_method"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type refundResponse struct {
	ID                     uint                     `json:"id"`
	OrderID                uint                     `json:"order_id"`
	TransactionID          uint                     `json:"transaction_id"`
	RefundID               *string                  `json:"refund_id,omitempty"`
	Amount                 float64                  `json:"amount"`
	RequestedAmount        float64                  `json:"requested_amount"`
	DiscountClawbackAmount float64                  `json:"discount_clawback_amount"`
	LineItems              []refundLineItemResponse `json:"line_items,omitempty"`
	Reason                 string                   `json:"reason"`
	Status                 string                   `json:"status"`
	CreatedAt              time.Time                `json:"created_at"`
	UpdatedAt              time.Time                `json:"updated_at"`
	CompletedAt            *time.Time               `json:"completed_at"`
}

type refundLineItemResponse struct {
	ID                 uint    `json:"id"`
	OrderItemID        uint    `json:"order_item_id"`
	ProductID          uint    `json:"product_id"`
	VariantID          *uint   `json:"variant_id,omitempty"`
	ProductName        string  `json:"product_name"`
	SKU                string  `json:"sku"`
	Quantity           int     `json:"quantity"`
	UnitPrice          float64 `json:"unit_price"`
	LineSubtotalAmount float64 `json:"line_subtotal_amount"`
	LineTaxAmount      float64 `json:"line_tax_amount"`
	LineDiscountAmount float64 `json:"line_discount_amount"`
	LineTotalAmount    float64 `json:"line_total_amount"`
	Restock            bool    `json:"restock"`
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
	return transactionResponse{
		ID:            transaction.ID,
		OrderID:       transaction.OrderID,
		TransactionID: transaction.TransactionID,
		PaymentMethod: transaction.PaymentMethod,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		Status:        transaction.Status,
		ErrorMessage:  transaction.ErrorMessage,
		CreatedAt:     transaction.CreatedAt,
		UpdatedAt:     transaction.UpdatedAt,
		CompletedAt:   transaction.CompletedAt,
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
	return refundResponse{
		ID:                     refund.ID,
		OrderID:                refund.OrderID,
		TransactionID:          refund.TransactionID,
		RefundID:               refund.RefundID,
		Amount:                 refund.Amount,
		RequestedAmount:        refund.RequestedAmount,
		DiscountClawbackAmount: refund.DiscountClawbackAmount,
		LineItems:              refundLineItemsToResponse(refund.LineItems),
		Reason:                 refund.Reason,
		Status:                 refund.Status,
		CreatedAt:              refund.CreatedAt,
		UpdatedAt:              refund.UpdatedAt,
		CompletedAt:            refund.CompletedAt,
	}
}

func refundLineItemsToResponse(lineItems []paymentdomain.RefundLineItem) []refundLineItemResponse {
	if len(lineItems) == 0 {
		return nil
	}
	items := make([]refundLineItemResponse, 0, len(lineItems))
	for _, item := range lineItems {
		items = append(items, refundLineItemResponse{
			ID:                 item.ID,
			OrderItemID:        item.OrderItemID,
			ProductID:          item.ProductID,
			VariantID:          item.VariantID,
			ProductName:        item.ProductName,
			SKU:                item.SKU,
			Quantity:           item.Quantity,
			UnitPrice:          item.UnitPrice,
			LineSubtotalAmount: item.LineSubtotalAmount,
			LineTaxAmount:      item.LineTaxAmount,
			LineDiscountAmount: item.LineDiscountAmount,
			LineTotalAmount:    item.LineTotalAmount,
			Restock:            item.Restock,
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
