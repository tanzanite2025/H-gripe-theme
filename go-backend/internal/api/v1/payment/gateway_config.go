package payment

import (
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"
)

func (h *Handler) loadPaymentGatewayConfiguration(gatewayType pgateway.GatewayType) (*pgateway.Config, error) {
	if h != nil && h.settingsService != nil {
		st, err := h.settingsService.GetDomainManagedSetting(pgateway.SecureGatewaySettingKey(gatewayType), "global")
		if err == nil {
			secureConfig, err := pgateway.DecodeStoredSecureGatewayConfig(st.Value, gatewayType)
			if err != nil {
				return nil, err
			}
			return pgateway.GatewayConfigFromSecureConfig(secureConfig), nil
		} else if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}
	return pgateway.LoadConfigFromEnv(gatewayType), nil
}
