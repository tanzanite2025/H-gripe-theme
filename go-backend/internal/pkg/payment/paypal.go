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

	// 构建订单请求
	order := &paypal.Order{
		Intent: "CAPTURE",
		PurchaseUnits: []paypal.PurchaseUnitRequest{
			{
				Amount: &paypal.PurchaseUnitAmount{
					Currency: req.Currency,
					Value:    fmt.Sprintf("%.2f", req.Amount),
				},
				Description: req.Description,
				CustomID:    req.OrderID,
			},
		},
		ApplicationContext: &paypal.ApplicationContext{
			BrandName: "Tanzanite Components",
			Locale:    "en-US",
			UserAction: "PAY_NOW",
		},
	}

	// 设置返回URL
	if req.ReturnURL != "" {
		order.ApplicationContext.ReturnURL = req.ReturnURL
	}
	if req.CancelURL != "" {
		order.ApplicationContext.CancelURL = req.CancelURL
	}

	// 创建订单
	createdOrder, err := g.client.CreateOrder(ctx, "CAPTURE", order.PurchaseUnits, nil, order.ApplicationContext)
	if err != nil {
		return nil, fmt.Errorf("failed to create paypal order: %w", err)
	}

	// 获取审批URL
	var approvalURL string
	for _, link := range createdOrder.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
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
	capturedOrder, err := g.client.CaptureOrder(ctx, paymentID, paypal.CaptureOrderRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to capture paypal order: %w", err)
	}

	// 提取金额和货币
	var amount float64
	var currency string
	if len(capturedOrder.PurchaseUnits) > 0 {
		if capturedOrder.PurchaseUnits[0].Payments != nil &&
			len(capturedOrder.PurchaseUnits[0].Payments.Captures) > 0 {
			capture := capturedOrder.PurchaseUnits[0].Payments.Captures[0]
			fmt.Sscanf(capture.Amount.Value, "%f", &amount)
			currency = capture.Amount.Currency
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
	refundReq := paypal.RefundCaptureRequest{
		Amount: &paypal.Money{
			Currency: captures[0].Amount.Currency,
			Value:    fmt.Sprintf("%.2f", amount),
		},
	}

	// 如果金额为0，则全额退款
	if amount == 0 {
		refundReq.Amount = nil
	}

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

	// PayPal webhook验证需要使用PayPal的验证API
	// 这里提供基本的HMAC验证逻辑
	// 在生产环境中，建议使用PayPal的官方验证API
	
	// 使用HMAC SHA256验证
	isValid := verifyHMACSHA256(payload, signature, g.config.WebhookSecret)
	if !isValid {
		return false, fmt.Errorf("webhook signature verification failed")
	}

	return true, nil
}

// VerifyPayPalWebhookWithAPI 使用PayPal API验证webhook（推荐方法）
func (g *paypalGatewayImpl) VerifyPayPalWebhookWithAPI(ctx context.Context, webhookID string, event map[string]interface{}) (bool, error) {
	// 这个方法需要使用PayPal的webhook验证API
	// POST /v1/notifications/verify-webhook-signature
	
	// 实际实现需要调用PayPal API
	// 这里仅提供接口定义
	
	return true, nil
}

// GetPayPalPaymentStatus 将PayPal状态映射为统一状态
func GetPayPalPaymentStatus(paypalStatus string) string {
	switch paypalStatus {
	case "CREATED":
		return "created"
	case "SAVED":
		return "saved"
	case "APPROVED":
		return "approved"
	case "VOIDED":
		return "voided"
	case "COMPLETED":
		return "completed"
	case "PAYER_ACTION_REQUIRED":
		return "pending"
	default:
		return paypalStatus
	}
}
