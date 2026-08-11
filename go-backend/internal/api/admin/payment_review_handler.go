package admin

import (
	"errors"
	"strconv"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) ListPaymentReviews(c *gin.Context) {
	params := pagination.ParsePagination(c)
	reviews, total, err := h.paymentService.ListPaymentReviews(
		strings.TrimSpace(c.Query("status")),
		params.Page,
		params.PageSize,
	)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Paged(c, reviews, params.Page, params.PageSize, total)
}

func (h *PaymentHandler) GetPaymentReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid payment review id")
		return
	}
	reviewRecord, err := h.paymentService.GetPaymentReview(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Payment review")
		return
	}
	response.Success(c, reviewRecord)
}

func (h *PaymentHandler) CreatePaymentReview(c *gin.Context) {
	startedAt := paymentAuditStartedAt()
	adminID, ok := currentAdminUserID(c)
	if !ok {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourcePaymentReview,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "admin user id is required",
			Changes:      paymentReviewAuditDetails(0, "pending", "", "", nil, nil, nil, "", nil, nil),
		})
		apierror.RespondUnauthorized(c)
		return
	}
	var req struct {
		OrderID         *uint  `json:"order_id"`
		TransactionID   *uint  `json:"transaction_id"`
		DisputeID       *uint  `json:"dispute_id"`
		PaymentIntentID string `json:"payment_intent_id"`
		Reason          string `json:"reason" binding:"required"`
		Notes           string `json:"notes"`
		AssignedToID    *uint  `json:"assigned_to_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourcePaymentReview,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentReviewAuditDetails(0, "pending", "", "", nil, nil, nil, "", nil, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	assignedToID := firstPaymentReviewAssignee(req.AssignedToID, adminID)
	if h == nil || h.paymentService == nil {
		err := errors.New("payment service is not configured")
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourcePaymentReview,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: paymentReviewAuditDetails(
				0,
				"pending",
				req.Reason,
				req.Notes,
				req.OrderID,
				req.TransactionID,
				req.DisputeID,
				req.PaymentIntentID,
				assignedToID,
				nil,
			),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	reviewRecord, err := h.paymentService.CreatePaymentReview(service.CreatePaymentReviewInput{
		OrderID:         req.OrderID,
		TransactionID:   req.TransactionID,
		DisputeID:       req.DisputeID,
		PaymentIntentID: req.PaymentIntentID,
		Status:          "pending",
		Reason:          req.Reason,
		Source:          "operator",
		Notes:           req.Notes,
		AssignedToID:    assignedToID,
	})
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourcePaymentReview,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: paymentReviewAuditDetails(
				0,
				"pending",
				req.Reason,
				req.Notes,
				req.OrderID,
				req.TransactionID,
				req.DisputeID,
				req.PaymentIntentID,
				assignedToID,
				nil,
			),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionCreate,
		Resource:   paymentAuditResourcePaymentReview,
		ResourceID: reviewRecord.ID,
		Status:     paymentAuditStatusSuccess,
		Changes: paymentReviewAuditDetails(
			reviewRecord.ID,
			"pending",
			req.Reason,
			req.Notes,
			req.OrderID,
			req.TransactionID,
			req.DisputeID,
			req.PaymentIntentID,
			assignedToID,
			reviewRecord,
		),
	})
	response.Created(c, reviewRecord)
}

func firstPaymentReviewAssignee(value *uint, fallback uint) *uint {
	if value != nil {
		return value
	}
	return &fallback
}

func (h *PaymentHandler) UpdatePaymentReview(c *gin.Context) {
	startedAt := paymentAuditStartedAt()
	adminID, ok := currentAdminUserID(c)
	if !ok {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentReview,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "admin user id is required",
			Changes:      paymentReviewAuditDetails(0, "", "", "", nil, nil, nil, "", nil, nil),
		})
		apierror.RespondUnauthorized(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentReview,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "invalid payment review id",
			Changes: map[string]interface{}{
				"raw_review_id": c.Param("id"),
			},
		})
		apierror.RespondBadRequest(c, "invalid payment review id")
		return
	}
	reviewID := uint(id)
	var req struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentReview,
			ResourceID:   reviewID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentReviewAuditDetails(reviewID, "", "", "", nil, nil, nil, "", nil, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.paymentService == nil {
		err := errors.New("payment service is not configured")
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentReview,
			ResourceID:   reviewID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentReviewAuditDetails(reviewID, req.Status, "", req.Notes, nil, nil, nil, "", nil, nil),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	reviewRecord, err := h.paymentService.UpdatePaymentReview(reviewID, req.Status, req.Notes, adminID)
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentReview,
			ResourceID:   reviewID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentReviewAuditDetails(reviewID, req.Status, "", req.Notes, nil, nil, nil, "", nil, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionUpdate,
		Resource:   paymentAuditResourcePaymentReview,
		ResourceID: reviewID,
		Status:     paymentAuditStatusSuccess,
		Changes:    paymentReviewAuditDetails(reviewID, req.Status, "", req.Notes, nil, nil, nil, "", nil, reviewRecord),
	})
	response.Success(c, reviewRecord)
}
