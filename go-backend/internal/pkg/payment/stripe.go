package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainmoney "commerce-platform/internal/domain/money"

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

	amountMoney, err := PaymentRequestMoney(req)
	if err != nil {
		return nil, err
	}
	amount := amountMoney.AmountMinor()
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
	if req.ShippingAddress != nil {
		params.Shipping = stripeShippingDetailsParams(req.ShippingAddress)
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
	responseAmount, err := paymentMajorStringFromMinor(pi.Amount, string(pi.Currency))
	if err != nil {
		return nil, err
	}

	return &PaymentResponse{
		ID:             pi.ID,
		Status:         string(pi.Status),
		Amount:         responseAmount,
		AmountMinor:    pi.Amount,
		Currency:       string(pi.Currency),
		ClientSecret:   pi.ClientSecret,
		PublishableKey: g.config.PublishableKey,
		PaymentURL:     paymentURL,
		TransactionID:  pi.ID,
		CreatedAt:      time.Unix(pi.Created, 0),
		Metadata:       pi.Metadata,
	}, nil
}

func stripeShippingDetailsParams(address *ShippingAddress) *stripe.ShippingDetailsParams {
	if address == nil {
		return nil
	}
	return &stripe.ShippingDetailsParams{
		Name:  stripe.String(strings.TrimSpace(address.Name)),
		Phone: stripe.String(strings.TrimSpace(address.Phone)),
		Address: &stripe.AddressParams{
			Line1:      stripe.String(strings.TrimSpace(address.Line1)),
			Line2:      stripe.String(strings.TrimSpace(address.Line2)),
			City:       stripe.String(strings.TrimSpace(address.City)),
			State:      stripe.String(strings.TrimSpace(address.State)),
			PostalCode: stripe.String(strings.TrimSpace(address.PostalCode)),
			Country:    stripe.String(strings.ToUpper(strings.TrimSpace(address.Country))),
		},
	}
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
	params.AddExpand("latest_charge")
	params.Context = ctx
	client, err := g.stripeClient()
	if err != nil {
		return nil, err
	}
	pi, err := client.PaymentIntents.Capture(paymentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to capture stripe payment: %w", err)
	}

	responseAmount, err := paymentMajorStringFromMinor(pi.Amount, string(pi.Currency))
	if err != nil {
		return nil, err
	}

	return &PaymentResponse{
		ID:               pi.ID,
		Status:           string(pi.Status),
		Amount:           responseAmount,
		AmountMinor:      pi.Amount,
		Currency:         string(pi.Currency),
		ClientSecret:     pi.ClientSecret,
		PublishableKey:   g.config.PublishableKey,
		TransactionID:    pi.ID,
		LiabilityShifted: g.resolveStripePaymentIntentLiabilityShifted(ctx, client, pi),
		CreatedAt:        time.Unix(pi.Created, 0),
		Metadata:         pi.Metadata,
	}, nil
}

// RefundPayment 退款Stripe支付
func (g *stripeGatewayImpl) RefundPayment(ctx context.Context, paymentID string, amountMinor int64) (*RefundResponse, error) {
	return g.RefundPaymentWithOptions(ctx, paymentID, amountMinor, RefundOptions{})
}

func (g *stripeGatewayImpl) RefundPaymentWithOptions(ctx context.Context, paymentID string, amountMinor int64, options RefundOptions) (*RefundResponse, error) {
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
	params.AddExpand("balance_transaction")
	if options.IdempotencyKey = strings.TrimSpace(options.IdempotencyKey); options.IdempotencyKey != "" {
		params.SetIdempotencyKey(options.IdempotencyKey)
	}

	// 如果指定了金额，设置部分退款。金额始终使用最小货币单位。
	if options.AmountMinor > 0 || amountMinor > 0 {
		refundCurrency := strings.TrimSpace(options.Currency)
		if refundCurrency == "" {
			refundCurrency = string(pi.Currency)
		}
		refundMinor := amountMinor
		if options.AmountMinor > 0 {
			refundMinor = options.AmountMinor
		}
		refundMoney, err := domainmoney.New(refundMinor, refundCurrency)
		if err != nil {
			return nil, err
		}
		params.Amount = stripe.Int64(refundMoney.AmountMinor())
	}

	// 创建退款
	r, err := client.Refunds.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe refund: %w", err)
	}

	responseAmount, err := paymentMajorStringFromMinor(r.Amount, string(pi.Currency))
	if err != nil {
		return nil, err
	}

	response := &RefundResponse{
		ID:          r.ID,
		PaymentID:   paymentID,
		Amount:      responseAmount,
		AmountMinor: r.Amount,
		Status:      string(r.Status),
		CreatedAt:   time.Unix(r.Created, 0),
	}
	if r.BalanceTransaction != nil {
		response.SettlementAmountMinor = absInt64(r.BalanceTransaction.Net)
		response.SettlementCurrency = string(r.BalanceTransaction.Currency)
		response.SettlementBalanceTransactionID = r.BalanceTransaction.ID
	}
	return response, nil
}

