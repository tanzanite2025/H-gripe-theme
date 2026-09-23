package order

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestOrderMoneyAccessorsUseMinorUnits(t *testing.T) {
	item := OrderItem{
		Currency:       "JPY",
		PriceMinor:     1250,
		SubtotalMinor:  2500,
		TaxAmountMinor: 100,
		DiscountMinor:  50,
		TotalMinor:     2550,
	}
	price, err := item.PriceMoney()
	require.NoError(t, err)
	require.Equal(t, int64(1250), price.AmountMinor())
	require.Equal(t, "JPY", price.Currency().String())

	orderRecord := Order{
		Currency:            "USD",
		SubtotalAmountMinor: 10000,
		ShippingFeeMinor:    125,
		TaxAmountMinor:      875,
		DiscountAmountMinor: 1000,
		TotalAmountMinor:    10000,
		PaymentCurrency:     "CNY",
		PaymentAmountMinor:  7200,
	}
	total, err := orderRecord.TotalMoney()
	require.NoError(t, err)
	require.Equal(t, int64(10000), total.AmountMinor())
	payment, err := orderRecord.PaymentMoney()
	require.NoError(t, err)
	require.Equal(t, int64(7200), payment.AmountMinor())
	require.Equal(t, "CNY", payment.Currency().String())
}

func TestConfigurationEvidenceUsesImmutableSnapshot(t *testing.T) {
	item := OrderItem{
		ConfigurationSnapshotData: datatypes.JSON([]byte(`{"schema_version":1,"selections":[]}`)),
	}
	require.JSONEq(t, `{"schema_version":1,"selections":[]}`, item.ConfigurationEvidenceJSON())

	item.ConfigurationSnapshotData = datatypes.JSON([]byte(`{}`))
	require.JSONEq(t, `{}`, item.ConfigurationEvidenceJSON())
}

func TestOrderBeforeCreateRequiresMinorAmounts(t *testing.T) {
	orderRecord := Order{
		OrderNumber:         "ORD-MONEY-1",
		Currency:            "USD",
		PaymentCurrency:     "USD",
		SubtotalAmountMinor: 1234,
		ShippingFeeMinor:    125,
		TaxAmountMinor:      41,
		DiscountAmountMinor: 200,
		TotalAmountMinor:    1200,
		PaymentAmountMinor:  1200,
	}
	require.NoError(t, orderRecord.BeforeCreate(nil))
	require.Equal(t, int64(1234), orderRecord.SubtotalAmountMinor)
	require.Equal(t, int64(125), orderRecord.ShippingFeeMinor)
	require.Equal(t, int64(41), orderRecord.TaxAmountMinor)
	require.Equal(t, int64(200), orderRecord.DiscountAmountMinor)
	require.Equal(t, int64(1200), orderRecord.TotalAmountMinor)
	require.Equal(t, int64(1200), orderRecord.PaymentAmountMinor)

	item := OrderItem{Currency: "JPY", PriceMinor: 1250, SubtotalMinor: 2500, TotalMinor: 2500}
	require.NoError(t, item.BeforeCreate(nil))
	require.Equal(t, int64(1250), item.PriceMinor)
	require.Equal(t, int64(2500), item.SubtotalMinor)
	require.Equal(t, int64(2500), item.TotalMinor)
}
