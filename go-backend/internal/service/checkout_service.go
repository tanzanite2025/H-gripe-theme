package service

import (
	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/order"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type CheckoutService struct {
	productRepo     *repository.ProductRepository
	couponRepo      *repository.CouponRepository
	paymentRepo     *repository.PaymentRepository
	loyaltyRepo     *repository.LoyaltyRepository
	shippingService *ShippingService
	loyaltyProgram  *LoyaltyProgramService
	currencyPolicy  *CurrencyPolicyService
	exchangeRates   *repository.ExchangeRateRepository
}

type CheckoutQuoteInput struct {
	UserID               uint
	Items                []order.OrderItem
	ShippingAddress      order.Address
	DisplayCurrency      string
	CouponCode           string
	PointsToUse          int
	LoyaltyProgramConfig *loyalty.ProgramConfig
}

type CheckoutQuote struct {
	Items           []order.OrderItem        `json:"items"`
	SubtotalAmount  float64                  `json:"subtotal_amount"`
	ShippingFee     float64                  `json:"shipping_fee"`
	ShippingQuote   *ShippingQuote           `json:"shipping_quote,omitempty"`
	TaxAmount       float64                  `json:"tax_amount"`
	MemberDiscount  float64                  `json:"member_discount"`
	PointsDiscount  float64                  `json:"points_discount"`
	CouponDiscount  float64                  `json:"coupon_discount"`
	DiscountAmount  float64                  `json:"discount_amount"`
	TotalAmount     float64                  `json:"total_amount"`
	CouponCode      string                   `json:"coupon_code"`
	PointsToUse     int                      `json:"points_to_use"`
	ProgramConfigID *uint                    `json:"loyalty_program_config_id,omitempty"`
	Coupon          *coupon.Coupon           `json:"coupon,omitempty"`
	Currency        string                   `json:"currency"`
	FXSnapshot      currency.OrderFXSnapshot `json:"-"`
}

