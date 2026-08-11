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

type PaymentRefundRecommendationHandler struct {
	recommendations *service.PaymentRefundRecommendationService
	auditService    paymentAuditRecorder
}

func NewPaymentRefundRecommendationHandler(recommendations *service.PaymentRefundRecommendationService) *PaymentRefundRecommendationHandler {
	return &PaymentRefundRecommendationHandler{recommendations: recommendations}
}

func (h *PaymentRefundRecommendationHandler) ListRecommendations(c *gin.Context) {
	if h == nil || h.recommendations == nil {
		apierror.RespondInternalError(c, errors.New("payment refund recommendation service is not configured"))
		return
	}
	params := pagination.ParsePagination(c)
	recommendations, total, err := h.recommendations.ListRecommendations(
		strings.TrimSpace(c.Query("status")),
		strings.TrimSpace(c.Query("provider")),
		params.Page,
		params.PageSize,
	)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Paged(c, recommendations, params.Page, params.PageSize, total)
}

func (h *PaymentRefundRecommendationHandler) UpdateRecommendation(c *gin.Context) {
	if h == nil {
		apierror.RespondInternalError(c, errors.New("payment refund recommendation service is not configured"))
		return
	}
	startedAt := paymentAuditStartedAt()
	if h.recommendations == nil {
		err := errors.New("payment refund recommendation service is not configured")
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourceRefundRecommendation,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundRecommendationAuditDetails(0, "", "", nil, nil),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	adminID, ok := currentAdminUserID(c)
	if !ok {
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourceRefundRecommendation,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "admin user id is required",
			Changes:      paymentRefundRecommendationAuditDetails(0, "", "", nil, nil),
		})
		apierror.RespondUnauthorized(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourceRefundRecommendation,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "invalid refund recommendation id",
			Changes: map[string]interface{}{
				"raw_recommendation_id": c.Param("id"),
			},
		})
		apierror.RespondBadRequest(c, "invalid refund recommendation id")
		return
	}
	recommendationID := uint(id)
	var req struct {
		Status        string `json:"status" binding:"required"`
		DecisionNotes string `json:"decision_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourceRefundRecommendation,
			ResourceID:   recommendationID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundRecommendationAuditDetails(recommendationID, "", "", nil, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	record, err := h.recommendations.UpdateRecommendationDecision(
		recommendationID,
		req.Status,
		req.DecisionNotes,
		adminID,
	)
	if err != nil {
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourceRefundRecommendation,
			ResourceID:   recommendationID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundRecommendationAuditDetails(recommendationID, req.Status, req.DecisionNotes, nil, nil),
		})
		if errors.Is(err, service.ErrPaymentRefundRecommendationNotFound) {
			apierror.RespondNotFound(c, "Refund recommendation")
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionUpdate,
		Resource:   paymentAuditResourceRefundRecommendation,
		ResourceID: recommendationID,
		Status:     paymentAuditStatusSuccess,
		Changes:    paymentRefundRecommendationAuditDetails(recommendationID, req.Status, req.DecisionNotes, record, nil),
	})
	response.Success(c, record)
}

func (h *PaymentRefundRecommendationHandler) CreatePendingRefund(c *gin.Context) {
	if h == nil {
		apierror.RespondInternalError(c, errors.New("payment refund recommendation service is not configured"))
		return
	}
	startedAt := paymentAuditStartedAt()
	if h.recommendations == nil {
		err := errors.New("payment refund recommendation service is not configured")
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundRecommendation,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundRecommendationAuditDetails(0, "", "", nil, nil),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	adminID, ok := currentAdminUserID(c)
	if !ok {
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundRecommendation,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "admin user id is required",
			Changes:      paymentRefundRecommendationAuditDetails(0, "", "", nil, nil),
		})
		apierror.RespondUnauthorized(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundRecommendation,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "invalid refund recommendation id",
			Changes: map[string]interface{}{
				"raw_recommendation_id": c.Param("id"),
			},
		})
		apierror.RespondBadRequest(c, "invalid refund recommendation id")
		return
	}
	recommendationID := uint(id)
	var req struct {
		Amount        float64 `json:"amount"`
		Reason        string  `json:"reason"`
		DecisionNotes string  `json:"decision_notes"`
		Confirm       bool    `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundRecommendation,
			ResourceID:   recommendationID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundRecommendationAuditDetails(recommendationID, "", "", nil, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if !req.Confirm {
		details := paymentRefundRecommendationAuditDetails(recommendationID, "", req.DecisionNotes, nil, nil)
		details["confirmation_matched"] = false
		details["requested_amount"] = req.Amount
		details["reason_present"] = strings.TrimSpace(req.Reason) != ""
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundRecommendation,
			ResourceID:   recommendationID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "confirmation is required before creating a pending refund",
			Changes:      details,
		})
		apierror.RespondBadRequest(c, "confirmation is required before creating a pending refund")
		return
	}

	recommendation, refund, err := h.recommendations.CreatePendingRefundFromRecommendation(
		service.CreatePendingRefundFromRecommendationInput{
			RecommendationID: recommendationID,
			Amount:           req.Amount,
			Reason:           req.Reason,
			DecisionNotes:    req.DecisionNotes,
			AdminID:          adminID,
		},
	)
	if err != nil {
		details := paymentRefundRecommendationAuditDetails(recommendationID, "", req.DecisionNotes, nil, nil)
		details["requested_amount"] = req.Amount
		details["reason_present"] = strings.TrimSpace(req.Reason) != ""
		h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceRefundRecommendation,
			ResourceID:   recommendationID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      details,
		})
		if errors.Is(err, service.ErrPaymentRefundRecommendationNotFound) {
			apierror.RespondNotFound(c, "Refund recommendation")
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	details := paymentRefundRecommendationAuditDetails(recommendationID, "", req.DecisionNotes, recommendation, refund)
	details["requested_amount"] = req.Amount
	details["reason_present"] = strings.TrimSpace(req.Reason) != ""
	h.recordRefundRecommendationAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionCreate,
		Resource:   paymentAuditResourceRefundRecommendation,
		ResourceID: recommendationID,
		Status:     paymentAuditStatusSuccess,
		Changes:    details,
	})
	response.Created(c, gin.H{
		"recommendation": recommendation,
		"refund":         refund,
	})
}
