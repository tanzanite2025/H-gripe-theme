package admin

import (
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/service"
)

type PaymentHandler struct {
	paymentService                    *service.PaymentService
	settingsService                   *service.AdminSettingsService
	paypalInvoiceSellerProfileService *service.PayPalDisputeInvoiceSellerProfileService
	auditService                      paymentAuditRecorder
	publicBaseURL                     string
	callbackProbeClient               paymentCallbackProbeHTTPClient
}

func NewPaymentHandler(paymentService *service.PaymentService, settingsService *service.AdminSettingsService, sellerProfileServices ...*service.PayPalDisputeInvoiceSellerProfileService) *PaymentHandler {
	var sellerProfileService *service.PayPalDisputeInvoiceSellerProfileService
	if len(sellerProfileServices) > 0 {
		sellerProfileService = sellerProfileServices[0]
	}
	return &PaymentHandler{
		paymentService:                    paymentService,
		settingsService:                   settingsService,
		paypalInvoiceSellerProfileService: sellerProfileService,
	}
}

func (h *PaymentHandler) ConfigurePublicBaseURL(baseURL string) {
	if h == nil {
		return
	}
	h.publicBaseURL = pgateway.NormalizePublicBaseURL(baseURL)
}
