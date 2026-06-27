package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockCouponRepository(t *testing.T) (*CouponRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		sqlDB.Close()
		t.Fatalf("failed to create gorm db: %v", err)
	}

	return NewCouponRepository(db), mock, func() {
		sqlDB.Close()
	}
}

func TestIncrementUsedCountRequiresAvailableUsageSlot(t *testing.T) {
	repo, mock, cleanup := newMockCouponRepository(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "coupons" SET "used_count"=used_count + $1 WHERE (id = $2 AND (usage_limit = 0 OR used_count < usage_limit)) AND "coupons"."deleted_at" IS NULL`)).
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.IncrementUsedCount(42)
	if !errors.Is(err, ErrCouponUsageLimitReached) {
		t.Fatalf("expected ErrCouponUsageLimitReached, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUpdateGiftCardBalanceRequiresSufficientBalanceForDebit(t *testing.T) {
	repo, mock, cleanup := newMockCouponRepository(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "gift_cards" SET "balance"=balance + $1 WHERE id = $2 AND balance >= $3 AND "gift_cards"."deleted_at" IS NULL`)).
		WithArgs(-8.0, sqlmock.AnyArg(), 8.0).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateGiftCardBalance(42, -8)
	if !errors.Is(err, ErrGiftCardInsufficientBalance) {
		t.Fatalf("expected ErrGiftCardInsufficientBalance, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
