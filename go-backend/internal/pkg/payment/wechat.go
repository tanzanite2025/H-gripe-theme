package payment

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// wechatGatewayImpl 微信支付网关完整实现
type wechatGatewayImpl struct {
	config        *Config
	client        *core.Client
	mchID         string          // 商户号
	appID         string          // 应用ID
	mchPrivateKey *rsa.PrivateKey // 商户私钥
}

// NewWechatGateway 创建微信支付网关实例
func NewWechatGateway(config *Config) (PaymentGateway, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if config.WechatAppID == "" {
		return nil, fmt.Errorf("wechat app_id is required")
	}
	if config.WechatAPIv3Key == "" {
		return nil, fmt.Errorf("wechat api_v3_key is required")
	}

	// config.APIKey = AppID
	// config.SecretKey = 商户API密钥（32字节）
	// config.WebhookSecret = 商户证书序列号

	// 加载商户私钥
	// 注意：在生产环境中，私钥应该从安全的密钥管理服务中加载
	// 这里假设SecretKey包含了私钥内容或路径
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load wechat merchant private key: %w", err)
	}

	// 从环境变量获取商户ID
	// 在实际使用中，可以添加到Config结构中
	mchID := config.APIKey // APIKey stores the merchant ID for the legacy gateway interface.

	// 创建微信支付客户端
	ctx := context.Background()
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, config.WebhookSecret, mchPrivateKey, config.WechatAPIv3Key),
	}

	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create wechat pay client: %w", err)
	}

	return &wechatGatewayImpl{
		config:        config,
		client:        client,
		mchID:         mchID,
		appID:         config.WechatAppID,
		mchPrivateKey: mchPrivateKey,
	}, nil
}

// CreatePayment 创建微信支付（Native扫码支付）
func (g *wechatGatewayImpl) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}
	if err := ValidateGatewayCurrency(g.config.Type, req.Currency); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// 创建Native支付服务
	svc := native.NativeApiService{Client: g.client}

	notifyURL := req.NotifyURL
	if notifyURL == "" {
		return nil, fmt.Errorf("wechat notify_url is required")
	}
	amount, err := MajorToMinorAmount(req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}

	prepayReq := native.PrepayRequest{
		Appid:       core.String(g.appID),
		Mchid:       core.String(g.mchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OrderID),
		NotifyUrl:   core.String(notifyURL),
		Amount: &native.Amount{
			Total:    core.Int64(amount),
			Currency: core.String(req.Currency),
		},
	}

	// 添加用户信息
	if req.Customer != nil && req.Customer.Email != "" {
		// 微信支付可以添加附加数据
		prepayReq.Attach = core.String(req.Customer.Email)
	}

	// 创建预支付订单
	resp, _, err := svc.Prepay(ctx, prepayReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create wechat payment: %w", err)
	}
	if resp.CodeUrl == nil || *resp.CodeUrl == "" {
		return nil, fmt.Errorf("wechat payment response did not include code_url")
	}

	// 构建元数据
	metadata := req.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["order_id"] = req.OrderID
	metadata["code_url"] = *resp.CodeUrl

	return &PaymentResponse{
		ID:            req.OrderID,
		Status:        "NOTPAY",
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentURL:    *resp.CodeUrl, // 二维码链接
		TransactionID: req.OrderID,
		CreatedAt:     time.Now(),
		Metadata:      metadata,
	}, nil
}

// CapturePayment 捕获微信支付（微信支付自动完成，此方法用于查询状态）
func (g *wechatGatewayImpl) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 查询订单状态
	return g.GetPayment(ctx, paymentID)
}

// RefundPayment 退款微信支付
func (g *wechatGatewayImpl) RefundPayment(ctx context.Context, paymentID string, amount float64) (*RefundResponse, error) {
	return g.RefundPaymentWithOptions(ctx, paymentID, amount, RefundOptions{})
}

