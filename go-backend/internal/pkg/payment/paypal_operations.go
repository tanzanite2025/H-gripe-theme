package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/plutov/paypal/v4"
)

// CapturePayment 捕获PayPal支付
func (g *paypalGatewayImpl) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 捕获订单
	captureReq := paypal.CaptureOrderRequest{}
	capturedOrder, err := g.client.CaptureOrder(ctx, paymentID, captureReq)
	if err != nil {
		return nil, fmt.Errorf("failed to capture paypal order: %w", err)
	}

	// 提取金额和货币
	var amount float64
	var currency string
	transactionID := capturedOrder.ID
	metadata := map[string]string{
		"paypal_order_id": capturedOrder.ID,
	}
	if len(capturedOrder.PurchaseUnits) > 0 {
		pu := capturedOrder.PurchaseUnits[0]
		if orderID := firstPayPalNonBlank(pu.ReferenceID); orderID != "" {
			metadata["order_id"] = orderID
		}
		if pu.Payments != nil && len(pu.Payments.Captures) > 0 {
			capture := pu.Payments.Captures[0]
			if strings.TrimSpace(capture.ID) != "" {
				transactionID = strings.TrimSpace(capture.ID)
				metadata["paypal_capture_id"] = transactionID
			}
			if orderID := firstPayPalNonBlank(capture.CustomID, metadata["order_id"]); orderID != "" {
				metadata["order_id"] = orderID
			}
			if capture.Amount != nil {
				amount, err = parsePaymentAmount("paypal capture amount", capture.Amount.Value)
				if err != nil {
					return nil, err
				}
				currency = capture.Amount.Currency
			}
		}
	}

	return &PaymentResponse{
		ID:            capturedOrder.ID,
		Status:        capturedOrder.Status,
		Amount:        amount,
		Currency:      currency,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
		Metadata:      metadata,
	}, nil
}

// RefundPayment 退款PayPal支付
func (g *paypalGatewayImpl) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	return g.RefundPaymentWithOptions(ctx, paymentID, amount, RefundOptions{})
}

func (g *paypalGatewayImpl) RefundPaymentWithOptions(ctx context.Context, paymentID string, amount float64, options RefundOptions) (*RefundResponse, error) {
	paymentID = strings.TrimSpace(paymentID)
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	directRefund, directErr := g.refundPayPalCapture(ctx, paymentID, paymentID, amount, options, "")
	if directErr == nil {
		return directRefund, nil
	}

	// The stored transaction id is normally a PayPal capture id. If an older row
	// contains the PayPal order id instead, fall back to resolving the capture.
	order, err := g.client.GetOrder(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to refund paypal capture directly: %v; failed to get paypal order: %w", directErr, err)
	}
	captureID, captureCurrency, ok := firstPayPalCaptureFromOrder(order)
	if !ok {
		return nil, fmt.Errorf("no captures found for paypal order")
	}
	refundResponse, err := g.refundPayPalCapture(ctx, paymentID, captureID, amount, options, captureCurrency)
	if err != nil {
		return nil, err
	}
	return refundResponse, nil
}

func (g *paypalGatewayImpl) refundPayPalCapture(ctx context.Context, paymentReference string, captureID string, amount float64, options RefundOptions, fallbackCurrency string) (*RefundResponse, error) {
	captureID = strings.TrimSpace(captureID)
	if captureID == "" {
		return nil, fmt.Errorf("paypal capture id is required")
	}

	// 构建退款请求
	refundReq := paypal.RefundCaptureRequest{}
	if amount > 0 {
		currency := firstPayPalNonBlank(options.Currency, fallbackCurrency)
		if currency == "" {
			return nil, fmt.Errorf("paypal refund currency is required for partial refunds")
		}
		refundAmount, err := FormatMajorAmount(amount, currency)
		if err != nil {
			return nil, err
		}
		refundReq.Amount = &paypal.Money{
			Currency: strings.ToUpper(strings.TrimSpace(currency)),
			Value:    refundAmount,
		}
	}

	// 执行退款
	refundResp, err := g.client.RefundCaptureWithPaypalRequestId(ctx, captureID, refundReq, options.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to refund paypal capture: %w", err)
	}

	// 解析退款金额
	var refundAmount float64
	if refundResp.Amount != nil {
		refundAmount, err = parsePaymentAmount("paypal refund amount", refundResp.Amount.Value)
		if err != nil {
			return nil, err
		}
	} else {
		refundAmount = amount
	}

	return &RefundResponse{
		ID:        refundResp.ID,
		PaymentID: paymentReference,
		Amount:    refundAmount,
		Status:    refundResp.Status,
		CreatedAt: time.Now(),
	}, nil
}

func firstPayPalCaptureFromOrder(order *paypal.Order) (string, string, bool) {
	if order == nil {
		return "", "", false
	}
	for _, unit := range order.PurchaseUnits {
		if unit.Payments == nil {
			continue
		}
		for _, capture := range unit.Payments.Captures {
			captureID := strings.TrimSpace(capture.ID)
			if captureID == "" {
				continue
			}
			currency := ""
			if capture.Amount != nil {
				currency = capture.Amount.Currency
			}
			return captureID, currency, true
		}
	}
	return "", "", false
}

// GetPayment 查询PayPal支付
func (g *paypalGatewayImpl) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 获取订单详情
	order, err := g.client.GetOrder(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get paypal order: %w", err)
	}

	// 提取金额和货币
	var amount float64
	var currency string
	if len(order.PurchaseUnits) > 0 {
		if order.PurchaseUnits[0].Amount != nil {
			amount, err = parsePaymentAmount("paypal order amount", order.PurchaseUnits[0].Amount.Value)
			if err != nil {
				return nil, err
			}
			currency = order.PurchaseUnits[0].Amount.Currency
		}
	}

	return &PaymentResponse{
		ID:            order.ID,
		Status:        order.Status,
		Amount:        amount,
		Currency:      currency,
		TransactionID: order.ID,
		CreatedAt:     time.Now(),
	}, nil
}

func firstPayPalNonBlank(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
