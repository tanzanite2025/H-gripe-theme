package payment

import (
	"context"
	"fmt"
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

func newStripeGateway(config *Config) (PaymentGateway, error) {
	return NewStripeGateway(config)
}

func newPayPalGateway(config *Config) (PaymentGateway, error) {
	return NewPayPalGateway(config)
}

func newAlipayGateway(config *Config) (PaymentGateway, error) {
	return NewAlipayGateway(config)
}

func newWechatGateway(config *Config) (PaymentGateway, error) {
	return NewWechatGateway(config)
}
