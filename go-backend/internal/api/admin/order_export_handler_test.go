package admin

import (
	orderdomain "commerce-platform/internal/domain/order"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOrderCustomsExportRow(t *testing.T) {
	declaredValue := 18.5
	record := orderdomain.Order{
		OrderNumber: "TZ-2026-ORDER",
		ShippingAddress: orderdomain.Address{
			FirstName:  "Ada",
			LastName:   "Lovelace",
			Country:    "US",
			PostalCode: "10001",
		},
	}
	item := orderdomain.OrderItem{
		ID:                     7,
		ProductName:            "Wheel rim",
		SKU:                    "RIM-001",
		Quantity:               2,
		HSCode:                 "871499",
		CNCode:                 "87149990",
		CountryOfOrigin:        "CN",
		CustomsDescription:     "Bicycle parts",
		DeclaredValue:          &declaredValue,
		DeclaredValueConfirmed: true,
	}

	row := orderCustomsExportRow(record, item)
	if got, want := row[0], "TZ-2026-ORDER"; got != want {
		t.Fatalf("order number = %q, want %q", got, want)
	}
	if got, want := row[1], "Ada Lovelace"; got != want {
		t.Fatalf("recipient = %q, want %q", got, want)
	}
	if got, want := row[14], "2"; got != want {
		t.Fatalf("quantity = %q, want %q", got, want)
	}
	if got, want := row[15], "871499"; got != want {
		t.Fatalf("HS code = %q, want %q", got, want)
	}
	if got, want := row[19], "18.50"; got != want {
		t.Fatalf("declared value = %q, want %q", got, want)
	}
	if got, want := row[20], "confirmed"; got != want {
		t.Fatalf("declared value status = %q, want %q", got, want)
	}
}

func TestExportOrderCustomsRejectsIncompleteDeclaredValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler, _, _ := newOrderFulfillmentAuditHandler(t)
	orderRecord := orderdomain.Order{
		OrderNumber:   "ORDER-CUSTOMS-INCOMPLETE",
		UserID:        42,
		Status:        "processing",
		PaymentStatus: "paid",
		TotalAmount:   100,
		Currency:      "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	variantID := uint(1)
	require.NoError(t, db.Create(&orderdomain.OrderItem{
		OrderID:                orderRecord.ID,
		ProductID:              1,
		VariantID:              &variantID,
		ProductName:            "Customs export test product",
		SKU:                    "CUSTOMS-EXPORT-SKU",
		Quantity:               1,
		Price:                  100,
		Subtotal:               100,
		Total:                  100,
		DeclaredValueConfirmed: true,
	}).Error)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{
		Key:   "id",
		Value: strconv.FormatUint(uint64(orderRecord.ID), 10),
	}}
	handler.ExportOrderCustoms(context)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), "order_customs_declared_value_incomplete") {
		t.Fatalf("response = %q, want customs completeness error", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "Declared Value") {
		t.Fatalf("incomplete customs export unexpectedly produced CSV: %q", recorder.Body.String())
	}
}
