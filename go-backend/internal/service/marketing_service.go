package service

import (
	"errors"
	"fmt"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/repository"
	"time"

	"github.com/google/uuid"
)

type MarketingService struct {
	couponRepo  *repository.CouponRepository
	loyaltyRepo *repository.LoyaltyRepository
}

func NewMarketingService(
	couponRepo *repository.CouponRepository,
	loyaltyRepo *repository.LoyaltyRepository,
) *MarketingService {
	return &MarketingService{
		couponRepo:  couponRepo,
		loyaltyRepo: loyaltyRepo,
	}
}

// Coupon 相关方法

// CreateCoupon 创建优惠券
func (s *MarketingService) CreateCoupon(c *coupon.Coupon) error {
	// 验证优惠券代码唯一性
	existing, _ := s.couponRepo.FindCouponByCode(c.Code)
	if existing != nil {
		return errors.New("coupon code already exists")
	}

	return s.couponRepo.CreateCoupon(c)
}

// ValidateCoupon 验证优惠券
func (s *MarketingService) ValidateCoupon(code string, userID uint, amount float64) (*coupon.Coupon, float64, error) {
	c, err := s.couponRepo.FindCouponByCode(code)
	if err != nil {
		return nil, 0, errors.New("coupon not found")
	}

	// 检查是否启用
	if !c.Enabled {
		return nil, 0, errors.New("coupon is disabled")
	}

	// 检查有效期
	now := time.Now()
	if now.Before(c.StartDate) || now.After(c.EndDate) {
		return nil, 0, errors.New("coupon is expired")
	}

	// 检查使用次数
	if c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit {
		return nil, 0, errors.New("coupon usage limit reached")
	}

	// 检查最小金额
	if amount < c.MinAmount {
		return nil, 0, fmt.Errorf("minimum amount %.2f required", c.MinAmount)
	}

	// 计算折扣
	var discount float64
	if c.Type == "fixed" {
		discount = c.Value
	} else if c.Type == "percentage" {
		discount = amount * c.Value / 100
		if c.MaxDiscount > 0 && discount > c.MaxDiscount {
			discount = c.MaxDiscount
		}
	}

	return c, discount, nil
}

// UseCoupon 使用优惠券
func (s *MarketingService) UseCoupon(couponID, userID, orderID uint, discountAmount float64) error {
	// 增加使用次数
	if err := s.couponRepo.IncrementUsedCount(couponID); err != nil {
		return err
	}

	// 创建使用记录
	usage := &coupon.CouponUsage{
		CouponID: couponID,
		UserID:   userID,
		OrderID:  orderID,
		Discount: discountAmount,
	}

	return s.couponRepo.CreateCouponUsage(usage)
}

// GetActiveCoupons 获取有效优惠券
func (s *MarketingService) GetActiveCoupons() ([]coupon.Coupon, error) {
	return s.couponRepo.FindActiveCoupons()
}

// GiftCard 相关方法

// CreateGiftCard 创建礼品卡
func (s *MarketingService) CreateGiftCard(userID uint, amount float64) (*coupon.GiftCard, error) {
	code := s.generateGiftCardCode()

	card := &coupon.GiftCard{
		Code:         code,
		Balance:      amount,
		InitialValue: amount,
		Status:       "active",
	}

	if err := s.couponRepo.CreateGiftCard(card); err != nil {
		return nil, err
	}

	return card, nil
}

// UseGiftCard 使用礼品卡
func (s *MarketingService) UseGiftCard(code string, amount float64, orderID uint) error {
	card, err := s.couponRepo.FindGiftCardByCode(code)
	if err != nil {
		return errors.New("gift card not found")
	}

	if card.Status != "active" {
		return errors.New("gift card is not active")
	}

	if card.Balance < amount {
		return errors.New("insufficient balance")
	}

	// 检查过期时间
	if card.ExpiresAt != nil && time.Now().After(*card.ExpiresAt) {
		return errors.New("gift card is expired")
	}

	// 扣除余额
	if err := s.couponRepo.UpdateGiftCardBalance(card.ID, -amount); err != nil {
		return err
	}

	// 创建交易记录
	transaction := &coupon.GiftCardTransaction{
		GiftCardID: card.ID,
		OrderID:    orderID,
		Amount:     -amount,
		Type:       "use",
		Balance:    card.Balance - amount,
	}

	return s.couponRepo.CreateGiftCardTransaction(transaction)
}

// Loyalty 相关方法

