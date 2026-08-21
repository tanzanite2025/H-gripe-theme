package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v76"
	stripeclient "github.com/stripe/stripe-go/v76/client"
)

// stripeGatewayImpl Stripe 支付网关完整实现
type stripeGatewayImpl struct {
	client *stripeclient.API
	config *Config
}

// NewStripeGateway 创建Stripe支付网关实例
func NewStripeGateway(config *Config) (PaymentGateway, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &stripeGatewayImpl{
		client: stripeclient.New(config.APIKey, nil),
		config: config,
	}, nil
}

func (g *stripeGatewayImpl) stripeClient() (*stripeclient.API, error) {
	if g == nil || g.client == nil {
		return nil, fmt.Errorf("stripe client is not initialized")
	}
	return g.client, nil
}

// CreatePayment 创建Stripe支付
func (g *stripeGatewayImpl) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	ctx, cancel := paymentGatewayContext(ctx)
	defer cancel()

	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}
	if err := ValidateGatewayCurrency(g.config.Type, req.Currency); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	amount, err := MajorToMinorAmount(req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}
	threeDSMode := NormalizeThreeDSecureMode(req.ThreeDSecure)
	if req.ThreeDSecure == "" {
		threeDSMode = NormalizeThreeDSecureMode(g.config.ThreeDSecure)
	}

	// 构建支付意图参数
	paymentMethodTypes := g.config.PaymentMethodTypes
	if len(paymentMethodTypes) == 0 {
		paymentMethodTypes = []string{"card"}
	}
	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(amount),
		Currency:           stripe.String(req.Currency),
		Description:        stripe.String(req.Description),
		PaymentMethodTypes: stripe.StringSlice(paymentMethodTypes),
		PaymentMethodOptions: &stripe.PaymentIntentPaymentMethodOptionsParams{
			Card: &stripe.PaymentIntentPaymentMethodOptionsCardParams{
				RequestThreeDSecure: stripe.String(threeDSMode),
			},
		},
	}
	params.Context = ctx
	if req.IdempotencyKey != "" {
		params.SetIdempotencyKey(req.IdempotencyKey)
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
	if strings.TrimSpace(params.Metadata["order_id"]) == "" {
		params.Metadata["order_id"] = req.OrderID
	}
	if strings.TrimSpace(params.Metadata["order_number"]) == "" {
		params.Metadata["order_number"] = req.OrderID
	}

	// 设置自动确认
	params.Confirm = stripe.Bool(false)

	// 如果提供了返回URL，设置支付方法选项
	if req.ReturnURL != "" && params.PaymentMethodOptions != nil && params.PaymentMethodOptions.Card != nil {
		params.PaymentMethodOptions.Card.SetupFutureUsage = stripe.String("off_session")
	}

	// 创建支付意图
	client, err := g.stripeClient()
	if err != nil {
		return nil, err
	}
	pi, err := client.PaymentIntents.New(params)
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
	responseAmount, err := MinorToMajorAmount(pi.Amount, string(pi.Currency))
	if err != nil {
		return nil, err
	}

	return &PaymentResponse{
		ID:             pi.ID,
		Status:         string(pi.Status),
		Amount:         responseAmount,
		Currency:       string(pi.Currency),
		ClientSecret:   pi.ClientSecret,
		PublishableKey: g.config.PublishableKey,
		PaymentURL:     paymentURL,
		TransactionID:  pi.ID,
		CreatedAt:      time.Unix(pi.Created, 0),
		Metadata:       pi.Metadata,
	}, nil
}

// CapturePayment 捕获Stripe支付
func (g *stripeGatewayImpl) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	ctx, cancel := paymentGatewayContext(ctx)
	defer cancel()

	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 捕获支付意图
	params := &stripe.PaymentIntentCaptureParams{}
	params.Context = ctx
	client, err := g.stripeClient()
	if err != nil {
		return nil, err
	}
	pi, err := client.PaymentIntents.Capture(paymentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to capture stripe payment: %w", err)
	}

	responseAmount, err := MinorToMajorAmount(pi.Amount, string(pi.Currency))
	if err != nil {
		return nil, err
	}

	return &PaymentResponse{
		ID:             pi.ID,
		Status:         string(pi.Status),
		Amount:         responseAmount,
		Currency:       string(pi.Currency),
		ClientSecret:   pi.ClientSecret,
		PublishableKey: g.config.PublishableKey,
		TransactionID:  pi.ID,
		CreatedAt:      time.Unix(pi.Created, 0),
		Metadata:       pi.Metadata,
	}, nil
}

// RefundPayment 退款Stripe支付
func (g *stripeGatewayImpl) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	return g.RefundPaymentWithOptions(ctx, paymentID, amount, RefundOptions{})
}

func (g *stripeGatewayImpl) RefundPaymentWithOptions(ctx context.Context, paymentID string, amount float64, options RefundOptions) (*RefundResponse, error) {
	ctx, cancel := paymentGatewayContext(ctx)
	defer cancel()

	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	client, err := g.stripeClient()
	if err != nil {
		return nil, err
	}
	getParams := &stripe.PaymentIntentParams{}
	getParams.Context = ctx
	pi, err := client.PaymentIntents.Get(paymentID, getParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe payment for refund: %w", err)
	}
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentID),
	}
	params.Context = ctx
	if options.IdempotencyKey = strings.TrimSpace(options.IdempotencyKey); options.IdempotencyKey != "" {
		params.SetIdempotencyKey(options.IdempotencyKey)
	}

	// 如果指定了金额，设置部分退款
	if amount > 0 {
		refundAmount, err := MajorToMinorAmount(amount, string(pi.Currency))
		if err != nil {
			return nil, err
		}
		params.Amount = stripe.Int64(refundAmount)
	}

	// 创建退款
	r, err := client.Refunds.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe refund: %w", err)
	}

	responseAmount, err := MinorToMajorAmount(r.Amount, string(pi.Currency))
	if err != nil {
		return nil, err
	}

	return &RefundResponse{
		ID:        r.ID,
		PaymentID: paymentID,
		Amount:    responseAmount,
		Status:    string(r.Status),
		CreatedAt: time.Unix(r.Created, 0),
	}, nil
}

// GetPayment 查询Stripe支付
func (g *stripeGatewayImpl) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	ctx, cancel := paymentGatewayContext(ctx)
	defer cancel()

	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 获取支付意图
	client, err := g.stripeClient()
	if err != nil {
		return nil, err
	}
	params := &stripe.PaymentIntentParams{}
	params.Context = ctx
	pi, err := client.PaymentIntents.Get(paymentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe payment: %w", err)
	}

	responseAmount, err := MinorToMajorAmount(pi.Amount, string(pi.Currency))
	if err != nil {
		return nil, err
	}

	return &PaymentResponse{
		ID:             pi.ID,
		Status:         string(pi.Status),
		Amount:         responseAmount,
		Currency:       string(pi.Currency),
		ClientSecret:   pi.ClientSecret,
		PublishableKey: g.config.PublishableKey,
		TransactionID:  pi.ID,
		CreatedAt:      time.Unix(pi.Created, 0),
		Metadata:       pi.Metadata,
	}, nil
}
