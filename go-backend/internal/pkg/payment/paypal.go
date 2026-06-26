package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/plutov/paypal/v4"
)

// paypalGatewayImpl PayPal 支付网关完整实现
type paypalGatewayImpl struct {
	config *Config
	client *paypal.Client
}

// NewPayPalGateway 创建PayPal支付网关实例
func NewPayPalGateway(config *Config) (PaymentGateway, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 确定环境
	var apiBase string
	if config.Environment == "production" {
		apiBase = paypal.APIBaseLive
	} else {
		apiBase = paypal.APIBaseSandBox
	}

	// 创建PayPal客户端
	client, err := paypal.NewClient(config.APIKey, config.SecretKey, apiBase)
	if err != nil {
		return nil, fmt.Errorf("failed to create paypal client: %w", err)
	}

	// 获取访问令牌
	_, err = client.GetAccessToken(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get paypal access token: %w", err)
	}

	return &paypalGatewayImpl{
		config: config,
		client: client,
	}, nil
}

// CreatePayment 创建PayPal支付
func (g *paypalGatewayImpl) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// 构建购买单元
	units := []paypal.PurchaseUnitRequest{
		{
			Amount: &paypal.PurchaseUnitAmount{
				Currency: req.Currency,
				Value:    fmt.Sprintf("%.2f", req.Amount),
			},
			Description: req.Description,
			CustomID:    req.OrderID,
		},
	}

	// 构建应用上下文
	appCtx := &paypal.ApplicationContext{
		BrandName:       "Tanzanite Components",
		Locale:          "en-US",
		UserAction:      "PAY_NOW",
		ShippingPreference: "NO_SHIPPING",
	}

	// 设置返回URL
	if req.ReturnURL != "" {
		appCtx.ReturnURL = req.ReturnURL
	}
	if req.CancelURL != "" {
		appCtx.CancelURL = req.CancelURL
	}

	// 创建订单（使用SDK v4的正确API）
	createdOrder, err := g.client.CreateOrder(
		ctx,
		paypal.OrderIntentCapture,
		units,
		nil, // PaymentSource可选
		appCtx,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create paypal order: %w", err)
	}

	// 提取approval URL
	var approvalURL string
	if createdOrder.Links != nil {
		for _, link := range createdOrder.Links {
			if link.Rel == "approve" {
				approvalURL = link.Href
				break
			}
		}
	}

	// 构建元数据
	metadata := req.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["order_id"] = req.OrderID
	metadata["paypal_order_id"] = createdOrder.ID

	return &PaymentResponse{
		ID:            createdOrder.ID,
		Status:        createdOrder.Status,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentURL:    approvalURL,
		TransactionID: createdOrder.ID,
		CreatedAt:     time.Now(),
		Metadata:      metadata,
	}, nil
}

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
				fmt.Sscanf(capture.Amount.Value, "%f", &amount)
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
		fmt.Sscanf(refundResp.Amount.Value, "%f", &refundAmount)
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
			fmt.Sscanf(order.PurchaseUnits[0].Amount.Value, "%f", &amount)
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

// VerifyWebhook 验证PayPal Webhook签名
func (g *paypalGatewayImpl) VerifyWebhook(payload []byte, signature string) (bool, error) {
	if g.config.WebhookSecret == "" {
		return false, fmt.Errorf("webhook secret is not configured")
	}

	// PayPal webhook验证使用HMAC SHA256
	// 注意：在生产环境中，建议使用PayPal的官方验证API
	// POST /v1/notifications/verify-webhook-signature
	
	// 基本的HMAC验证
	isValid := verifyHMACSHA256(payload, signature, g.config.WebhookSecret)
	if !isValid {
		return false, fmt.Errorf("webhook signature verification failed")
	}

	return true, nil
}
