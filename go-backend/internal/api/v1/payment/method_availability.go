package payment

import (
	"strings"

	paymentdomain "tanzanite/internal/domain/payment"
	pgateway "tanzanite/internal/pkg/payment"

	"github.com/gin-gonic/gin"
)

func (h *Handler) paymentMethodsToAvailabilityResponse(
	c *gin.Context,
	methods []paymentdomain.PaymentMethod,
) ([]paymentMethodResponse, error) {
	items := make([]paymentMethodResponse, 0, len(methods))
	for _, method := range methods {
		item, err := h.paymentMethodToAvailabilityResponse(c, method)
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
) (paymentMethodResponse, error) {
	item := paymentMethodToResponse(method)
	if !method.Enabled {
		return item, nil
	}
	if err := h.applyDefaultCurrencyAvailability(&item); err != nil {
		return paymentMethodResponse{}, err
	}
	if h == nil || h.protection == nil {
		return item, nil
	}

	decision, err := h.protection.Evaluate(paymentdomain.PaymentProtectionEvaluationInput{
		Provider:      item.Provider,
		Country:       paymentMethodAvailabilityCountry(c),
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

func (h *Handler) applyDefaultCurrencyAvailability(item *paymentMethodResponse) error {
	if h == nil || h.currencyPolicy == nil || item == nil || item.Provider == "" || !item.Available {
		return nil
	}
	defaultCurrency, err := h.currencyPolicy.DefaultOrderCurrency()
	if err != nil {
		return err
	}
	if !pgateway.GatewaySupportsCurrency(pgateway.GatewayType(item.Provider), defaultCurrency) {
		item.Available = false
		item.UnavailableReason = "currency_not_supported"
	}
	return nil
}

func paymentMethodAvailabilityCountry(c *gin.Context) string {
	if c != nil {
		if country := strings.TrimSpace(c.Query("country")); country != "" {
			return country
		}
	}
	return paymentRequestCountry(c)
}

func paymentMethodProvider(code string) string {
	return pgateway.ProviderForPaymentMethod(code)
}
