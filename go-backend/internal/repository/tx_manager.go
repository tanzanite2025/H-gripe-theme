package repository

import "gorm.io/gorm"

type TxManager struct {
	db                            *gorm.DB
	orderRepo                     *OrderRepository
	orderIdempotencyRepo          *OrderIdempotencyRepository
	attributionRepo               *OrderAttributionRepository
	productRepo                   *ProductRepository
	couponRepo                    *CouponRepository
	loyaltyRepo                   *LoyaltyRepository
	programRepo                   *LoyaltyProgramRepository
	referralRepo                  *ReferralRepository
	referralProgramRepo           *ReferralProgramRepository
	paymentRepo                   *PaymentRepository
	refundReviewRepo              *PaymentRefundRecommendationRepository
	refundExecRepo                *PaymentRefundExecutionRepository
	refundIdempotencyRepo         *PaymentRefundIdempotencyRepository
	afterSalesRefundRepo          *AfterSalesRefundReviewRepository
	afterSalesCaseRepo            *AfterSalesCaseRepository
	shippingRepo                  *ShippingRepository
	settingRepo                   *SettingRepository
	exchangeRateRepo              *ExchangeRateRepository
	policyDisclosureRepo          *OrderPolicyDisclosureRepository
	productQualityRequirementRepo *ProductQualityRequirementRepository
	orderEvidenceSnapshotRepo     *OrderEvidenceSnapshotRepository
	orderEvidenceRepo             *OrderEvidenceRepository
	orderEvidenceSubmissionRepo   *OrderEvidenceSubmissionSnapshotRepository
	outboxRepo                    *OutboxRepository
	productBrandRepo              *ProductBrandRepository
	cartRepo                      *CartRepository
}

func (m *TxManager) OrderRepository() *OrderRepository {
	if m == nil {
		return nil
	}
	return m.orderRepo
}

