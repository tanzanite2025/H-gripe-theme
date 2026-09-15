package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func TestCreateOrderRequestRequiresExpectedTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		"POST",
		"/api/v1/orders",
		bytes.NewBufferString(`{
			"items":[{"product_id":1,"quantity":1}],
			"shipping_address":{
				"first_name":"Test",
				"last_name":"Buyer",
				"address1":"1 Test Street",
				"city":"Austin",
				"postal_code":"78701",
				"country":"US",
				"phone":"+15555550123",
				"email":"buyer@example.com"
			},
			"billing_address":{
				"first_name":"Test",
				"last_name":"Buyer",
				"address1":"1 Test Street",
				"city":"Austin",
				"postal_code":"78701",
				"country":"US",
				"phone":"+15555550123",
				"email":"buyer@example.com"
			},
			"payment_method":"card",
			"shipping_method":"standard",
			"shipping_quote_id":"00000000-0000-0000-0000-000000000001",
			"selected_quote_plan_id":"00000000-0000-0000-0000-000000000002"
		}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	var req CreateOrderRequest
	err := context.ShouldBindJSON(&req)

	if err == nil {
		t.Fatal("expected missing expected_total to fail request validation")
	}
	if req.ExpectedTotal != nil {
		t.Fatal("expected_total should remain nil when the field is omitted")
	}
}

func TestCreateOrderRequestAcceptsZeroExpectedTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		"POST",
		"/api/v1/orders",
		bytes.NewBufferString(`{
			"items":[{"product_id":1,"quantity":1}],
			"shipping_address":{
				"first_name":"Test",
				"last_name":"Buyer",
				"address1":"1 Test Street",
				"city":"Austin",
				"postal_code":"78701",
				"country":"US",
				"phone":"+15555550123",
				"email":"buyer@example.com"
			},
			"billing_address":{
				"first_name":"Test",
				"last_name":"Buyer",
				"address1":"1 Test Street",
				"city":"Austin",
				"postal_code":"78701",
				"country":"US",
				"phone":"+15555550123",
				"email":"buyer@example.com"
			},
			"payment_method":"card",
			"shipping_method":"standard",
			"expected_total":0,
			"shipping_quote_id":"00000000-0000-0000-0000-000000000001",
			"selected_quote_plan_id":"00000000-0000-0000-0000-000000000002",
			"gift_card_code":"REDEEM-EXAMPLE"
		}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	var req CreateOrderRequest
	err := context.ShouldBindJSON(&req)

	if err != nil {
		t.Fatalf("expected valid zero expected_total request to bind, got %v", err)
	}
	if req.ExpectedTotal == nil {
		t.Fatal("expected_total should bind as a non-nil pointer for zero")
	}
	if *req.ExpectedTotal != 0 {
		t.Fatalf("expected_total = %v, want 0", *req.ExpectedTotal)
	}
	if req.GiftCardCode != "REDEEM-EXAMPLE" {
		t.Fatalf("gift_card_code = %q, want REDEEM-EXAMPLE", req.GiftCardCode)
	}
}

func TestCreateOrderRequestDefaultsBillingAddressToShippingAddress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/orders",
		bytes.NewBufferString(`{
			"items":[{"product_id":1,"quantity":1}],
			"shipping_address":{
				"first_name":"Test",
				"last_name":"Buyer",
				"address1":"1 Test Street",
				"city":"Austin",
				"postal_code":"78701",
				"country":"US",
				"phone":"+15555550123",
				"email":"buyer@example.com"
			},
			"payment_method":"card",
			"shipping_method":"standard",
			"expected_total":84,
			"shipping_quote_id":"00000000-0000-0000-0000-000000000001",
			"selected_quote_plan_id":"00000000-0000-0000-0000-000000000002"
		}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	var req CreateOrderRequest
	requireNoBindError(t, context.ShouldBindJSON(&req))
	if req.BillingAddress != nil {
		t.Fatalf("billing_address = %#v, want nil when omitted", req.BillingAddress)
	}
	shipping := addressFromRequest(req.ShippingAddress)
	billing := billingAddressFromRequest(shipping, req.BillingAddress)
	if billing != shipping {
		t.Fatalf("billing address = %#v, want shipping address %#v", billing, shipping)
	}
	if country := billingCountryFromRequest(req.ShippingAddress.Country, req.BillingAddress); country != "US" {
		t.Fatalf("billing country = %q, want US", country)
	}
}

func TestCreateOrderRequestAcceptsIndependentBillingAddress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/orders",
		bytes.NewBufferString(`{
			"items":[{"product_id":1,"quantity":1}],
			"shipping_address":{
				"first_name":"Test",
				"last_name":"Buyer",
				"address1":"1 Test Street",
				"city":"Austin",
				"postal_code":"78701",
				"country":"US",
				"phone":"+15555550123",
				"email":"buyer@example.com"
			},
			"billing_address":{
				"first_name":"Billing",
				"last_name":"Buyer",
				"address1":"2 Billing Street",
				"city":"Berlin",
				"postal_code":"10115",
				"country":"DE",
				"phone":"+4930123456",
				"email":"billing@example.com"
			},
			"payment_method":"card",
			"shipping_method":"standard",
			"expected_total":84,
			"shipping_quote_id":"00000000-0000-0000-0000-000000000001",
			"selected_quote_plan_id":"00000000-0000-0000-0000-000000000002"
		}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	var req CreateOrderRequest
	requireNoBindError(t, context.ShouldBindJSON(&req))
	if req.BillingAddress == nil {
		t.Fatal("billing_address should be present")
	}
	billing := billingAddressFromRequest(addressFromRequest(req.ShippingAddress), req.BillingAddress)
	if billing.Country != "DE" || billing.City != "Berlin" || billing.Email != "billing@example.com" {
		t.Fatalf("billing address = %#v, want independent DE billing address", billing)
	}
	if country := billingCountryFromRequest(req.ShippingAddress.Country, req.BillingAddress); country != "DE" {
		t.Fatalf("billing country = %q, want DE", country)
	}
}

