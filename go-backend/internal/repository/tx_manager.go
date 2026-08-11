package repository

import "gorm.io/gorm"

type TxManager struct {
	db               *gorm.DB
	orderRepo        *OrderRepository
	attributionRepo  *OrderAttributionRepository
	productRepo      *ProductRepository
	couponRepo       *CouponRepository
	loyaltyRepo      *LoyaltyRepository
	programRepo      *LoyaltyProgramRepository
	redemptionRepo   *GiftCardRedemptionRepository
	paymentRepo      *PaymentRepository
	refundReviewRepo *PaymentRefundRecommendationRepository
	refundExecRepo   *PaymentRefundExecutionRepository
	shippingRepo     *ShippingRepository
	settingRepo      *SettingRepository
	exchangeRateRepo *ExchangeRateRepository
	outboxRepo       *OutboxRepository
}

type TxRepositories struct {
	Order            *OrderRepository
	OrderAttribution *OrderAttributionRepository
	Product          *ProductRepository
	Coupon           *CouponRepository
	Loyalty          *LoyaltyRepository
	Program          *LoyaltyProgramRepository
	Redemption       *GiftCardRedemptionRepository
	Payment          *PaymentRepository
	RefundReview     *PaymentRefundRecommendationRepository
	RefundExecution  *PaymentRefundExecutionRepository
	Shipping         *ShippingRepository
	Setting          *SettingRepository
	ExchangeRate     *ExchangeRateRepository
	Outbox           *OutboxRepository
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

func (m *TxManager) ConfigureGiftCardRedemptionRepository(repo *GiftCardRedemptionRepository) {
	m.redemptionRepo = repo
}

func (m *TxManager) ConfigureLoyaltyProgramRepository(repo *LoyaltyProgramRepository) {
	m.programRepo = repo
}

func (m *TxManager) ConfigureOutboxRepository(repo *OutboxRepository) {
	m.outboxRepo = repo
}

func (m *TxManager) ConfigurePaymentRefundRecommendationRepository(repo *PaymentRefundRecommendationRepository) {
	m.refundReviewRepo = repo
}

func (m *TxManager) ConfigurePaymentRefundExecutionRepository(repo *PaymentRefundExecutionRepository) {
	m.refundExecRepo = repo
}

func (m *TxManager) ConfigureOrderAttributionRepository(repo *OrderAttributionRepository) {
	m.attributionRepo = repo
}

func (m *TxManager) ConfigureSettingRepository(repo *SettingRepository) {
	m.settingRepo = repo
}

func (m *TxManager) ConfigureExchangeRateRepository(repo *ExchangeRateRepository) {
	m.exchangeRateRepo = repo
}

func (m *TxManager) WithinTx(fn func(TxRepositories) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		var shippingRepo *ShippingRepository
		if m.shippingRepo != nil {
			shippingRepo = m.shippingRepo.WithTx(tx)
		}
		var redemptionRepo *GiftCardRedemptionRepository
		if m.redemptionRepo != nil {
			redemptionRepo = m.redemptionRepo.WithTx(tx)
		}
		var programRepo *LoyaltyProgramRepository
		if m.programRepo != nil {
			programRepo = m.programRepo.WithTx(tx)
		}
		var outboxRepo *OutboxRepository
		if m.outboxRepo != nil {
			outboxRepo = m.outboxRepo.WithTx(tx)
		}
		var attributionRepo *OrderAttributionRepository
		if m.attributionRepo != nil {
			attributionRepo = m.attributionRepo.WithTx(tx)
		}
		var refundReviewRepo *PaymentRefundRecommendationRepository
		if m.refundReviewRepo != nil {
			refundReviewRepo = m.refundReviewRepo.WithTx(tx)
		}
		var refundExecRepo *PaymentRefundExecutionRepository
		if m.refundExecRepo != nil {
			refundExecRepo = m.refundExecRepo.WithTx(tx)
		}
		var settingRepo *SettingRepository
		if m.settingRepo != nil {
			settingRepo = m.settingRepo.WithTx(tx)
		}
		var exchangeRateRepo *ExchangeRateRepository
		if m.exchangeRateRepo != nil {
			exchangeRateRepo = m.exchangeRateRepo.WithTx(tx)
		}
		return fn(TxRepositories{
			Order:            m.orderRepo.WithTx(tx),
			OrderAttribution: attributionRepo,
			Product:          m.productRepo.WithTx(tx),
			Coupon:           m.couponRepo.WithTx(tx),
			Loyalty:          m.loyaltyRepo.WithTx(tx),
			Program:          programRepo,
			Redemption:       redemptionRepo,
			Payment:          m.paymentRepo.WithTx(tx),
			RefundReview:     refundReviewRepo,
			RefundExecution:  refundExecRepo,
			Shipping:         shippingRepo,
			Setting:          settingRepo,
			ExchangeRate:     exchangeRateRepo,
			Outbox:           outboxRepo,
		})
	})
}