type TxRepositories struct {
	Order                     *OrderRepository
	OrderIdempotency          *OrderIdempotencyRepository
	OrderAttribution          *OrderAttributionRepository
	Product                   *ProductRepository
	Coupon                    *CouponRepository
	Loyalty                   *LoyaltyRepository
	Program                   *LoyaltyProgramRepository
	Referral                  *ReferralRepository
	ReferralProgram           *ReferralProgramRepository
	Payment                   *PaymentRepository
	RefundReview              *PaymentRefundRecommendationRepository
	RefundExecution           *PaymentRefundExecutionRepository
	RefundIdempotency         *PaymentRefundIdempotencyRepository
	AfterSalesRefund          *AfterSalesRefundReviewRepository
	AfterSalesCase            *AfterSalesCaseRepository
	Shipping                  *ShippingRepository
	Setting                   *SettingRepository
	ExchangeRate              *ExchangeRateRepository
	PolicyDisclosure          *OrderPolicyDisclosureRepository
	ProductQualityRequirement *ProductQualityRequirementRepository
	OrderEvidenceSnapshot     *OrderEvidenceSnapshotRepository
	OrderEvidence             *OrderEvidenceRepository
	OrderEvidenceSubmission   *OrderEvidenceSubmissionSnapshotRepository
	Outbox                    *OutboxRepository
	ProductBrand              *ProductBrandRepository
	Cart                      *CartRepository
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

func (m *TxManager) ConfigureLoyaltyProgramRepository(repo *LoyaltyProgramRepository) {
	m.programRepo = repo
}

func (m *TxManager) ConfigureReferralRepositories(referralRepo *ReferralRepository, programRepo *ReferralProgramRepository) {
	m.referralRepo = referralRepo
	m.referralProgramRepo = programRepo
}

func (m *TxManager) ConfigureOutboxRepository(repo *OutboxRepository) {
	m.outboxRepo = repo
}

func (m *TxManager) ConfigureProductBrandRepository(repo *ProductBrandRepository) {
	m.productBrandRepo = repo
}

func (m *TxManager) ConfigureCartRepository(repo *CartRepository) {
	m.cartRepo = repo
}

func (m *TxManager) ConfigurePaymentRefundRecommendationRepository(repo *PaymentRefundRecommendationRepository) {
	m.refundReviewRepo = repo
}

func (m *TxManager) ConfigurePaymentRefundExecutionRepository(repo *PaymentRefundExecutionRepository) {
	m.refundExecRepo = repo
}

func (m *TxManager) ConfigurePaymentRefundIdempotencyRepository(repo *PaymentRefundIdempotencyRepository) {
	m.refundIdempotencyRepo = repo
}

func (m *TxManager) ConfigureAfterSalesRefundReviewRepository(repo *AfterSalesRefundReviewRepository) {
	m.afterSalesRefundRepo = repo
}

func (m *TxManager) ConfigureAfterSalesCaseRepository(repo *AfterSalesCaseRepository) {
	m.afterSalesCaseRepo = repo
}

func (m *TxManager) ConfigureOrderAttributionRepository(repo *OrderAttributionRepository) {
	m.attributionRepo = repo
}

func (m *TxManager) ConfigureOrderIdempotencyRepository(repo *OrderIdempotencyRepository) {
	m.orderIdempotencyRepo = repo
}

func (m *TxManager) ConfigureSettingRepository(repo *SettingRepository) {
	m.settingRepo = repo
}

func (m *TxManager) ConfigureExchangeRateRepository(repo *ExchangeRateRepository) {
	m.exchangeRateRepo = repo
}

func (m *TxManager) ConfigureOrderPolicyDisclosureRepository(repo *OrderPolicyDisclosureRepository) {
	m.policyDisclosureRepo = repo
}

func (m *TxManager) ConfigureProductQualityRequirementRepository(repo *ProductQualityRequirementRepository) {
	m.productQualityRequirementRepo = repo
}

func (m *TxManager) ConfigureOrderEvidenceSnapshotRepository(repo *OrderEvidenceSnapshotRepository) {
	m.orderEvidenceSnapshotRepo = repo
}

func (m *TxManager) ConfigureOrderEvidenceRepository(repo *OrderEvidenceRepository) {
	m.orderEvidenceRepo = repo
}

func (m *TxManager) ConfigureOrderEvidenceSubmissionSnapshotRepository(repo *OrderEvidenceSubmissionSnapshotRepository) {
	m.orderEvidenceSubmissionRepo = repo
}

func (m *TxManager) WithinTx(fn func(TxRepositories) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		var shippingRepo *ShippingRepository
		if m.shippingRepo != nil {
			shippingRepo = m.shippingRepo.WithTx(tx)
		}
		var programRepo *LoyaltyProgramRepository
		if m.programRepo != nil {
			programRepo = m.programRepo.WithTx(tx)
		}
		var referralRepo *ReferralRepository
		if m.referralRepo != nil {
			referralRepo = m.referralRepo.WithTx(tx)
		}
		var referralProgramRepo *ReferralProgramRepository
		if m.referralProgramRepo != nil {
			referralProgramRepo = m.referralProgramRepo.WithTx(tx)
		}
		var outboxRepo *OutboxRepository
		if m.outboxRepo != nil {
			outboxRepo = m.outboxRepo.WithTx(tx)
		}
		var attributionRepo *OrderAttributionRepository
		if m.attributionRepo != nil {
			attributionRepo = m.attributionRepo.WithTx(tx)
		}
		var orderIdempotencyRepo *OrderIdempotencyRepository
		if m.orderIdempotencyRepo != nil {
			orderIdempotencyRepo = m.orderIdempotencyRepo.WithTx(tx)
		}
		var refundReviewRepo *PaymentRefundRecommendationRepository
		if m.refundReviewRepo != nil {
			refundReviewRepo = m.refundReviewRepo.WithTx(tx)
		}
		var refundExecRepo *PaymentRefundExecutionRepository
		if m.refundExecRepo != nil {
			refundExecRepo = m.refundExecRepo.WithTx(tx)
		}
		var refundIdempotencyRepo *PaymentRefundIdempotencyRepository
		if m.refundIdempotencyRepo != nil {
			refundIdempotencyRepo = m.refundIdempotencyRepo.WithTx(tx)
		}
		var afterSalesRefundRepo *AfterSalesRefundReviewRepository
		if m.afterSalesRefundRepo != nil {
			afterSalesRefundRepo = m.afterSalesRefundRepo.WithTx(tx)
		}
		var afterSalesCaseRepo *AfterSalesCaseRepository
		if m.afterSalesCaseRepo != nil {
			afterSalesCaseRepo = m.afterSalesCaseRepo.WithTx(tx)
		}
		var settingRepo *SettingRepository
		if m.settingRepo != nil {
			settingRepo = m.settingRepo.WithTx(tx)
		}
		var exchangeRateRepo *ExchangeRateRepository
		if m.exchangeRateRepo != nil {
			exchangeRateRepo = m.exchangeRateRepo.WithTx(tx)
		}
		var policyDisclosureRepo *OrderPolicyDisclosureRepository
		if m.policyDisclosureRepo != nil {
			policyDisclosureRepo = m.policyDisclosureRepo.WithTx(tx)
		}
		var productQualityRequirementRepo *ProductQualityRequirementRepository
		if m.productQualityRequirementRepo != nil {
			productQualityRequirementRepo = m.productQualityRequirementRepo.WithTx(tx)
		}
		var orderEvidenceSnapshotRepo *OrderEvidenceSnapshotRepository
		if m.orderEvidenceSnapshotRepo != nil {
			orderEvidenceSnapshotRepo = m.orderEvidenceSnapshotRepo.WithTx(tx)
		}
		var orderEvidenceRepo *OrderEvidenceRepository
		if m.orderEvidenceRepo != nil {
			orderEvidenceRepo = m.orderEvidenceRepo.WithTx(tx)
		}
		var orderEvidenceSubmissionRepo *OrderEvidenceSubmissionSnapshotRepository
		if m.orderEvidenceSubmissionRepo != nil {
			orderEvidenceSubmissionRepo = m.orderEvidenceSubmissionRepo.WithTx(tx)
		}
		var productBrandRepo *ProductBrandRepository
		if m.productBrandRepo != nil {
			productBrandRepo = m.productBrandRepo.WithTx(tx)
		}
		var cartRepo *CartRepository
		if m.cartRepo != nil {
			cartRepo = m.cartRepo.WithTx(tx)
		}
		return fn(TxRepositories{
			Order:                     m.orderRepo.WithTx(tx),
			OrderIdempotency:          orderIdempotencyRepo,
			OrderAttribution:          attributionRepo,
			Product:                   m.productRepo.WithTx(tx),
			Coupon:                    m.couponRepo.WithTx(tx),
			Loyalty:                   m.loyaltyRepo.WithTx(tx),
			Program:                   programRepo,
			Referral:                  referralRepo,
			ReferralProgram:           referralProgramRepo,
			Payment:                   m.paymentRepo.WithTx(tx),
			RefundReview:              refundReviewRepo,
			RefundExecution:           refundExecRepo,
			RefundIdempotency:         refundIdempotencyRepo,
			AfterSalesRefund:          afterSalesRefundRepo,
			AfterSalesCase:            afterSalesCaseRepo,
			Shipping:                  shippingRepo,
			Setting:                   settingRepo,
			ExchangeRate:              exchangeRateRepo,
			PolicyDisclosure:          policyDisclosureRepo,
			ProductQualityRequirement: productQualityRequirementRepo,
			OrderEvidenceSnapshot:     orderEvidenceSnapshotRepo,
			OrderEvidence:             orderEvidenceRepo,
			OrderEvidenceSubmission:   orderEvidenceSubmissionRepo,
			Outbox:                    outboxRepo,
			ProductBrand:              productBrandRepo,
			Cart:                      cartRepo,
		})
	})
}
