package service

import (
	"errors"
	"fmt"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/order"
	"tanzanite/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderService struct {
	db           *gorm.DB
	orderRepo    *repository.OrderRepository
	productRepo  *repository.ProductRepository
	couponRepo   *repository.CouponRepository
	paymentRepo  *repository.PaymentRepository
	shippingRepo *repository.ShippingRepository
	auditRepo    *repository.AuditRepository
	loyaltyRepo  *repository.LoyaltyRepository
}

func NewOrderService(
	db *gorm.DB,
	orderRepo *repository.OrderRepository,
	productRepo *repository.ProductRepository,
	couponRepo *repository.CouponRepository,
	paymentRepo *repository.PaymentRepository,
	shippingRepo *repository.ShippingRepository,
	auditRepo *repository.AuditRepository,
	loyaltyRepo *repository.LoyaltyRepository,
) *OrderService {
	return &OrderService{
		db:           db,
		orderRepo:    orderRepo,
		productRepo:  productRepo,
		couponRepo:   couponRepo,
		paymentRepo:  paymentRepo,
		shippingRepo: shippingRepo,
		auditRepo:    auditRepo,
		loyaltyRepo:  loyaltyRepo,
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(userID uint, items []order.OrderItem, shippingAddress, billingAddress order.Address, paymentMethod, shippingMethod string, couponCode string, pointsToUse int) (*order.Order, error) {
	// 生成订单号
	orderNumber := s.generateOrderNumber()
	
	// 计算订单金额
	var totalAmount float64
	for i := range items {
		product, err := s.productRepo.FindByID(items[i].ProductID)
		if err != nil {
			return nil, fmt.Errorf("[CRITICAL] Product ID %d not found in database: %w", items[i].ProductID, err)
		}
		
		items[i].Price = product.Price
		items[i].Subtotal = product.Price * float64(items[i].Quantity)
		totalAmount += items[i].Subtotal
	}
	
	// 计算运费
	shippingFee, err := s.calculateShippingFee(shippingMethod, totalAmount, shippingAddress.Country)
	if err != nil {
		shippingFee = 0 // 默认免运费
	}
	
	// 计算税费
	taxAmount, err := s.calculateTax(totalAmount, shippingAddress.Country, shippingAddress.State)
	if err != nil {
		taxAmount = 0
	}
	
	// 应用会员折扣 (Member Tier Discount)
	memberDiscount := 0.0
	userLoyalty, err := s.loyaltyRepo.FindUserLoyaltyByUserID(userID)
	if err == nil && userLoyalty != nil {
		tierDiscount := 0.0
		switch {
		case userLoyalty.TotalPoints >= 10000:
			tierDiscount = 0.20 // Platinum
		case userLoyalty.TotalPoints >= 5000:
			tierDiscount = 0.15 // Gold
		case userLoyalty.TotalPoints >= 2000:
			tierDiscount = 0.10 // Silver
		case userLoyalty.TotalPoints >= 500:
			tierDiscount = 0.05 // Bronze
		}
		if tierDiscount > 0 {
			memberDiscount = totalAmount * tierDiscount
		}
	}

	// 应用积分抵扣 (Points Discount)
	pointsDiscount := 0.0
	if pointsToUse > 0 {
		if userLoyalty == nil || userLoyalty.AvailablePoints < pointsToUse {
			return nil, fmt.Errorf("[CRITICAL] Insufficient points: available %d, requested %d", 
				func() int { if userLoyalty != nil { return userLoyalty.AvailablePoints } return 0 }(), pointsToUse)
		}
		
		// 1积分 = 0.01元
		pointsDiscount = float64(pointsToUse) * 0.01
		// 限制最多抵扣 50% 的订单金额
		if maxPointsDiscount := totalAmount * 0.5; pointsDiscount > maxPointsDiscount {
			pointsDiscount = maxPointsDiscount
			pointsToUse = int(maxPointsDiscount * 100)
		}
	}

	// 应用优惠券
	discountAmount := memberDiscount + pointsDiscount
	var targetCoupon *coupon.Coupon
	if couponCode != "" {
		discount, err := s.applyCoupon(couponCode, userID, totalAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to apply coupon %s: %w", couponCode, err)
		}
		discountAmount += discount
		cp, cpErr := s.couponRepo.FindCouponByCode(couponCode)
		if cpErr != nil {
			return nil, fmt.Errorf("referenced coupon code %s details not found: %w", couponCode, cpErr)
		}
		targetCoupon = cp
	}
	
	// 计算最终金额
	finalAmount := totalAmount + shippingFee + taxAmount - discountAmount
	if finalAmount < 0 {
		finalAmount = 0
	}
	
	// 创建订单
	o := &order.Order{
		OrderNumber:     orderNumber,
		UserID:          userID,
		Status:          "pending",
		PaymentMethod:   paymentMethod,
		PaymentStatus:   "unpaid",
		ShippingMethod:  shippingMethod,
		ShippingStatus:  "pending",
		TotalAmount:     finalAmount,
		ShippingFee:     shippingFee,
		TaxAmount:       taxAmount,
		DiscountAmount:  discountAmount,
		Items:           items,
		ShippingAddress: shippingAddress,
		BillingAddress:  billingAddress,
	}
	
	// 用事务原子化订单创建和优惠券扣减
	var createdOrder *order.Order
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		txOrderRepo := s.orderRepo.WithTx(tx)
		
		// 1. 创建订单和明细记录
		if err := txOrderRepo.Create(o); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to create order in database: %w", err)
		}
		createdOrder = o
		
		// 2. 优惠券扣减和记录
		if targetCoupon != nil {
			txCouponRepo := s.couponRepo.WithTx(tx)
			
			// 增加使用计数
			if err := txCouponRepo.IncrementUsedCount(targetCoupon.ID); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to increment usage count for coupon ID %d: %w", targetCoupon.ID, err)
			}
			
			// 创建使用凭证记录
			usage := &coupon.CouponUsage{
				CouponID:  targetCoupon.ID,
				UserID:    userID,
				OrderID:   o.ID,
				Discount:  discountAmount,
				CreatedAt: time.Now(),
			}
			if err := txCouponRepo.CreateCouponUsage(usage); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to record coupon usage for coupon ID %d: %w", targetCoupon.ID, err)
			}
		}

		// 3. 积分扣减和记录
		if pointsToUse > 0 {
			txLoyaltyRepo := s.loyaltyRepo.WithTx(tx)
			
			// 修改 UserLoyalty 表记录
			if err := tx.Exec("UPDATE user_loyalties SET available_points = available_points - ?, used_points = used_points + ? WHERE user_id = ?", pointsToUse, pointsToUse, userID).Error; err != nil {
				return fmt.Errorf("[CRITICAL] Failed to deduct points: %w", err)
			}
			
			// 记录积分流水
			txTransaction := &loyalty.LoyaltyTransaction{
				UserID:      userID,
				Points:      -pointsToUse,
				Type:        "redemption",
				Description: fmt.Sprintf("Order #%s points redemption", orderNumber),
				Balance:     userLoyalty.AvailablePoints - pointsToUse,
				CreatedAt:   time.Now(),
			}
			if err := txLoyaltyRepo.CreateTransaction(txTransaction); err != nil {
				return fmt.Errorf("[CRITICAL] Failed to record points transaction: %w", err)
			}
		}

		return nil
	})
	
	if txErr != nil {
		return nil, txErr
	}
	
	return createdOrder, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(id uint, userID uint) (*order.Order, error) {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	
	// 验证权限
	if o.UserID != userID {
		return nil, errors.New("unauthorized")
	}
	
	return o, nil
}