func TestCreateOrderRequestRejectsIncompleteBillingAddress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/orders",
		bytes.NewBufferString(`{
			"items":[{"product_id":1,"quantity":1}],
			"shipping_address":{
				"first_name":"Test",
				"last_name":"Buyer",
				"address1":"1 Test Street",
				"city":"Austin",
				"postal_code":"78701",
				"country":"US",
				"phone":"+15555550123",
				"email":"buyer@example.com"
			},
			"billing_address":{},
			"payment_method":"card",
			"shipping_method":"standard",
			"expected_total":84,
			"shipping_quote_id":"00000000-0000-0000-0000-000000000001",
			"selected_quote_plan_id":"00000000-0000-0000-0000-000000000002"
		}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	var req CreateOrderRequest
	if err := context.ShouldBindJSON(&req); err == nil {
		t.Fatal("expected incomplete billing_address to fail request validation")
	}
}

func requireNoBindError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected request to bind, got %v", err)
	}
}

func TestRespondCreateOrderErrorMapsChangedTotalToConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondCreateOrderError(context, service.ErrOrderTotalChanged)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}

	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["code"] != "order_total_changed" {
		t.Fatalf("code = %q, want order_total_changed", payload["code"])
	}
	if payload["message"] != "Price has been updated; please review the order total" {
		t.Fatalf("message = %q", payload["message"])
	}
}

func TestRespondCreateOrderErrorMapsUnavailableShippingRate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondCreateOrderError(context, fmt.Errorf("%w: country BR", service.ErrShippingRateUnavailable))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["code"] != "shipping_rate_unavailable" {
		t.Fatalf("code = %q, want shipping_rate_unavailable", payload["code"])
	}
}

func TestRespondCreateOrderErrorMapsConsumedCheckoutCartToConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	respondCreateOrderError(context, service.ErrCheckoutCartAlreadyConsumed)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}

	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["code"] != "checkout_cart_already_consumed" {
		t.Fatalf("code = %q, want checkout_cart_already_consumed", payload["code"])
	}
	if payload["message"] != "Your cart was already submitted as an order. Check your order status before trying again." {
		t.Fatalf("message = %q", payload["message"])
	}
}

func TestPublicOrderResponseOmitsInternalDatabaseIDs(t *testing.T) {
	variantID := uint(17)
	item := orderdomain.Order{
		ID:                  42,
		OrderNumber:         "TZ-2026-ABCDEFGHIJKLMNOPQRST",
		UserID:              99,
		Currency:            "USD",
		SubtotalAmountMinor: 12345,
		TotalAmountMinor:    12345,
		Items: []orderdomain.OrderItem{
			{
				ID:            84,
				OrderID:       42,
				ProductID:     5,
				VariantID:     &variantID,
				Currency:      "USD",
				PriceMinor:    12345,
				SubtotalMinor: 12345,
				TotalMinor:    12345,
			},
		},
		CreatedAt: time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC),
	}

	payload, err := json.Marshal(publicOrderResponse(item))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"id", "user_id", "order_id"} {
		if _, exists := decoded[forbidden]; exists {
			t.Fatalf("public order response leaked %s: %s", forbidden, payload)
		}
	}
	var orderNumber string
	if err := json.Unmarshal(decoded["order_number"], &orderNumber); err != nil {
		t.Fatal(err)
	}
	if orderNumber != "TZ-2026-ABCDEFGHIJKLMNOPQRST" {
		t.Fatalf("public order response order_number = %q", orderNumber)
	}
	for _, required := range []string{"subtotal_minor", "shipping_fee_minor", "tax_minor", "discount_minor", "total_minor", "points_value_minor"} {
		if _, exists := decoded[required]; !exists {
			t.Fatalf("public order response missing %s: %s", required, payload)
		}
	}
	for _, forbidden := range []string{"subtotal_amount", "shipping_fee", "tax_amount", "discount_amount", "total_amount", "points_value"} {
		if _, exists := decoded[forbidden]; exists {
			t.Fatalf("public order response exposed legacy amount %s: %s", forbidden, payload)
		}
	}
	var publicItems []map[string]json.RawMessage
	if err := json.Unmarshal(decoded["items"], &publicItems); err != nil {
		t.Fatal(err)
	}
	for _, publicItem := range publicItems {
		for _, forbidden := range []string{"id", "order_id"} {
			if _, exists := publicItem[forbidden]; exists {
				t.Fatalf("public order item leaked %s: %s", forbidden, payload)
			}
		}
		for _, required := range []string{"currency", "price_minor", "subtotal_minor", "discount_minor", "tax_minor", "total_minor"} {
			if _, exists := publicItem[required]; !exists {
				t.Fatalf("public order item missing %s: %s", required, payload)
			}
		}
		for _, forbidden := range []string{"price", "subtotal", "discount", "tax_amount", "total"} {
			if _, exists := publicItem[forbidden]; exists {
				t.Fatalf("public order item exposed legacy amount %s: %s", forbidden, payload)
			}
		}
	}
}
