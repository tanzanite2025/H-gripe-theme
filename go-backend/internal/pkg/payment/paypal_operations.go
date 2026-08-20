package payment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/plutov/paypal/v4"
)

// CapturePayment 捕获PayPal支付
func (g *paypalGatewayImpl) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return g.CapturePaymentWithOptions(ctx, paymentID, CaptureOptions{})
}

func (g *paypalGatewayImpl) CapturePaymentWithOptions(ctx context.Context, paymentID string, options CaptureOptions) (*PaymentResponse, error) {
	ctx, cancel := paymentGatewayContext(ctx)
	defer cancel()

	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 捕获订单
	captureReq := paypal.CaptureOrderRequest{}
	var capturedOrder *paypal.CaptureOrderResponse
	var err error
	requestID := strings.TrimSpace(options.IdempotencyKey)
	if requestID != "" {
		if client, ok := g.client.(interface {
			CaptureOrderWithPaypalRequestId(context.Context, string, paypal.CaptureOrderRequest, string, *paypal.CaptureOrderMockResponse) (*paypal.CaptureOrderResponse, error)
		}); ok {
			capturedOrder, err = client.CaptureOrderWithPaypalRequestId(ctx, paymentID, captureReq, requestID, nil)
		} else {
			capturedOrder, err = g.client.CaptureOrder(ctx, paymentID, captureReq)
		}
	} else {
		capturedOrder, err = g.client.CaptureOrder(ctx, paymentID, captureReq)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to capture paypal order: %w", err)
	}

	// 提取金额和货币
	var amount float64
	var currency string
	transactionID := ""
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
	if strings.EqualFold(strings.TrimSpace(capturedOrder.Status), "COMPLETED") {
		if transactionID == "" {
			return nil, fmt.Errorf("paypal capture response did not include a completed capture id")
		}
		if amount <= 0 {
			return nil, fmt.Errorf("paypal capture response did not include a positive captured amount")
		}
		if strings.TrimSpace(currency) == "" {
			return nil, fmt.Errorf("paypal capture response did not include capture currency")
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

func PayPalCaptureRequestID(paymentID string) string {
	sum := sha256.Sum256([]byte("paypal:capture:" + strings.TrimSpace(paymentID)))
	return "cap-" + hex.EncodeToString(sum[:16])
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

	return g.refundPayPalCapture(ctx, paymentID, paymentID, amount, options)
}

func (g *paypalGatewayImpl) refundPayPalCapture(ctx context.Context, paymentReference string, captureID string, amount float64, options RefundOptions) (*RefundResponse, error) {
	ctx, cancel := paymentGatewayContext(ctx)
	defer cancel()

	captureID = strings.TrimSpace(captureID)
	if captureID == "" {
		return nil, fmt.Errorf("paypal capture id is required")
	}

	// 构建退款请求
	refundReq := paypal.RefundCaptureRequest{}
	if amount > 0 {
		currency := firstPayPalNonBlank(options.Currency)
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

// GetPayment 查询PayPal支付
func (g *paypalGatewayImpl) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	ctx, cancel := paymentGatewayContext(ctx)
	defer cancel()

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
