package service

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	domainpricing "commerce-platform/internal/domain/pricing"
	productdomain "commerce-platform/internal/domain/product"
	paymentpkg "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type CheckoutService struct {
	productRepo         *repository.ProductRepository
	couponRepo          *repository.CouponRepository
	paymentRepo         *repository.PaymentRepository
	loyaltyRepo         *repository.LoyaltyRepository
	referralRepo        *repository.ReferralRepository
	referralProgramRepo *repository.ReferralProgramRepository
	shippingService     *ShippingService
	loyaltyProgram      *LoyaltyProgramService
	currencyPolicy      *CurrencyPolicyService
	exchangeRates       *repository.ExchangeRateRepository
}

type CheckoutQuoteInput struct {
	UserID               uint
	Items                []order.OrderItem
	ShippingAddress      order.Address
	DisplayCurrency      string
	PaymentMethod        string
	ShippingQuoteID      string
	SelectedQuotePlanID  string
	CouponCode           string
	GiftCardCode         string
	PointsToUse          int
	LoyaltyProgramConfig *loyalty.ProgramConfig
}

type CheckoutQuote struct {
	Items                 []order.OrderItem        `json:"items"`
	SubtotalMinor         int64                    `json:"subtotal_minor"`
	ShippingFeeMinor      int64                    `json:"shipping_fee_minor"`
	TaxMinor              int64                    `json:"tax_minor"`
	MemberDiscountMinor   int64                    `json:"member_discount_minor"`
	PointsDiscountMinor   int64                    `json:"points_discount_minor"`
	CouponDiscountMinor   int64                    `json:"coupon_discount_minor"`
	GiftCardDiscountMinor int64                    `json:"gift_card_discount_minor"`
	DiscountMinor         int64                    `json:"discount_minor"`
	TotalMinor            int64                    `json:"total_minor"`
	SubtotalAmount        float64                  `json:"subtotal_amount"`
	ShippingFee           float64                  `json:"shipping_fee"`
	ShippingQuote         *ShippingQuote           `json:"shipping_quote,omitempty"`
	TaxAmount             float64                  `json:"tax_amount"`
	MemberDiscount        float64                  `json:"member_discount"`
	PointsDiscount        float64                  `json:"points_discount"`
	CouponDiscount        float64                  `json:"coupon_discount"`
	GiftCardDiscount      float64                  `json:"gift_card_discount"`
	DiscountAmount        float64                  `json:"discount_amount"`
	TotalAmount           float64                  `json:"total_amount"`
	CouponCode            string                   `json:"coupon_code"`
	GiftCardCode          string                   `json:"gift_card_code"`
	PointsToUse           int                      `json:"points_to_use"`
	ProgramConfigID       *uint                    `json:"loyalty_program_config_id,omitempty"`
	Coupon                *coupon.Coupon           `json:"coupon,omitempty"`
	Currency              string                   `json:"currency"`
	PaymentCurrency       string                   `json:"payment_currency,omitempty"`
	PaymentAmountMinor    int64                    `json:"payment_amount_minor,omitempty"`
	PaymentAmount         float64                  `json:"payment_amount,omitempty"`
	FXSnapshot            currency.OrderFXSnapshot `json:"-"`
	GiftCard              *coupon.GiftCard         `json:"-"`
	GiftCardDiscountCents int64                    `json:"-"`
	PricingSnapshot       domainpricing.Snapshot   `json:"-"`
}

const checkoutMaxPointsDiscountSubtotalRate = 0.5

type checkoutRepositories struct {
	productRepo         *repository.ProductRepository
	couponRepo          *repository.CouponRepository
	paymentRepo         *repository.PaymentRepository
	loyaltyRepo         *repository.LoyaltyRepository
	referralRepo        *repository.ReferralRepository
	referralProgramRepo *repository.ReferralProgramRepository
	shippingService     *ShippingService
	currencyPolicy      *CurrencyPolicyService
	exchangeRates       *repository.ExchangeRateRepository
	lockCoupon          bool
}

func NewCheckoutService(
	productRepo *repository.ProductRepository,
	couponRepo *repository.CouponRepository,
	paymentRepo *repository.PaymentRepository,
	loyaltyRepo *repository.LoyaltyRepository,
	shippingServices ...*ShippingService,
) *CheckoutService {
	checkoutService := &CheckoutService{
		productRepo: productRepo,
		couponRepo:  couponRepo,
		paymentRepo: paymentRepo,
		loyaltyRepo: loyaltyRepo,
	}
	if len(shippingServices) > 0 {
		checkoutService.shippingService = shippingServices[0]
	}
	return checkoutService
}

func (s *CheckoutService) ConfigureLoyaltyProgram(program *LoyaltyProgramService) {
	s.loyaltyProgram = program
}

func (s *CheckoutService) ConfigureCurrencyPolicy(policy *CurrencyPolicyService) {
	if s != nil {
		s.currencyPolicy = policy
	}
}

func (s *CheckoutService) ConfigureExchangeRateRepository(repo *repository.ExchangeRateRepository) {
	if s != nil {
		s.exchangeRates = repo
	}
}

func (s *CheckoutService) ConfigureReferralRepositories(referralRepo *repository.ReferralRepository, programRepo *repository.ReferralProgramRepository) {
	if s == nil {
		return
	}
	s.referralRepo = referralRepo
	s.referralProgramRepo = programRepo
}

func (s *CheckoutService) Quote(input CheckoutQuoteInput) (*CheckoutQuote, error) {
	return s.quote(input, checkoutRepositories{
		productRepo:         s.productRepo,
		couponRepo:          s.couponRepo,
		paymentRepo:         s.paymentRepo,
		loyaltyRepo:         s.loyaltyRepo,
		shippingService:     s.shippingService,
		currencyPolicy:      s.currencyPolicy,
		exchangeRates:       s.exchangeRates,
		referralRepo:        s.referralRepo,
		referralProgramRepo: s.referralProgramRepo,
	})
}

func (s *CheckoutService) QuoteWithRepositories(input CheckoutQuoteInput, repos repository.TxRepositories) (*CheckoutQuote, error) {
	if repos.Setting == nil {
		return nil, errors.New("transactional checkout currency policy repository is not configured")
	}
	shippingService := s.shippingService
	if repos.Shipping != nil {
		shippingService = NewShippingService(repos.Shipping, repos.Product)
		shippingService.ConfigureExchangeRateService(NewExchangeRateService(repos.ExchangeRate, repos.Setting))
	}
	return s.quote(input, checkoutRepositories{
		productRepo:         repos.Product,
		couponRepo:          repos.Coupon,
		paymentRepo:         repos.Payment,
		loyaltyRepo:         repos.Loyalty,
		shippingService:     shippingService,
		currencyPolicy:      NewCurrencyPolicyService(repos.Setting),
		exchangeRates:       repos.ExchangeRate,
		referralRepo:        repos.Referral,
		referralProgramRepo: repos.ReferralProgram,
		lockCoupon:          true,
	})
}

