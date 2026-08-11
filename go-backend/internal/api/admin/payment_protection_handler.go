package admin

import (
	"errors"
	"strconv"
	"time"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentProtectionHandler struct {
	protection   *service.PaymentProtectionService
	auditService paymentAuditRecorder
}

func NewPaymentProtectionHandler(protection *service.PaymentProtectionService) *PaymentProtectionHandler {
	return &PaymentProtectionHandler{protection: protection}
}

func (h *PaymentProtectionHandler) ListControls(c *gin.Context) {
	if h == nil || h.protection == nil {
		apierror.RespondInternalError(c, errors.New("payment protection service is not configured"))
		return
	}
	controls, err := h.protection.ListControls(
		time.Now().UTC(),
		c.Query("include_expired") == "true",
	)
	if err != nil {
		respondPaymentProtectionError(c, err)
		return
	}
	response.Success(c, gin.H{
		"enabled":  h.protection.Enabled(),
		"controls": controls,
		"policy":   h.protection.PolicySummary(),
	})
}

func (h *PaymentProtectionHandler) CreateControl(c *gin.Context) {
	if h == nil {
		apierror.RespondInternalError(c, errors.New("payment protection service is not configured"))
		return
	}
	startedAt := paymentAuditStartedAt()
	if h.protection == nil {
		err := errors.New("payment protection service is not configured")
		h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceProtectionControl,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentProtectionControlRequestAuditDetails("", "", "", "", time.Time{}),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	var req struct {
		Action     string    `json:"action" binding:"required"`
		ScopeType  string    `json:"scope_type" binding:"required"`
		ScopeValue string    `json:"scope_value"`
		Reason     string    `json:"reason" binding:"required"`
		ExpiresAt  time.Time `json:"expires_at" binding:"required"`
		Confirm    bool      `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceProtectionControl,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentProtectionControlRequestAuditDetails("", "", "", "", time.Time{}),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if !req.Confirm {
		details := paymentProtectionControlRequestAuditDetails(req.Action, req.ScopeType, req.ScopeValue, req.Reason, req.ExpiresAt)
		details["confirmation_matched"] = false
		h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceProtectionControl,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "confirmation is required before creating a payment protection control",
			Changes:      details,
		})
		apierror.RespondBadRequest(c, "confirmation is required before creating a payment protection control")
		return
	}

	control, err := h.protection.CreateControl(
		service.CreatePaymentProtectionControlInput{
			Action:     req.Action,
			ScopeType:  req.ScopeType,
			ScopeValue: req.ScopeValue,
			Reason:     req.Reason,
			ExpiresAt:  req.ExpiresAt,
		},
		paymentProtectionActor(c),
	)
	if err != nil {
		h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourceProtectionControl,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentProtectionControlRequestAuditDetails(req.Action, req.ScopeType, req.ScopeValue, req.Reason, req.ExpiresAt),
		})
		respondPaymentProtectionError(c, err)
		return
	}
	h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionCreate,
		Resource:   paymentAuditResourceProtectionControl,
		ResourceID: control.ID,
		Status:     paymentAuditStatusSuccess,
		Changes:    paymentProtectionControlAuditDetails(control.ID, control),
	})
	response.Created(c, control)
}

func (h *PaymentProtectionHandler) RevokeControl(c *gin.Context) {
	if h == nil {
		apierror.RespondInternalError(c, errors.New("payment protection service is not configured"))
		return
	}
	startedAt := paymentAuditStartedAt()
	if h.protection == nil {
		err := errors.New("payment protection service is not configured")
		h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionRevoke,
			Resource:     paymentAuditResourceProtectionControl,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentProtectionControlAuditDetails(0, nil),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionRevoke,
			Resource:     paymentAuditResourceProtectionControl,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "invalid payment protection control id",
			Changes: map[string]interface{}{
				"raw_control_id": c.Param("id"),
			},
		})
		apierror.RespondBadRequest(c, "invalid payment protection control id")
		return
	}
	controlID := uint(id)
	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionRevoke,
			Resource:     paymentAuditResourceProtectionControl,
			ResourceID:   controlID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentProtectionControlAuditDetails(controlID, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if !req.Confirm {
		details := paymentProtectionControlAuditDetails(controlID, nil)
		details["confirmation_matched"] = false
		h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionRevoke,
			Resource:     paymentAuditResourceProtectionControl,
			ResourceID:   controlID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "confirmation is required before revoking a payment protection control",
			Changes:      details,
		})
		apierror.RespondBadRequest(c, "confirmation is required before revoking a payment protection control")
		return
	}
	control, err := h.protection.RevokeControl(controlID, paymentProtectionActor(c))
	if err != nil {
		h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionRevoke,
			Resource:     paymentAuditResourceProtectionControl,
			ResourceID:   controlID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentProtectionControlAuditDetails(controlID, nil),
		})
		respondPaymentProtectionError(c, err)
		return
	}
	h.recordProtectionControlAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionRevoke,
		Resource:   paymentAuditResourceProtectionControl,
		ResourceID: controlID,
		Status:     paymentAuditStatusSuccess,
		Changes:    paymentProtectionControlAuditDetails(controlID, control),
	})
	response.Success(c, control)
}

func (h *PaymentProtectionHandler) ListControlAuditLogs(c *gin.Context) {
	if h == nil || h.protection == nil {
		apierror.RespondInternalError(c, errors.New("payment protection service is not configured"))
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, "invalid payment protection control id")
		return
	}
	page, pageSize := paymentProtectionPagination(c)
	logs, total, err := h.protection.ListControlAuditLogs(uint(id), page, pageSize)
	if err != nil {
		respondPaymentProtectionError(c, err)
		return
	}
	response.Success(c, gin.H{
		"logs": logs,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages(total, pageSize),
		},
	})
}

func paymentProtectionActor(c *gin.Context) service.PaymentProtectionActor {
	username, _ := c.Get("username")
	return service.PaymentProtectionActor{
		UserID:    c.GetUint("user_id"),
		Username:  stringValue(username),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
	}
}

func respondPaymentProtectionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPaymentProtectionDisabled):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrInvalidPaymentProtectionControl):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrPaymentProtectionControlNotFound):
		apierror.RespondNotFound(c, "Payment protection control")
	case errors.Is(err, service.ErrPaymentProtectionControlRevoked):
		apierror.RespondConflict(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}

func paymentProtectionPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func totalPages(total int64, pageSize int) int64 {
	if total == 0 {
		return 0
	}
	return (total + int64(pageSize) - 1) / int64(pageSize)
}

func stringValue(value interface{}) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
