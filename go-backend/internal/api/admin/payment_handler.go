package admin

import (
	"errors"
	"strconv"
	"strings"
	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/pkg/apierror"
	pgateway "tanzanite/internal/pkg/payment"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService  *service.PaymentService
	settingsService *service.AdminSettingsService
}

func NewPaymentHandler(paymentService *service.PaymentService, settingsService *service.AdminSettingsService) *PaymentHandler {
	return &PaymentHandler{
		paymentService:  paymentService,
		settingsService: settingsService,
	}
}

func (h *PaymentHandler) ListPaymentMethods(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	methods, err := h.paymentService.ListPaymentMethods(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": methods})
}

func (h *PaymentHandler) GetGatewayRuntimeStatus(c *gin.Context) {
	readiness := pgateway.BuildRuntimeReadiness(adminRequestBaseURL(c))
	pgateway.ApplySecureGatewayStatuses(&readiness, h.secureGatewayStatuses())
	response.Success(c, readiness)
}

func (h *PaymentHandler) UpsertGatewayConfig(c *gin.Context) {
	provider, err := pgateway.ParseGatewayType(c.Param("provider"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if !pgateway.PaymentConfigMasterKeyConfigured() {
		apierror.RespondBadRequest(c, pgateway.PaymentConfigMasterKeyEnv+" is required before saving payment secrets")
		return
	}
	if h.settingsService == nil {
		apierror.RespondInternalError(c, errPaymentSettingsUnavailable())
		return
	}

	var req struct {
		Environment string            `json:"environment"`
		Credentials map[string]string `json:"credentials"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	config := pgateway.SecureGatewayConfig{
		Provider:    provider,
		Environment: pgateway.NormalizeGatewayEnvironment(req.Environment),
		Credentials: map[string]string{},
	}
	if existingConfig, found, err := h.readSecureGatewayConfig(provider); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	} else if found {
		config = existingConfig
		if req.Environment != "" {
			config.Environment = pgateway.NormalizeGatewayEnvironment(req.Environment)
		}
	}

	for key, value := range req.Credentials {
		if !secureGatewayFieldAllowed(provider, key) {
			apierror.RespondBadRequest(c, "unsupported credential field: "+key)
			return
		}
		if strings.TrimSpace(value) == "" {
			continue
		}
		config.Credentials[key] = strings.TrimSpace(value)
	}

	encrypted, err := pgateway.EncryptSecureGatewayConfig(config, pgateway.PaymentConfigMasterKey())
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if _, err := h.settingsService.UpdateSetting(setting.UpdateSettingRequest{
		Key:         pgateway.SecureGatewaySettingKey(provider),
		Value:       encrypted,
		Type:        "encrypted_json",
		Group:       "payment_secret",
		Locale:      "global",
		IsPublic:    false,
		Description: "Encrypted payment gateway runtime config",
	}); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"status": h.secureGatewayStatus(provider)})
}

func (h *PaymentHandler) DeleteGatewayConfig(c *gin.Context) {
	provider, err := pgateway.ParseGatewayType(c.Param("provider"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h.settingsService == nil {
		apierror.RespondInternalError(c, errPaymentSettingsUnavailable())
		return
	}
	if err := h.settingsService.DeleteSetting(pgateway.SecureGatewaySettingKey(provider), "global"); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

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

	response.Success(c, method)
}

func (h *PaymentHandler) CreatePaymentMethod(c *gin.Context) {
	var req paymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	method, err := req.toDomain()
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if err := validatePaymentMethod(method); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.paymentService.CreatePaymentMethod(&method); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, method)
}

func (h *PaymentHandler) UpdatePaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid payment method id")
		return
	}

	var req paymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	method, err := req.toDomain()
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	method.ID = uint(id)
	if err := validatePaymentMethod(method); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.paymentService.UpdatePaymentMethod(&method); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, method)
}

func (h *PaymentHandler) DeletePaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid payment method id")
		return
	}

	if err := h.paymentService.DeletePaymentMethod(uint(id)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

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
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	var req struct {
		OrderID       uint    `json:"order_id" binding:"required"`
		TransactionID uint    `json:"transaction_id" binding:"required"`
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		Reason        string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	refund := paymentdomain.Refund{
		OrderID:       req.OrderID,
		TransactionID: req.TransactionID,
		Amount:        req.Amount,
		Reason:        req.Reason,
	}

	if err := h.paymentService.CreateAdminRefund(&refund, userIDValue.(uint)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, refund)
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

	st, err := h.settingsService.GetSetting(pgateway.SecureGatewaySettingKey(provider), "global")
	if err != nil {
		return pgateway.SecureGatewayConfig{}, false, nil
	}
	if !pgateway.PaymentConfigMasterKeyConfigured() {
		return pgateway.SecureGatewayConfig{}, true, errors.New(pgateway.PaymentConfigMasterKeyEnv + " is required to read encrypted payment config")
	}

	config, err := pgateway.DecryptSecureGatewayConfig(st.Value, pgateway.PaymentConfigMasterKey())
	if err != nil {
		return pgateway.SecureGatewayConfig{}, true, err
	}
	if config.Provider != provider {
		return pgateway.SecureGatewayConfig{}, true, errors.New("payment config provider mismatch")
	}
	return config, true, nil
}

func secureGatewayFieldAllowed(provider pgateway.GatewayType, field string) bool {
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
