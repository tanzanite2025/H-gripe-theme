package admin

import (
	"commerce-platform/internal/api/middleware"
	"errors"
	"strconv"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) GetRefund(c *gin.Context) {
	id, err := parseAdminUintParam(c, "id", "invalid refund id")
	if err != nil {
		return
	}

	refund, err := h.paymentService.GetRefund(id)
	if err != nil {
		apierror.RespondNotFound(c, "Refund")
		return
	}

	response.Success(c, adminRefundResponse(*refund))
}

func (h *PaymentHandler) GetOrderRefunds(c *gin.Context) {
	orderID, err := parseAdminUintParam(c, "order_id", "invalid order id")
	if err != nil {
		return
	}

	refunds, err := h.paymentService.GetOrderRefunds(orderID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	items := make([]interface{}, 0, len(refunds))
	for _, refund := range refunds {
		items = append(items, adminRefundResponse(refund))
	}
	response.Success(c, gin.H{"data": items})
}

func adminRefundResponse(refund paymentdomain.Refund) gin.H {
	return gin.H{
		"id":                             refund.ID,
		"order_id":                       refund.OrderID,
		"transaction_id":                 refund.TransactionID,
		"refund_id":                      refund.RefundID,
		"amount_minor":                   refund.AmountMinor,
		"gift_card_refund_amount_minor":  refund.GiftCardRefundAmountMinor,
		"requested_amount_minor":         refund.RequestedAmountMinor,
		"discount_clawback_amount_minor": refund.DiscountClawbackAmountMinor,
		"currency":                       refund.Currency,
		"line_items":                     refund.LineItems,
		"reason":                         refund.Reason,
		"status":                         refund.Status,
		"refunded_by":                    refund.RefundedBy,
		"created_at":                     refund.CreatedAt,
		"updated_at":                     refund.UpdatedAt,
		"completed_at":                   refund.CompletedAt,
	}
}

func (h *PaymentHandler) CreateRefund(c *gin.Context) {
	startedAt := paymentAuditStartedAt()
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundDraft,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "admin user id is required",
			Changes:      paymentRefundDraftAuditDetails(0, 0, 0, "", 0, 0, nil),
		})
		apierror.RespondUnauthorized(c)
		return
	}
	adminID, ok := userIDValue.(uint)
	if !ok || adminID == 0 {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundDraft,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "admin user id is invalid",
			Changes:      paymentRefundDraftAuditDetails(0, 0, 0, "", 0, 0, nil),
		})
		apierror.RespondUnauthorized(c)
		return
	}

	var req struct {
		OrderID       uint    `json:"order_id" binding:"required"`
		TransactionID uint    `json:"transaction_id" binding:"required"`
		Amount        float64 `json:"amount"`
		Reason        string  `json:"reason"`
		LineItems     []struct {
			OrderItemID uint `json:"order_item_id" binding:"required"`
			Quantity    int  `json:"quantity" binding:"required,gt=0"`
			Restock     bool `json:"restock"`
		} `json:"line_items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundDraft,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundDraftAuditDetails(0, 0, 0, "", 0, 0, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	restockCount := 0
	for _, item := range req.LineItems {
		if item.Restock {
			restockCount++
		}
	}

	refund := paymentdomain.Refund{
		OrderID:       req.OrderID,
		TransactionID: req.TransactionID,
		Amount:        req.Amount,
		Reason:        req.Reason,
	}
	if len(req.LineItems) > 0 {
		refund.LineItems = make([]paymentdomain.RefundLineItem, 0, len(req.LineItems))
		for _, item := range req.LineItems {
			refund.LineItems = append(refund.LineItems, paymentdomain.RefundLineItem{
				OrderItemID: item.OrderItemID,
				Quantity:    item.Quantity,
				Restock:     item.Restock,
			})
		}
	}

	if h == nil || h.paymentService == nil {
		err := errors.New("payment service is not configured")
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundDraft,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: paymentRefundDraftAuditDetails(
				req.OrderID,
				req.TransactionID,
				req.Amount,
				req.Reason,
				len(req.LineItems),
				restockCount,
				&refund,
			),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	createdRefund, err := h.paymentService.CreateAdminRefundWithIdempotency(
		&refund,
		adminID,
		middleware.GetIdempotencyKey(c),
		middleware.GetIdempotencyRequestHash(c),
	)
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundDraft,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: paymentRefundDraftAuditDetails(
				req.OrderID,
				req.TransactionID,
				req.Amount,
				req.Reason,
				len(req.LineItems),
				restockCount,
				&refund,
			),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionCreate,
		Resource:   paymentAuditResourceRefundDraft,
		ResourceID: createdRefund.ID,
		Status:     paymentAuditStatusSuccess,
		Changes: paymentRefundDraftAuditDetails(
			req.OrderID,
			req.TransactionID,
			req.Amount,
			req.Reason,
			len(req.LineItems),
			restockCount,
			createdRefund,
		),
	})
	response.Created(c, createdRefund)
}

func parseAdminUintParam(c *gin.Context, name, message string) (uint, error) {
	value, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, message)
		return 0, err
	}
	return uint(value), nil
}