// lockedRefereeCouponCode returns the one-time coupon provisioned when a
// customer accepted a referral. It is intentionally only considered while the
// referral is still pending; after payment the outbox handler releases the
// reward log and subsequent quotes must not reuse it.
func (s *CheckoutService) lockedRefereeCouponCode(repos checkoutRepositories, userID uint) (string, error) {
	if userID == 0 || repos.referralRepo == nil || repos.referralProgramRepo == nil {
		return "", nil
	}
	record, err := repos.referralRepo.FindRecordByRefereeID(userID)
	if repository.IsRecordNotFound(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if record.Status != loyalty.ReferralStatusPending || record.OrderID != nil {
		return "", nil
	}
	config, err := repos.referralProgramRepo.FindByID(record.ProgramConfigID)
	if err != nil {
		return "", err
	}
	if config.RefereeBenefitType != loyalty.ReferralBenefitPercentCoupon && config.RefereeBenefitType != loyalty.ReferralBenefitFixedCoupon {
		return "", nil
	}
	rewardKey := fmt.Sprintf("referral:%d:referee:%s:v1", record.ID, config.RefereeBenefitType)
	reward, err := repos.referralRepo.FindRewardByIdempotencyKey(rewardKey)
	if repository.IsRecordNotFound(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if reward.Status != loyalty.ReferralRewardStatusLocked || reward.CouponID == nil || reward.RecipientUserID != userID {
		return "", nil
	}
	if repos.couponRepo == nil {
		return "", errors.New("coupon repository is not configured")
	}
	couponRecord, err := repos.couponRepo.FindCouponByID(*reward.CouponID)
	if repository.IsRecordNotFound(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(couponRecord.Code), nil
}

func (s *CheckoutService) quote(input CheckoutQuoteInput, repos checkoutRepositories) (*CheckoutQuote, error) {
	if len(input.Items) == 0 {
		return nil, errors.New("cart is empty")
	}
	items := make([]order.OrderItem, len(input.Items))
	pricingItems := make([]checkoutPricingLineInput, len(input.Items))
	shippingItems := make([]ShippingQuoteItemInput, 0, len(input.Items))
	var subtotalMoney domainmoney.Money
	var subtotalMoneyInitialized bool
	quoteCurrency := ""
	requestedCurrency := currency.NormalizeCode(input.DisplayCurrency)
	if requestedCurrency != "" && (!currency.IsValidCode(requestedCurrency) || !currency.IsCatalogCode(requestedCurrency)) {
		return nil, fmt.Errorf("unsupported checkout display currency %s", requestedCurrency)
	}
	for i, item := range input.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product ID %d", item.ProductID)
		}

		product, variant, err := repos.productRepo.FindPurchasableVariant(item.ProductID, item.VariantID)
		if err != nil {
			return nil, fmt.Errorf("[CRITICAL] Product ID %d not found in database: %w", item.ProductID, err)
		}
		if variant == nil {
			return nil, fmt.Errorf("[CRITICAL] Product ID %d has no purchasable variant", item.ProductID)
		}
		selectedOptions, configErr := SelectedOptionsFromConfiguration(item.ConfigurationData)
		if configErr != nil {
			return nil, fmt.Errorf("invalid configuration for product ID %d: %w", item.ProductID, configErr)
		}
		configuration, configErr := ResolveProductConfiguration(product, variant, selectedOptions)
		if configErr != nil {
			return nil, fmt.Errorf("configuration conflict for product ID %d: %w", item.ProductID, configErr)
		}

		resolvedVariantID := variant.ID
		variantID := &resolvedVariantID
		itemCurrency := currency.NormalizeCode(variant.Currency)
		if itemCurrency == "" {
			itemCurrency = productdomain.DefaultPriceCurrency
		}
		if !currency.IsValidCode(itemCurrency) || !currency.IsCatalogCode(itemCurrency) {
			return nil, fmt.Errorf("unsupported price currency for SKU %s", variant.SKU)
		}
		if quoteCurrency == "" {
			quoteCurrency = requestedCurrency
			if quoteCurrency == "" {
				quoteCurrency = itemCurrency
			}
		}
		priceMoney, err := variant.EffectivePriceMoney()
		if err != nil {
			return nil, fmt.Errorf("invalid price for SKU %s: %w", variant.SKU, err)
		}
		if priceMoney.Currency().String() != itemCurrency {
			return nil, fmt.Errorf("price currency mismatch for SKU %s", variant.SKU)
		}
		basePriceMoney := priceMoney
		priceMoney, err = priceMoney.Add(configuration.Delta)
		if err != nil {
			return nil, fmt.Errorf("calculate configured price for SKU %s: %w", variant.SKU, err)
		}
		hasStoredConfiguration := len(item.ConfigurationData) > 0 || strings.TrimSpace(item.ConfigurationHash) != ""
		if strings.TrimSpace(item.ConfigurationHash) != "" && item.ConfigurationHash != configuration.Hash {
			return nil, fmt.Errorf("%w: configuration hash for SKU %s does not match the catalog", ErrProductConfigurationConflict, variant.SKU)
		}
		if hasStoredConfiguration {
			expectedMoney, expectedErr := domainmoney.FromMajorFloat(item.Price, itemCurrency)
			if expectedErr != nil || expectedMoney.AmountMinor() != priceMoney.AmountMinor() {
				return nil, fmt.Errorf("%w: SKU %s", ErrProductConfigurationPriceChanged, variant.SKU)
			}
		}
		if quoteCurrency != itemCurrency {
			if repos.exchangeRates == nil {
				return nil, fmt.Errorf("exchange rate repository is required to convert %s to %s", itemCurrency, quoteCurrency)
			}
			conversionService := NewExchangeRateService(repos.exchangeRates, nil)
			conversionService.ConfigureCurrencyPolicy(repos.currencyPolicy)
			convertedMoney, conversionErr := conversionService.ConvertMoneyStrict(priceMoney, quoteCurrency)
			if conversionErr != nil {
				return nil, fmt.Errorf("exchange rate unavailable for %s to %s checkout pricing: %w", itemCurrency, quoteCurrency, conversionErr)
			}
			priceMoney = convertedMoney
			basePriceMoney, conversionErr = conversionService.ConvertMoneyStrict(basePriceMoney, quoteCurrency)
			if conversionErr != nil {
				return nil, fmt.Errorf("exchange rate unavailable for %s to %s checkout base pricing: %w", itemCurrency, quoteCurrency, conversionErr)
			}
		}
		if priceMoney.Currency().String() != quoteCurrency {
			return nil, fmt.Errorf("converted price currency mismatch for SKU %s", variant.SKU)
		}
		price, err := priceMoney.MajorFloat()
		if err != nil {
			return nil, fmt.Errorf("format price for SKU %s: %w", variant.SKU, err)
		}
		configuration.Snapshot.ProductID = product.ID
		configuration.Snapshot.VariantID = variant.ID
		configuration.Snapshot.ProductName = product.Name
		configuration.Snapshot.VariantName = variant.Title
		configuration.Snapshot.ConfigurationHash = configuration.Hash
		configuration.Snapshot.NormalizedConfiguration = append([]byte(nil), configuration.Data...)
		configuration.Snapshot.Currency = quoteCurrency
		configuration.Snapshot.PriceBreakdown.BaseUnitPriceMinor = basePriceMoney.AmountMinor()
		configuration.Snapshot.PriceBreakdown.FinalUnitPriceMinor = priceMoney.AmountMinor()
		configuration.Snapshot.PriceBreakdown.OptionsUnitPriceMinor = priceMoney.AmountMinor() - basePriceMoney.AmountMinor()
		configurationSnapshotData, snapshotErr := json.Marshal(configuration.Snapshot)
		if snapshotErr != nil {
			return nil, fmt.Errorf("encode configuration snapshot for SKU %s: %w", variant.SKU, snapshotErr)
		}
		lineSubtotalMoney, err := priceMoney.MultiplyInt(int64(item.Quantity))
		if err != nil {
			return nil, fmt.Errorf("calculate subtotal for SKU %s: %w", variant.SKU, err)
		}
		lineSubtotal, err := lineSubtotalMoney.MajorFloat()
		if err != nil {
			return nil, fmt.Errorf("format subtotal for SKU %s: %w", variant.SKU, err)
		}
		if !subtotalMoneyInitialized {
			subtotalMoney, err = domainmoney.New(0, quoteCurrency)
			if err != nil {
				return nil, fmt.Errorf("initialize checkout subtotal: %w", err)
			}
			subtotalMoneyInitialized = true
		}
		subtotalMoney, err = subtotalMoney.Add(lineSubtotalMoney)
		if err != nil {
			return nil, fmt.Errorf("accumulate checkout subtotal: %w", err)
		}
		sku := variant.SKU
		attributes := variant.OptionValues
		availableStock := variant.Stock
		fulfillmentMode := productdomain.NormalizeFulfillmentMode(product.FulfillmentMode)
		if fulfillmentMode == productdomain.FulfillmentModeStock && availableStock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product ID %d", item.ProductID)
		}

		items[i] = item
		items[i].VariantID = variantID
		items[i].Currency = quoteCurrency
		items[i].ConfigurationData = configuration.Data
		items[i].ConfigurationHash = configuration.Hash
		items[i].ConfigurationSnapshotData = configurationSnapshotData
		items[i].ProductCategoryID = product.ProductCategoryID
		if product.ProductCategory != nil {
			items[i].ProductCategorySlug = product.ProductCategory.Slug
		}
		items[i].Price = price
		items[i].PriceMinor = priceMoney.AmountMinor()
		items[i].Subtotal = lineSubtotal
		items[i].SubtotalMinor = lineSubtotalMoney.AmountMinor()
		items[i].ProductName = product.Name
		items[i].SKU = sku
		items[i].Attributes = attributes
		items[i].WeightGrams = variant.Weight
		items[i].FulfillmentMode = fulfillmentMode
		if !productdomain.IsValidFulfillmentMode(items[i].FulfillmentMode) {
			return nil, fmt.Errorf("invalid fulfillment mode for product ID %d", product.ID)
		}
		items[i].Total = items[i].Subtotal
		items[i].TotalMinor = items[i].SubtotalMinor
		items[i].HSCode = product.HSCode
		items[i].CNCode = product.CNCode
		items[i].CountryOfOrigin = product.CountryOfOrigin
		items[i].CustomsDescription = product.CustomsDescription
		pricingItems[i] = checkoutPricingLineInput{
			ProductID:           items[i].ProductID,
			VariantID:           orderItemVariantID(items[i]),
			Quantity:            items[i].Quantity,
			ProductCategoryID:   items[i].ProductCategoryID,
			ProductCategorySlug: items[i].ProductCategorySlug,
			UnitPrice:           priceMoney,
			Subtotal:            lineSubtotalMoney,
		}
		items[i].DeclaredValue = nil
		items[i].DeclaredValueConfirmed = false
		if variant.Weight <= 0 {
			return nil, fmt.Errorf("shipping weight is missing for SKU %s", variant.SKU)
		}
		shippingTemplateID, err := resolveProductShippingTemplateID(product, variant)
		if err != nil {
			return nil, err
		}
		shippingItems = append(shippingItems, ShippingQuoteItemInput{
			ProductID:                      product.ID,
			VariantID:                      variantID,
			ProductSpecificationTemplateID: product.ProductSpecificationTemplateID,
			ShippingTemplateID:             uintPtr(shippingTemplateID),
			Quantity:                       item.Quantity,
			UnitPrice:                      price,
			WeightGrams:                    variant.Weight,
		})
	}
	if quoteCurrency == "" {
		return nil, errors.New("product price currency is required")
	}
	subtotal, err := subtotalMoney.MajorFloat()
	if err != nil {
		return nil, fmt.Errorf("format checkout subtotal: %w", err)
	}

	// Capture the order-time FX contract before valuing loyalty points. The
	// redemption rate is defined in the configured points currency, while the
	// quote and its discount cap are expressed in order currency.
	fxSnapshot, err := s.resolveOrderFXSnapshot(quoteCurrency, repos.currencyPolicy, repos.exchangeRates)
	if err != nil {
		return nil, err
	}
	loyaltyConfig := input.LoyaltyProgramConfig
	pointsFXSnapshot := fxSnapshot
	if input.PointsToUse > 0 {
		if loyaltyConfig == nil {
			loyaltyConfig, err = s.currentLoyaltyProgramConfig()
			if err != nil {
				return nil, err
			}
			if loyaltyConfig == nil {
				return nil, ErrLoyaltyProgramConfigNotFound
			}
		}
		pointsFXSnapshot, err = s.resolvePointsFXSnapshot(
			quoteCurrency,
			loyaltyConfig.Currency,
			repos.exchangeRates,
		)
		if err != nil {
			return nil, err
		}
	}

	if repos.shippingService == nil {
		return nil, errors.New("shipping quote service is not configured")
	}
	shippingQuote, err := repos.shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:             input.ShippingAddress.Country,
		PostalCode:          input.ShippingAddress.PostalCode,
		Amount:              subtotal,
		Currency:            quoteCurrency,
		DisplayCurrency:     input.DisplayCurrency,
		ShippingQuoteID:     input.ShippingQuoteID,
		SelectedQuotePlanID: input.SelectedQuotePlanID,
		Items:               shippingItems,
	})
	if err != nil {
		return nil, err
	}
	shippingFeeMoney, err := domainmoney.FromMajorFloat(shippingQuote.ShippingFee, quoteCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid shipping fee: %w", err)
	}
	shippingFee, err := shippingFeeMoney.MajorFloat()
	if err != nil {
		return nil, fmt.Errorf("format shipping fee: %w", err)
	}
	memberDiscountMoney, err := s.calculateMemberDiscountMoney(repos.loyaltyRepo, input.UserID, subtotalMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate member discount: %w", err)
	}
	if memberDiscountMoney.AmountMinor() > subtotalMoney.AmountMinor() {
		memberDiscountMoney = subtotalMoney
	}
	memberDiscount, err := memberDiscountMoney.MajorFloat()
	if err != nil {
		return nil, fmt.Errorf("format member discount: %w", err)
	}
	remainingMerchandiseMoney, err := subtotalMoney.Subtract(memberDiscountMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate remaining merchandise after member discount: %w", err)
	}
	var targetCoupon *coupon.Coupon
	couponDiscountMoney, err := domainmoney.New(0, quoteCurrency)
	if err != nil {
		return nil, fmt.Errorf("initialize coupon discount: %w", err)
	}
	couponDiscount := 0.0
	couponCode := strings.TrimSpace(input.CouponCode)
	if couponCode == "" {
		couponCode, err = s.lockedRefereeCouponCode(repos, input.UserID)
		if err != nil {
			return nil, fmt.Errorf("resolve referral benefit: %w", err)
		}
	}
	if couponCode != "" {
		targetCoupon, couponDiscountMoney, err = s.validateCouponWithFXMoney(
			repos.couponRepo,
			couponCode,
			input.UserID,
			input.ShippingAddress.Email,
			remainingMerchandiseMoney,
			pricingItems,
			fxSnapshot,
			repos.lockCoupon,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to apply coupon %s: %w", couponCode, err)
		}
		if couponDiscountMoney.AmountMinor() > remainingMerchandiseMoney.AmountMinor() {
			couponDiscountMoney = remainingMerchandiseMoney
		}
		couponDiscount, err = couponDiscountMoney.MajorFloat()
		if err != nil {
			return nil, fmt.Errorf("format coupon discount: %w", err)
		}
		remainingMerchandiseMoney, err = remainingMerchandiseMoney.Subtract(couponDiscountMoney)
		if err != nil {
			return nil, fmt.Errorf("calculate remaining merchandise after coupon discount: %w", err)
		}
	}

	pointsToUse, pointsDiscountMoney, programConfigID, err := s.calculatePointsDiscountMoney(
		repos.loyaltyRepo,
		input.UserID,
		input.PointsToUse,
		remainingMerchandiseMoney,
		loyaltyConfig,
		pointsFXSnapshot,
	)
	if err != nil {
		return nil, err
	}
	if pointsDiscountMoney.AmountMinor() > remainingMerchandiseMoney.AmountMinor() {
		pointsDiscountMoney = remainingMerchandiseMoney
	}
	pointsDiscount, err := pointsDiscountMoney.MajorFloat()
	if err != nil {
		return nil, fmt.Errorf("format points discount: %w", err)
	}
	remainingMerchandiseMoney, err = remainingMerchandiseMoney.Subtract(pointsDiscountMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate remaining merchandise after points discount: %w", err)
	}
	pricingSnapshot, err := buildCheckoutPricingSnapshot(checkoutPricingInput{
		Items:               pricingItems,
		Currency:            quoteCurrency,
		Subtotal:            subtotalMoney,
		MemberDiscount:      memberDiscountMoney,
		CouponDiscount:      couponDiscountMoney,
		Coupon:              targetCoupon,
		PointsDiscount:      pointsDiscountMoney,
		PointsToUse:         pointsToUse,
		ProgramConfigID:     programConfigID,
		MerchandiseNetTotal: remainingMerchandiseMoney,
	})
	if err != nil {
		return nil, fmt.Errorf("validate checkout pricing pipeline: %w", err)
	}
	taxMoney, err := s.calculateTaxMoney(
		repos.paymentRepo,
		remainingMerchandiseMoney,
		input.ShippingAddress.Country,
		input.ShippingAddress.State,
		input.ShippingAddress.PostalCode,
		quoteCurrency,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate tax: %w", err)
	}
	pricingSnapshot, err = pricingSnapshot.AllocateTax(taxMoney)
	if err != nil {
		return nil, fmt.Errorf("allocate checkout tax across pricing lines: %w", err)
	}
	taxAmount, err := taxMoney.MajorFloat()
	if err != nil {
		return nil, fmt.Errorf("format tax amount: %w", err)
	}

	giftCardCode := strings.TrimSpace(input.GiftCardCode)
	var giftCard *coupon.GiftCard
	giftCardDiscountCents := int64(0)
	grossMoney, err := subtotalMoney.Add(shippingFeeMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate checkout gross amount: %w", err)
	}
	grossMoney, err = grossMoney.Add(taxMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate checkout gross amount: %w", err)
	}
	remainingAmountMoney, err := grossMoney.Subtract(memberDiscountMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate remaining checkout amount: %w", err)
	}
	remainingAmountMoney, err = remainingAmountMoney.Subtract(pointsDiscountMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate remaining checkout amount: %w", err)
	}
	remainingAmountMoney, err = remainingAmountMoney.Subtract(couponDiscountMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate remaining checkout amount: %w", err)
	}
	if giftCardCode != "" {
		giftCard, err = s.validateGiftCard(
			repos.couponRepo,
			giftCardCode,
			input.UserID,
			quoteCurrency,
			repos.lockCoupon,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to apply gift card: %w", err)
		}

		if remainingAmountMoney.AmountMinor() > 0 {
			giftCardDiscountCents = minInt64(giftCard.BalanceCents, remainingAmountMoney.AmountMinor())
		}
	}
	giftCardDiscountMoney, err := domainmoney.New(giftCardDiscountCents, quoteCurrency)
	if err != nil {
		return nil, fmt.Errorf("calculate gift card discount: %w", err)
	}
	discountAmountMoney, err := memberDiscountMoney.Add(pointsDiscountMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate checkout discount total: %w", err)
	}
	discountAmountMoney, err = discountAmountMoney.Add(couponDiscountMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate checkout discount total: %w", err)
	}
	discountAmountMoney, err = discountAmountMoney.Add(giftCardDiscountMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate checkout discount total: %w", err)
	}
	totalMoney, err := grossMoney.Subtract(discountAmountMoney)
	if err != nil {
		return nil, fmt.Errorf("calculate checkout total: %w", err)
	}
	if totalMoney.AmountMinor() < 0 {
		totalMoney, err = domainmoney.New(0, quoteCurrency)
		if err != nil {
			return nil, fmt.Errorf("clamp checkout total: %w", err)
		}
	}
	giftCardDiscount, err := giftCardDiscountMoney.MajorFloat()
	if err != nil {
		return nil, fmt.Errorf("format gift card discount: %w", err)
	}
	discountAmount, err := discountAmountMoney.MajorFloat()
	if err != nil {
		return nil, fmt.Errorf("format checkout discount total: %w", err)
	}
	totalAmount, err := totalMoney.MajorFloat()
	if err != nil {
		return nil, fmt.Errorf("format checkout total: %w", err)
	}

	paymentCurrency := quoteCurrency
	paymentAmount := totalAmount
	paymentAmountMinor := totalMoney.AmountMinor()
	if provider := paymentpkg.ProviderForPaymentMethod(input.PaymentMethod); provider == string(paymentpkg.GatewayAlipay) || provider == string(paymentpkg.GatewayWechat) {
		paymentCurrency = "CNY"
		if quoteCurrency != paymentCurrency {
			if repos.exchangeRates == nil {
				return nil, fmt.Errorf("exchange rate repository is required to convert %s to %s payment settlement", quoteCurrency, paymentCurrency)
			}
			conversionService := NewExchangeRateService(repos.exchangeRates, nil)
			conversionService.ConfigureCurrencyPolicy(repos.currencyPolicy)
			convertedMoney, conversionErr := conversionService.ConvertMoneyStrict(totalMoney, paymentCurrency)
			if conversionErr != nil {
				return nil, fmt.Errorf("exchange rate unavailable for %s to %s payment settlement: %w", quoteCurrency, paymentCurrency, conversionErr)
			}
			paymentAmount, err = convertedMoney.MajorFloat()
			if err != nil {
				return nil, fmt.Errorf("format payment settlement amount: %w", err)
			}
			paymentAmountMinor = convertedMoney.AmountMinor()
		}
	}

	return &CheckoutQuote{
		Items:                 items,
		SubtotalMinor:         subtotalMoney.AmountMinor(),
		ShippingFeeMinor:      shippingFeeMoney.AmountMinor(),
		TaxMinor:              taxMoney.AmountMinor(),
		MemberDiscountMinor:   memberDiscountMoney.AmountMinor(),
		PointsDiscountMinor:   pointsDiscountMoney.AmountMinor(),
		CouponDiscountMinor:   couponDiscountMoney.AmountMinor(),
		GiftCardDiscountMinor: giftCardDiscountMoney.AmountMinor(),
		DiscountMinor:         discountAmountMoney.AmountMinor(),
		TotalMinor:            totalMoney.AmountMinor(),
		SubtotalAmount:        subtotal,
		ShippingFee:           shippingFee,
		ShippingQuote:         shippingQuote,
		TaxAmount:             taxAmount,
		MemberDiscount:        memberDiscount,
		PointsDiscount:        pointsDiscount,
		CouponDiscount:        couponDiscount,
		GiftCardDiscount:      giftCardDiscount,
		DiscountAmount:        discountAmount,
		TotalAmount:           totalAmount,
		CouponCode:            couponCode,
		GiftCardCode:          giftCardCode,
		PointsToUse:           pointsToUse,
		ProgramConfigID:       programConfigID,
		Coupon:                targetCoupon,
		Currency:              quoteCurrency,
		PaymentCurrency:       paymentCurrency,
		PaymentAmount:         paymentAmount,
		PaymentAmountMinor:    paymentAmountMinor,
		FXSnapshot:            fxSnapshot,
		GiftCard:              giftCard,
		GiftCardDiscountCents: giftCardDiscountCents,
		PricingSnapshot:       pricingSnapshot,
	}, nil
}

