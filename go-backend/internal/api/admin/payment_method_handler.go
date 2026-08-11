package admin

import (
	"errors"
	"strconv"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) ListPaymentMethods(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	methods, err := h.paymentService.ListPaymentMethods(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": paymentMethodsToResponse(methods)})
}

func (h *PaymentHandler) GetPaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid payment method id")
		return
	}

	method, err := h.paymentService.GetPaymentMethod(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Payment method")
		return
	}

	response.Success(c, paymentMethodToResponse(*method))
}

func (h *PaymentHandler) CreatePaymentMethod(c *gin.Context) {
	startedAt := paymentAuditStartedAt()
	var req paymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourcePaymentMethod,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(0, paymentdomain.PaymentMethod{}),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	method, err := req.toDomain()
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourcePaymentMethod,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(0, method),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if err := validatePaymentMethod(method); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourcePaymentMethod,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(0, method),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.paymentService == nil {
		err := errors.New("payment service is not configured")
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourcePaymentMethod,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(0, method),
		})
		apierror.RespondInternalError(c, err)
		return
	}

	if err := h.paymentService.CreatePaymentMethod(&method); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionCreate,
			Resource:     paymentAuditResourcePaymentMethod,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(0, method),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionCreate,
		Resource:   paymentAuditResourcePaymentMethod,
		ResourceID: method.ID,
		Status:     paymentAuditStatusSuccess,
		Changes:    paymentMethodAuditDetails(method.ID, method),
	})
	response.Created(c, paymentMethodToResponse(method))
}

func (h *PaymentHandler) UpdatePaymentMethod(c *gin.Context) {
	startedAt := paymentAuditStartedAt()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentMethod,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "invalid payment method id",
			Changes: map[string]interface{}{
				"raw_payment_method_id": c.Param("id"),
			},
		})
		apierror.RespondBadRequest(c, "invalid payment method id")
		return
	}
	methodID := uint(id)

	var req paymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentMethod,
			ResourceID:   methodID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(methodID, paymentdomain.PaymentMethod{ID: methodID}),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	method, err := req.toDomain()
	if err != nil {
		method.ID = methodID
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentMethod,
			ResourceID:   methodID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(methodID, method),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	method.ID = methodID
	if err := validatePaymentMethod(method); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentMethod,
			ResourceID:   methodID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(methodID, method),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.paymentService == nil {
		err := errors.New("payment service is not configured")
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentMethod,
			ResourceID:   methodID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(methodID, method),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	oldMethod, _ := h.paymentService.GetPaymentMethod(methodID)

	if err := h.paymentService.UpdatePaymentMethod(&method); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionUpdate,
			Resource:     paymentAuditResourcePaymentMethod,
			ResourceID:   methodID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(methodID, method),
			OldValue:     paymentMethodOldValue(oldMethod),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionUpdate,
		Resource:   paymentAuditResourcePaymentMethod,
		ResourceID: methodID,
		Status:     paymentAuditStatusSuccess,
		Changes:    paymentMethodAuditDetails(methodID, method),
		OldValue:   paymentMethodOldValue(oldMethod),
		NewValue:   paymentMethodAuditDetails(methodID, method),
	})
	response.Success(c, paymentMethodToResponse(method))
}

func (h *PaymentHandler) DeletePaymentMethod(c *gin.Context) {
	startedAt := paymentAuditStartedAt()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionDelete,
			Resource:     paymentAuditResourcePaymentMethod,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "invalid payment method id",
			Changes: map[string]interface{}{
				"raw_payment_method_id": c.Param("id"),
			},
		})
		apierror.RespondBadRequest(c, "invalid payment method id")
		return
	}
	methodID := uint(id)
	if h == nil || h.paymentService == nil {
		err := errors.New("payment service is not configured")
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionDelete,
			Resource:     paymentAuditResourcePaymentMethod,
			ResourceID:   methodID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(methodID, paymentdomain.PaymentMethod{ID: methodID}),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	oldMethod, _ := h.paymentService.GetPaymentMethod(methodID)

	if err := h.paymentService.DeletePaymentMethod(methodID); err != nil {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionDelete,
			Resource:     paymentAuditResourcePaymentMethod,
			ResourceID:   methodID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentMethodAuditDetails(methodID, paymentdomain.PaymentMethod{ID: methodID}),
			OldValue:     paymentMethodOldValue(oldMethod),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionDelete,
		Resource:   paymentAuditResourcePaymentMethod,
		ResourceID: methodID,
		Status:     paymentAuditStatusSuccess,
		Changes:    paymentMethodAuditDetails(methodID, paymentdomain.PaymentMethod{ID: methodID}),
		OldValue:   paymentMethodOldValue(oldMethod),
	})
	response.SuccessWithMessage(c, "payment method deleted", nil)
}
