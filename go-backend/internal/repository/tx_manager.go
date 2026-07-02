package repository

import "gorm.io/gorm"

type TxManager struct {
	db          *gorm.DB
	orderRepo   *OrderRepository
	productRepo *ProductRepository
	couponRepo  *CouponRepository
	loyaltyRepo *LoyaltyRepository
	paymentRepo *PaymentRepository
}

type TxRepositories struct {
	Order   *OrderRepository
	Product *ProductRepository
	Coupon  *CouponRepository
	Loyalty *LoyaltyRepository
	Payment *PaymentRepository
}

func NewTxManager(
	db *gorm.DB,
	orderRepo *OrderRepository,
	productRepo *ProductRepository,
	couponRepo *CouponRepository,
	loyaltyRepo *LoyaltyRepository,
	paymentRepo *PaymentRepository,
) *TxManager {
	return &TxManager{
		db:          db,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		couponRepo:  couponRepo,
		loyaltyRepo: loyaltyRepo,
		paymentRepo: paymentRepo,
	}
}

func (m *TxManager) WithinTx(fn func(TxRepositories) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		return fn(TxRepositories{
			Order:   m.orderRepo.WithTx(tx),
			Product: m.productRepo.WithTx(tx),
			Coupon:  m.couponRepo.WithTx(tx),
			Loyalty: m.loyaltyRepo.WithTx(tx),
			Payment: m.paymentRepo.WithTx(tx),
		})
	})
}
