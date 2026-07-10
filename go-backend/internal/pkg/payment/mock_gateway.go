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
