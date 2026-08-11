package invoice

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	orderdomain "commerce-platform/internal/domain/order"

	"github.com/go-pdf/fpdf"
)

// SellerProfile contains the seller identity printed on a commercial invoice.
// TaxID is optional because this document is not a statutory tax invoice.
type SellerProfile struct {
	Name    string
	Address string
	Email   string
	Phone   string
	Website string
	TaxID   string
}

func (p SellerProfile) Configured() bool {
	return strings.TrimSpace(p.Name) != "" && strings.TrimSpace(p.Address) != ""
}

// Address is the printable address snapshot used by the PDF renderer.
type Address struct {
	Name       string
	Company    string
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
	Phone      string
	Email      string
}

// LineItem is copied from the order at document generation time.
type LineItem struct {
	Description string
	SKU         string
	Quantity    int
	UnitPrice   float64
	Subtotal    float64
	Tax         float64
	Discount    float64
	Total       float64
}

// CommercialInvoice is an immutable invoice/receipt snapshot for evidence.
type CommercialInvoice struct {
	DocumentNumber   string
	DocumentDate     time.Time
	Currency         string
	Seller           SellerProfile
	BillTo           Address
	ShipTo           Address
	Items            []LineItem
	PaymentMethod    string
	PaymentStatus    string
	PaymentDate      *time.Time
	PaymentReference string
	Subtotal         float64
	Shipping         float64
	Tax              float64
	Discount         float64
	Total            float64
	Disclaimer       string
}

func SellerProfileFromEnv() SellerProfile {
	return SellerProfile{
		Name:    strings.TrimSpace(os.Getenv("PAYPAL_DISPUTE_INVOICE_SELLER_NAME")),
		Address: strings.TrimSpace(os.Getenv("PAYPAL_DISPUTE_INVOICE_SELLER_ADDRESS")),
		Email:   strings.TrimSpace(os.Getenv("PAYPAL_DISPUTE_INVOICE_SELLER_EMAIL")),
		Phone:   strings.TrimSpace(os.Getenv("PAYPAL_DISPUTE_INVOICE_SELLER_PHONE")),
		Website: strings.TrimSpace(os.Getenv("PAYPAL_DISPUTE_INVOICE_SELLER_WEBSITE")),
		TaxID:   strings.TrimSpace(os.Getenv("PAYPAL_DISPUTE_INVOICE_SELLER_TAX_ID")),
	}
}

func BuildFromOrder(orderRecord *orderdomain.Order, seller SellerProfile, paymentReference string, generatedAt time.Time) (CommercialInvoice, error) {
	if orderRecord == nil {
		return CommercialInvoice{}, errors.New("order is required")
	}
	if !seller.Configured() {
		return CommercialInvoice{}, errors.New("seller name and address are required for a commercial invoice")
	}
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}
	documentDate := orderRecord.CreatedAt
	if documentDate.IsZero() {
		documentDate = generatedAt
	}

	items := make([]LineItem, 0, len(orderRecord.Items))
	for _, item := range orderRecord.Items {
		items = append(items, LineItem{
			Description: item.ProductName,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.Price,
			Subtotal:    item.Subtotal,
			Tax:         item.TaxAmount,
			Discount:    item.Discount,
			Total:       item.Total,
		})
	}

	document := CommercialInvoice{
		DocumentNumber:   commercialInvoiceNumber(orderRecord.OrderNumber),
		DocumentDate:     documentDate.UTC(),
		Currency:         strings.ToUpper(strings.TrimSpace(orderRecord.Currency)),
		Seller:           seller,
		BillTo:           addressFromOrder(orderRecord.BillingAddress),
		ShipTo:           addressFromOrder(orderRecord.ShippingAddress),
		Items:            items,
		PaymentMethod:    orderRecord.PaymentMethod,
		PaymentStatus:    orderRecord.PaymentStatus,
		PaymentDate:      orderRecord.PaidAt,
		PaymentReference: strings.TrimSpace(paymentReference),
		Subtotal:         orderRecord.SubtotalAmount,
		Shipping:         orderRecord.ShippingFee,
		Tax:              orderRecord.TaxAmount,
		Discount:         orderRecord.DiscountAmount,
		Total:            orderRecord.TotalAmount,
		Disclaimer:       "Commercial invoice / order receipt prepared for payment dispute evidence. This document is not a statutory tax invoice unless the seller's applicable tax requirements are satisfied.",
	}
	if document.Currency == "" {
		document.Currency = "USD"
	}
	return document, document.Validate()
}

func (d CommercialInvoice) Validate() error {
	if strings.TrimSpace(d.DocumentNumber) == "" {
		return errors.New("commercial invoice document number is required")
	}
	if d.DocumentDate.IsZero() {
		return errors.New("commercial invoice document date is required")
	}
	if !d.Seller.Configured() {
		return errors.New("commercial invoice seller profile is incomplete")
	}
	if strings.TrimSpace(d.Currency) == "" {
		return errors.New("commercial invoice currency is required")
	}
	if len(d.Items) == 0 {
		return errors.New("commercial invoice requires at least one line item")
	}
	return nil
}

