package payment

import (
	"context"
	"fmt"
	"time"
)

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