type checkoutRepositories struct {
	productRepo     *repository.ProductRepository
	couponRepo      *repository.CouponRepository
	paymentRepo     *repository.PaymentRepository
	loyaltyRepo     *repository.LoyaltyRepository
	shippingService *ShippingService
	currencyPolicy  *CurrencyPolicyService
	exchangeRates   *repository.ExchangeRateRepository
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

func (s *CheckoutService) Quote(input CheckoutQuoteInput) (*CheckoutQuote, error) {
	return s.quote(input, checkoutRepositories{
		productRepo:     s.productRepo,
		couponRepo:      s.couponRepo,
		paymentRepo:     s.paymentRepo,
		loyaltyRepo:     s.loyaltyRepo,
		shippingService: s.shippingService,
		currencyPolicy:  s.currencyPolicy,
		exchangeRates:   s.exchangeRates,
	})
}

func (s *CheckoutService) QuoteWithRepositories(input CheckoutQuoteInput, repos repository.TxRepositories) (*CheckoutQuote, error) {
	if repos.Setting == nil {
		return nil, errors.New("transactional checkout currency policy repository is not configured")
	}
	shippingService := s.shippingService
	if repos.Shipping != nil {
		shippingService = NewShippingService(repos.Shipping, repos.Product)
	}
	return s.quote(input, checkoutRepositories{
		productRepo:     repos.Product,
		couponRepo:      repos.Coupon,
		paymentRepo:     repos.Payment,
		loyaltyRepo:     repos.Loyalty,
		shippingService: shippingService,
		currencyPolicy:  NewCurrencyPolicyService(repos.Setting),
		exchangeRates:   repos.ExchangeRate,
	})
}

func (s *CheckoutService) quote(input CheckoutQuoteInput, repos checkoutRepositories) (*CheckoutQuote, error) {
	if len(input.Items) == 0 {
		return nil, errors.New("cart is empty")
	}
	items := make([]order.OrderItem, len(input.Items))
	shippingItems := make([]ShippingQuoteItemInput, 0, len(input.Items))
	var subtotal float64
	quoteCurrency := ""
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

		resolvedVariantID := variant.ID
		variantID := &resolvedVariantID
		price := variant.EffectivePrice()
		itemCurrency := currency.NormalizeCode(variant.Currency)
		if itemCurrency == "" {
			itemCurrency = productdomain.DefaultPriceCurrency
		}
		if !currency.IsValidCode(itemCurrency) || !currency.IsCatalogCode(itemCurrency) {
			return nil, fmt.Errorf("unsupported price currency for SKU %s", variant.SKU)
		}
		if quoteCurrency == "" {
			quoteCurrency = itemCurrency
		} else if quoteCurrency != itemCurrency {
			return nil, errors.New("cart contains multiple price currencies")
		}
		sku := variant.SKU
		attributes := variant.OptionValues
		availableStock := variant.Stock
		if availableStock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product ID %d", item.ProductID)
		}

		items[i] = item
		items[i].VariantID = variantID
		items[i].Price = price
		items[i].Subtotal = price * float64(item.Quantity)
		items[i].ProductName = product.Name
		items[i].SKU = sku
		items[i].Attributes = attributes
		items[i].Total = items[i].Subtotal
		subtotal += items[i].Subtotal

		if variant.Weight <= 0 {
			return nil, fmt.Errorf("shipping weight is missing for SKU %s", variant.SKU)
		}
		shippingTemplateID, err := resolveProductShippingTemplateID(product, variant)
		if err != nil {
			return nil, err
		}
		shippingItems = append(shippingItems, ShippingQuoteItemInput{
			ProductID:          product.ID,
			VariantID:          variantID,
			ProductTypeID:      product.ProductTypeID,
			ShippingTemplateID: uintPtr(shippingTemplateID),
			Quantity:           item.Quantity,
			UnitPrice:          price,
			WeightGrams:        variant.Weight,
		})
	}
	if quoteCurrency == "" {
		return nil, errors.New("product price currency is required")
	}

	if repos.shippingService == nil {
		return nil, errors.New("shipping quote service is not configured")
	}
	shippingQuote, err := repos.shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country:         input.ShippingAddress.Country,
		Amount:          subtotal,
		Currency:        quoteCurrency,
		DisplayCurrency: input.DisplayCurrency,
		Items:           shippingItems,
	})
	if err != nil {
		return nil, err
	}
	shippingFee := shippingQuote.ShippingFee
	taxAmount := s.calculateTax(repos.paymentRepo, subtotal, input.ShippingAddress.Country, input.ShippingAddress.State)
	memberDiscount := s.calculateMemberDiscount(repos.loyaltyRepo, input.UserID, subtotal)
	pointsToUse, pointsDiscount, programConfigID, err := s.calculatePointsDiscount(
		repos.loyaltyRepo,
		input.UserID,
		input.PointsToUse,
		subtotal,
		input.LoyaltyProgramConfig,
	)
	if err != nil {
		return nil, err
	}

	var targetCoupon *coupon.Coupon
	couponDiscount := 0.0
	if input.CouponCode != "" {
		targetCoupon, couponDiscount, err = s.validateCoupon(repos.couponRepo, input.CouponCode, subtotal)
		if err != nil {
			return nil, fmt.Errorf("failed to apply coupon %s: %w", input.CouponCode, err)
		}
	}

	discountAmount := memberDiscount + pointsDiscount + couponDiscount
	totalAmount := subtotal + shippingFee + taxAmount - discountAmount
	if totalAmount < 0 {
		totalAmount = 0
	}

	fxSnapshot, err := s.resolveOrderFXSnapshot(quoteCurrency, repos.currencyPolicy, repos.exchangeRates)
	if err != nil {
		return nil, err
	}

	return &CheckoutQuote{
		Items:           items,
		SubtotalAmount:  subtotal,
		ShippingFee:     shippingFee,
		ShippingQuote:   shippingQuote,
		TaxAmount:       taxAmount,
		MemberDiscount:  memberDiscount,
		PointsDiscount:  pointsDiscount,
		CouponDiscount:  couponDiscount,
		DiscountAmount:  discountAmount,
		TotalAmount:     totalAmount,
		CouponCode:      input.CouponCode,
		PointsToUse:     pointsToUse,
		ProgramConfigID: programConfigID,
		Coupon:          targetCoupon,
		Currency:        quoteCurrency,
		FXSnapshot:      fxSnapshot,
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
		if record, err := exchangeRates.Find(baseCurrency, orderCurrency); err == nil && record.Rate > 0 {
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
		if record, err := exchangeRates.Find(orderCurrency, baseCurrency); err == nil && record.Rate > 0 {
			return currency.OrderFXSnapshot{
				Version:         currency.OrderFXSnapshotVersion,
				BaseCurrency:    baseCurrency,
				OrderCurrency:   orderCurrency,
				BaseToOrderRate: 1 / record.Rate,
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

func (s *CheckoutService) calculateMemberDiscount(loyaltyRepo *repository.LoyaltyRepository, userID uint, subtotal float64) float64 {
	userLoyalty, err := loyaltyRepo.FindUserLoyaltyByUserID(userID)
	if err != nil || userLoyalty == nil {
		return 0
	}

	level, err := loyaltyRepo.FindMemberLevelByPoints(userLoyalty.TotalPoints)
	if err != nil || level == nil || level.DiscountRate <= 0 {
		return 0
	}

	return subtotal * (level.DiscountRate / 100)
}

func (s *CheckoutService) calculatePointsDiscount(
	loyaltyRepo *repository.LoyaltyRepository,
	userID uint,
	requestedPoints int,
	subtotal float64,
	config *loyalty.ProgramConfig,
) (int, float64, *uint, error) {
	if requestedPoints <= 0 {
		return 0, 0, nil, nil
	}

	if config == nil {
		var err error
		config, err = s.currentLoyaltyProgramConfig()
		if err != nil {
			return 0, 0, nil, err
		}
	}
	if !config.Enabled {
		return 0, 0, nil, errors.New("point redemption is disabled")
	}

	userLoyalty, err := loyaltyRepo.FindUserLoyaltyByUserID(userID)
	if err != nil || userLoyalty == nil {
		return 0, 0, nil, fmt.Errorf("[CRITICAL] Insufficient points: available %d, requested %d", 0, requestedPoints)
	}
	if userLoyalty.AvailablePoints < requestedPoints {
		return 0, 0, nil, fmt.Errorf("[CRITICAL] Insufficient points: available %d, requested %d", userLoyalty.AvailablePoints, requestedPoints)
	}

	pointsDiscount := float64(requestedPoints) / float64(config.ExchangeRatePoints)
	pointsToUse := requestedPoints
	if maxPointsDiscount := subtotal * 0.5; pointsDiscount > maxPointsDiscount {
		pointsToUse = int(math.Floor(maxPointsDiscount * float64(config.ExchangeRatePoints)))
		pointsDiscount = float64(pointsToUse) / float64(config.ExchangeRatePoints)
	}

	return pointsToUse, pointsDiscount, programConfigID(config), nil
}

func (s *CheckoutService) currentLoyaltyProgramConfig() (*loyalty.ProgramConfig, error) {
	if s.loyaltyProgram != nil {
		return s.loyaltyProgram.GetActive()
	}
	return nil, ErrLoyaltyProgramConfigNotFound
}

func (s *CheckoutService) validateCoupon(couponRepo *repository.CouponRepository, code string, amount float64) (*coupon.Coupon, float64, error) {
	c, err := couponRepo.FindCouponByCode(code)
	if err != nil {
		return nil, 0, errors.New("coupon not found")
	}
	if !c.Enabled {
		return nil, 0, errors.New("coupon is disabled")
	}

	now := time.Now()
	if now.Before(c.StartDate) || now.After(c.EndDate) {
		return nil, 0, errors.New("coupon is expired")
	}
	if c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit {
		return nil, 0, errors.New("coupon usage limit reached")
	}
	if amount < c.MinAmount {
		return nil, 0, fmt.Errorf("minimum amount %.2f required", c.MinAmount)
	}

	return c, c.CalculateDiscount(amount), nil
}

func (s *CheckoutService) calculateTax(paymentRepo *repository.PaymentRepository, amount float64, country, state string) float64 {
	taxRate, err := paymentRepo.FindTaxRateByLocation(country, state)
	if err != nil {
		return 0
	}
	return amount * taxRate.Rate / 100
}