func (s *CheckoutService) resolveOrderFXSnapshot(
	orderCurrency string,
	currencyPolicy *CurrencyPolicyService,
	exchangeRates *repository.ExchangeRateRepository,
) (currency.OrderFXSnapshot, error) {
	orderCurrency = currency.NormalizeCode(orderCurrency)
	if !currency.IsCatalogCode(orderCurrency) {
		return currency.OrderFXSnapshot{}, fmt.Errorf("unsupported order currency %s", orderCurrency)
	}

	baseCurrency := currency.DefaultPrimaryCurrency
	policyConfigured := currencyPolicy != nil
	if policyConfigured {
		primary, err := currencyPolicy.PrimaryCurrency()
		if err != nil {
			return currency.OrderFXSnapshot{}, fmt.Errorf("resolve primary currency for order FX snapshot: %w", err)
		}
		baseCurrency = currency.NormalizeCode(primary)
	}

	capturedAt := time.Now().UTC()
	if baseCurrency == orderCurrency {
		return currency.OrderFXSnapshot{
			Version:         currency.OrderFXSnapshotVersion,
			BaseCurrency:    baseCurrency,
			OrderCurrency:   orderCurrency,
			BaseToOrderRate: 1,
			Source:          "same_currency",
			CapturedAt:      capturedAt,
		}, nil
	}

	if exchangeRates != nil {
		now := time.Now().UTC()
		if record, err := exchangeRates.FindFresh(baseCurrency, orderCurrency, now); err == nil && record.Rate > 0 {
			return currency.OrderFXSnapshot{
				Version:         currency.OrderFXSnapshotVersion,
				BaseCurrency:    baseCurrency,
				OrderCurrency:   orderCurrency,
				BaseToOrderRate: record.Rate,
				Source:          nonEmptyFXSource(record.Source, "cached_exchange_rate"),
				CapturedAt:      capturedAt,
				RateFetchedAt:   snapshotTimePtr(record.FetchedAt),
			}, nil
		}
		if record, err := exchangeRates.FindFresh(orderCurrency, baseCurrency, now); err == nil && record.Rate > 0 {
			inverseRate, inverseErr := invertExchangeRate(record.Rate)
			if inverseErr != nil {
				return currency.OrderFXSnapshot{}, inverseErr
			}
			return currency.OrderFXSnapshot{
				Version:         currency.OrderFXSnapshotVersion,
				BaseCurrency:    baseCurrency,
				OrderCurrency:   orderCurrency,
				BaseToOrderRate: inverseRate,
				Source:          nonEmptyFXSource(record.Source, "cached_exchange_rate_reverse"),
				CapturedAt:      capturedAt,
				RateFetchedAt:   snapshotTimePtr(record.FetchedAt),
			}, nil
		}
	}

	if !policyConfigured {
		return currency.OrderFXSnapshot{}, fmt.Errorf("historical FX snapshot is unavailable for %s to %s order", baseCurrency, orderCurrency)
	}
	return currency.OrderFXSnapshot{}, fmt.Errorf(
		"historical FX snapshot is unavailable for %s to %s order",
		baseCurrency,
		orderCurrency,
	)
}

