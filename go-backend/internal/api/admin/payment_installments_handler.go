package admin

import (
	"errors"
	"strings"

	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) GetPaymentProviderInstallments(c *gin.Context) {
	provider, err := pgateway.ParseGatewayType(c.Param("provider"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.settingsService == nil {
		apierror.RespondInternalError(c, errors.New("payment settings service is not configured"))
		return
	}

	key := settingdomain.PaymentInstallmentsSettingKey(string(provider))
	record, err := h.settingsService.GetDomainManagedSetting(key, "global")
	if err != nil {
		if !repository.IsRecordNotFound(err) {
			apierror.RespondInternalError(c, err)
			return
		}
		response.Success(c, settingdomain.PaymentProviderInstallmentsSettings{
			Provider: string(provider),
		})
		return
	}

	settings, err := settingdomain.PaymentProviderInstallmentsSettingsFromValue(string(provider), record.Value)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *PaymentHandler) UpdatePaymentProviderInstallments(c *gin.Context) {
	provider, err := pgateway.ParseGatewayType(c.Param("provider"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.settingsService == nil {
		apierror.RespondInternalError(c, errors.New("payment settings service is not configured"))
		return
	}

	var req settingdomain.PaymentProviderInstallmentsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	settings := req.Settings(string(provider))
	value, err := settings.Value()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	if _, err := h.settingsService.UpdateDomainManagedSetting(settingdomain.UpdateSettingRequest{
		Key:         settingdomain.PaymentInstallmentsSettingKey(string(provider)),
		Value:       value,
		Type:        "json",
		Group:       settingdomain.PaymentInstallmentsGroup,
		Locale:      "global",
		IsPublic:    false,
		Description: strings.TrimSpace(string(provider)) + " installments configuration",
	}); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, settings)
}

func (h *PaymentHandler) DeletePaymentProviderInstallments(c *gin.Context) {
	provider, err := pgateway.ParseGatewayType(c.Param("provider"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.settingsService == nil {
		apierror.RespondInternalError(c, errors.New("payment settings service is not configured"))
		return
	}

	if err := h.settingsService.DeleteDomainManagedSetting(settingdomain.PaymentInstallmentsSettingKey(string(provider)), "global"); err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.SuccessWithMessage(c, "payment provider installments deleted", gin.H{"provider": provider})
}
