package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tanzanite/internal/domain/setting"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/invoice"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

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
	Subtotal         float64                        `json:"subtotal"`
	Shipping         float64                        `json:"shipping"`
	Tax              float64                        `json:"tax"`
	Discount         float64                        `json:"discount"`
	Total            float64                        `json:"total"`
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
	Description string  `json:"description"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
	Tax         float64 `json:"tax"`
	Discount    float64 `json:"discount"`
	Total       float64 `json:"total"`
}

func (r paypalCommercialInvoicePreviewRequest) commercialInvoice() (invoice.CommercialInvoice, error) {
	documentDate, err := parseInvoicePreviewDate(r.DocumentDate, time.Now().UTC())
	if err != nil {
		return invoice.CommercialInvoice{}, err
	}

	items := make([]invoice.LineItem, 0, len(r.Items))
	var calculatedSubtotal float64
	for _, item := range r.Items {
		quantity := item.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		subtotal := item.Subtotal
		if subtotal == 0 && item.UnitPrice != 0 {
			subtotal = item.UnitPrice * float64(quantity)
		}
		total := item.Total
		if total == 0 {
			total = subtotal + item.Tax - item.Discount
		}
		calculatedSubtotal += subtotal
		items = append(items, invoice.LineItem{
			Description: item.Description,
			SKU:         item.SKU,
			Quantity:    quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    subtotal,
			Tax:         item.Tax,
			Discount:    item.Discount,
			Total:       total,
		})
	}

	subtotal := r.Subtotal
	if subtotal == 0 {
		subtotal = calculatedSubtotal
	}
	total := r.Total
	if total == 0 {
		total = subtotal + r.Shipping + r.Tax - r.Discount
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
		Shipping:         r.Shipping,
		Tax:              r.Tax,
		Discount:         r.Discount,
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