func RenderCommercialInvoicePDF(document CommercialInvoice, fontPath string) ([]byte, error) {
	if err := document.Validate(); err != nil {
		return nil, err
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(16, 16, 16)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AliasNbPages("")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-14)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(110, 110, 110)
		pdf.CellFormat(178, 5, fmt.Sprintf("Commercial invoice %s | Page %d/{nb}", document.DocumentNumber, pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	fontFamily := "Arial"
	if strings.TrimSpace(fontPath) != "" {
		pdf.AddUTF8Font("Invoice", "", fontPath)
		pdf.AddUTF8Font("Invoice", "B", fontPath)
		if err := pdf.Error(); err != nil {
			return nil, fmt.Errorf("load invoice font: %w", err)
		}
		fontFamily = "Invoice"
	}
	setFont := func(style string, size float64) {
		pdf.SetFont(fontFamily, style, size)
	}
	text := func(value string) string {
		if fontFamily == "Invoice" {
			return normalizePDFWhitespace(value)
		}
		return asciiPDFText(value)
	}

	pdf.AddPage()
	pdf.SetTextColor(28, 36, 48)
	setFont("B", 18)
	pdf.CellFormat(108, 9, text(document.Seller.Name), "", 0, "L", false, 0, "")
	setFont("B", 20)
	pdf.CellFormat(70, 9, "COMMERCIAL INVOICE", "", 1, "R", false, 0, "")
	setFont("", 9)
	pdf.SetTextColor(80, 88, 98)
	pdf.MultiCell(108, 4.5, text(document.Seller.Address), "", "L", false)
	for _, line := range []string{
		document.Seller.Email,
		document.Seller.Phone,
		document.Seller.Website,
	} {
		if strings.TrimSpace(line) != "" {
			pdf.CellFormat(108, 4.5, text(line), "", 1, "L", false, 0, "")
		}
	}
	if strings.TrimSpace(document.Seller.TaxID) != "" {
		pdf.CellFormat(108, 4.5, text("Tax ID: "+document.Seller.TaxID), "", 1, "L", false, 0, "")
	}

	pdf.SetY(31)
	pdf.SetX(124)
	setFont("", 9)
	pdf.SetTextColor(45, 52, 62)
	pdf.CellFormat(62, 5, text("Document number: "+document.DocumentNumber), "", 1, "R", false, 0, "")
	pdf.SetX(124)
	pdf.CellFormat(62, 5, text("Document date: "+document.DocumentDate.Format("2006-01-02")), "", 1, "R", false, 0, "")
	if document.PaymentReference != "" {
		pdf.SetX(124)
		pdf.CellFormat(62, 5, text("Payment reference: "+document.PaymentReference), "", 1, "R", false, 0, "")
	}

	pdf.SetY(66)
	drawAddressBox(pdf, "BILL TO", document.BillTo, text, setFont)
	pdf.SetXY(108, 66)
	drawAddressBox(pdf, "SHIP TO", document.ShipTo, text, setFont)

	pdf.SetY(111)
	setFont("B", 10)
	pdf.SetTextColor(28, 36, 48)
	pdf.CellFormat(178, 7, "ORDER DETAILS", "", 1, "L", false, 0, "")
	setFont("", 9)
	pdf.SetTextColor(80, 88, 98)
	paymentLine := fmt.Sprintf("Payment method: %s | Payment status: %s", textOrDash(document.PaymentMethod), textOrDash(document.PaymentStatus))
	if document.PaymentDate != nil {
		paymentLine += " | Paid: " + document.PaymentDate.UTC().Format("2006-01-02")
	}
	pdf.CellFormat(178, 5, text(paymentLine), "", 1, "L", false, 0, "")

	pdf.Ln(5)
	drawTableHeader(pdf, text, setFont)
	for _, item := range document.Items {
		rowHeight := itemRowHeight(pdf, item, text)
		if pdf.GetY()+rowHeight > 267 {
			pdf.AddPage()
			pdf.SetY(18)
			drawTableHeader(pdf, text, setFont)
		}
		drawItemRow(pdf, item, text, setFont)
	}

	pdf.Ln(5)
	drawTotals(pdf, document, text, setFont)

	pdf.Ln(7)
	setFont("I", 8)
	pdf.SetTextColor(100, 108, 118)
	pdf.MultiCell(178, 4.5, text(document.Disclaimer), "", "L", false)

	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, fmt.Errorf("render commercial invoice PDF: %w", err)
	}
	return output.Bytes(), nil
}

func addressFromOrder(address orderdomain.Address) Address {
	return Address{
		Name:       strings.TrimSpace(strings.Join([]string{address.FirstName, address.LastName}, " ")),
		Company:    strings.TrimSpace(address.Company),
		Line1:      strings.TrimSpace(address.Address1),
		Line2:      strings.TrimSpace(address.Address2),
		City:       strings.TrimSpace(address.City),
		State:      strings.TrimSpace(address.State),
		PostalCode: strings.TrimSpace(address.PostalCode),
		Country:    strings.TrimSpace(address.Country),
		Phone:      strings.TrimSpace(address.Phone),
		Email:      strings.TrimSpace(address.Email),
	}
}

func commercialInvoiceNumber(orderNumber string) string {
	orderNumber = strings.TrimSpace(orderNumber)
	if orderNumber == "" {
		return "CI-ORDER"
	}
	return "CI-" + orderNumber
}

func normalizePDFWhitespace(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	var builder strings.Builder
	for _, r := range value {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			builder.WriteRune(r)
		}
	}
	return strings.TrimSpace(builder.String())
}

func asciiPDFText(value string) string {
	value = normalizePDFWhitespace(value)
	var builder strings.Builder
	for _, r := range value {
		if r == '\n' || r == '\t' || (r >= 32 && r <= 126) {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('?')
	}
	return builder.String()
}

func textOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
