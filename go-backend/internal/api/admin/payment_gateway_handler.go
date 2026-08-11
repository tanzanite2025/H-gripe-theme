package admin

import (
	"errors"
	"io"
	"strings"

	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"

	"github.com/gin-gonic/gin"
)

const (
	gatewayEnvironmentProduction         = "production"
	paymentGatewayProductionConfirmation = "PRODUCTION"
)

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
