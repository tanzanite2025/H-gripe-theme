package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// PaymentGateway 支付网关接口
type PaymentGateway interface {
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error)
	RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error)
	GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)
	VerifyWebhook(payload []byte, signature string) (bool, error)
}

// PaymentRequest 支付请求
type PaymentRequest struct {
	Amount      float64           `json:"amount"`
	Currency    string            `json:"currency"`
	OrderID     string            `json:"order_id"`
	Description string            `json:"description"`
	Customer    *Customer         `json:"customer"`
	ReturnURL   string            `json:"return_url"`
	CancelURL   string            `json:"cancel_url"`
	Metadata    map[string]string `json:"metadata"`
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	ID            string            `json:"id"`
	Status        string            `json:"status"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	PaymentURL    string            `json:"payment_url,omitempty"`
	TransactionID string            `json:"transaction_id,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// RefundResponse 退款响应
type RefundResponse struct {
	ID        string    `json:"id"`
	PaymentID string    `json:"payment_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Customer 客户信息
type Customer struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone,omitempty"`
}

// GatewayType 支付网关类型
type GatewayType string

const (
	GatewayStripe GatewayType = "stripe"
	GatewayPayPal GatewayType = "paypal"
	GatewayAlipay GatewayType = "alipay"
	GatewayWechat GatewayType = "wechat"
)

// Config 支付网关配置
type Config struct {
	Type          GatewayType
	APIKey        string
	SecretKey     string
	WebhookSecret string
	Environment   string // sandbox, production
}

// NewPaymentGateway 创建支付网关
func NewPaymentGateway(config *Config) (PaymentGateway, error) {
	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	switch config.Type {
	case GatewayStripe:
		return newStripeGateway(config)
	case GatewayPayPal:
		return newPayPalGateway(config)
	case GatewayAlipay:
		return newAlipayGateway(config)
	case GatewayWechat:
		return newWechatGateway(config)
	default:
		return nil, fmt.Errorf("unsupported gateway type: %s", config.Type)
	}
}

// stripeGateway Stripe 支付网关实现
type stripeGateway struct {
	config *Config
}

func newStripeGateway(config *Config) (PaymentGateway, error) {
	// 使用真实的Stripe实现
	return NewStripeGateway(config)
}

func (g *stripeGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// TODO: 实现 Stripe 支付创建
	// 示例代码:
	// params := &stripe.PaymentIntentParams{
	//     Amount:   stripe.Int64(int64(req.Amount * 100)),
	//     Currency: stripe.String(req.Currency),
	//     Description: stripe.String(req.Description),
	// }
	// pi, err := paymentintent.New(params)

	return &PaymentResponse{
		ID:         "pi_mock_" + req.OrderID,
		Status:     "pending",
		Amount:     req.Amount,
		Currency:   req.Currency,
		PaymentURL: "https://checkout.stripe.com/pay/mock",
		CreatedAt:  time.Now(),
		Metadata:   req.Metadata,
	}, nil
}

func (g *stripeGateway) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	// TODO: 实现 Stripe 支付捕获
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "succeeded",
		CreatedAt: time.Now(),
	}, nil
}

func (g *stripeGateway) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	// TODO: 实现 Stripe 退款
	return &RefundResponse{
		ID:        "re_mock_" + paymentID,
		PaymentID: paymentID,
		Amount:    amount,
		Status:    "succeeded",
		CreatedAt: time.Now(),
	}, nil
}

func (g *stripeGateway) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	// TODO: 实现 Stripe 支付查询
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "succeeded",
		CreatedAt: time.Now(),
	}, nil
}

