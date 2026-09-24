package admin

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentRefundExecutionHandler struct {
	paymentService *service.PaymentService
	auditService   paymentAuditRecorder
}

func NewPaymentRefundExecutionHandler(paymentService *service.PaymentService) *PaymentRefundExecutionHandler {
	return &PaymentRefundExecutionHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentRefundExecutionHandler) RequestPendingRefundExecution(c *gin.Context) {
	if h == nil {
		apierror.RespondInternalError(c, errors.New("payment service is not configured"))
		return
	}
	startedAt := paymentAuditStartedAt()
	if h.paymentService == nil {
		err := errors.New("payment service is not configured")
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundExecutionAuditDetails(0, "", nil, nil, nil),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	adminID, ok := currentAdminUserID(c)
	if !ok {
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "admin user id is required",
			Changes:      paymentRefundExecutionAuditDetails(0, "", nil, nil, nil),
		})
		apierror.RespondUnauthorized(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "invalid refund id",
			Changes: map[string]interface{}{
				"raw_refund_id": c.Param("id"),
			},
		})
		apierror.RespondBadRequest(c, "invalid refund id")
		return
	}
	refundID := uint(id)
	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			ResourceID:   refundID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundExecutionAuditDetails(refundID, "", nil, nil, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if !req.Confirm {
		details := paymentRefundExecutionAuditDetails(refundID, "", nil, nil, nil)
		details["confirmation_matched"] = false
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			ResourceID:   refundID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "confirmation is required before executing provider refund",
			Changes:      details,
		})
		apierror.RespondBadRequest(c, "confirmation is required before executing provider refund")
		return
	}

	refund, err := h.paymentService.GetRefund(refundID)
	if err != nil {
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			ResourceID:   refundID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "refund not found",
			Changes:      paymentRefundExecutionAuditDetails(refundID, "", nil, nil, nil),
		})
		apierror.RespondNotFound(c, "Refund")
		return
	}
	transaction, err := h.paymentService.GetTransaction(refund.TransactionID)
	if err != nil {
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			ResourceID:   refundID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "refund transaction not found",
			Changes:      paymentRefundExecutionAuditDetails(refundID, "", refund, nil, nil),
		})
		apierror.RespondBadRequest(c, "refund transaction not found")
		return
	}
	provider, err := pgateway.ParseGatewayType(transaction.PaymentMethod)
	if err != nil {
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			ResourceID:   refundID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundExecutionAuditDetails(refundID, transaction.PaymentMethod, refund, transaction, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	requestedRefund, execution, err := h.paymentService.RequestPendingRefundExecution(c.Request.Context(), service.RequestPendingRefundExecutionInput{
		RefundID: refundID,
		AdminID:  adminID,
		Provider: string(provider),
	})
	if err != nil {
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			ResourceID:   refundID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundExecutionAuditDetails(refundID, string(provider), refund, transaction, execution),
		})
		if errors.Is(err, service.ErrPaymentRefundExecutionInProgress) {
			apierror.RespondBadRequest(c, "refund execution is already in progress")
			return
		}
		if respondHistoricalRefundFXSnapshotError(c, err) {
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionExecute,
		Resource:   paymentAuditResourceRefundExecution,
		ResourceID: refundID,
		Status:     paymentAuditStatusSuccess,
		Changes:    paymentRefundExecutionAuditDetails(refundID, string(provider), requestedRefund, transaction, execution),
	})
	c.JSON(http.StatusAccepted, gin.H{
		"refund":    requestedRefund,
		"execution": execution,
		"status":    "processing",
	})
}