func (s *CheckoutService) resolvePointsFXSnapshot(
	orderCurrency string,
	pointsCurrency string,
	exchangeRates *repository.ExchangeRateRepository,
) (currency.OrderFXSnapshot, error) {
	orderCurrency = currency.NormalizeCode(orderCurrency)
	pointsCurrency = currency.NormalizeCode(pointsCurrency)
	if !currency.IsCatalogCode(orderCurrency) {
		return currency.OrderFXSnapshot{}, fmt.Errorf("unsupported order currency %s", orderCurrency)
	}
	if !currency.IsCatalogCode(pointsCurrency) {
		return currency.OrderFXSnapshot{}, fmt.Errorf("unsupported loyalty points currency %s", pointsCurrency)
	}

	capturedAt := time.Now().UTC()
	if pointsCurrency == orderCurrency {
		return currency.OrderFXSnapshot{
			Version:         currency.OrderFXSnapshotVersion,
			BaseCurrency:    pointsCurrency,
			OrderCurrency:   orderCurrency,
			BaseToOrderRate: 1,
			Source:          "same_currency",
			CapturedAt:      capturedAt,
		}, nil
	}

	if exchangeRates == nil {
		return currency.OrderFXSnapshot{}, fmt.Errorf(
			"historical FX snapshot is unavailable for %s to %s points redemption",
			pointsCurrency,
			orderCurrency,
		)
	}

	now := time.Now().UTC()
	if record, err := exchangeRates.FindFresh(pointsCurrency, orderCurrency, now); err == nil && record.Rate > 0 {
		return currency.OrderFXSnapshot{
			Version:         currency.OrderFXSnapshotVersion,
			BaseCurrency:    pointsCurrency,
			OrderCurrency:   orderCurrency,
			BaseToOrderRate: record.Rate,
			Source:          nonEmptyFXSource(record.Source, "cached_exchange_rate"),
			CapturedAt:      capturedAt,
			RateFetchedAt:   snapshotTimePtr(record.FetchedAt),
		}, nil
	}
	if record, err := exchangeRates.FindFresh(orderCurrency, pointsCurrency, now); err == nil && record.Rate > 0 {
		inverseRate, inverseErr := invertExchangeRate(record.Rate)
		if inverseErr != nil {
			return currency.OrderFXSnapshot{}, inverseErr
		}
		return currency.OrderFXSnapshot{
			Version:         currency.OrderFXSnapshotVersion,
			BaseCurrency:    pointsCurrency,
			OrderCurrency:   orderCurrency,
			BaseToOrderRate: inverseRate,
			Source:          nonEmptyFXSource(record.Source, "cached_exchange_rate_reverse"),
			CapturedAt:      capturedAt,
			RateFetchedAt:   snapshotTimePtr(record.FetchedAt),
		}, nil
	}

	return currency.OrderFXSnapshot{}, fmt.Errorf(
		"historical FX snapshot is unavailable for %s to %s points redemption",
		pointsCurrency,
		orderCurrency,
	)
}

