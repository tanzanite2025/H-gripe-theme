package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/smartwalle/alipay/v3"
)

// CreateAlipayAppPayment 创建支付宝APP支付
func (g *alipayGatewayImpl) CreateAlipayAppPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// 构建APP支付请求
	var p = alipay.TradeAppPay{}
	p.OutTradeNo = req.OrderID
	p.Subject = req.Description
	amountMoney, err := PaymentRequestMoney(req)
	if err != nil {
		return nil, err
	}
	p.TotalAmount, err = amountMoney.FormatMajor()
	if err != nil {
		return nil, err
	}
	p.ProductCode = "QUICK_MSECURITY_PAY"

	// 生成支付字符串（给APP使用）
	paymentString, err := g.client.TradeAppPay(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create alipay app payment: %w", err)
	}

	responseAmount, err := amountMoney.FormatMajor()
	if err != nil {
		return nil, err
	}
	return &PaymentResponse{
		ID:            req.OrderID,
		Status:        "WAIT_BUYER_PAY",
		Amount:        responseAmount,
		AmountMinor:   amountMoney.AmountMinor(),
		Currency:      req.Currency,
		PaymentURL:    paymentString, // APP端使用这个字符串调起支付
		TransactionID: req.OrderID,
		CreatedAt:     time.Now(),
		Metadata:      req.Metadata,
	}, nil
}

// CreateAlipayWapPayment 创建支付宝WAP支付（手机网站支付）
func (g *alipayGatewayImpl) CreateAlipayWapPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	if err := ValidatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// 构建WAP支付请求
	var p = alipay.TradeWapPay{}
	p.OutTradeNo = req.OrderID
	p.Subject = req.Description
	amountMoney, err := PaymentRequestMoney(req)
	if err != nil {
		return nil, err
	}
	p.TotalAmount, err = amountMoney.FormatMajor()
	if err != nil {
		return nil, err
	}
	p.ProductCode = "QUICK_WAP_WAY"

	if req.ReturnURL != "" {
		p.ReturnURL = req.ReturnURL
	}
	if req.CancelURL != "" {
		p.QuitURL = req.CancelURL
	}
	if req.NotifyURL != "" {
		p.NotifyURL = req.NotifyURL
	}

	// 生成支付URL
	paymentURL, err := g.client.TradeWapPay(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create alipay wap payment: %w", err)
	}

	responseAmount, err := amountMoney.FormatMajor()
	if err != nil {
		return nil, err
	}
	return &PaymentResponse{
		ID:            req.OrderID,
		Status:        "WAIT_BUYER_PAY",
		Amount:        responseAmount,
		AmountMinor:   amountMoney.AmountMinor(),
		Currency:      req.Currency,
		PaymentURL:    paymentURL.String(),
		TransactionID: req.OrderID,
		CreatedAt:     time.Now(),
		Metadata:      req.Metadata,
	}, nil
}
