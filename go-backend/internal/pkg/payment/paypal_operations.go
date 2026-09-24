package payment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	domainmoney "commerce-platform/internal/domain/money"

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
	requestID := strings.TrimSpace(options.IdempotencyKey)
	capturedOrder, err := retryPayPalOperation(g, ctx, func(callCtx context.Context) (*paypal.CaptureOrderResponse, error) {
		if requestID != "" {
			if client, ok := g.client.(interface {
				CaptureOrderWithPaypalRequestId(context.Context, string, paypal.CaptureOrderRequest, string, *paypal.CaptureOrderMockResponse) (*paypal.CaptureOrderResponse, error)
			}); ok {
				return client.CaptureOrderWithPaypalRequestId(callCtx, paymentID, captureReq, requestID, nil)
			}
		}
		return g.client.CaptureOrder(callCtx, paymentID, captureReq)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to capture paypal order: %w", err)
	}

	// 提取金额和货币
	amount := "0"
	var amountMinor int64
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
				captureMoney, moneyErr := domainmoney.ParseMajor(capture.Amount.Value, capture.Amount.Currency)
				if moneyErr != nil {
					return nil, fmt.Errorf("invalid paypal capture amount: %w", moneyErr)
				}
				amount, err = captureMoney.FormatMajor()
				if err != nil {
					return nil, err
				}
				amountMinor = captureMoney.AmountMinor()
				currency = capture.Amount.Currency
			}
		}
	}
	if strings.EqualFold(strings.TrimSpace(capturedOrder.Status), "COMPLETED") {
		if transactionID == "" {
			return nil, fmt.Errorf("paypal capture response did not include a completed capture id")
		}
		if amountMinor <= 0 {
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
		AmountMinor:   amountMinor,
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
func (g *paypalGatewayImpl) RefundPayment(ctx context.Context, paymentID string, amountMinor int64) (*RefundResponse, error) {
	return g.RefundPaymentWithOptions(ctx, paymentID, amountMinor, RefundOptions{})
}

func (g *paypalGatewayImpl) RefundPaymentWithOptions(ctx context.Context, paymentID string, amountMinor int64, options RefundOptions) (*RefundResponse, error) {
	paymentID = strings.TrimSpace(paymentID)
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	return g.refundPayPalCapture(ctx, paymentID, paymentID, amountMinor, options)
}

func (g *paypalGatewayImpl) refundPayPalCapture(ctx context.Context, paymentReference string, captureID string, amountMinor int64, options RefundOptions) (*RefundResponse, error) {
	ctx, cancel := paymentGatewayContext(ctx)
	defer cancel()

	captureID = strings.TrimSpace(captureID)
	if captureID == "" {
		return nil, fmt.Errorf("paypal capture id is required")
	}

	// 构建退款请求
	refundReq := paypal.RefundCaptureRequest{}
	if options.AmountMinor > 0 || amountMinor > 0 {
		currency := firstPayPalNonBlank(options.Currency)
		if currency == "" {
			return nil, fmt.Errorf("paypal refund currency is required for partial refunds")
		}
		refundMinor := amountMinor
		if options.AmountMinor > 0 {
			refundMinor = options.AmountMinor
		}
		refundMoney, err := domainmoney.New(refundMinor, currency)
		if err != nil {
			return nil, err
		}
		refundAmount, err := refundMoney.FormatMajor()
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
	refundAmount := "0"
	var refundAmountMinor int64
	if refundResp.Amount != nil {
		refundMoney, moneyErr := domainmoney.ParseMajor(refundResp.Amount.Value, refundResp.Amount.Currency)
		if moneyErr != nil {
			return nil, fmt.Errorf("invalid paypal refund amount: %w", moneyErr)
		}
		refundAmount, err = refundMoney.FormatMajor()
		if err != nil {
			return nil, err
		}
		refundAmountMinor = refundMoney.AmountMinor()
	} else {
		refundAmountMinor = amountMinor
		if options.AmountMinor > 0 {
			refundAmountMinor = options.AmountMinor
		}
		if refundAmountMinor > 0 {
			refundMoney, moneyErr := domainmoney.New(refundAmountMinor, firstPayPalNonBlank(options.Currency, "USD"))
			if moneyErr != nil {
				return nil, moneyErr
			}
			refundAmount, moneyErr = refundMoney.FormatMajor()
			if moneyErr != nil {
				return nil, moneyErr
			}
		}
	}

	return &RefundResponse{
		ID:          refundResp.ID,
		PaymentID:   paymentReference,
		Amount:      refundAmount,
		AmountMinor: refundAmountMinor,
		Status:      refundResp.Status,
		CreatedAt:   time.Now(),
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
	amount := "0"
	var amountMinor int64
	var currency string
	transactionID := ""
	metadata := map[string]string{
		"paypal_order_id": strings.TrimSpace(order.ID),
	}
	if len(order.PurchaseUnits) > 0 {
		purchaseUnit := order.PurchaseUnits[0]
		if orderID := firstPayPalNonBlank(purchaseUnit.CustomID, purchaseUnit.ReferenceID); orderID != "" {
			metadata["order_id"] = orderID
			metadata["custom_id"] = orderID
		}
		if purchaseUnit.Amount != nil {
			orderMoney, moneyErr := domainmoney.ParseMajor(purchaseUnit.Amount.Value, purchaseUnit.Amount.Currency)
			if moneyErr != nil {
				return nil, fmt.Errorf("invalid paypal order amount: %w", moneyErr)
			}
			amount, err = orderMoney.FormatMajor()
			if err != nil {
				return nil, err
			}
			amountMinor = orderMoney.AmountMinor()
			currency = purchaseUnit.Amount.Currency
		}
		if purchaseUnit.Payments != nil {
			var captured *paypal.CaptureAmount
			for index := range purchaseUnit.Payments.Captures {
				candidate := &purchaseUnit.Payments.Captures[index]
				if captured == nil || strings.EqualFold(strings.TrimSpace(candidate.Status), "COMPLETED") {
					captured = candidate
				}
				if strings.EqualFold(strings.TrimSpace(candidate.Status), "COMPLETED") {
					break
				}
			}
			if captured != nil {
				transactionID = strings.TrimSpace(captured.ID)
				if transactionID != "" {
					metadata["paypal_capture_id"] = transactionID
				}
				if orderID := strings.TrimSpace(captured.CustomID); orderID != "" {
					metadata["order_id"] = orderID
					metadata["custom_id"] = orderID
				}
				if captured.Amount != nil {
					captureMoney, moneyErr := domainmoney.ParseMajor(captured.Amount.Value, captured.Amount.Currency)
					if moneyErr != nil {
						return nil, fmt.Errorf("invalid paypal capture amount: %w", moneyErr)
					}
					amount, err = captureMoney.FormatMajor()
					if err != nil {
						return nil, err
					}
					amountMinor = captureMoney.AmountMinor()
					currency = captured.Amount.Currency
				}
			}
		}
	}

	return &PaymentResponse{
		ID:            order.ID,
		Status:        order.Status,
		Amount:        amount,
		AmountMinor:   amountMinor,
		Currency:      currency,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
		Metadata:      metadata,
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
