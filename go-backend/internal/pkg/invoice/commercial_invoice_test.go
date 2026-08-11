package invoice

import (
	"bytes"
	"testing"
	"time"

	"tanzanite/internal/domain/order"

	"github.com/go-pdf/fpdf"
	"github.com/stretchr/testify/require"
)

func TestBuildFromOrderCreatesCommercialInvoiceSnapshot(t *testing.T) {
	paidAt := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	orderRecord := &order.Order{
		OrderNumber:    "ORD-1001",
		PaymentMethod:  "paypal",
		PaymentStatus:  "paid",
		SubtotalAmount: 100,
		ShippingFee:    15,
		TaxAmount:      10,
		DiscountAmount: 5,
		TotalAmount:    120,
		Currency:       "usd",
		CreatedAt:      time.Date(2026, 8, 9, 9, 0, 0, 0, time.UTC),
		PaidAt:         &paidAt,
		BillingAddress: order.Address{
			FirstName:  "Jane",
			LastName:   "Buyer",
			Address1:   "1 Main Street",
			City:       "Austin",
			State:      "TX",
			PostalCode: "78701",
			Country:    "US",
		},
		ShippingAddress: order.Address{
			FirstName:  "Jane",
			LastName:   "Buyer",
			Address1:   "1 Main Street",
			City:       "Austin",
			State:      "TX",
			PostalCode: "78701",
			Country:    "US",
		},
		Items: []order.OrderItem{{
			ProductName: "Carbon wheel",
			SKU:         "CW-01",
			Quantity:    2,
			Price:       50,
			Subtotal:    100,
			Total:       100,
		}},
	}

	document, err := BuildFromOrder(orderRecord, SellerProfile{
		Name:    "Tanzanite Factory",
		Address: "100 Factory Road\nAustin, TX 78701\nUS",
	}, "PAYPAL-TXN-1", time.Time{})
	require.NoError(t, err)
	require.Equal(t, "CI-ORD-1001", document.DocumentNumber)
	require.Equal(t, "USD", document.Currency)
	require.Equal(t, "Jane Buyer", document.BillTo.Name)
	require.Equal(t, "PAYPAL-TXN-1", document.PaymentReference)
	require.Len(t, document.Items, 1)
}

func TestRenderCommercialInvoicePDFProducesReadablePDF(t *testing.T) {
	document := CommercialInvoice{
		DocumentNumber: "CI-ORD-1002",
		DocumentDate:   time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC),
		Currency:       "USD",
		Seller: SellerProfile{
			Name:    "Tanzanite Factory",
			Address: "100 Factory Road\nAustin, TX 78701\nUS",
		},
		BillTo: Address{Name: "Jane Buyer", Line1: "1 Main Street", City: "Austin", State: "TX", PostalCode: "78701", Country: "US"},
		ShipTo: Address{Name: "Jane Buyer", Line1: "1 Main Street", City: "Austin", State: "TX", PostalCode: "78701", Country: "US"},
		Items: []LineItem{{
			Description: "Carbon wheel",
			SKU:         "CW-01",
			Quantity:    2,
			UnitPrice:   50,
			Total:       100,
		}},
		Subtotal: 100,
		Total:    100,
	}

	pdfBytes, err := RenderCommercialInvoicePDF(document, "")
	require.NoError(t, err)
	require.Greater(t, len(pdfBytes), 500)
	require.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")))

	reader := fpdf.New("P", "mm", "A4", "")
	require.NotNil(t, reader)
}
