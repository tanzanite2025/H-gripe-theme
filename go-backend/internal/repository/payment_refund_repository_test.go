package repository

import (
	"testing"

	domainmoney "commerce-platform/internal/domain/money"
	paymentdomain "commerce-platform/internal/domain/payment"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestFindPendingRefundByTransactionAndAmountMatchesExactMinorUnits(t *testing.T) {
	db := newPaymentRefundRepositoryTestDB(t)
	repo := NewPaymentRepository(db)

	require.NoError(t, db.Create(&[]paymentdomain.Refund{
		{TransactionID: 10, Currency: "USD", AmountMinor: 1001, RequestedAmountMinor: 1200, Status: "pending"},
		{TransactionID: 10, Currency: "USD", AmountMinor: 1000, RequestedAmountMinor: 1200, Status: "pending"},
	}).Error)

	refund, err := repo.FindPendingRefundByTransactionAndAmount(10, domainmoney.MustNew(1000, "USD"))
	require.NoError(t, err)
	require.Equal(t, uint(2), refund.ID)

	_, err = repo.FindPendingRefundByTransactionAndAmount(10, domainmoney.MustNew(1001, "USD"))
	require.NoError(t, err)
}

func TestFindPendingRefundByTransactionAndAmountUsesRequestedAmountWhenNetAmountDiffers(t *testing.T) {
	db := newPaymentRefundRepositoryTestDB(t)
	repo := NewPaymentRepository(db)
	require.NoError(t, db.Create(&paymentdomain.Refund{
		TransactionID:        11,
		Currency:             "USD",
		AmountMinor:          999,
		RequestedAmountMinor: 1000,
		Status:               "pending",
	}).Error)

	refund, err := repo.FindPendingRefundByTransactionAndAmount(11, domainmoney.MustNew(1000, "USD"))
	require.NoError(t, err)
	require.NotNil(t, refund)
}

func TestFindPendingRefundByTransactionAndAmountHonorsZeroMinorUnitCurrencies(t *testing.T) {
	db := newPaymentRefundRepositoryTestDB(t)
	repo := NewPaymentRepository(db)
	require.NoError(t, db.Create(&paymentdomain.Refund{
		TransactionID: 12,
		Currency:      "JPY",
		AmountMinor:   1000,
		Status:        "pending",
	}).Error)

	refund, err := repo.FindPendingRefundByTransactionAndAmount(12, domainmoney.MustNew(1000, "JPY"))
	require.NoError(t, err)
	require.NotNil(t, refund)

	_, err = repo.FindPendingRefundByTransactionAndAmount(12, domainmoney.MustNew(1001, "JPY"))
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func newPaymentRefundRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&paymentdomain.Refund{}, &paymentdomain.RefundLineItem{}))
	return db
}
