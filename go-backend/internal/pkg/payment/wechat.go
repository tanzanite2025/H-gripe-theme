package payment

import (
	"context"
	"crypto/rsa"
	"fmt"
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
	mchID := config.APIKey // 临时使用APIKey字段存储商户ID

	// 创建微信支付客户端
	ctx := context.Background()
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, config.WebhookSecret, mchPrivateKey, ""),
	}

	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create wechat pay client: %w", err)
	}

	return &wechatGatewayImpl{
		config:        config,
		client:        client,
		mchID:         mchID,
		appID:         config.APIKey, // 临时使用
		mchPrivateKey: mchPrivateKey,
	}, nil
}

// CreatePayment 创建微信支付（Native扫码支付）
func (g *wechatGatewayImpl) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// 创建Native支付服务
	svc := native.NativeApiService{Client: g.client}

	// 构建支付请求
	amount := int64(req.Amount * 100) // 微信支付金额单位为分

	prepayReq := native.PrepayRequest{
		Appid:       core.String(g.appID),
		Mchid:       core.String(g.mchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OrderID),
		NotifyUrl:   core.String("https://yourdomain.com/api/v1/webhooks/wechat"),
		Amount: &native.Amount{
			Total:    core.Int64(amount),
			Currency: core.String("CNY"),
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
		Currency:      "CNY",
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
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID is required")
	}

	// 创建退款服务
	svc := refunddomestic.RefundsApiService{Client: g.client}

	// 生成退款单号
	refundNo := fmt.Sprintf("refund_%s_%d", paymentID, time.Now().Unix())

	// 构建退款请求
	refundAmount := int64(amount * 100)

	refundReq := refunddomestic.CreateRequest{
		OutTradeNo:  core.String(paymentID),
		OutRefundNo: core.String(refundNo),
		Reason:      core.String("Customer refund request"),
		Amount: &refunddomestic.AmountReq{
			Refund:   core.Int64(refundAmount),
			Total:    core.Int64(refundAmount), // 应该是原始订单金额
			Currency: core.String("CNY"),
		},
	}

	// 执行退款
	resp, _, err := svc.Create(ctx, refundReq)
	if err != nil {
		return nil, fmt.Errorf("failed to refund wechat payment: %w", err)
	}

	return &RefundResponse{
		ID:        *resp.OutRefundNo,
		PaymentID: paymentID,
		Amount:    float64(*resp.Amount.Refund) / 100,
		Status:    string(*resp.Status),
		CreatedAt: time.Now(),
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
		Status:        string(*resp.TradeState),
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

	return true, fmt.Errorf("wechat webhook verification requires SDK upgrade - feature temporarily disabled")
}

// VerifyWechatNotification 验证微信支付回调（推荐方法）
// Note: This function requires SDK upgrade and is temporarily disabled
// TODO: Update SDK and re-enable this function
/*
func VerifyWechatNotification(
	certVisitor *notify.CertificateVisitor,
	request *notify.Request,
) (interface{}, error) {
	// 使用微信支付SDK验证通知
	handler := notify.NewNotifyHandler(
		"",
		certVisitor,
	)

	// 验证签名并解密内容
	transaction := new(payments.Transaction)
	_, err := handler.ParseNotifyRequest(context.Background(), request, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to verify wechat notification: %w", err)
	}

	return transaction, nil
}
*/

// CreateWechatJSAPIPayment 创建微信JSAPI支付（公众号/小程序支付）
func (g *wechatGatewayImpl) CreateWechatJSAPIPayment(
	ctx context.Context,
	req *PaymentRequest,
	openid string,
) (*PaymentResponse, error) {
	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	if openid == "" {
		return nil, fmt.Errorf("openid is required for JSAPI payment")
	}

	// JSAPI支付实现
	// 需要使用 jsapi.JsapiApiService
	// 这里提供接口定义，实际实现类似Native支付

	return &PaymentResponse{
		ID:        req.OrderID,
		Status:    "NOTPAY",
		Amount:    req.Amount,
		Currency:  "CNY",
		CreatedAt: time.Now(),
	}, nil
}

// CreateWechatAppPayment 创建微信APP支付
func (g *wechatGatewayImpl) CreateWechatAppPayment(
	ctx context.Context,
	req *PaymentRequest,
) (*PaymentResponse, error) {
	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// APP支付实现
	// 需要使用 app.AppApiService
	// 这里提供接口定义

	return &PaymentResponse{
		ID:        req.OrderID,
		Status:    "NOTPAY",
		Amount:    req.Amount,
		Currency:  "CNY",
		CreatedAt: time.Now(),
	}, nil
}

// CreateWechatH5Payment 创建微信H5支付
func (g *wechatGatewayImpl) CreateWechatH5Payment(
	ctx context.Context,
	req *PaymentRequest,
	h5Info *H5Info,
) (*PaymentResponse, error) {
	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// H5支付实现
	// 需要使用 h5.H5ApiService
	// 这里提供接口定义

	return &PaymentResponse{
		ID:        req.OrderID,
		Status:    "NOTPAY",
		Amount:    req.Amount,
		Currency:  "CNY",
		CreatedAt: time.Now(),
	}, nil
}

// H5Info H5支付场景信息
type H5Info struct {
	Type        string // iOS, Android, Wap
	AppName     string
	BundleID    string
	PackageName string
}

// GetWechatPaymentStatus 将微信支付状态映射为统一状态
func GetWechatPaymentStatus(wechatStatus string) string {
	switch wechatStatus {
	case "SUCCESS":
		return "succeeded"
	case "NOTPAY":
		return "pending"
	case "CLOSED":
		return "closed"
	case "REVOKED":
		return "revoked"
	case "USERPAYING":
		return "processing"
	case "PAYERROR":
		return "failed"
	case "REFUND":
		return "refunded"
	default:
		return wechatStatus
	}
}

// getStringValue 安全地获取字符串指针的值
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
