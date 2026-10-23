package service

import (
	"testing"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestReadOnlyRecoveryShippingUsesActualRateCalculationWithoutCheckoutSnapshots(t *testing.T) {
	database, ordinary := newTestShippingQuoteService(t)
	template := seedWeightQuoteTemplate(t, database)
	product, variant := seedQuoteProduct(t, database, 50, 900, template.ID)
	carrier := seedQuoteCarrier(t, database, "Read-only recovery carrier", "RECOVERY")
	seedQuoteCustomerBillingTemplateAndCarrierRoute(t, database, carrier.ID, template.ID, shippingdomain.CarrierService{
		ServiceCode: "RECOVERY-STANDARD", ServiceName: "Recovery standard", Countries: `["US"]`, BillingMode: "actual_weight", FirstWeightGrams: 500, AdditionalWeightGrams: 500,
	})
	input := ShippingQuoteInput{Country: "US", Currency: "USD", Items: []ShippingQuoteItemInput{{ProductID: product.ID, VariantID: &variant.ID, Quantity: 1}}}
	readOnly := NewShippingServiceForReadOnlyRecoveryVerification(repository.NewShippingRepository(database), repository.NewProductRepository(database))
	var before, after int64
	require.NoError(t, database.Model(&shippingdomain.QuoteSnapshot{}).Count(&before).Error)
	quote, err := readOnly.QuoteCart(input)
	require.NoError(t, err)
	require.NotNil(t, quote.SelectedPlan)
	require.Equal(t, int64(500), quote.ShippingFeeMinor)
	require.Empty(t, quote.ID)
	require.Empty(t, quote.SelectedPlan.ID)
	require.True(t, quote.ExpiresAt.IsZero())
	require.NoError(t, database.Model(&shippingdomain.QuoteSnapshot{}).Count(&after).Error)
	require.Equal(t, before, after)
	checkoutQuote, err := ordinary.QuoteCart(input)
	require.NoError(t, err)
	require.Equal(t, quote.ShippingFeeMinor, checkoutQuote.ShippingFeeMinor)
	require.NotEmpty(t, checkoutQuote.ID)
	require.NotEmpty(t, checkoutQuote.SelectedPlan.ID)
	require.NoError(t, database.Model(&shippingdomain.QuoteSnapshot{}).Count(&after).Error)
	require.Equal(t, before+1, after, "ordinary checkout must still persist its snapshot")
}