func absInt64(value int64) int64 {
	if value < 0 {
		if value == -1<<63 {
			return 1<<63 - 1
		}
		return -value
	}
	return value
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
	// Stripe webhooks commonly carry latest_charge as an ID-only relationship.
	// Expanding it here gives the 3DS outcome needed for high-value fulfilment
	// decisions without depending on a removed PaymentIntent.charges list.
	params.AddExpand("latest_charge")
	params.Context = ctx
	pi, err := client.PaymentIntents.Get(paymentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe payment: %w", err)
	}

	responseAmount, err := paymentMajorStringFromMinor(pi.Amount, string(pi.Currency))
	if err != nil {
		return nil, err
	}

	return &PaymentResponse{
		ID:               pi.ID,
		Status:           string(pi.Status),
		Amount:           responseAmount,
		AmountMinor:      pi.Amount,
		Currency:         string(pi.Currency),
		ClientSecret:     pi.ClientSecret,
		PublishableKey:   g.config.PublishableKey,
		TransactionID:    pi.ID,
		LiabilityShifted: g.resolveStripePaymentIntentLiabilityShifted(ctx, client, pi),
		CreatedAt:        time.Unix(pi.Created, 0),
		Metadata:         pi.Metadata,
	}, nil
}

func (g *stripeGatewayImpl) resolveStripePaymentIntentLiabilityShifted(
	ctx context.Context,
	client *stripeclient.API,
	intent *stripe.PaymentIntent,
) *bool {
	if shifted := stripePaymentIntentLiabilityShifted(intent); shifted != nil {
		return shifted
	}
	if client == nil || intent == nil || intent.LatestCharge == nil || strings.TrimSpace(intent.LatestCharge.ID) == "" {
		return nil
	}

	// Some Stripe API versions still return latest_charge as an ID even when
	// the expand request is accepted. Retrieve the Charge explicitly so the
	// liability decision never depends on a webhook's expansion shape.
	chargeParams := &stripe.ChargeParams{}
	chargeParams.Context = ctx
	charge, err := client.Charges.Get(strings.TrimSpace(intent.LatestCharge.ID), chargeParams)
	if err != nil {
		return nil
	}
	intent.LatestCharge = charge
	if charge.LastResponse != nil {
		if shifted := stripeChargeRawLiabilityShifted(charge.LastResponse.RawJSON); shifted != nil {
			return shifted
		}
	}
	return stripePaymentIntentLiabilityShifted(intent)
}

func stripeChargeRawLiabilityShifted(raw []byte) *bool {
	var payload struct {
		PaymentMethodDetails *struct {
			Card *struct {
				ThreeDSecure *struct {
					LiabilityShifted json.RawMessage `json:"liability_shifted"`
					LiabilityShift   json.RawMessage `json:"liability_shift"`
				} `json:"three_d_secure"`
			} `json:"card"`
		} `json:"payment_method_details"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil || payload.PaymentMethodDetails == nil ||
		payload.PaymentMethodDetails.Card == nil || payload.PaymentMethodDetails.Card.ThreeDSecure == nil {
		return nil
	}
	threeDSecure := payload.PaymentMethodDetails.Card.ThreeDSecure
	if shifted := parseStripeLiabilityBool(threeDSecure.LiabilityShifted); shifted != nil {
		return shifted
	}
	return parseStripeLiabilityOutcome(threeDSecure.LiabilityShift)
}

func parseStripeLiabilityBool(raw json.RawMessage) *bool {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err == nil {
		return stripe.Bool(value)
	}
	var textValue string
	if err := json.Unmarshal(raw, &textValue); err != nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(textValue)) {
	case "true", "t", "1", "yes", "y":
		return stripe.Bool(true)
	case "false", "f", "0", "no", "n":
		return stripe.Bool(false)
	default:
		return nil
	}
}

func parseStripeLiabilityOutcome(raw json.RawMessage) *bool {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	switch strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.TrimSpace(value))) {
	case "issuer", "issuershifted", "shifted", "liabilityshifted":
		return stripe.Bool(true)
	case "merchant", "merchantliability", "notshifted", "noliabilityshift", "none":
		return stripe.Bool(false)
	default:
		return nil
	}
}

// stripePaymentIntentLiabilityShifted converts Stripe's expanded Charge 3DS
// result into the gateway contract. stripe-go v76 does not expose a
// liability_shifted field on the Charge type; Stripe's supported 3DS result
// values are the durable source available through the typed API response.
func stripePaymentIntentLiabilityShifted(intent *stripe.PaymentIntent) *bool {
	if intent == nil || intent.LatestCharge == nil || intent.LatestCharge.PaymentMethodDetails == nil ||
		intent.LatestCharge.PaymentMethodDetails.Card == nil ||
		intent.LatestCharge.PaymentMethodDetails.Card.ThreeDSecure == nil {
		return nil
	}

	switch intent.LatestCharge.PaymentMethodDetails.Card.ThreeDSecure.Result {
	case stripe.ChargePaymentMethodDetailsCardThreeDSecureResultAuthenticated,
		stripe.ChargePaymentMethodDetailsCardThreeDSecureResultAttemptAcknowledged:
		return stripe.Bool(true)
	case stripe.ChargePaymentMethodDetailsCardThreeDSecureResultFailed,
		stripe.ChargePaymentMethodDetailsCardThreeDSecureResultNotSupported,
		stripe.ChargePaymentMethodDetailsCardThreeDSecureResultProcessingError:
		return stripe.Bool(false)
	default:
		// Exempted and empty results do not prove a liability shift. Keep the
		// value unknown so the high-value policy can hold conservatively.
		return nil
	}
}
