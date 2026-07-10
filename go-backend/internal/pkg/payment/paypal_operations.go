package payment

import (
	"context"
	"fmt"
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
	if len(capturedOrder.PurchaseUnits) > 0 {
		pu := capturedOrder.PurchaseUnits[0]
		if pu.Payments != nil && len(pu.Payments.Captures) > 0 {
			capture := pu.Payments.Captures[0]
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
		TransactionID: capturedOrder.ID,
		CreatedAt:     time.Now(),
	}, nil
}

// RefundPayment 退款PayPal支付
func (g *paypalGatewayImpl) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 首先获取订单详情以获取capture ID
	order, err := g.client.GetOrder(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get paypal order: %w", err)
	}

	if len(order.PurchaseUnits) == 0 || order.PurchaseUnits[0].Payments == nil {
		return nil, fmt.Errorf("no payment captures found for order")
	}

	captures := order.PurchaseUnits[0].Payments.Captures
	if len(captures) == 0 {
		return nil, fmt.Errorf("no captures found for order")
	}

	// 使用第一个capture进行退款
	captureID := captures[0].ID

	// 构建退款请求
	refundReq := paypal.RefundCaptureRequest{}
	if amount > 0 {
		refundReq.Amount = &paypal.Money{
			Currency: captures[0].Amount.Currency,
			Value:    fmt.Sprintf("%.2f", amount),
		}
	}
	// 如果amount为0，则全额退款（不设置Amount字段）

	// 执行退款
	refundResp, err := g.client.RefundCapture(ctx, captureID, refundReq)
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
	}

	return &RefundResponse{
		ID:        refundResp.ID,
		PaymentID: paymentID,
		Amount:    refundAmount,
		Status:    refundResp.Status,
		CreatedAt: time.Now(),
	}, nil
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
