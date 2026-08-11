package payment

import (
	"strings"

	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/gin-gonic/gin"
)

type paymentMethodAvailabilityContext struct {
	Country string
}

func (h *Handler) paymentMethodsToAvailabilityResponse(
	c *gin.Context,
	methods []paymentdomain.PaymentMethod,
) ([]paymentMethodResponse, error) {
	context, err := h.resolvePaymentMethodAvailabilityContext(c)
	if err != nil {
		return nil, err
	}
	items := make([]paymentMethodResponse, 0, len(methods))
	for _, method := range methods {
		item, err := h.paymentMethodToAvailabilityResponse(c, method, context)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (h *Handler) paymentMethodToAvailabilityResponse(
	c *gin.Context,
	method paymentdomain.PaymentMethod,
	context paymentMethodAvailabilityContext,
) (paymentMethodResponse, error) {
	item := paymentMethodToResponse(method)
	if !method.Enabled {
		item.Available = false
		item.UnavailableReason = "disabled"
		return item, nil
	}

	if provider := strings.TrimSpace(item.Provider); provider != "" {
		available, reason := h.gatewayAvailability(pgateway.GatewayType(provider))
		if !available {
			item.Available = false
			item.UnavailableReason = reason
			return item, nil
		}
	} else {
		item.Available = false
		item.UnavailableReason = "payment_provider_not_supported"
		return item, nil
	}

	if h == nil || h.protection == nil {
		return item, nil
	}

	decision, err := h.protection.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider:      item.Provider,
		Country:       context.Country,
		PaymentMethod: method.Code,
	})
	if err != nil {
		return paymentMethodResponse{}, err
	}
	if decision.PausePayment {
		item.Available = false
		item.UnavailableReason = "temporarily_unavailable"
	}
	return item, nil
}

func (h *Handler) gatewayAvailability(provider pgateway.GatewayType) (bool, string) {
	config, err := h.loadGatewayConfig(provider)
	if err != nil {
		return false, "gateway_config_invalid"
	}

	switch provider {
	case pgateway.GatewayStripe:
		if strings.TrimSpace(config.APIKey) == "" ||
			strings.TrimSpace(config.PublishableKey) == "" ||
			strings.TrimSpace(config.WebhookSecret) == "" {
			return false, "gateway_not_configured"
		}
	case pgateway.GatewayPayPal:
		if strings.TrimSpace(config.APIKey) == "" ||
			strings.TrimSpace(config.SecretKey) == "" ||
			strings.TrimSpace(config.WebhookSecret) == "" {
			return false, "gateway_not_configured"
		}
	case pgateway.GatewayAlipay:
		if strings.TrimSpace(config.APIKey) == "" ||
			strings.TrimSpace(config.SecretKey) == "" ||
			strings.TrimSpace(config.WebhookSecret) == "" {
			return false, "gateway_not_configured"
		}
	case pgateway.GatewayWechat:
		platformVerifierConfigured := strings.TrimSpace(config.WechatPayPlatformCertificate) != "" ||
			(strings.TrimSpace(config.WechatPayPlatformPublicKey) != "" &&
				strings.TrimSpace(config.WechatPayPlatformPublicKeyID) != "")
		if strings.TrimSpace(config.APIKey) == "" ||
			strings.TrimSpace(config.WechatAppID) == "" ||
			strings.TrimSpace(config.SecretKey) == "" ||
			strings.TrimSpace(config.WebhookSecret) == "" ||
			strings.TrimSpace(config.WechatAPIv3Key) == "" ||
			!platformVerifierConfigured {
			return false, "gateway_not_configured"
		}
	default:
		return false, "payment_provider_not_supported"
	}

	return true, ""
}

func (h *Handler) resolvePaymentMethodAvailabilityContext(c *gin.Context) (paymentMethodAvailabilityContext, error) {
	return paymentMethodAvailabilityContext{Country: paymentMethodAvailabilityCountry(c)}, nil
}

func paymentMethodAvailabilityCountry(c *gin.Context) string {
	if c != nil {
		if country := strings.TrimSpace(c.Query("country")); country != "" {
			return country
		}
		if country := strings.TrimSpace(c.GetHeader("X-Market-Country")); country != "" {
			return country
		}
	}
	return paymentRequestCountry(c)
}

func paymentMethodProvider(code string) string {
	return pgateway.ProviderForPaymentMethod(code)
}
