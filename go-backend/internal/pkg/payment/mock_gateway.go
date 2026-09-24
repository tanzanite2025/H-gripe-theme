package payment

import (
	"context"
	"time"
)

// MockPaymentGateway 模拟支付网关 (用于测试)
type MockPaymentGateway struct{}

func NewMockPaymentGateway() PaymentGateway {
	return &MockPaymentGateway{}
}

func (g *MockPaymentGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	amountMoney, err := PaymentRequestMoney(req)
	if err != nil {
		return nil, err
	}
	amount, err := amountMoney.FormatMajor()
	if err != nil {
		return nil, err
	}
	return &PaymentResponse{
		ID:          "mock_" + req.OrderID,
		Status:      "succeeded",
		Amount:      amount,
		AmountMinor: amountMoney.AmountMinor(),
		Currency:    req.Currency,
		PaymentURL:  "https://mock.payment.com/checkout",
		CreatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}, nil
}

func (g *MockPaymentGateway) CapturePayment(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	return &PaymentResponse{
		ID:        paymentID,
		Status:    "succeeded",
		CreatedAt: time.Now(),
	}, nil
}

func (g *MockPaymentGateway) RefundPayment(ctx context.Context, paymentID string, amountMinor int64) (*RefundResponse, error) {
	return g.RefundPaymentWithOptions(ctx, paymentID, amountMinor, RefundOptions{})
}

func (g *MockPaymentGateway) RefundPaymentWithOptions(ctx context.Context, paymentID string, amountMinor int64, options RefundOptions) (*RefundResponse, error) {
	refundID := "refund_" + paymentID
	if options.IdempotencyKey != "" {
		refundID = options.IdempotencyKey
	}
	refundCurrency := options.Currency
	if refundCurrency == "" {
		refundCurrency = "USD"
	}
	options.Currency = refundCurrency
	refundMoney, err := RefundOptionMoney(amountMinor, options)
	if err != nil {
		return nil, err
	}
	refundAmount, err := refundMoney.FormatMajor()
	if err != nil {
		return nil, err
	}
	return &RefundResponse{
		ID:          refundID,
		PaymentID:   paymentID,
		Amount:      refundAmount,
		AmountMinor: refundMoney.AmountMinor(),
		Status:      "succeeded",
		CreatedAt:   time.Now(),
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