// verifyHMACSHA256 验证 HMAC SHA256 签名
func verifyHMACSHA256(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (g *stripeGateway) VerifyWebhook(payload []byte, signature string) (bool, error) {
	// Stripe 真实验签（未来引入官方 stripe-go 后可以替换为 stripe.Webhook.ConstructEvent）
	if g.config.WebhookSecret == "" {
		return false, fmt.Errorf("webhook secret is not configured")
	}

	// Stripe signature 格式: t=1492774577,v1=5257a869e7ecebeda32affa62cdca3fa51cad7e77a0e56ff536d0ce8e108d8bd
	// 这里提供基础的 v1 提取逻辑（简化版）
	parts := strings.Split(signature, ",")
	var t, v1 string
	for _, part := range parts {
		if strings.HasPrefix(part, "t=") {
			t = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			v1 = strings.TrimPrefix(part, "v1=")
		}
	}

	if t == "" || v1 == "" {
		return false, fmt.Errorf("invalid stripe signature format")
	}

	// 拼接待签名的字符串: timestamp.payload
	signedPayload := fmt.Sprintf("%s.%s", t, string(payload))

	isValid := verifyHMACSHA256([]byte(signedPayload), v1, g.config.WebhookSecret)
	if !isValid {
		return false, fmt.Errorf("signature verification failed")
	}

	return true, nil
}

// paypalGateway PayPal 支付网关实现
type paypalGateway struct {
	config *Config
}

func newPayPalGateway(config *Config) (PaymentGateway, error) {
	// 使用真实的PayPal实现
	return NewPayPalGateway(config)
}

func (g *paypalGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// TODO: 实现 PayPal 支付创建
	return &PaymentResponse{
		ID:         "PAYID-" + req.OrderID,
		Status:     "created",
		Amount:     req.Amount,
		Currency:   req.Currency,
		PaymentURL: "https://www.paypal.com/checkoutnow?token=mock",
		CreatedAt:  time.Now(),
	}, nil
}

func (g *paypalGateway) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	// TODO: 实现 PayPal 支付捕获
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "completed",
		CreatedAt: time.Now(),
	}, nil
}

func (g *paypalGateway) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	// TODO: 实现 PayPal 退款
	return &RefundResponse{
		ID:        "refund_" + paymentID,
		PaymentID: paymentID,
		Amount:    amount,
		Status:    "completed",
		CreatedAt: time.Now(),
	}, nil
}

func (g *paypalGateway) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	// TODO: 实现 PayPal 支付查询
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "completed",
		CreatedAt: time.Now(),
	}, nil
}

func (g *paypalGateway) VerifyWebhook(payload []byte, signature string) (bool, error) {
	// TODO: 实现 PayPal Webhook 验证
	return true, nil
}

// alipayGateway 支付宝支付网关实现
type alipayGateway struct {
	config *Config
}

func newAlipayGateway(config *Config) (PaymentGateway, error) {
	// 使用真实的支付宝实现
	return NewAlipayGateway(config)
}

func (g *alipayGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// TODO: 实现支付宝支付创建
	return &PaymentResponse{
		ID:         "alipay_" + req.OrderID,
		Status:     "WAIT_BUYER_PAY",
		Amount:     req.Amount,
		Currency:   req.Currency,
		PaymentURL: "https://openapi.alipay.com/gateway.do?mock",
		CreatedAt:  time.Now(),
	}, nil
}

func (g *alipayGateway) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "TRADE_SUCCESS",
		CreatedAt: time.Now(),
	}, nil
}

func (g *alipayGateway) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	return &RefundResponse{
		ID:        "refund_" + paymentID,
		PaymentID: paymentID,
		Amount:    amount,
		Status:    "REFUND_SUCCESS",
		CreatedAt: time.Now(),
	}, nil
}

func (g *alipayGateway) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "TRADE_SUCCESS",
		CreatedAt: time.Now(),
	}, nil
}

func (g *alipayGateway) VerifyWebhook(payload []byte, signature string) (bool, error) {
	// TODO: 实现支付宝异步通知验证
	return true, nil
}

// wechatGateway 微信支付网关实现
type wechatGateway struct {
	config *Config
}

func newWechatGateway(config *Config) (PaymentGateway, error) {
	// 使用真实的微信支付实现
	return NewWechatGateway(config)
}

