package service

import (
	"testing"

	coupondomain "commerce-platform/internal/domain/coupon"
	domainmoney "commerce-platform/internal/domain/money"
	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestCalculateRefundPaymentSplitUsesTransactionCurrencyMinorUnits(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 10000)
	transaction.Amount = 7000
	transaction.Currency = "JPY"
	require.NoError(t, db.Save(&transaction).Error)
	require.NoError(t, db.Model(&orderdomain.Order{}).Where("id = ?", orderRecord.ID).Update("currency", "JPY").Error)

	card := coupondomain.GiftCard{
		Code:         "GC-JPY-SPLIT",
		InitialCents: 3000,
		BalanceCents: 0,
		Currency:     "JPY",
		Status:       "used",
	}
	require.NoError(t, db.Create(&card).Error)
	require.NoError(t, db.Create(&coupondomain.GiftCardTransaction{
		GiftCardID:   card.ID,
		OrderID:      orderRecord.ID,
		Currency:     "JPY",
		Type:         "use",
		AmountCents:  -3000,
		BalanceCents: 0,
	}).Error)

	repos := repository.TxRepositories{
		Coupon:  repository.NewCouponRepository(db),
		Payment: repository.NewPaymentRepository(db),
	}
	split, err := calculateRefundPaymentSplit(
		repos,
		orderRecord.ID,
		domainmoney.MustNew(10000, "JPY"),
		domainmoney.MustNew(7000, "JPY"),
		domainmoney.MustNew(0, "JPY"),
	)
	require.NoError(t, err)
	require.Equal(t, int64(7000), split.GatewayAmount.AmountMinor())
	require.Equal(t, int64(3000), split.GiftCardAmount.AmountMinor())
}

func TestCalculateRefundPaymentSplitRejectsGiftCardCurrencyMismatch(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	orderRecord, _ := createRefundRecommendationPaidOrder(t, db, 100)
	seedRefundGiftCardPayment(t, db, orderRecord.ID, "GC-CURRENCY-MISMATCH", 1000)
	repos := repository.TxRepositories{
		Coupon:  repository.NewCouponRepository(db),
		Payment: repository.NewPaymentRepository(db),
	}

	_, err := calculateRefundPaymentSplit(
		repos,
		orderRecord.ID,
		domainmoney.MustNew(100, "JPY"),
		domainmoney.MustNew(100, "JPY"),
		domainmoney.MustNew(0, "JPY"),
	)
	require.ErrorContains(t, err, "does not match transaction currency")
}
