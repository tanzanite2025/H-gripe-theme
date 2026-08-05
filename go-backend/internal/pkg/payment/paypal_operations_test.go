package payment

import (
	"context"
	"errors"
	"testing"

	"github.com/plutov/paypal/v4"
)

func TestPayPalRefundUsesStoredCaptureIDDirectly(t *testing.T) {
	client := &fakePayPalCheckoutClient{
		refundResponse: &paypal.RefundResponse{
			ID:     "PAYPAL-REFUND-1",
			Status: "COMPLETED",
			Amount: &paypal.PurchaseUnitAmount{Currency: "USD", Value: "25.50"},
		},
	}
	gateway := &paypalGatewayImpl{config: &Config{Type: GatewayPayPal}, client: client}

	response, err := gateway.RefundPaymentWithOptions(context.Background(), "CAPTURE-1", 25.5, RefundOptions{
		IdempotencyKey: "refund-key-1",
		Currency:       "USD",
	})

	if err != nil {
		t.Fatalf("RefundPaymentWithOptions() error = %v", err)
	}
	if response.ID != "PAYPAL-REFUND-1" || response.Amount != 25.5 {
		t.Fatalf("unexpected refund response: %#v", response)
	}
	if client.refundCaptureID != "CAPTURE-1" {
		t.Fatalf("expected direct capture refund, got capture id %q", client.refundCaptureID)
	}
	if client.getOrderID != "" {
		t.Fatalf("expected no order lookup for direct capture refund, got %q", client.getOrderID)
	}
	if client.refundRequestID != "refund-key-1" {
		t.Fatalf("expected idempotency key to be forwarded")
	}
	if client.refundRequest.Amount == nil || client.refundRequest.Amount.Currency != "USD" || client.refundRequest.Amount.Value != "25.50" {
		t.Fatalf("unexpected refund amount request: %#v", client.refundRequest.Amount)
	}
}

func TestPayPalRefundFallsBackFromOrderIDToCaptureID(t *testing.T) {
	client := &fakePayPalCheckoutClient{
		order: &paypal.Order{PurchaseUnits: []paypal.PurchaseUnit{
			{
				Payments: &paypal.CapturedPayments{Captures: []paypal.CaptureAmount{
					{ID: "CAPTURE-FROM-ORDER", Amount: &paypal.PurchaseUnitAmount{Currency: "EUR", Value: "88.00"}},
				}},
			},
		}},
		fallbackRefundResponse: &paypal.RefundResponse{
			ID:     "PAYPAL-REFUND-2",
			Status: "COMPLETED",
			Amount: &paypal.PurchaseUnitAmount{Currency: "EUR", Value: "10.00"},
		},
	}
	gateway := &paypalGatewayImpl{config: &Config{Type: GatewayPayPal}, client: client}

	response, err := gateway.RefundPaymentWithOptions(context.Background(), "ORDER-1", 10, RefundOptions{})

	if err != nil {
		t.Fatalf("RefundPaymentWithOptions() error = %v", err)
	}
	if response.ID != "PAYPAL-REFUND-2" {
		t.Fatalf("unexpected refund response: %#v", response)
	}
	if client.getOrderID != "ORDER-1" {
		t.Fatalf("expected order lookup for fallback, got %q", client.getOrderID)
	}
	if client.refundCaptureID != "CAPTURE-FROM-ORDER" {
		t.Fatalf("expected fallback capture id, got %q", client.refundCaptureID)
	}
	if client.refundRequest.Amount == nil || client.refundRequest.Amount.Currency != "EUR" || client.refundRequest.Amount.Value != "10.00" {
		t.Fatalf("unexpected fallback refund amount request: %#v", client.refundRequest.Amount)
	}
}

type fakePayPalCheckoutClient struct {
	order                  *paypal.Order
	getOrderID             string
	refundCaptureID        string
	refundRequest          paypal.RefundCaptureRequest
	refundRequestID        string
	refundErr              error
	refundResponse         *paypal.RefundResponse
	fallbackRefundResponse *paypal.RefundResponse
}

func (c *fakePayPalCheckoutClient) CreateOrder(context.Context, string, []paypal.PurchaseUnitRequest, *paypal.PaymentSource, *paypal.ApplicationContext) (*paypal.Order, error) {
	return nil, errors.New("not implemented")
}

func (c *fakePayPalCheckoutClient) CaptureOrder(context.Context, string, paypal.CaptureOrderRequest) (*paypal.CaptureOrderResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *fakePayPalCheckoutClient) GetOrder(_ context.Context, orderID string) (*paypal.Order, error) {
	c.getOrderID = orderID
	if c.order == nil {
		return nil, errors.New("order not found")
	}
	return c.order, nil
}

func (c *fakePayPalCheckoutClient) RefundCaptureWithPaypalRequestId(_ context.Context, captureID string, request paypal.RefundCaptureRequest, requestID string) (*paypal.RefundResponse, error) {
	c.refundCaptureID = captureID
	c.refundRequest = request
	c.refundRequestID = requestID
	if c.refundErr != nil {
		err := c.refundErr
		c.refundErr = nil
		return nil, err
	}
	if c.fallbackRefundResponse != nil {
		return c.fallbackRefundResponse, nil
	}
	return c.refundResponse, nil
}