func (g *wechatGatewayImpl) RefundPaymentWithOptions(ctx context.Context, paymentID string, amount float64, options RefundOptions) (*RefundResponse, error) {
	paymentID = strings.TrimSpace(paymentID)
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 创建退款服务
	svc := refunddomestic.RefundsApiService{Client: g.client}

	// 生成退款单号
	refundNo := fmt.Sprintf("refund_%s_%d", paymentID, time.Now().Unix())
	if options.IdempotencyKey != "" {
		refundNo = options.IdempotencyKey
	}

	refundReq, err := buildWechatRefundRequest(paymentID, amount, refundNo, options)
	if err != nil {
		return nil, err
	}

	// 执行退款
	resp, _, err := svc.Create(ctx, refundReq)
	if err != nil {
		return nil, fmt.Errorf("failed to refund wechat payment: %w", err)
	}

	return &RefundResponse{
		ID:        *resp.OutRefundNo,
		PaymentID: paymentID,
		Amount:    amount,
		Status:    string(*resp.Status),
		CreatedAt: time.Now(),
	}, nil
}

func buildWechatRefundRequest(paymentID string, amount float64, refundNo string, options RefundOptions) (refunddomestic.CreateRequest, error) {
	providerTransactionID := strings.TrimSpace(options.ProviderTransactionID)
	merchantOrderNumber := strings.TrimSpace(options.MerchantOrderNumber)
	if providerTransactionID == "" {
		return refunddomestic.CreateRequest{}, fmt.Errorf("wechat transaction_id is required for refunds")
	}
	if merchantOrderNumber == "" {
		return refunddomestic.CreateRequest{}, fmt.Errorf("merchant order number is required for wechat refunds")
	}
	if amount <= 0 {
		return refunddomestic.CreateRequest{}, fmt.Errorf("refund amount must be greater than zero")
	}
	if options.OriginalAmount <= 0 {
		return refunddomestic.CreateRequest{}, fmt.Errorf("original payment amount is required for wechat refunds")
	}
	if refundNo = strings.TrimSpace(refundNo); refundNo == "" {
		return refunddomestic.CreateRequest{}, fmt.Errorf("refund request id is required")
	}
	currency := strings.ToUpper(strings.TrimSpace(options.Currency))
	if currency == "" {
		currency = "CNY"
	}
	refundAmount, err := MajorToMinorAmount(amount, currency)
	if err != nil {
		return refunddomestic.CreateRequest{}, err
	}
	totalAmount, err := MajorToMinorAmount(options.OriginalAmount, currency)
	if err != nil {
		return refunddomestic.CreateRequest{}, err
	}
	reason := strings.TrimSpace(options.Reason)
	if reason == "" {
		reason = "Customer refund request"
	}
	return refunddomestic.CreateRequest{
		TransactionId: core.String(providerTransactionID),
		OutRefundNo:   core.String(refundNo),
		Reason:        core.String(reason),
		Amount: &refunddomestic.AmountReq{
			Refund:   core.Int64(refundAmount),
			Total:    core.Int64(totalAmount),
			Currency: core.String(currency),
		},
	}, nil
}

// GetPayment 查询微信支付
func (g *wechatGatewayImpl) GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 创建查询服务
	svc := native.NativeApiService{Client: g.client}

	// 查询订单
	queryReq := native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(paymentID),
		Mchid:      core.String(g.mchID),
	}

	resp, _, err := svc.QueryOrderByOutTradeNo(ctx, queryReq)
	if err != nil {
		return nil, fmt.Errorf("failed to query wechat payment: %w", err)
	}

	// 提取金额
	var amount float64
	if resp.Amount != nil && resp.Amount.Total != nil {
		amount = float64(*resp.Amount.Total) / 100
	}

	// 构建元数据
	metadata := make(map[string]string)
	if resp.TransactionId != nil {
		metadata["transaction_id"] = *resp.TransactionId
	}
	if resp.Attach != nil {
		metadata["attach"] = *resp.Attach
	}

	return &PaymentResponse{
		ID:            *resp.OutTradeNo,
		Status:        *resp.TradeState,
		Amount:        amount,
		Currency:      "CNY",
		TransactionID: getStringValue(resp.TransactionId),
		CreatedAt:     time.Now(),
		Metadata:      metadata,
	}, nil
}

// VerifyWebhook 验证微信支付回调签名
func (g *wechatGatewayImpl) VerifyWebhook(payload []byte, signature string) (bool, error) {
	// 微信支付V3使用更复杂的验签方式
	// 需要从HTTP头中提取多个字段：
	// Wechatpay-Signature, Wechatpay-Timestamp, Wechatpay-Nonce, Wechatpay-Serial

	// 这里提供基础验证框架
	// 实际使用时需要传入完整的HTTP头信息

	return false, fmt.Errorf("wechat webhook verification requires SDK upgrade - feature temporarily disabled")
}