func nonEmptyFXSource(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

func snapshotTimePtr(value time.Time) *time.Time {
	value = value.UTC()
	return &value
}

func invertExchangeRate(rate float64) (float64, error) {
	rateRat, err := exchangeRateRat(rate)
	if err != nil {
		return 0, err
	}
	inverse := new(big.Rat).Inv(rateRat)
	value, _ := inverse.Float64()
	if value <= 0 || math.IsInf(value, 0) || math.IsNaN(value) {
		return 0, errors.New("inverse exchange rate overflows float64")
	}
	return value, nil
}

// calculateMemberDiscountMoney computes the member discount directly in the
// order currency's minor units. DiscountRate remains a persistence field for
// now, but is parsed into an exact rational at this boundary so no floating
// point arithmetic participates in pricing.
func (s *CheckoutService) calculateMemberDiscountMoney(
	loyaltyRepo *repository.LoyaltyRepository,
	userID uint,
	subtotal domainmoney.Money,
) (domainmoney.Money, error) {
	zero, err := domainmoney.New(0, subtotal.Currency().String())
	if err != nil {
		return domainmoney.Money{}, err
	}
	if loyaltyRepo == nil {
		return zero, nil
	}

	userLoyalty, err := loyaltyRepo.FindUserLoyaltyByUserID(userID)
	if err != nil || userLoyalty == nil {
		return zero, nil
	}
	level, err := loyaltyRepo.FindMemberLevelByPoints(userLoyalty.TotalPoints)
	if err != nil || level == nil || level.DiscountRate <= 0 {
		return zero, nil
	}
	if level.DiscountRate >= 100 {
		return subtotal, nil
	}

	rate, ok := new(big.Rat).SetString(strconv.FormatFloat(level.DiscountRate, 'f', -1, 64))
	if !ok || rate.Sign() < 0 {
		return domainmoney.Money{}, fmt.Errorf("invalid member discount rate %v", level.DiscountRate)
	}
	rate.Quo(rate, big.NewRat(100, 1))
	return subtotal.MultiplyRat(rate)
}

// calculatePointsDiscountMoney keeps the checkout path in minor units.
func (s *CheckoutService) calculatePointsDiscountMoney(
	loyaltyRepo *repository.LoyaltyRepository,
	userID uint,
	requestedPoints int,
	subtotal domainmoney.Money,
	config *loyalty.ProgramConfig,
	fxSnapshot currency.OrderFXSnapshot,
) (int, domainmoney.Money, *uint, error) {
	zero, err := domainmoney.New(0, subtotal.Currency().String())
	if err != nil {
		return 0, domainmoney.Money{}, nil, err
	}
	if requestedPoints <= 0 {
		return 0, zero, nil, nil
	}
	if config == nil {
		config, err = s.currentLoyaltyProgramConfig()
		if err != nil {
			return 0, domainmoney.Money{}, nil, err
		}
		if config == nil {
			return 0, domainmoney.Money{}, nil, ErrLoyaltyProgramConfigNotFound
		}
	}
	if !config.Enabled {
		return 0, domainmoney.Money{}, nil, errors.New("point redemption is disabled")
	}
	pointsCurrency := currency.NormalizeCode(config.Currency)
	if !currency.IsCatalogCode(pointsCurrency) {
		return 0, domainmoney.Money{}, nil, errors.New("point redemption currency is invalid")
	}
	if config.ExchangeRatePoints <= 0 {
		return 0, domainmoney.Money{}, nil, errors.New("point redemption exchange rate is invalid")
	}
	if currency.NormalizeCode(fxSnapshot.BaseCurrency) != pointsCurrency {
		return 0, domainmoney.Money{}, nil, fmt.Errorf("point redemption FX base currency must be %s", pointsCurrency)
	}
	if err := fxSnapshot.Validate(subtotal.Currency().String()); err != nil {
		return 0, domainmoney.Money{}, nil, fmt.Errorf("point redemption FX snapshot is invalid: %w", err)
	}
	userLoyalty, err := loyaltyRepo.FindUserLoyaltyByUserID(userID)
	if err != nil || userLoyalty == nil {
		return 0, domainmoney.Money{}, nil, fmt.Errorf("[CRITICAL] Insufficient points: available %d, requested %d", 0, requestedPoints)
	}
	if userLoyalty.AvailablePoints < requestedPoints {
		return 0, domainmoney.Money{}, nil, fmt.Errorf("[CRITICAL] Insufficient points: available %d, requested %d", userLoyalty.AvailablePoints, requestedPoints)
	}
	rate, ok := new(big.Rat).SetString(strconv.FormatFloat(fxSnapshot.BaseToOrderRate, 'f', -1, 64))
	if !ok || rate.Sign() <= 0 {
		return 0, domainmoney.Money{}, nil, errors.New("point redemption FX rate is invalid")
	}
	calculate := func(points int64) (domainmoney.Money, error) {
		value := new(big.Rat).SetFrac(big.NewInt(points), big.NewInt(int64(config.ExchangeRatePoints)))
		value.Mul(value, rate)
		return domainmoney.FromMajorRat(value, subtotal.Currency().String())
	}
	pointsToUse := requestedPoints
	pointsDiscount, err := calculate(int64(pointsToUse))
	if err != nil {
		return 0, domainmoney.Money{}, nil, fmt.Errorf("calculate points discount: %w", err)
	}
	maxDiscount, err := subtotal.MultiplyRatio(1, 2)
	if err != nil {
		return 0, domainmoney.Money{}, nil, fmt.Errorf("calculate points discount cap: %w", err)
	}
	if pointsDiscount.AmountMinor() > maxDiscount.AmountMinor() {
		// Find the largest whole-point redemption whose rounded minor-unit value
		// remains within the 50% merchandise cap.
		low, high := int64(0), int64(requestedPoints)
		for low < high {
			mid := low + (high-low+1)/2
			candidate, candidateErr := calculate(mid)
			if candidateErr != nil {
				return 0, domainmoney.Money{}, nil, fmt.Errorf("calculate capped points discount: %w", candidateErr)
			}
			if candidate.AmountMinor() <= maxDiscount.AmountMinor() {
				low = mid
			} else {
				high = mid - 1
			}
		}
		pointsToUse = int(low)
		pointsDiscount, err = calculate(low)
		if err != nil {
			return 0, domainmoney.Money{}, nil, fmt.Errorf("calculate capped points discount: %w", err)
		}
	}
	return pointsToUse, pointsDiscount, programConfigID(config), nil
}

func (s *CheckoutService) currentLoyaltyProgramConfig() (*loyalty.ProgramConfig, error) {
	if s.loyaltyProgram != nil {
		return s.loyaltyProgram.GetActive()
	}
	return nil, ErrLoyaltyProgramConfigNotFound
}

func (s *CheckoutService) validateCouponWithFXMoney(
	couponRepo *repository.CouponRepository,
	code string,
	userID uint,
	email string,
	amountMoney domainmoney.Money,
	items []checkoutPricingLineInput,
	fxSnapshot currency.OrderFXSnapshot,
	lockForUpdate bool,
) (*coupon.Coupon, domainmoney.Money, error) {
	zero, err := domainmoney.New(0, amountMoney.Currency().String())
	if err != nil {
		return nil, domainmoney.Money{}, err
	}
	var c *coupon.Coupon
	if lockForUpdate {
		c, err = couponRepo.FindCouponByCodeForUpdate(code)
	} else {
		c, err = couponRepo.FindCouponByCode(code)
	}
	if err != nil {
		return nil, domainmoney.Money{}, errors.New("coupon not found")
	}
	if !c.Enabled {
		return nil, domainmoney.Money{}, errors.New("coupon is disabled")
	}
	if err := validateCouponRecipient(c, userID); err != nil {
		return nil, domainmoney.Money{}, err
	}

	now := time.Now()
	if now.Before(c.StartDate) || now.After(c.EndDate) {
		return nil, domainmoney.Money{}, errors.New("coupon is expired")
	}
	if c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit {
		return nil, domainmoney.Money{}, errors.New("coupon usage limit reached")
	}
	if err := validateCouponPerUserUsageLimit(couponRepo, c, userID, email); err != nil {
		return nil, domainmoney.Money{}, err
	}
	applicableProducts, err := parseCouponProductIDs(c.ApplicableProducts)
	if err != nil {
		return nil, domainmoney.Money{}, fmt.Errorf("invalid applicable_products: %w", err)
	}
	excludedProducts, err := parseCouponProductIDs(c.ExcludedProducts)
	if err != nil {
		return nil, domainmoney.Money{}, fmt.Errorf("invalid excluded_products: %w", err)
	}
	applicableCategories, err := parseCouponCategoryValues(c.ApplicableCategories)
	if err != nil {
		return nil, domainmoney.Money{}, fmt.Errorf("invalid applicable_categories: %w", err)
	}
	if len(items) == 0 && (len(applicableProducts) > 0 || len(excludedProducts) > 0 || len(applicableCategories.IDs) > 0 || len(applicableCategories.Slugs) > 0) {
		return nil, domainmoney.Money{}, errors.New("cart items are required to validate coupon product restrictions")
	}

	pricingCurrency := amountMoney.Currency().String()
	qualifiedAmountMoney := amountMoney
	if len(items) > 0 {
		qualifiedSubtotalMoney, totalSubtotalMoney, subtotalErr := couponQualifiedSubtotalMoneyWithCategories(items, applicableProducts, excludedProducts, applicableCategories, pricingCurrency)
		if subtotalErr != nil {
			return nil, domainmoney.Money{}, subtotalErr
		}
		if qualifiedSubtotalMoney.AmountMinor() <= 0 {
			return nil, domainmoney.Money{}, errors.New("coupon does not apply to cart items")
		}
		// `amount` is merchandise after earlier discounts (for example member
		// discount). Allocate those discounts pro-rata so excluded products do
		// not indirectly increase the coupon's taxable/discountable base.
		qualifiedAmountMoney = qualifiedSubtotalMoney
		if totalSubtotalMoney.AmountMinor() > 0 && amountMoney.AmountMinor() < totalSubtotalMoney.AmountMinor() {
			qualifiedAmountMoney, err = amountMoney.MultiplyRatio(
				qualifiedSubtotalMoney.AmountMinor(),
				totalSubtotalMoney.AmountMinor(),
			)
			if err != nil {
				return nil, domainmoney.Money{}, fmt.Errorf("allocate coupon merchandise amount: %w", err)
			}
		}
		if qualifiedAmountMoney.AmountMinor() > amountMoney.AmountMinor() {
			qualifiedAmountMoney = amountMoney
		}
	}
	conversionRate, err := couponConversionRateRat(c, fxSnapshot, pricingCurrency)
	if err != nil {
		return nil, domainmoney.Money{}, err
	}
	couponCurrency := couponCurrencyForSnapshot(c, fxSnapshot)
	minimumMoney, err := domainmoney.FromMajorFloat(c.MinAmount, couponCurrency)
	if err != nil {
		return nil, domainmoney.Money{}, fmt.Errorf("invalid coupon minimum amount: %w", err)
	}
	if couponCurrency != pricingCurrency {
		minimumMoney, err = minimumMoney.ConvertAtRat(conversionRate, pricingCurrency)
		if err != nil {
			return nil, domainmoney.Money{}, fmt.Errorf("convert coupon minimum amount: %w", err)
		}
	}
	if qualifiedAmountMoney.AmountMinor() < minimumMoney.AmountMinor() {
		minimumMajor, _ := minimumMoney.FormatMajor()
		return nil, domainmoney.Money{}, fmt.Errorf("minimum amount %s required", minimumMajor)
	}
	discount := zero
	if c.Type == "percentage" {
		rate, ok := new(big.Rat).SetString(strconv.FormatFloat(c.Value, 'f', -1, 64))
		if !ok || rate.Sign() < 0 {
			return nil, domainmoney.Money{}, errors.New("invalid coupon percentage")
		}
		rate.Quo(rate, big.NewRat(100, 1))
		discount, err = qualifiedAmountMoney.MultiplyRat(rate)
		if err != nil {
			return nil, domainmoney.Money{}, fmt.Errorf("calculate coupon percentage discount: %w", err)
		}
		if c.MaxDiscount > 0 {
			maxMoney, maxErr := domainmoney.FromMajorFloat(c.MaxDiscount, couponCurrency)
			if maxErr != nil {
				return nil, domainmoney.Money{}, fmt.Errorf("invalid coupon max discount: %w", maxErr)
			}
			if couponCurrency != pricingCurrency {
				maxMoney, maxErr = maxMoney.ConvertAtRat(conversionRate, pricingCurrency)
				if maxErr != nil {
					return nil, domainmoney.Money{}, fmt.Errorf("convert coupon max discount: %w", maxErr)
				}
			}
			if discount.AmountMinor() > maxMoney.AmountMinor() {
				discount = maxMoney
			}
		}
	} else {
		discount, err = domainmoney.FromMajorFloat(c.Value, couponCurrency)
		if err != nil {
			return nil, domainmoney.Money{}, fmt.Errorf("invalid coupon value: %w", err)
		}
		if couponCurrency != pricingCurrency {
			discount, err = discount.ConvertAtRat(conversionRate, pricingCurrency)
			if err != nil {
				return nil, domainmoney.Money{}, fmt.Errorf("convert coupon value: %w", err)
			}
		}
		if c.MaxDiscount > 0 {
			maxMoney, maxErr := domainmoney.FromMajorFloat(c.MaxDiscount, couponCurrency)
			if maxErr != nil {
				return nil, domainmoney.Money{}, fmt.Errorf("invalid coupon max discount: %w", maxErr)
			}
			if couponCurrency != pricingCurrency {
				maxMoney, maxErr = maxMoney.ConvertAtRat(conversionRate, pricingCurrency)
				if maxErr != nil {
					return nil, domainmoney.Money{}, fmt.Errorf("convert coupon max discount: %w", maxErr)
				}
			}
			if discount.AmountMinor() > maxMoney.AmountMinor() {
				discount = maxMoney
			}
		}
	}
	if discount.AmountMinor() > qualifiedAmountMoney.AmountMinor() {
		discount = qualifiedAmountMoney
	}
	return c, discount, nil
}

func parseCouponProductIDs(raw string) (map[uint]struct{}, error) {
	return parseCouponIDs(raw, "product IDs")
}

func parseCouponIDs(raw, label string) (map[uint]struct{}, error) {
	result := make(map[uint]struct{})
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result, nil
	}
	if !strings.HasPrefix(raw, "[") {
		return nil, fmt.Errorf("must be a JSON array of positive %s", label)
	}
	var ids []uint
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, fmt.Errorf("must be a JSON array of positive %s", label)
	}
	for _, id := range ids {
		if id == 0 {
			return nil, fmt.Errorf("%s must be positive", label)
		}
		result[id] = struct{}{}
	}
	return result, nil
}