func (g *wechatGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// TODO: 实现微信支付创建
	return &PaymentResponse{
		ID:         "wx_" + req.OrderID,
		Status:     "NOTPAY",
		Amount:     req.Amount,
		Currency:   req.Currency,
		PaymentURL: "weixin://wxpay/bizpayurl?mock",
		CreatedAt:  time.Now(),
	}, nil
}

func (g *wechatGateway) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "SUCCESS",
		CreatedAt: time.Now(),
	}, nil
}

func (g *wechatGateway) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	return &RefundResponse{
		ID:        "refund_" + paymentID,
		PaymentID: paymentID,
		Amount:    amount,
		Status:    "SUCCESS",
		CreatedAt: time.Now(),
	}, nil
}

func (g *wechatGateway) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "SUCCESS",
		CreatedAt: time.Now(),
	}, nil
}

func (g *wechatGateway) VerifyWebhook(payload []byte, signature string) (bool, error) {
	// TODO: 实现微信支付回调验证
	return true, nil
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv(gatewayType GatewayType) *Config {
	prefix := string(gatewayType)
	return &Config{
		Type:          gatewayType,
		APIKey:        os.Getenv(prefix + "_API_KEY"),
		SecretKey:     os.Getenv(prefix + "_SECRET_KEY"),
		WebhookSecret: os.Getenv(prefix + "_WEBHOOK_SECRET"),
		Environment:   getEnv(prefix+"_ENVIRONMENT", "sandbox"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// MockPaymentGateway 模拟支付网关 (用于测试)
type MockPaymentGateway struct{}

func NewMockPaymentGateway() PaymentGateway {
	return &MockPaymentGateway{}
}

func (g *MockPaymentGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:         "mock_" + req.OrderID,
		Status:     "succeeded",
		Amount:     req.Amount,
		Currency:   req.Currency,
		PaymentURL: "https://mock.payment.com/checkout",
		CreatedAt:  time.Now(),
		Metadata:   req.Metadata,
	}, nil
}

func (g *MockPaymentGateway) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "succeeded",
		CreatedAt: time.Now(),
	}, nil
}

func (g *MockPaymentGateway) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	return &RefundResponse{
		ID:        "refund_" + paymentID,
		PaymentID: paymentID,
		Amount:    amount,
		Status:    "succeeded",
		CreatedAt: time.Now(),
	}, nil
}

func (g *MockPaymentGateway) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "succeeded",
		CreatedAt: time.Now(),
	}, nil
}

func (g *MockPaymentGateway) VerifyWebhook(payload []byte, signature string) (bool, error) {
	return true, nil
}

// validateConfig 验证支付网关配置
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Type == "" {
		return fmt.Errorf("gateway type is required")
	}

	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if config.SecretKey == "" {
		return fmt.Errorf("secret key is required")
	}

	if config.Environment != "sandbox" && config.Environment != "production" {
		return fmt.Errorf("environment must be 'sandbox' or 'production'")
	}

	return nil
}

// ValidatePaymentRequest 验证支付请求
func ValidatePaymentRequest(req *PaymentRequest) error {
	if req == nil {
		return fmt.Errorf("payment request cannot be nil")
	}

	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	// 验证货币代码格式 (ISO 4217)
	currencyRegex := regexp.MustCompile(`^[A-Z]{3}$`)
	if !currencyRegex.MatchString(req.Currency) {
		return fmt.Errorf("invalid currency code: must be 3 uppercase letters")
	}

	if req.OrderID == "" {
		return fmt.Errorf("order ID is required")
	}

	if req.Customer == nil {
		return fmt.Errorf("customer information is required")
	}

	if req.Customer.Email == "" {
		return fmt.Errorf("customer email is required")
	}

	// 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Customer.Email) {
		return fmt.Errorf("invalid customer email format")
	}

	return nil
}

// ValidateRefundAmount 验证退款金额
func ValidateRefundAmount(amount, originalAmount float64) error {
	if amount <= 0 {
		return fmt.Errorf("refund amount must be greater than 0")
	}

	if amount > originalAmount {
		return fmt.Errorf("refund amount cannot exceed original payment amount")
	}

	return nil
}
