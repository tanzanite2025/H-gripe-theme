package repository

import "gorm.io/gorm"

type TxManager struct {
	db           *gorm.DB
	orderRepo    *OrderRepository
	productRepo  *ProductRepository
	couponRepo   *CouponRepository
	loyaltyRepo  *LoyaltyRepository
	paymentRepo  *PaymentRepository
	shippingRepo *ShippingRepository
}

type TxRepositories struct {
	Order    *OrderRepository
	Product  *ProductRepository
	Coupon   *CouponRepository
	Loyalty  *LoyaltyRepository
	Payment  *PaymentRepository
	Shipping *ShippingRepository
}

func NewTxManager(
	db *gorm.DB,
	orderRepo *OrderRepository,
	productRepo *ProductRepository,
	couponRepo *CouponRepository,
	loyaltyRepo *LoyaltyRepository,
	paymentRepo *PaymentRepository,
	shippingRepo ...*ShippingRepository,
) *TxManager {
	manager := &TxManager{
		db:          db,
		orderRepo:   orderRepo,
		productRepo: productRepo,
		couponRepo:  couponRepo,
		loyaltyRepo: loyaltyRepo,
		paymentRepo: paymentRepo,
	}
	if len(shippingRepo) > 0 {
		manager.shippingRepo = shippingRepo[0]
	}
	return manager
}

func (m *TxManager) WithinTx(fn func(TxRepositories) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		var shippingRepo *ShippingRepository
		if m.shippingRepo != nil {
			shippingRepo = m.shippingRepo.WithTx(tx)
		}
		return fn(TxRepositories{
			Order:    m.orderRepo.WithTx(tx),
			Product:  m.productRepo.WithTx(tx),
			Coupon:   m.couponRepo.WithTx(tx),
			Loyalty:  m.loyaltyRepo.WithTx(tx),
			Payment:  m.paymentRepo.WithTx(tx),
			Shipping: shippingRepo,
		})
	})
}