type couponCategoryValues struct {
	IDs   map[uint]struct{}
	Slugs map[string]struct{}
}

func couponQualifiedSubtotalMoneyWithCategories(items []checkoutPricingLineInput, applicable, excluded map[uint]struct{}, applicableCategories couponCategoryValues, currencyCode string) (domainmoney.Money, domainmoney.Money, error) {
	zero, err := domainmoney.New(0, currencyCode)
	if err != nil {
		return domainmoney.Money{}, domainmoney.Money{}, err
	}
	qualified, total := zero, zero
	for _, line := range items {
		if err := validateCheckoutPricingMoney(line.UnitPrice, currencyCode, "coupon line unit price"); err != nil {
			return domainmoney.Money{}, domainmoney.Money{}, err
		}
		if err := validateCheckoutPricingMoney(line.Subtotal, currencyCode, "coupon line subtotal"); err != nil {
			return domainmoney.Money{}, domainmoney.Money{}, err
		}
		lineAmount := line.Subtotal
		if lineAmount.AmountMinor() == 0 {
			var err error
			lineAmount, err = line.UnitPrice.MultiplyInt(int64(line.Quantity))
			if err != nil {
				return domainmoney.Money{}, domainmoney.Money{}, fmt.Errorf("calculate coupon line amount: %w", err)
			}
		}
		total, err = total.Add(lineAmount)
		if err != nil {
			return domainmoney.Money{}, domainmoney.Money{}, err
		}
		if _, blocked := excluded[line.ProductID]; blocked {
			continue
		}
		if len(applicable) > 0 {
			if _, ok := applicable[line.ProductID]; !ok {
				continue
			}
		}
		if len(applicableCategories.IDs) > 0 || len(applicableCategories.Slugs) > 0 {
			if line.ProductCategoryID == nil {
				if _, ok := applicableCategories.Slugs[strings.ToLower(strings.TrimSpace(line.ProductCategorySlug))]; !ok {
					continue
				}
			} else {
				_, idOK := applicableCategories.IDs[*line.ProductCategoryID]
				_, slugOK := applicableCategories.Slugs[strings.ToLower(strings.TrimSpace(line.ProductCategorySlug))]
				if !idOK && !slugOK {
					continue
				}
			}
		}
		qualified, err = qualified.Add(lineAmount)
		if err != nil {
			return domainmoney.Money{}, domainmoney.Money{}, err
		}
	}
	return qualified, total, nil
}

