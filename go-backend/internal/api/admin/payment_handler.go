package admin

import (
	"errors"
	"io"
	"strconv"
	"strings"
	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	pgateway "tanzanite/internal/pkg/payment"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService      *service.PaymentService
	settingsService     *service.AdminSettingsService
	auditService        paymentAuditRecorder
	publicBaseURL       string
	callbackProbeClient paymentCallbackProbeHTTPClient
}

const (
	gatewayEnvironmentProduction         = "production"
	paymentGatewayProductionConfirmation = "PRODUCTION"
)

func NewPaymentHandler(paymentService *service.PaymentService, settingsService *service.AdminSettingsService) *PaymentHandler {
	return &PaymentHandler{
		paymentService:  paymentService,
		settingsService: settingsService,
	}
}

func (h *PaymentHandler) ConfigurePublicBaseURL(baseURL string) {
	if h == nil {
		return
	}
	h.publicBaseURL = pgateway.NormalizePublicBaseURL(baseURL)
}

func (h *PaymentHandler) ListPaymentMethods(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	methods, err := h.paymentService.ListPaymentMethods(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": paymentMethodsToResponse(methods)})
}

func (h *PaymentHandler) GetGatewayRuntimeStatus(c *gin.Context) {
	readiness := pgateway.BuildRuntimeReadiness(h.runtimeReadinessBaseURL(c))
	pgateway.ApplySecureGatewayStatuses(&readiness, h.secureGatewayStatuses())
	response.Success(c, readiness)
}

func (h *PaymentHandler) runtimeReadinessBaseURL(c *gin.Context) string {
	if h != nil && h.publicBaseURL != "" {
		return h.publicBaseURL
	}
	return adminRequestBaseURL(c)
}

func (h *PaymentHandler) UpsertGatewayConfig(c *gin.Context) {
	provider, err := pgateway.ParseGatewayType(c.Param("provider"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	startedAt := paymentAuditStartedAt()
	if !pgateway.PaymentConfigMasterKeyConfigured() {
		message := pgateway.PaymentConfigMasterKeyEnv + " is required before saving payment secrets"
		h.recordPaymentGatewayConfigAudit(
			c,
			startedAt,
			provider,
			paymentAuditActionUpdate,
			paymentAuditStatusFailed,
			"master_key_missing",
			message,
			map[string]interface{}{"master_key_configured": false},
		)
		apierror.RespondBadRequest(c, message)
		return
	}

	var req struct {
		Environment  string            `json:"environment"`
		Credentials  map[string]string `json:"credentials"`
		Confirmation string            `json:"confirmation"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordPaymentGatewayConfigAudit(
			c,
			startedAt,
			provider,
			paymentAuditActionUpdate,
			paymentAuditStatusFailed,
			"invalid_request",
			err.Error(),
			map[string]interface{}{},
		)
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	config := pgateway.SecureGatewayConfig{
		Provider:    provider,
		Environment: pgateway.NormalizeGatewayEnvironment(req.Environment),
		Credentials: map[string]string{},
	}
	if config.Environment == gatewayEnvironmentProduction && strings.TrimSpace(req.Confirmation) != paymentGatewayProductionConfirmation {
		message := "type " + paymentGatewayProductionConfirmation + " to save a production payment gateway config"
		details := paymentGatewayConfigRequestAuditDetails(provider, config.Environment, req.Credentials)
		details["confirmation_matched"] = false
		h.recordPaymentGatewayConfigAudit(c, startedAt, provider, paymentAuditActionUpdate, paymentAuditStatusFailed, "confirmation_mismatch", message, details)
		apierror.RespondBadRequest(c, message)
		return
	}
	if h.settingsService == nil {
		err := errPaymentSettingsUnavailable()
		h.recordPaymentGatewayConfigAudit(c, startedAt, provider, paymentAuditActionUpdate, paymentAuditStatusFailed, "settings_unavailable", err.Error(), paymentGatewayConfigRequestAuditDetails(provider, config.Environment, req.Credentials))
		apierror.RespondInternalError(c, err)
		return
	}
	if existingConfig, found, err := h.readSecureGatewayConfig(provider); err != nil {
		details := paymentGatewayConfigRequestAuditDetails(provider, config.Environment, req.Credentials)
		details["existing_config"] = found
		h.recordPaymentGatewayConfigAudit(c, startedAt, provider, paymentAuditActionUpdate, paymentAuditStatusFailed, "read_existing_failed", err.Error(), details)
		apierror.RespondBadRequest(c, err.Error())
		return
	} else if found {
		config = existingConfig
		if req.Environment != "" {
			config.Environment = pgateway.NormalizeGatewayEnvironment(req.Environment)
		}
		if config.Environment == gatewayEnvironmentProduction && strings.TrimSpace(req.Confirmation) != paymentGatewayProductionConfirmation {
			message := "type " + paymentGatewayProductionConfirmation + " to save a production payment gateway config"
			details := paymentGatewayConfigRequestAuditDetails(provider, config.Environment, req.Credentials)
			details["existing_config"] = true
			details["confirmation_matched"] = false
			h.recordPaymentGatewayConfigAudit(c, startedAt, provider, paymentAuditActionUpdate, paymentAuditStatusFailed, "confirmation_mismatch", message, details)
			apierror.RespondBadRequest(c, message)
			return
		}
	}

	for key, value := range req.Credentials {
		if !secureGatewayFieldAllowed(provider, key) {
			message := "unsupported credential field: " + key
			details := paymentGatewayConfigRequestAuditDetails(provider, config.Environment, req.Credentials)
			details["unsupported_field"] = strings.TrimSpace(key)
			h.recordPaymentGatewayConfigAudit(c, startedAt, provider, paymentAuditActionUpdate, paymentAuditStatusFailed, "unsupported_field", message, details)
			apierror.RespondBadRequest(c, message)
			return
		}
		if strings.TrimSpace(value) == "" {
			continue
		}
		config.Credentials[key] = strings.TrimSpace(value)
	}

	encrypted, err := pgateway.EncryptSecureGatewayConfig(config, pgateway.PaymentConfigMasterKey())
	if err != nil {
		h.recordPaymentGatewayConfigAudit(c, startedAt, provider, paymentAuditActionUpdate, paymentAuditStatusFailed, "encrypt_failed", err.Error(), paymentGatewayConfigStoredAuditDetails(config))
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if _, err := h.settingsService.UpdateDomainManagedSetting(setting.UpdateSettingRequest{
		Key:         pgateway.SecureGatewaySettingKey(provider),
		Value:       encrypted,
		Type:        "encrypted_json",
		Group:       "payment_secret",
		Locale:      "global",
		IsPublic:    false,
		Description: "Encrypted payment gateway runtime config",
	}); err != nil {
		h.recordPaymentGatewayConfigAudit(c, startedAt, provider, paymentAuditActionUpdate, paymentAuditStatusFailed, "save_setting_failed", err.Error(), paymentGatewayConfigStoredAuditDetails(config))
		apierror.RespondInternalError(c, err)
		return
	}

	h.recordPaymentGatewayConfigAudit(c, startedAt, provider, paymentAuditActionUpdate, paymentAuditStatusSuccess, "", "", paymentGatewayConfigStoredAuditDetails(config))
	response.Success(c, gin.H{"status": h.secureGatewayStatus(provider)})
}

func (h *PaymentHandler) DeleteGatewayConfig(c *gin.Context) {
	provider, err := pgateway.ParseGatewayType(c.Param("provider"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	startedAt := paymentAuditStartedAt()
	var req struct {
		Confirmation string `json:"confirmation"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		h.recordPaymentGatewayConfigAudit(
			c,
			startedAt,
			provider,
			paymentAuditActionDelete,
			paymentAuditStatusFailed,
			"invalid_request",
			err.Error(),
			paymentGatewayDeleteAuditDetails(provider, false),
		)
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	expectedConfirmation := paymentGatewayDeleteConfirmation(provider)
	if strings.TrimSpace(req.Confirmation) != expectedConfirmation {
		message := "type " + expectedConfirmation + " to delete this payment gateway config"
		h.recordPaymentGatewayConfigAudit(
			c,
			startedAt,
			provider,
			paymentAuditActionDelete,
			paymentAuditStatusFailed,
			"confirmation_mismatch",
			message,
			paymentGatewayDeleteAuditDetails(provider, false),
		)
		apierror.RespondBadRequest(c, message)
		return
	}
	if h.settingsService == nil {
		err := errPaymentSettingsUnavailable()
		h.recordPaymentGatewayConfigAudit(
			c,
			startedAt,
			provider,
			paymentAuditActionDelete,
			paymentAuditStatusFailed,
			"settings_unavailable",
			err.Error(),
			paymentGatewayDeleteAuditDetails(provider, true),
		)
		apierror.RespondInternalError(c, err)
		return
	}
	if err := h.settingsService.DeleteDomainManagedSetting(pgateway.SecureGatewaySettingKey(provider), "global"); err != nil {
		h.recordPaymentGatewayConfigAudit(
			c,
			startedAt,
			provider,
			paymentAuditActionDelete,
			paymentAuditStatusFailed,
			"delete_setting_failed",
			err.Error(),
			paymentGatewayDeleteAuditDetails(provider, true),
		)
		apierror.RespondInternalError(c, err)
		return
	}

	h.recordPaymentGatewayConfigAudit(
		c,
		startedAt,
		provider,
		paymentAuditActionDelete,
		paymentAuditStatusSuccess,
		"",
		"",
		paymentGatewayDeleteAuditDetails(provider, true),
	)
	response.SuccessWithMessage(c, "payment gateway config deleted", gin.H{"provider": provider})
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

func (h *PaymentHandler) GetTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid transaction id")
		return
	}

	transaction, err := h.paymentService.GetTransaction(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Transaction")
		return
	}

	response.Success(c, transaction)
}

func (h *PaymentHandler) GetOrderTransactions(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid order id")
		return
	}

	transactions, err := h.paymentService.GetOrderTransactions(uint(orderID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": transactions})
}

func (h *PaymentHandler) GetRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid refund id")
		return
	}

	refund, err := h.paymentService.GetRefund(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Refund")
		return
	}

	response.Success(c, refund)
}

func (h *PaymentHandler) GetOrderRefunds(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid order id")
		return
	}

	refunds, err := h.paymentService.GetOrderRefunds(uint(orderID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": refunds})
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
	if err := h.paymentService.CreateAdminRefund(&refund, adminID); err != nil {
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
		ResourceID: refund.ID,
		Status:     paymentAuditStatusSuccess,
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
	response.Created(c, refund)
}

func (h *PaymentHandler) ListStripeDisputes(c *gin.Context) {
	params := pagination.ParsePagination(c)
	disputes, total, err := h.paymentService.ListStripeDisputes(
		strings.TrimSpace(c.Query("status")),
		params.Page,
		params.PageSize,
	)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Paged(c, disputes, params.Page, params.PageSize, total)
}

func (h *PaymentHandler) GetStripeDispute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid dispute id")
		return
	}
	dispute, err := h.paymentService.GetStripeDispute(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Stripe dispute")
		return
	}
	response.Success(c, dispute)
}

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

func adminRequestBaseURL(c *gin.Context) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	if host == "" {
		return ""
	}

	return scheme + "://" + host
}

func (h *PaymentHandler) secureGatewayStatuses() []pgateway.SecureGatewayConfigStatus {
	statuses := make([]pgateway.SecureGatewayConfigStatus, 0, 4)
	for _, provider := range []pgateway.GatewayType{pgateway.GatewayStripe, pgateway.GatewayPayPal, pgateway.GatewayAlipay, pgateway.GatewayWechat} {
		statuses = append(statuses, h.secureGatewayStatus(provider))
	}
	return statuses
}

func (h *PaymentHandler) secureGatewayStatus(provider pgateway.GatewayType) pgateway.SecureGatewayConfigStatus {
	status := pgateway.SecureGatewayConfigStatus{
		Provider:      provider,
		RuntimeSource: "environment",
	}
	config, found, err := h.readSecureGatewayConfig(provider)
	if !found {
		return status
	}

	status.Configured = true
	if err != nil {
		status.Error = err.Error()
		return status
	}

	status.Readable = true
	status.RuntimeSource = "admin-encrypted"
	status.Environment = config.Environment
	status.ConfiguredFields = pgateway.SecureGatewayConfiguredFields(config)
	return status
}

func (h *PaymentHandler) readSecureGatewayConfig(provider pgateway.GatewayType) (pgateway.SecureGatewayConfig, bool, error) {
	if h.settingsService == nil {
		return pgateway.SecureGatewayConfig{}, false, errPaymentSettingsUnavailable()
	}

	st, err := h.settingsService.GetDomainManagedSetting(pgateway.SecureGatewaySettingKey(provider), "global")
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

func (h *PaymentHandler) adminStripeGatewayConfig() (*pgateway.Config, error) {
	if h.settingsService != nil {
		config, found, err := h.readSecureGatewayConfig(pgateway.GatewayStripe)
		if err != nil {
			return nil, err
		}
		if found {
			gatewayConfig := pgateway.GatewayConfigFromSecureConfig(config)
			if strings.TrimSpace(gatewayConfig.APIKey) == "" {
				return nil, errors.New("Stripe API key is not configured")
			}
			return gatewayConfig, nil
		}
	}
	config := pgateway.LoadConfigFromEnv(pgateway.GatewayStripe)
	if strings.TrimSpace(config.APIKey) == "" {
		return nil, errors.New("Stripe API key is not configured")
	}
	return config, nil
}

func secureGatewayFieldAllowed(provider pgateway.GatewayType, field string) bool {
	if provider == pgateway.GatewayStripe && field == "three_ds_mode" {
		return true
	}
	for _, allowed := range pgateway.SecureGatewayCredentialFields(provider) {
		if field == allowed {
			return true
		}
	}
	return false
}

func errPaymentSettingsUnavailable() error {
	return errors.New("payment settings service is unavailable")
}

func paymentGatewayDeleteConfirmation(provider pgateway.GatewayType) string {
	return "DELETE " + strings.ToUpper(string(provider))
}
