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
	client payPalCheckoutClient
}

type payPalCheckoutClient interface {
	CreateOrder(context.Context, string, []paypal.PurchaseUnitRequest, *paypal.PaymentSource, *paypal.ApplicationContext) (*paypal.Order, error)
	CaptureOrder(context.Context, string, paypal.CaptureOrderRequest) (*paypal.CaptureOrderResponse, error)
	GetOrder(context.Context, string) (*paypal.Order, error)
	RefundCaptureWithPaypalRequestId(context.Context, string, paypal.RefundCaptureRequest, string) (*paypal.RefundResponse, error)
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
	if err := ValidateGatewayCurrency(g.config.Type, req.Currency); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}
	amountValue, err := FormatMajorAmount(req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}

	// 构建购买单元
	units := []paypal.PurchaseUnitRequest{
		{
			ReferenceID: req.OrderID,
			Amount: &paypal.PurchaseUnitAmount{
				Currency: req.Currency,
				Value:    amountValue,
			},
			Description: req.Description,
			CustomID:    req.OrderID,
		},
	}

	// 构建应用上下文
	appCtx := &paypal.ApplicationContext{
		Locale:             "en-US",
		UserAction:         "PAY_NOW",
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
