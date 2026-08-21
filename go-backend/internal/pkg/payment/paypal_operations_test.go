package payment

import (
	"context"
	"errors"
	"net/http"
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

func TestPayPalRefundDoesNotResolveCaptureFromOrderID(t *testing.T) {
	client := &fakePayPalCheckoutClient{
		refundErr: errors.New("capture not found"),
		order:     &paypal.Order{ID: "ORDER-1"},
	}
	gateway := &paypalGatewayImpl{config: &Config{Type: GatewayPayPal}, client: client}

	_, err := gateway.RefundPaymentWithOptions(context.Background(), "ORDER-1", 10, RefundOptions{Currency: "EUR"})

	if err == nil {
		t.Fatalf("RefundPaymentWithOptions() expected error")
	}
	if client.getOrderID != "" {
		t.Fatalf("expected no order lookup for refund, got %q", client.getOrderID)
	}
	if client.refundCaptureID != "ORDER-1" {
		t.Fatalf("expected stored reference to be used as capture id, got %q", client.refundCaptureID)
	}
}

func TestPayPalCreatePaymentRetriesOnUnauthorized(t *testing.T) {
	client := &fakePayPalCheckoutClient{
		createOrderErrs: []error{paypalUnauthorizedError(t)},
		createOrderResponse: &paypal.Order{
			ID:     "ORDER-2",
			Status: "CREATED",
			Links: []paypal.Link{
				{Rel: "approve", Href: "https://paypal.example/approve"},
			},
		},
	}
	gateway := &paypalGatewayImpl{config: &Config{Type: GatewayPayPal}, client: client}

	response, err := gateway.CreatePayment(context.Background(), &PaymentRequest{
		Amount:      12.34,
		Currency:    "USD",
		OrderID:     "ORDER-1",
		Description: "test order",
		Customer:    &Customer{Email: "buyer@example.com", Name: "Buyer"},
	})
	if err != nil {
		t.Fatalf("CreatePayment() error = %v", err)
	}
	if client.createOrderCalls != 2 {
		t.Fatalf("expected create order retry, got %d calls", client.createOrderCalls)
	}
	if client.accessTokenRefreshes != 1 {
		t.Fatalf("expected one access token refresh, got %d", client.accessTokenRefreshes)
	}
	if response.ID != "ORDER-2" || response.PaymentURL != "https://paypal.example/approve" {
		t.Fatalf("unexpected payment response: %#v", response)
	}
}

func TestPayPalCapturePaymentRetriesOnUnauthorized(t *testing.T) {
	client := &fakePayPalCheckoutClient{
		captureOrderErrs: []error{paypalUnauthorizedError(t)},
		captureOrderResponse: &paypal.CaptureOrderResponse{
			ID:     "ORDER-1",
			Status: "PENDING",
		},
	}
	gateway := &paypalGatewayImpl{config: &Config{Type: GatewayPayPal}, client: client}

	response, err := gateway.CapturePayment(context.Background(), "ORDER-1")
	if err != nil {
		t.Fatalf("CapturePayment() error = %v", err)
	}
	if client.captureOrderCalls != 2 {
		t.Fatalf("expected capture retry, got %d calls", client.captureOrderCalls)
	}
	if client.accessTokenRefreshes != 1 {
		t.Fatalf("expected one access token refresh, got %d", client.accessTokenRefreshes)
	}
	if response.ID != "ORDER-1" || response.Status != "PENDING" {
		t.Fatalf("unexpected capture response: %#v", response)
	}
}

type fakePayPalCheckoutClient struct {
	order                *paypal.Order
	getOrderID           string
	refundCaptureID      string
	refundRequest        paypal.RefundCaptureRequest
	refundRequestID      string
	refundErr            error
	refundResponse       *paypal.RefundResponse
	createOrderCalls     int
	captureOrderCalls    int
	createOrderErrs      []error
	captureOrderErrs     []error
	createOrderResponse  *paypal.Order
	captureOrderResponse *paypal.CaptureOrderResponse
	accessTokenRefreshes int
}

func (c *fakePayPalCheckoutClient) CreateOrder(context.Context, string, []paypal.PurchaseUnitRequest, *paypal.PaymentSource, *paypal.ApplicationContext) (*paypal.Order, error) {
	c.createOrderCalls++
	if len(c.createOrderErrs) > 0 {
		err := c.createOrderErrs[0]
		c.createOrderErrs = c.createOrderErrs[1:]
		return nil, err
	}
	if c.createOrderResponse != nil {
		return c.createOrderResponse, nil
	}
	return nil, errors.New("not implemented")
}

func (c *fakePayPalCheckoutClient) CaptureOrder(context.Context, string, paypal.CaptureOrderRequest) (*paypal.CaptureOrderResponse, error) {
	c.captureOrderCalls++
	if len(c.captureOrderErrs) > 0 {
		err := c.captureOrderErrs[0]
		c.captureOrderErrs = c.captureOrderErrs[1:]
		return nil, err
	}
	if c.captureOrderResponse != nil {
		return c.captureOrderResponse, nil
	}
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
	return c.refundResponse, nil
}

func (c *fakePayPalCheckoutClient) GetAccessToken(context.Context) (*paypal.TokenResponse, error) {
	c.accessTokenRefreshes++
	return &paypal.TokenResponse{Token: "refreshed-token"}, nil
}

func paypalUnauthorizedError(t *testing.T) error {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "https://api-m.paypal.com/v2/checkout/orders", nil)
	if err != nil {
		t.Fatalf("failed to build paypal unauthorized request: %v", err)
	}
	return &paypal.ErrorResponse{
		Response: &http.Response{
			StatusCode: http.StatusUnauthorized,
			Request:    req,
		},
		Message: "Unauthorized",
	}
}
