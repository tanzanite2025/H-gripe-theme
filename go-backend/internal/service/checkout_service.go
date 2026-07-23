package service

import (
	"errors"
	"fmt"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/order"
	"tanzanite/internal/repository"
	"time"
)

type CheckoutService struct {
	productRepo     *repository.ProductRepository
	couponRepo      *repository.CouponRepository
	paymentRepo     *repository.PaymentRepository
	loyaltyRepo     *repository.LoyaltyRepository
	shippingService *ShippingService
}

type CheckoutQuoteInput struct {
	UserID          uint
	Items           []order.OrderItem
	ShippingAddress order.Address
	CouponCode      string
	PointsToUse     int
}

type CheckoutQuote struct {
	Items          []order.OrderItem `json:"items"`
	SubtotalAmount float64           `json:"subtotal_amount"`
	ShippingFee    float64           `json:"shipping_fee"`
	ShippingQuote  *ShippingQuote    `json:"shipping_quote,omitempty"`
	TaxAmount      float64           `json:"tax_amount"`
	MemberDiscount float64           `json:"member_discount"`
	PointsDiscount float64           `json:"points_discount"`
	CouponDiscount float64           `json:"coupon_discount"`
	DiscountAmount float64           `json:"discount_amount"`
	TotalAmount    float64           `json:"total_amount"`
	CouponCode     string            `json:"coupon_code"`
	PointsToUse    int               `json:"points_to_use"`
	Coupon         *coupon.Coupon    `json:"coupon,omitempty"`
}

type checkoutRepositories struct {
	productRepo     *repository.ProductRepository
	couponRepo      *repository.CouponRepository
	paymentRepo     *repository.PaymentRepository
	loyaltyRepo     *repository.LoyaltyRepository
	shippingService *ShippingService
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

func (s *CheckoutService) Quote(input CheckoutQuoteInput) (*CheckoutQuote, error) {
	return s.quote(input, checkoutRepositories{
		productRepo:     s.productRepo,
		couponRepo:      s.couponRepo,
		paymentRepo:     s.paymentRepo,
		loyaltyRepo:     s.loyaltyRepo,
		shippingService: s.shippingService,
	})
}

func (s *CheckoutService) QuoteWithRepositories(input CheckoutQuoteInput, repos repository.TxRepositories) (*CheckoutQuote, error) {
	shippingService := s.shippingService
	if repos.Shipping != nil {
		shippingService = NewShippingService(repos.Shipping)
	}
	return s.quote(input, checkoutRepositories{
		productRepo:     repos.Product,
		couponRepo:      repos.Coupon,
		paymentRepo:     repos.Payment,
		loyaltyRepo:     repos.Loyalty,
		shippingService: shippingService,
	})
}

func (s *CheckoutService) quote(input CheckoutQuoteInput, repos checkoutRepositories) (*CheckoutQuote, error) {
	if len(input.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	items := make([]order.OrderItem, len(input.Items))
	shippingItems := make([]ShippingQuoteItemInput, 0, len(input.Items))
	var subtotal float64
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
		shippingItems = append(shippingItems, ShippingQuoteItemInput{
			ProductID:     product.ID,
			VariantID:     variantID,
			ProductTypeID: product.ProductTypeID,
			Quantity:      item.Quantity,
			UnitPrice:     price,
			WeightGrams:   variant.Weight,
		})
	}

	if repos.shippingService == nil {
		return nil, errors.New("shipping quote service is not configured")
	}
	shippingQuote, err := repos.shippingService.QuoteResolvedItems(ShippingQuoteInput{
		Country: input.ShippingAddress.Country,
		Amount:  subtotal,
		Items:   shippingItems,
	})
	if err != nil {
		return nil, err
	}
	shippingFee := shippingQuote.ShippingFee
	taxAmount := s.calculateTax(repos.paymentRepo, subtotal, input.ShippingAddress.Country, input.ShippingAddress.State)
	memberDiscount := s.calculateMemberDiscount(repos.loyaltyRepo, input.UserID, subtotal)
	pointsToUse, pointsDiscount, err := s.calculatePointsDiscount(repos.loyaltyRepo, input.UserID, input.PointsToUse, subtotal)
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

	return &CheckoutQuote{
		Items:          items,
		SubtotalAmount: subtotal,
		ShippingFee:    shippingFee,
		ShippingQuote:  shippingQuote,
		TaxAmount:      taxAmount,
		MemberDiscount: memberDiscount,
		PointsDiscount: pointsDiscount,
		CouponDiscount: couponDiscount,
		DiscountAmount: discountAmount,
		TotalAmount:    totalAmount,
		CouponCode:     input.CouponCode,
		PointsToUse:    pointsToUse,
		Coupon:         targetCoupon,
	}, nil
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

func (s *CheckoutService) calculatePointsDiscount(loyaltyRepo *repository.LoyaltyRepository, userID uint, requestedPoints int, subtotal float64) (int, float64, error) {
	if requestedPoints <= 0 {
		return 0, 0, nil
	}

	userLoyalty, err := loyaltyRepo.FindUserLoyaltyByUserID(userID)
	if err != nil || userLoyalty == nil {
		return 0, 0, fmt.Errorf("[CRITICAL] Insufficient points: available %d, requested %d", 0, requestedPoints)
	}
	if userLoyalty.AvailablePoints < requestedPoints {
		return 0, 0, fmt.Errorf("[CRITICAL] Insufficient points: available %d, requested %d", userLoyalty.AvailablePoints, requestedPoints)
	}

	pointsDiscount := float64(requestedPoints) * 0.01
	pointsToUse := requestedPoints
	if maxPointsDiscount := subtotal * 0.5; pointsDiscount > maxPointsDiscount {
		pointsDiscount = maxPointsDiscount
		pointsToUse = int(maxPointsDiscount * 100)
	}

	return pointsToUse, pointsDiscount, nil
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