func parseCouponCategoryValues(raw string) (couponCategoryValues, error) {
	result := couponCategoryValues{IDs: make(map[uint]struct{}), Slugs: make(map[string]struct{})}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result, nil
	}
	if !strings.HasPrefix(raw, "[") {
		return result, errors.New("must be a JSON array of category IDs or slugs")
	}
	var values []interface{}
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return result, errors.New("must be a JSON array of category IDs or slugs")
	}
	for _, value := range values {
		switch typed := value.(type) {
		case float64:
			if typed <= 0 || typed != math.Trunc(typed) {
				return result, errors.New("category IDs must be positive integers")
			}
			result.IDs[uint(typed)] = struct{}{}
		case string:
			slug := strings.ToLower(strings.TrimSpace(typed))
			if slug == "" {
				return result, errors.New("category slugs must not be empty")
			}
			result.Slugs[slug] = struct{}{}
		default:
			return result, errors.New("category values must be IDs or slugs")
		}
	}
	return result, nil
}

func couponCurrencyForSnapshot(c *coupon.Coupon, snapshot currency.OrderFXSnapshot) string {
	if c == nil {
		return ""
	}
	if code := currency.NormalizeCode(c.Currency); code != "" {
		return code
	}
	if code := currency.NormalizeCode(snapshot.BaseCurrency); code != "" {
		return code
	}
	return currency.DefaultPrimaryCurrency
}

