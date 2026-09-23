package payment

import (
	domainmoney "commerce-platform/internal/domain/money"
	"context"
	"fmt"
	"time"
)

// PaymentGateway 支付网关接口
type PaymentGateway interface {
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error)
	// amountMinor is expressed in the payment currency's smallest unit. A zero
	// amount requests a full refund; positive values request a partial refund.
	RefundPayment(ctx context.Context, paymentID string, amountMinor int64) (*RefundResponse, error)
	RefundPaymentWithOptions(ctx context.Context, paymentID string, amountMinor int64, options RefundOptions) (*RefundResponse, error)
	GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)
	VerifyWebhook(payload []byte, signature string) (bool, error)
}

type CaptureOptions struct {
	IdempotencyKey string `json:"-"`
}

type CapturePaymentWithOptions interface {
	CapturePaymentWithOptions(ctx context.Context, paymentID string, options CaptureOptions) (*PaymentResponse, error)
}

// PaymentRequest 支付请求
type PaymentRequest struct {
	AmountMinor     int64             `json:"amount_minor"`
	Currency        string            `json:"currency"`
	OrderID         string            `json:"order_id"`
	Description     string            `json:"description"`
	Customer        *Customer         `json:"customer"`
	ShippingAddress *ShippingAddress  `json:"shipping_address,omitempty"`
	ReturnURL       string            `json:"return_url"`
	CancelURL       string            `json:"cancel_url"`
	NotifyURL       string            `json:"notify_url,omitempty"`
	ThreeDSecure    string            `json:"three_ds_secure,omitempty"`
	CardBIN         string            `json:"card_bin,omitempty"`
	IdempotencyKey  string            `json:"-"`
	Metadata        map[string]string `json:"metadata"`
}

// ShippingAddress contains the complete delivery address needed by providers
// for fraud screening, AVS checks, and delivery-risk evaluation.
type ShippingAddress struct {
	Name       string `json:"name"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	Phone      string `json:"phone,omitempty"`
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	ID               string            `json:"id"`
	Status           string            `json:"status"`
	Amount           string            `json:"amount"`
	AmountMinor      int64             `json:"amount_minor,omitempty"`
	Currency         string            `json:"currency"`
	ClientSecret     string            `json:"client_secret,omitempty"`
	PublishableKey   string            `json:"publishable_key,omitempty"`
	PaymentURL       string            `json:"payment_url,omitempty"`
	TransactionID    string            `json:"transaction_id,omitempty"`
	LiabilityShifted *bool             `json:"liability_shifted,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// RefundResponse 退款响应
type RefundResponse struct {
	ID          string    `json:"id"`
	PaymentID   string    `json:"payment_id"`
	Amount      string    `json:"amount"`
	AmountMinor int64     `json:"amount_minor,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	// Settlement fields are populated when the provider returns a balance
	// transaction. Amount is a positive minor-unit deduction from the
	// merchant settlement balance, not the customer-facing refund amount.
	SettlementAmountMinor          int64  `json:"settlement_amount_minor,omitempty"`
	SettlementCurrency             string `json:"settlement_currency,omitempty"`
	SettlementBalanceTransactionID string `json:"settlement_balance_transaction_id,omitempty"`
}

type RefundOptions struct {
	IdempotencyKey        string `json:"-"`
	AmountMinor           int64  `json:"amount_minor,omitempty"`
	OriginalAmountMinor   int64  `json:"original_amount_minor,omitempty"`
	Reason                string `json:"reason,omitempty"`
	Currency              string `json:"currency,omitempty"`
	MerchantOrderNumber   string `json:"merchant_order_number,omitempty"`
	ProviderTransactionID string `json:"provider_transaction_id,omitempty"`
}

// PaymentRequestMoney resolves the exact request amount at the provider
// boundary. Payment requests are minor-unit-only; conversion from major units
// belongs at the transport boundary before constructing PaymentRequest.
func PaymentRequestMoney(req *PaymentRequest) (domainmoney.Money, error) {
	if req == nil {
		return domainmoney.Money{}, fmt.Errorf("payment request cannot be nil")
	}
	if req.AmountMinor <= 0 {
		return domainmoney.Money{}, fmt.Errorf("payment amount minor must be greater than zero")
	}
	return domainmoney.New(req.AmountMinor, req.Currency)
}

// RefundOptionMoney resolves an exact refund amount when a provider adapter
// receives a minor-unit amount from the refund execution worker.
func RefundOptionMoney(amountMinor int64, options RefundOptions) (domainmoney.Money, error) {
	if options.AmountMinor < 0 {
		return domainmoney.Money{}, fmt.Errorf("refund minor amount cannot be negative")
	}
	if amountMinor < 0 {
		return domainmoney.Money{}, fmt.Errorf("refund minor amount cannot be negative")
	}
	if options.AmountMinor > 0 {
		amountMinor = options.AmountMinor
	}
	return domainmoney.New(amountMinor, options.Currency)
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

const defaultPaymentGatewayTimeout = 15 * time.Second

// Config 支付网关配置
type Config struct {
	Type               GatewayType
	APIKey             string
	SecretKey          string
	PublishableKey     string
	WebhookSecret      string
	ThreeDSecure       string
	Environment        string // sandbox, production
	PaymentMethodTypes []string

	// Provider-specific webhook credentials. Keep these out of the generic
	// fields so callback verification can evolve without overloading names.
	WechatAPIv3Key               string
	WechatAppID                  string
	WechatPayPlatformCertificate string
	WechatPayPlatformPublicKey   string
	WechatPayPlatformPublicKeyID string
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

func paymentGatewayContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, defaultPaymentGatewayTimeout)
}
