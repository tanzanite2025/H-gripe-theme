package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
)

// stripeGatewayImpl Stripe 支付网关完整实现
type stripeGatewayImpl struct {
	config *Config
}

// NewStripeGateway 创建Stripe支付网关实例
func NewStripeGateway(config *Config) (PaymentGateway, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 设置Stripe API密钥
	stripe.Key = config.APIKey

	return &stripeGatewayImpl{config: config}, nil
}

// CreatePayment 创建Stripe支付
func (g *stripeGatewayImpl) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// Stripe金额以最小货币单位计算（美分）
	amount := int64(req.Amount * 100)

	// 构建支付意图参数
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(req.Currency),
		Description: stripe.String(req.Description),
	}

	// 设置客户信息
	if req.Customer != nil {
		params.ReceiptEmail = stripe.String(req.Customer.Email)

		// 如果有客户ID，使用它
		if req.Customer.ID != "" {
			params.Customer = stripe.String(req.Customer.ID)
		}
	}

	// 设置元数据
	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}

	// 添加订单ID到元数据
	if params.Metadata == nil {
		params.Metadata = make(map[string]string)
	}
	params.Metadata["order_id"] = req.OrderID

	// 设置自动确认
	params.Confirm = stripe.Bool(false)

	// 如果提供了返回URL，设置支付方法选项
	if req.ReturnURL != "" {
		params.PaymentMethodOptions = &stripe.PaymentIntentPaymentMethodOptionsParams{
			Card: &stripe.PaymentIntentPaymentMethodOptionsCardParams{
				SetupFutureUsage: stripe.String("off_session"),
			},
		}
	}

	// 创建支付意图
	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe payment intent: %w", err)
	}

	// 构建支付URL（如果需要）
	paymentURL := ""
	if pi.ClientSecret != "" {
		// 客户端需要使用client_secret来完成支付
		// 在实际应用中，你可能需要构建一个自定义的结账页面URL
		if req.ReturnURL != "" {
			paymentURL = fmt.Sprintf("%s?payment_intent=%s&payment_intent_client_secret=%s",
				req.ReturnURL, pi.ID, pi.ClientSecret)
		}
	}

	// 返回响应
	return &PaymentResponse{
		ID:            pi.ID,
		Status:        string(pi.Status),
		Amount:        float64(pi.Amount) / 100,
		Currency:      string(pi.Currency),
		PaymentURL:    paymentURL,
		TransactionID: pi.ID,
		CreatedAt:     time.Unix(pi.Created, 0),
		Metadata:      pi.Metadata,
	}, nil
}

// CapturePayment 捕获Stripe支付
func (g *stripeGatewayImpl) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 捕获支付意图
	params := &stripe.PaymentIntentCaptureParams{}
	pi, err := paymentintent.Capture(paymentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to capture stripe payment: %w", err)
	}

	return &PaymentResponse{
		ID:            pi.ID,
		Status:        string(pi.Status),
		Amount:        float64(pi.Amount) / 100,
		Currency:      string(pi.Currency),
		TransactionID: pi.ID,
		CreatedAt:     time.Unix(pi.Created, 0),
		Metadata:      pi.Metadata,
	}, nil
}

// RefundPayment 退款Stripe支付
func (g *stripeGatewayImpl) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 构建退款参数
	refundAmount := int64(amount * 100)
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentID),
	}

	// 如果指定了金额，设置部分退款
	if amount > 0 {
		params.Amount = stripe.Int64(refundAmount)
	}

	// 创建退款
	r, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe refund: %w", err)
	}

	return &RefundResponse{
		ID:        r.ID,
		PaymentID: paymentID,
		Amount:    float64(r.Amount) / 100,
		Status:    string(r.Status),
		CreatedAt: time.Unix(r.Created, 0),
	}, nil
}

// GetPayment 查询Stripe支付
func (g *stripeGatewayImpl) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 获取支付意图
	pi, err := paymentintent.Get(paymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe payment: %w", err)
	}

	return &PaymentResponse{
		ID:            pi.ID,
		Status:        string(pi.Status),
		Amount:        float64(pi.Amount) / 100,
		Currency:      string(pi.Currency),
		TransactionID: pi.ID,
		CreatedAt:     time.Unix(pi.Created, 0),
		Metadata:      pi.Metadata,
	}, nil
}
