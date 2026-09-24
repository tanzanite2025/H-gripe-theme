package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/invoice"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type paypalCommercialInvoicePreviewRequest struct {
	DocumentNumber   string                         `json:"document_number"`
	DocumentDate     string                         `json:"document_date"`
	Currency         string                         `json:"currency"`
	Seller           paypalInvoicePreviewSeller     `json:"seller"`
	BillTo           paypalInvoicePreviewAddress    `json:"bill_to"`
	ShipTo           paypalInvoicePreviewAddress    `json:"ship_to"`
	Items            []paypalInvoicePreviewLineItem `json:"items"`
	PaymentMethod    string                         `json:"payment_method"`
	PaymentStatus    string                         `json:"payment_status"`
	PaymentDate      string                         `json:"payment_date"`
	PaymentReference string                         `json:"payment_reference"`
	Subtotal         string                         `json:"subtotal"`
	Shipping         string                         `json:"shipping"`
	Tax              string                         `json:"tax"`
	Discount         string                         `json:"discount"`
	Total            string                         `json:"total"`
}

// parse accepts decimal strings at the API boundary.
func (r *paypalCommercialInvoicePreviewRequest) UnmarshalJSON(data []byte) error {
	type alias paypalCommercialInvoicePreviewRequest
	var raw alias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = paypalCommercialInvoicePreviewRequest(raw)
	if r.Subtotal == "" {
		r.Subtotal = "0"
	}
	if r.Shipping == "" {
		r.Shipping = "0"
	}
	if r.Tax == "" {
		r.Tax = "0"
	}
	if r.Discount == "" {
		r.Discount = "0"
	}
	if r.Total == "" {
		r.Total = "0"
	}
	return nil
}

type paypalInvoicePreviewSeller struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
	TaxID   string `json:"tax_id"`
}

type paypalInvoicePreviewAddress struct {
	Name       string `json:"name"`
	Company    string `json:"company"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
}

type paypalInvoicePreviewLineItem struct {
	Description string `json:"description"`
	SKU         string `json:"sku"`
	Quantity    int    `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
	Subtotal    string `json:"subtotal"`
	Tax         string `json:"tax"`
	Discount    string `json:"discount"`
	Total       string `json:"total"`
}

func normalizeInvoiceAmount(value string) string {
	if strings.TrimSpace(value) == "" {
		return "0"
	}
	return value
}

func (r paypalCommercialInvoicePreviewRequest) commercialInvoice() (invoice.CommercialInvoice, error) {
	documentDate, err := parseInvoicePreviewDate(r.DocumentDate, time.Now().UTC())
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	currencyCode, err := currency.ParseCode(r.Currency)
	if err != nil {
		return invoice.CommercialInvoice{}, fmt.Errorf("invalid invoice currency: %w", err)
	}
	toMoney := func(amount string) (domainmoney.Money, error) {
		return domainmoney.ParseMajor(amount, currencyCode.String())
	}
	formatMoney := func(amount domainmoney.Money) (string, error) {
		return amount.FormatMajor()
	}
	items := make([]invoice.LineItem, 0, len(r.Items))
	calculatedSubtotal := domainmoney.MustNew(0, currencyCode.String())
	for _, item := range r.Items {
		item.UnitPrice = normalizeInvoiceAmount(item.UnitPrice)
		item.Subtotal = normalizeInvoiceAmount(item.Subtotal)
		item.Tax = normalizeInvoiceAmount(item.Tax)
		item.Discount = normalizeInvoiceAmount(item.Discount)
		item.Total = normalizeInvoiceAmount(item.Total)
		quantity := item.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		unitPriceMoney, err := toMoney(item.UnitPrice)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		subtotalMoney, err := toMoney(item.Subtotal)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		if item.Subtotal == "0" && item.UnitPrice != "0" {
			subtotalMoney, err = unitPriceMoney.MultiplyInt(int64(quantity))
			if err != nil {
				return invoice.CommercialInvoice{}, err
			}
		}
		taxMoney, err := toMoney(item.Tax)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		discountMoney, err := toMoney(item.Discount)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		totalMoney, err := toMoney(item.Total)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		if item.Total == "0" {
			totalMoney, err = subtotalMoney.Add(taxMoney)
			if err == nil {
				totalMoney, err = totalMoney.Subtract(discountMoney)
			}
			if err != nil {
				return invoice.CommercialInvoice{}, err
			}
		}
		calculatedSubtotal, err = calculatedSubtotal.Add(subtotalMoney)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		unitPrice, err := formatMoney(unitPriceMoney)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		subtotal, err := formatMoney(subtotalMoney)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		tax, err := formatMoney(taxMoney)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		discount, err := formatMoney(discountMoney)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		total, err := formatMoney(totalMoney)
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		items = append(items, invoice.LineItem{
			Description: item.Description,
			SKU:         item.SKU,
			Quantity:    quantity,
			UnitPrice:   unitPrice,
			Subtotal:    subtotal,
			Tax:         tax,
			Discount:    discount,
			Total:       total,
		})
	}

	subtotalMoney, err := toMoney(r.Subtotal)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	if r.Subtotal == "0" {
		subtotalMoney = calculatedSubtotal
	}
	shippingMoney, err := toMoney(r.Shipping)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	taxMoney, err := toMoney(r.Tax)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	discountMoney, err := toMoney(r.Discount)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	totalMoney, err := toMoney(r.Total)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	if r.Total == "0" {
		totalMoney, err = subtotalMoney.Add(shippingMoney)
		if err == nil {
			totalMoney, err = totalMoney.Add(taxMoney)
		}
		if err == nil {
			totalMoney, err = totalMoney.Subtract(discountMoney)
		}
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
	}
	subtotal, err := formatMoney(subtotalMoney)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	shipping, err := formatMoney(shippingMoney)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	tax, err := formatMoney(taxMoney)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	discount, err := formatMoney(discountMoney)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	total, err := formatMoney(totalMoney)
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}
	var paymentDate *time.Time
	if strings.TrimSpace(r.PaymentDate) != "" {
		parsed, err := parseInvoicePreviewDate(r.PaymentDate, time.Time{})
		if err != nil {
			return invoice.CommercialInvoice{}, err
		}
		paymentDate = &parsed
	}

	return invoice.CommercialInvoice{
		DocumentNumber: r.DocumentNumber,
		DocumentDate:   documentDate,
		Currency:       r.Currency,
		Seller: invoice.SellerProfile{
			Name:    r.Seller.Name,
			Address: r.Seller.Address,
			Email:   r.Seller.Email,
			Phone:   r.Seller.Phone,
			Website: r.Seller.Website,
			TaxID:   r.Seller.TaxID,
		},
		BillTo:           previewInvoiceAddress(r.BillTo),
		ShipTo:           previewInvoiceAddress(r.ShipTo),
		Items:            items,
		PaymentMethod:    r.PaymentMethod,
		PaymentStatus:    r.PaymentStatus,
		PaymentDate:      paymentDate,
		PaymentReference: r.PaymentReference,
		Subtotal:         subtotal,
		Shipping:         shipping,
		Tax:              tax,
		Discount:         discount,
		Total:            total,
	}, nil
}