// GetUserOrders 获取用户订单列表
func (s *OrderService) GetUserOrders(userID uint, page, pageSize int) ([]order.Order, int64, error) {
	return s.orderRepo.FindByUserID(userID, page, pageSize)
}

// GetAllOrders 获取所有订单（管理员）
func (s *OrderService) GetAllOrders(page, pageSize int, status string) ([]order.Order, int64, error) {
	return s.orderRepo.FindAll(page, pageSize, status)
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(id uint, status string) error {
	// 验证状态转换
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return err
	}
	
	if !s.isValidStatusTransition(o.Status, status) {
		return fmt.Errorf("invalid status transition from %s to %s", o.Status, status)
	}
	
	return s.orderRepo.UpdateStatus(id, status)
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(id uint, userID uint) error {
	o, err := s.orderRepo.FindByID(id)
	if err != nil {
		return err
	}
	
	// 验证权限
	if o.UserID != userID {
		return errors.New("unauthorized")
	}
	
	// 只有pending和paid状态可以取消
	if o.Status != "pending" && o.Status != "paid" {
		return errors.New("order cannot be cancelled")
	}
	
	return s.orderRepo.UpdateStatus(id, "cancelled")
}

// GetOrderStats 获取订单统计
func (s *OrderService) GetOrderStats(userID uint) (map[string]int64, error) {
	return s.orderRepo.GetOrderStats(userID)
}

// 辅助方法

// generateOrderNumber 生成订单号
func (s *OrderService) generateOrderNumber() string {
	return fmt.Sprintf("ORD%s%s", time.Now().Format("20060102"), uuid.New().String()[:8])
}

// calculateShippingFee 计算运费
func (s *OrderService) calculateShippingFee(method string, amount float64, country string) (float64, error) {
	// 这里应该根据运费模板计算
	// 简化实现：固定运费
	if amount >= 100 {
		return 0, nil // 满100免运费
	}
	return 10.0, nil
}

// calculateTax 计算税费
func (s *OrderService) calculateTax(amount float64, country, state string) (float64, error) {
	taxRate, err := s.paymentRepo.FindTaxRateByLocation(country, state)
	if err != nil {
		return 0, nil // 没有税率配置则不收税
	}
	
	return amount * taxRate.Rate / 100, nil
}

// applyCoupon 应用优惠券
func (s *OrderService) applyCoupon(code string, userID uint, amount float64) (float64, error) {
	coupon, err := s.couponRepo.FindCouponByCode(code)
	if err != nil {
		return 0, err
	}
	
	// 验证优惠券
	if !coupon.Enabled {
		return 0, errors.New("coupon is disabled")
	}
	
	now := time.Now()
	if now.Before(coupon.StartDate) || now.After(coupon.EndDate) {
		return 0, errors.New("coupon is expired")
	}
	
	if coupon.UsageLimit > 0 && coupon.UsedCount >= coupon.UsageLimit {
		return 0, errors.New("coupon usage limit reached")
	}
	
	if amount < coupon.MinAmount {
		return 0, fmt.Errorf("minimum amount %.2f required", coupon.MinAmount)
	}
	
	// 计算折扣
	var discount float64
	if coupon.Type == "fixed" {
		discount = coupon.Value
	} else if coupon.Type == "percentage" {
		discount = amount * coupon.Value / 100
		if coupon.MaxDiscount > 0 && discount > coupon.MaxDiscount {
			discount = coupon.MaxDiscount
		}
	}
	
	return discount, nil
}

// isValidStatusTransition 验证状态转换是否有效
func (s *OrderService) isValidStatusTransition(from, to string) bool {
	validTransitions := map[string][]string{
		"pending":   {"paid", "cancelled"},
		"paid":      {"shipped", "cancelled"},
		"shipped":   {"completed"},
		"completed": {},
		"cancelled": {},
	}
	
	allowedStatuses, ok := validTransitions[from]
	if !ok {
		return false
	}
	
	for _, status := range allowedStatuses {
		if status == to {
			return true
		}
	}
	
	return false
}
