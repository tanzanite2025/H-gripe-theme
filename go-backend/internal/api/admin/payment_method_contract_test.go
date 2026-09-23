package admin

import (
	paymentdomain "commerce-platform/internal/domain/payment"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePaymentMethodRejectsImpreciseFeeRate(t *testing.T) {
	for _, value := range []string{"100.000000000000001", "5/2"} {
		method := paymentdomain.PaymentMethod{
			Name:           "Card",
			Code:           "card",
			FeeType:        "percentage",
			FeeRateDecimal: value,
		}
		require.Error(t, validatePaymentMethod(method), value)
	}
}

func TestValidatePaymentMethodAllowsFixedMinorFee(t *testing.T) {
	method := paymentdomain.PaymentMethod{
		Name:          "Card",
		Code:          "card",
		FeeType:       "fixed",
		FeeValueMinor: 125,
	}
	require.NoError(t, validatePaymentMethod(method))
}