func previewInvoiceAddress(address paypalInvoicePreviewAddress) invoice.Address {
	return invoice.Address{
		Name:       address.Name,
		Company:    address.Company,
		Line1:      address.Line1,
		Line2:      address.Line2,
		City:       address.City,
		State:      address.State,
		PostalCode: address.PostalCode,
		Country:    address.Country,
		Phone:      address.Phone,
		Email:      address.Email,
	}
}

func parseInvoicePreviewDate(value string, fallback time.Time) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, errors.New("invoice preview date must use YYYY-MM-DD or RFC3339")
}

func (h *PaymentHandler) PreviewPayPalDisputeCommercialInvoicePDF(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid PayPal dispute id")
		return
	}
	pdf, err := h.paymentService.BuildPayPalDisputeCommercialInvoicePDF(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrPayPalDisputeNotFound) {
			apierror.RespondNotFound(c, "PayPal dispute")
			return
		}
		if errors.Is(err, service.ErrPayPalDisputeInvoiceUnavailable) {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "paypal_dispute_invoice_unavailable", err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	c.Header("Cache-Control", "private, no-store")
	c.Header("Content-Disposition", "inline; filename="+strconv.Quote(pdf.Filename))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "application/pdf", pdf.Bytes)
}

func (h *PaymentHandler) PreviewPayPalCommercialInvoicePDF(c *gin.Context) {
	var req paypalCommercialInvoicePreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	document, err := req.commercialInvoice()
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	pdf, err := h.paymentService.RenderPayPalCommercialInvoicePreview(document)
	if err != nil {
		apierror.RespondError(c, http.StatusUnprocessableEntity, "paypal_invoice_preview_unavailable", err.Error())
		return
	}

	c.Header("Cache-Control", "private, no-store")
	c.Header("Content-Disposition", "inline; filename="+strconv.Quote(pdf.Filename))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "application/pdf", pdf.Bytes)
}

func (h *PaymentHandler) GetPayPalDisputeInvoiceSellerProfile(c *gin.Context) {
	if h == nil || h.paypalInvoiceSellerProfileService == nil {
		apierror.RespondInternalError(c, errors.New("PayPal commercial invoice seller profile service is unavailable"))
		return
	}
	profile, err := h.paypalInvoiceSellerProfileService.Get()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, profile)
}

func (h *PaymentHandler) UpdatePayPalDisputeInvoiceSellerProfile(c *gin.Context) {
	if h == nil || h.paypalInvoiceSellerProfileService == nil {
		apierror.RespondInternalError(c, errors.New("PayPal commercial invoice seller profile service is unavailable"))
		return
	}
	var request setting.PayPalDisputeInvoiceSellerProfileUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	profile, err := h.paypalInvoiceSellerProfileService.Update(request)
	if err != nil {
		if errors.Is(err, service.ErrPayPalDisputeInvoiceSellerProfileIncomplete) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, profile)
}
