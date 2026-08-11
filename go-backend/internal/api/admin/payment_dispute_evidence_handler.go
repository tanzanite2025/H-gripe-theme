package admin

import (
	"errors"
	"strconv"

	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) GetStripeDisputeEvidence(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid dispute id")
		return
	}
	evidencePackage, err := h.paymentService.BuildStripeDisputeEvidencePackage(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrStripeDisputeNotFound) {
			apierror.RespondNotFound(c, "Stripe dispute")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, evidencePackage)
}

func (h *PaymentHandler) SubmitStripeDisputeEvidence(c *gin.Context) {
	startedAt := paymentAuditStartedAt()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionSubmit,
			Resource:     paymentAuditResourceDisputeEvidence,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "invalid dispute id",
			Changes: map[string]interface{}{
				"raw_dispute_id": c.Param("id"),
			},
		})
		apierror.RespondBadRequest(c, "invalid dispute id")
		return
	}
	disputeID := uint(id)
	var req struct {
		Confirm                      bool   `json:"confirm"`
		Submit                       *bool  `json:"submit"`
		IncludeCustomerCommunication bool   `json:"include_customer_communication"`
		AdditionalStatement          string `json:"additional_statement"`
		ShippingDocumentationFileID  string `json:"shipping_documentation_file_id"`
		CustomerCommunicationFileID  string `json:"customer_communication_file_id"`
		ReceiptFileID                string `json:"receipt_file_id"`
		UncategorizedFileID          string `json:"uncategorized_file_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionSubmit,
			Resource:     paymentAuditResourceDisputeEvidence,
			ResourceID:   disputeID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      stripeDisputeEvidenceAuditDetails(disputeID, true, false, "", "", "", "", "", nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	submit := true
	if req.Submit != nil {
		submit = *req.Submit
	}
	if !req.Confirm {
		details := stripeDisputeEvidenceAuditDetails(
			disputeID,
			submit,
			req.IncludeCustomerCommunication,
			req.AdditionalStatement,
			req.ShippingDocumentationFileID,
			req.CustomerCommunicationFileID,
			req.ReceiptFileID,
			req.UncategorizedFileID,
			nil,
		)
		details["confirmation_matched"] = false
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionSubmit,
			Resource:     paymentAuditResourceDisputeEvidence,
			ResourceID:   disputeID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "confirmation is required before submitting dispute evidence",
			Changes:      details,
		})
		apierror.RespondBadRequest(c, "confirmation is required before submitting dispute evidence")
		return
	}

	config, err := h.adminStripeGatewayConfig()
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionSubmit,
			Resource:     paymentAuditResourceDisputeEvidence,
			ResourceID:   disputeID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: stripeDisputeEvidenceAuditDetails(
				disputeID,
				submit,
				req.IncludeCustomerCommunication,
				req.AdditionalStatement,
				req.ShippingDocumentationFileID,
				req.CustomerCommunicationFileID,
				req.ReceiptFileID,
				req.UncategorizedFileID,
				nil,
			),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.paymentService.SubmitStripeDisputeEvidence(c.Request.Context(), service.SubmitStripeDisputeEvidenceInput{
		DisputeID:                    disputeID,
		APIKey:                       config.APIKey,
		Confirm:                      req.Confirm,
		Submit:                       submit,
		IncludeCustomerCommunication: req.IncludeCustomerCommunication,
		AdditionalStatement:          req.AdditionalStatement,
		ShippingDocumentationFileID:  req.ShippingDocumentationFileID,
		CustomerCommunicationFileID:  req.CustomerCommunicationFileID,
		ReceiptFileID:                req.ReceiptFileID,
		UncategorizedFileID:          req.UncategorizedFileID,
	})
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionSubmit,
			Resource:     paymentAuditResourceDisputeEvidence,
			ResourceID:   disputeID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: stripeDisputeEvidenceAuditDetails(
				disputeID,
				submit,
				req.IncludeCustomerCommunication,
				req.AdditionalStatement,
				req.ShippingDocumentationFileID,
				req.CustomerCommunicationFileID,
				req.ReceiptFileID,
				req.UncategorizedFileID,
				result,
			),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionSubmit,
		Resource:   paymentAuditResourceDisputeEvidence,
		ResourceID: disputeID,
		Status:     paymentAuditStatusSuccess,
		Changes: stripeDisputeEvidenceAuditDetails(
			disputeID,
			submit,
			req.IncludeCustomerCommunication,
			req.AdditionalStatement,
			req.ShippingDocumentationFileID,
			req.CustomerCommunicationFileID,
			req.ReceiptFileID,
			req.UncategorizedFileID,
			result,
		),
	})
	response.Success(c, result)
}

func (h *PaymentHandler) GetPayPalDisputeEvidence(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid PayPal dispute id")
		return
	}
	evidencePackage, err := h.paymentService.BuildPayPalDisputeEvidencePackage(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrPayPalDisputeNotFound) {
			apierror.RespondNotFound(c, "PayPal dispute")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, evidencePackage)
}