func couponConversionRateRat(c *coupon.Coupon, snapshot currency.OrderFXSnapshot, pricingCurrency string) (*big.Rat, error) {
	if c == nil {
		return nil, errors.New("coupon is required")
	}
	couponCurrency := couponCurrencyForSnapshot(c, snapshot)
	if !currency.IsCatalogCode(couponCurrency) {
		return nil, fmt.Errorf("unsupported coupon currency %s", couponCurrency)
	}
	orderCurrency := currency.NormalizeCode(pricingCurrency)
	if orderCurrency == "" {
		orderCurrency = currency.NormalizeCode(snapshot.OrderCurrency)
	}
	if orderCurrency == "" || couponCurrency == orderCurrency {
		return big.NewRat(1, 1), nil
	}
	if err := snapshot.Validate(orderCurrency); err != nil {
		return nil, fmt.Errorf("coupon FX snapshot is invalid: %w", err)
	}
	base := currency.NormalizeCode(snapshot.BaseCurrency)
	if couponCurrency != base {
		return nil, fmt.Errorf("coupon currency %s is incompatible with order FX base %s", couponCurrency, base)
	}
	rate, ok := new(big.Rat).SetString(strconv.FormatFloat(snapshot.BaseToOrderRate, 'f', -1, 64))
	if !ok || rate.Sign() <= 0 {
		return nil, errors.New("coupon FX conversion rate is invalid")
	}
	return rate, nil
}

func (s *CheckoutService) validateGiftCard(
	couponRepo *repository.CouponRepository,
	code string,
	userID uint,
	orderCurrency string,
	lockForUpdate bool,
) (*coupon.GiftCard, error) {
	var card *coupon.GiftCard
	var err error
	if lockForUpdate {
		card, err = couponRepo.FindGiftCardByCodeForUpdate(code)
	} else {
		card, err = couponRepo.FindGiftCardByCode(code)
	}
	if err != nil {
		return nil, errors.New("gift card not found")
	}
	if !card.IsValid() {
		return nil, errors.New("gift card is not active or has no available balance")
	}
	if card.OwnerUserID != nil && *card.OwnerUserID != userID {
		return nil, errors.New("gift card does not belong to this customer")
	}
	if currency.NormalizeCode(card.Currency) != currency.NormalizeCode(orderCurrency) {
		return nil, errors.New("gift card currency does not match the order currency")
	}
	return card, nil
}

func minInt64(left, right int64) int64 {
	if left < right {
		return left
	}
	return right
}

func (s *CheckoutService) calculateTaxMoney(
	paymentRepo *repository.PaymentRepository,
	amount domainmoney.Money,
	country, state, postalCode, currencyCode string,
) (domainmoney.Money, error) {
	zero, err := domainmoney.New(0, currencyCode)
	if err != nil {
		return domainmoney.Money{}, fmt.Errorf("initialize tax amount: %w", err)
	}
	if amount.Currency().String() != zero.Currency().String() {
		return domainmoney.Money{}, fmt.Errorf("tax amount currency mismatch: %s and %s", amount.Currency(), zero.Currency())
	}
	if paymentRepo == nil {
		return domainmoney.Money{}, errors.New("payment repository is not configured")
	}
	taxRate, err := paymentRepo.FindTaxRateByLocation(country, state, postalCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return zero, nil
		}
		return domainmoney.Money{}, fmt.Errorf("failed to load tax rate for %s/%s/%s: %w", country, state, postalCode, err)
	}
	if taxRate == nil {
		return domainmoney.Money{}, errors.New("tax rate lookup returned no result")
	}
	rate, ok := new(big.Rat).SetString(strconv.FormatFloat(taxRate.Rate, 'f', -1, 64))
	if !ok {
		return domainmoney.Money{}, errors.New("tax rate is invalid")
	}
	rate.Quo(rate, big.NewRat(100, 1))
	scale := int64(1)
	if minorUnits, exists := currency.MinorUnits(currencyCode); exists {
		for i := 0; i < minorUnits; i++ {
			scale *= 10
		}
	}
	baseMajor := new(big.Rat).SetFrac(big.NewInt(amount.AmountMinor()), big.NewInt(scale))
	baseMajor.Mul(baseMajor, rate)
	return domainmoney.FromMajorRat(baseMajor, currencyCode)
}