// EarnPoints 赚取积分
func (s *MarketingService) EarnPoints(userID uint, points int, source string, sourceID uint, description string) error {
	// 获取当前余额
	balance, err := s.loyaltyRepo.GetUserPointsBalance(userID)
	if err != nil {
		balance = 0
	}

	// 创建交易记录
	transaction := &loyalty.LoyaltyTransaction{
		UserID:      userID,
		Type:        "earn",
		Points:      points,
		Balance:     balance + points,
		Source:      source,
		SourceID:    sourceID,
		Description: description,
	}

	if err := s.loyaltyRepo.CreateTransaction(transaction); err != nil {
		return err
	}

	// 更新用户积分
	return s.loyaltyRepo.UpdateUserPoints(userID, points)
}

// SpendPoints 消费积分
func (s *MarketingService) SpendPoints(userID uint, points int, orderID uint) error {
	// 检查余额
	balance, err := s.loyaltyRepo.GetUserPointsBalance(userID)
	if err != nil {
		return err
	}

	if balance < points {
		return errors.New("insufficient points")
	}

	// 创建交易记录
	transaction := &loyalty.LoyaltyTransaction{
		UserID:      userID,
		Type:        "spend",
		Points:      -points,
		Balance:     balance - points,
		Source:      "order",
		SourceID:    orderID,
		Description: fmt.Sprintf("Spent %d points on order", points),
	}

	if err := s.loyaltyRepo.CreateTransaction(transaction); err != nil {
		return err
	}

	// 更新用户积分
	return s.loyaltyRepo.UpdateUserPoints(userID, -points)
}

// CheckIn 签到
func (s *MarketingService) CheckIn(userID uint) (int, error) {
	// 检查今天是否已签到
	today := time.Now()
	existing, _ := s.loyaltyRepo.FindCheckInByUserAndDate(userID, today)
	if existing != nil {
		return 0, errors.New("already checked in today")
	}

	// 获取连续签到天数
	streak, _ := s.loyaltyRepo.GetUserCheckInStreak(userID)

	// 计算奖励积分（连续签到奖励更多）
	points := 10 + (streak / 7 * 5) // 每连续7天多奖励5积分
	if points > 50 {
		points = 50 // 最多50积分
	}

	// 创建签到记录
	checkIn := &loyalty.CheckIn{
		UserID:          userID,
		CheckInDate:     today.Format("2006-01-02"),
		PointsEarned:    points,
		ConsecutiveDays: streak + 1,
	}

	if err := s.loyaltyRepo.CreateCheckIn(checkIn); err != nil {
		return 0, err
	}

	// 奖励积分
	if err := s.EarnPoints(userID, points, "checkin", checkIn.ID, "Daily check-in reward"); err != nil {
		return 0, err
	}

	return points, nil
}

// CreateReferral 创建推荐
func (s *MarketingService) CreateReferral(referrerID, refereeID uint) error {
	// 检查是否已经被推荐过
	existing, _ := s.loyaltyRepo.FindReferralByRefereeID(refereeID)
	if existing != nil {
		return errors.New("user already referred")
	}

	referral := &loyalty.Referral{
		ReferrerID: referrerID,
		ReferredID: refereeID,
		Status:     "pending",
	}

	return s.loyaltyRepo.CreateReferral(referral)
}

// CompleteReferral 完成推荐（被推荐人首次购买后）
func (s *MarketingService) CompleteReferral(refereeID uint, orderID uint) error {
	referral, err := s.loyaltyRepo.FindReferralByRefereeID(refereeID)
	if err != nil {
		return err
	}

	if referral.Status != "pending" {
		return errors.New("referral already completed")
	}

	// 更新推荐状态
	referral.Status = "completed"
	referral.CompletedAt = &time.Time{}
	*referral.CompletedAt = time.Now()

	if err := s.loyaltyRepo.UpdateReferral(referral); err != nil {
		return err
	}

	// 奖励推荐人积分
	referrerPoints := 100
	if err := s.EarnPoints(referral.ReferrerID, referrerPoints, "referral", referral.ID, "Referral reward"); err != nil {
		return err
	}

	// 奖励被推荐人积分
	refereePoints := 50
	return s.EarnPoints(refereeID, refereePoints, "referral", referral.ID, "New user referral bonus")
}

// GetUserLoyalty 获取用户会员信息
func (s *MarketingService) GetUserLoyalty(userID uint) (*loyalty.UserLoyalty, error) {
	return s.loyaltyRepo.FindUserLoyaltyByUserID(userID)
}

// 辅助方法

// generateGiftCardCode 生成礼品卡代码
func (s *MarketingService) generateGiftCardCode() string {
	return fmt.Sprintf("GC%s", uuid.New().String()[:12])
}
