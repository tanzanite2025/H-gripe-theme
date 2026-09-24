package payment

import (
	domainmoney "commerce-platform/internal/domain/money"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTaxRateUsesDecimal(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&TaxRate{}))

	rate := TaxRate{Name: "exact", Country: "US", RateDecimal: "10"}
	require.NoError(t, db.Create(&rate).Error)
	amount := domainmoney.MustNew(1250, "USD")
	tax, err := rate.CalculateTaxMoney(amount)
	require.NoError(t, err)
	require.Equal(t, int64(125), tax.AmountMinor())
}

func TestTaxRateRejectsPrecisionBeyondDatabaseScale(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&TaxRate{}))

	rate := TaxRate{Name: "overprecise", Country: "US", RateDecimal: "7.1234567890123456"}
	require.Error(t, db.Create(&rate).Error)
}
