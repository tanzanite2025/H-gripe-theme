package admin

import (
	"errors"
	"strconv"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentRefundExecutionHandler struct {
	paymentService  *service.PaymentService
	settingsService *service.AdminSettingsService
	auditService    paymentAuditRecorder
}

func NewPaymentRefundExecutionHandler(paymentService *service.PaymentService, settingsService *service.AdminSettingsService) *PaymentRefundExecutionHandler {
	return &PaymentRefundExecutionHandler{
		paymentService:  paymentService,
		settingsService: settingsService,
	}
}

func (h *PaymentRefundExecutionHandler) ExecutePendingRefund(c *gin.Context) {
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
	config, err := h.gatewayConfig(provider)
	if err != nil {
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			ResourceID:   refundID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundExecutionAuditDetails(refundID, string(provider), refund, transaction, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	gateway, err := pgateway.NewPaymentGateway(config)
	if err != nil {
		h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionExecute,
			Resource:     paymentAuditResourceRefundExecution,
			ResourceID:   refundID,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      paymentRefundExecutionAuditDetails(refundID, string(provider), refund, transaction, nil),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	completedRefund, execution, err := h.paymentService.ExecutePendingRefund(c.Request.Context(), service.ExecutePendingRefundInput{
		RefundID: refundID,
		AdminID:  adminID,
		Provider: string(provider),
		Gateway:  gateway,
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
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	h.recordRefundExecutionAudit(c, paymentAdminAuditEvent{
		StartedAt:  startedAt,
		Action:     paymentAuditActionExecute,
		Resource:   paymentAuditResourceRefundExecution,
		ResourceID: refundID,
		Status:     paymentAuditStatusSuccess,
		Changes:    paymentRefundExecutionAuditDetails(refundID, string(provider), completedRefund, transaction, execution),
	})
	response.Success(c, gin.H{
		"refund":    completedRefund,
		"execution": execution,
	})
}

func (h *PaymentRefundExecutionHandler) gatewayConfig(provider pgateway.GatewayType) (*pgateway.Config, error) {
	if h.settingsService != nil {
		config, found, err := readAdminSecureGatewayConfig(h.settingsService, provider)
		if err != nil {
			return nil, err
		}
		if found {
			gatewayConfig := pgateway.GatewayConfigFromSecureConfig(config)
			if strings.TrimSpace(gatewayConfig.APIKey) == "" {
				return nil, errors.New(string(provider) + " API key is not configured")
			}
			return gatewayConfig, nil
		}
	}
	config := pgateway.LoadConfigFromEnv(provider)
	if strings.TrimSpace(config.APIKey) == "" {
		return nil, errors.New(string(provider) + " API key is not configured")
	}
	return config, nil
}

func readAdminSecureGatewayConfig(
	settingsService *service.AdminSettingsService,
	provider pgateway.GatewayType,
) (pgateway.SecureGatewayConfig, bool, error) {
	if settingsService == nil {
		return pgateway.SecureGatewayConfig{}, false, errors.New("payment settings service is unavailable")
	}
	st, err := settingsService.GetDomainManagedSetting(pgateway.SecureGatewaySettingKey(provider), "global")
	if err != nil {
		if !repository.IsRecordNotFound(err) {
			return pgateway.SecureGatewayConfig{}, false, err
		}
		return pgateway.SecureGatewayConfig{}, false, nil
	}
	config, err := pgateway.DecodeStoredSecureGatewayConfig(st.Value, provider)
	if err != nil {
		return pgateway.SecureGatewayConfig{}, true, err
	}
	return config, true, nil
}
