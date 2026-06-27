package service

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MarketingService struct {
	couponRepo  *repository.CouponRepository
	loyaltyRepo *repository.LoyaltyRepository
}

var (
	ErrMarketingNotFound  = errors.New("marketing resource not found")
	ErrCouponCodeExists   = errors.New("coupon code already exists")
	ErrGiftCardCodeExists = errors.New("gift card code already exists")
)

type CouponCreateInput struct {
	Code                 string
	Type                 string
	Value                float64
	Description          string
	MinAmount            float64
	MaxDiscount          float64
	UsageLimit           int
	UsageLimitPerUser    int
	StartDate            time.Time
	EndDate              time.Time
	ApplicableProducts   string
	ExcludedProducts     string
	ApplicableCategories string
	Enabled              bool
}

type CouponUpdateInput struct {
	Code                 *string
	Type                 *string
	Value                *float64
	Description          *string
	MinAmount            *float64
	MaxDiscount          *float64
	UsageLimit           *int
	UsageLimitPerUser    *int
	StartDate            *time.Time
	EndDate              *time.Time
	ApplicableProducts   *string
	ExcludedProducts     *string
	ApplicableCategories *string
	Enabled              *bool
}

type GiftCardCreateInput struct {
	Code           string
	InitialValue   float64
	Currency       string
	RecipientEmail string
	RecipientName  string
	SenderName     string
	Message        string
	CoverImage     string
	ExpiresAt      *time.Time
}

type GiftCardDetail struct {
	GiftCard     *coupon.GiftCard
	Transactions []coupon.GiftCardTransaction
}

type MemberLevelCreateInput struct {
	Name             string
	MinPoints        int
	MaxPoints        int
	DiscountRate     float64
	PointsMultiplier float64
	Benefits         string
	Icon             string
	Color            string
	SortOrder        int
}

type MemberLevelUpdateInput struct {
	Name             *string
	MinPoints        *int
	MaxPoints        *int
	DiscountRate     *float64
	PointsMultiplier *float64
	Benefits         *string
	Icon             *string
	Color            *string
	SortOrder        *int
}

func NewMarketingService(
	couponRepo *repository.CouponRepository,
	loyaltyRepo *repository.LoyaltyRepository,
) *MarketingService {
	s := &MarketingService{
		couponRepo:  couponRepo,
		loyaltyRepo: loyaltyRepo,
	}
	return s
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
	switch c.Type {
	case "fixed":
		discount = c.Value
	case "percentage":
		discount = amount * c.Value / 100
		if c.MaxDiscount > 0 && discount > c.MaxDiscount {
			discount = c.MaxDiscount
		}
	}

	return c, discount, nil
}

// UseCoupon 使用优惠券
func (s *MarketingService) UseCoupon(couponID, userID, orderID uint, discountAmount float64) error {
	return s.couponRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		txCouponRepo := s.couponRepo.WithTx(tx)

		// 增加使用次数
		if err := txCouponRepo.IncrementUsedCount(couponID); err != nil {
			return err
		}

		// 创建使用记录
		usage := &coupon.CouponUsage{
			CouponID: couponID,
			UserID:   userID,
			OrderID:  orderID,
			Discount: discountAmount,
		}

		return txCouponRepo.CreateCouponUsage(usage)
	})
}

// GetActiveCoupons 获取有效优惠券
func (s *MarketingService) GetActiveCoupons() ([]coupon.Coupon, error) {
	return s.couponRepo.FindActiveCoupons()
}

// GetAllCoupons 获取全部优惠券 (供管理端使用)
func (s *MarketingService) GetAllCoupons(page, pageSize int) ([]coupon.Coupon, int64, error) {
	return s.couponRepo.FindAllCoupons(page, pageSize)
}

func (s *MarketingService) ListCouponsAdmin(page, pageSize int, status string) ([]coupon.Coupon, int64, error) {
	coupons, total, err := s.couponRepo.FindAllCoupons(page, pageSize)
	if err != nil || status == "" || status == "all" {
		return coupons, total, err
	}

	filtered := make([]coupon.Coupon, 0, len(coupons))
	now := time.Now()
	for _, cp := range coupons {
		switch status {
		case "active":
			if cp.Enabled && now.After(cp.StartDate) && now.Before(cp.EndDate) {
				filtered = append(filtered, cp)
			}
		case "expired":
			if now.After(cp.EndDate) {
				filtered = append(filtered, cp)
			}
		case "disabled":
			if !cp.Enabled {
				filtered = append(filtered, cp)
			}
		default:
			return nil, 0, fmt.Errorf("unsupported coupon status filter %s", status)
		}
	}

	return filtered, int64(len(filtered)), nil
}

func (s *MarketingService) GetCoupon(id uint) (*coupon.Coupon, error) {
	cp, err := s.couponRepo.FindCouponByID(id)
	if err != nil {
		return nil, normalizeMarketingError(err)
	}
	return cp, nil
}

func (s *MarketingService) CreateCouponAdmin(input CouponCreateInput) (*coupon.Coupon, error) {
	if err := s.ensureCouponCodeAvailable(input.Code, 0); err != nil {
		return nil, err
	}

	cp := &coupon.Coupon{
		Code:                 input.Code,
		Type:                 input.Type,
		Value:                input.Value,
		Description:          input.Description,
		MinAmount:            input.MinAmount,
		MaxDiscount:          input.MaxDiscount,
		UsageLimit:           input.UsageLimit,
		UsageLimitPerUser:    input.UsageLimitPerUser,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		ApplicableProducts:   input.ApplicableProducts,
		ExcludedProducts:     input.ExcludedProducts,
		ApplicableCategories: input.ApplicableCategories,
		Enabled:              input.Enabled,
	}

	if err := s.couponRepo.CreateCoupon(cp); err != nil {
		return nil, err
	}

	return cp, nil
}

func (s *MarketingService) UpdateCouponAdmin(id uint, input CouponUpdateInput) (*coupon.Coupon, error) {
	cp, err := s.GetCoupon(id)
	if err != nil {
		return nil, err
	}

	if input.Code != nil && *input.Code != cp.Code {
		if err := s.ensureCouponCodeAvailable(*input.Code, cp.ID); err != nil {
			return nil, err
		}
		cp.Code = *input.Code
	}
	if input.Type != nil {
		cp.Type = *input.Type
	}
	if input.Value != nil {
		cp.Value = *input.Value
	}
	if input.Description != nil {
		cp.Description = *input.Description
	}
	if input.MinAmount != nil {
		cp.MinAmount = *input.MinAmount
	}
	if input.MaxDiscount != nil {
		cp.MaxDiscount = *input.MaxDiscount
	}
	if input.UsageLimit != nil {
		cp.UsageLimit = *input.UsageLimit
	}
	if input.UsageLimitPerUser != nil {
		cp.UsageLimitPerUser = *input.UsageLimitPerUser
	}
	if input.StartDate != nil {
		cp.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		cp.EndDate = *input.EndDate
	}
	if input.ApplicableProducts != nil {
		cp.ApplicableProducts = *input.ApplicableProducts
	}
	if input.ExcludedProducts != nil {
		cp.ExcludedProducts = *input.ExcludedProducts
	}
	if input.ApplicableCategories != nil {
		cp.ApplicableCategories = *input.ApplicableCategories
	}
	if input.Enabled != nil {
		cp.Enabled = *input.Enabled
	}

	if err := s.couponRepo.UpdateCoupon(cp); err != nil {
		return nil, err
	}

	return cp, nil
}

func (s *MarketingService) DeleteCouponAdmin(id uint) error {
	if _, err := s.GetCoupon(id); err != nil {
		return err
	}
	return s.couponRepo.DeleteCoupon(id)
}

func (s *MarketingService) GetCouponStats() (map[string]interface{}, error) {
	coupons, _, err := s.couponRepo.FindAllCoupons(1, 1000)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	stats := map[string]interface{}{
		"total":    len(coupons),
		"active":   0,
		"expired":  0,
		"disabled": 0,
		"used":     0,
	}

	totalUsed := 0
	for _, cp := range coupons {
		if cp.Enabled && now.After(cp.StartDate) && now.Before(cp.EndDate) {
			stats["active"] = stats["active"].(int) + 1
		} else if now.After(cp.EndDate) {
			stats["expired"] = stats["expired"].(int) + 1
		} else if !cp.Enabled {
			stats["disabled"] = stats["disabled"].(int) + 1
		}
		totalUsed += cp.UsedCount
	}
	stats["used"] = totalUsed

	return stats, nil
}

func (s *MarketingService) ListGiftCardsAdmin(page, pageSize int, status string) ([]coupon.GiftCard, int64, error) {
	return s.couponRepo.FindAllGiftCards(page, pageSize, status)
}

func (s *MarketingService) GetGiftCard(id uint) (*GiftCardDetail, error) {
	card, err := s.couponRepo.FindGiftCardByID(id)
	if err != nil {
		return nil, normalizeMarketingError(err)
	}

	transactions, err := s.couponRepo.FindGiftCardTransactionsByCardID(id)
	if err != nil {
		return nil, err
	}

	return &GiftCardDetail{
		GiftCard:     card,
		Transactions: transactions,
	}, nil
}

func (s *MarketingService) CreateGiftCardAdmin(input GiftCardCreateInput) (*coupon.GiftCard, error) {
	if input.Currency == "" {
		input.Currency = "USD"
	}

	var card coupon.GiftCard
	err := s.loyaltyRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		txCouponRepo := s.couponRepo.WithTx(tx)
		if err := ensureGiftCardCodeAvailable(txCouponRepo, input.Code, 0); err != nil {
			return err
		}

		card = coupon.GiftCard{
			Code:           input.Code,
			InitialValue:   input.InitialValue,
			Balance:        input.InitialValue,
			Currency:       input.Currency,
			Status:         "active",
			RecipientEmail: input.RecipientEmail,
			RecipientName:  input.RecipientName,
			SenderName:     input.SenderName,
			Message:        input.Message,
			CoverImage:     input.CoverImage,
			ExpiresAt:      input.ExpiresAt,
		}
		if err := txCouponRepo.CreateGiftCard(&card); err != nil {
			return err
		}

		transaction := &coupon.GiftCardTransaction{
			GiftCardID: card.ID,
			Type:       "issue",
			Amount:     input.InitialValue,
			Balance:    input.InitialValue,
			Note:       "Admin issued gift card",
		}
		return txCouponRepo.CreateGiftCardTransaction(transaction)
	})
	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (s *MarketingService) UpdateGiftCardStatus(id uint, status string) (*coupon.GiftCard, error) {
	card, err := s.couponRepo.FindGiftCardByID(id)
	if err != nil {
		return nil, normalizeMarketingError(err)
	}

	card.Status = status
	if err := s.couponRepo.UpdateGiftCard(card); err != nil {
		return nil, err
	}

	return card, nil
}

func (s *MarketingService) ListLoyaltyTransactions(userID uint, page, pageSize int) ([]loyalty.LoyaltyTransaction, int64, error) {
	return s.loyaltyRepo.FindTransactionsByUserID(userID, page, pageSize)
}

func (s *MarketingService) AdminAdjustPointsWithTransaction(userID uint, points int, description string) (*loyalty.LoyaltyTransaction, error) {
	if description == "" {
		description = "Admin adjustment"
	}
	return s.loyaltyRepo.AdjustUserPoints(userID, points, "adjust", "admin", 0, description)
}

func (s *MarketingService) ListCheckIns(userID uint, page, pageSize int) ([]loyalty.CheckIn, int64, error) {
	return s.loyaltyRepo.FindCheckInsByUserID(userID, page, pageSize)
}

func (s *MarketingService) ListReferrals(referrerID uint) ([]loyalty.Referral, error) {
	return s.loyaltyRepo.FindReferralsByReferrerID(referrerID)
}

func (s *MarketingService) UpdateReferralStatus(id uint, status string) (*loyalty.Referral, error) {
	referral, err := s.loyaltyRepo.FindReferralByID(id)
	if err != nil {
		return nil, normalizeMarketingError(err)
	}

	referral.Status = status
	if status == "completed" && referral.CompletedAt == nil {
		now := time.Now()
		referral.CompletedAt = &now
	}

	if err := s.loyaltyRepo.UpdateReferral(referral); err != nil {
		return nil, err
	}

	return referral, nil
}

func (s *MarketingService) GetMemberLevel(id uint) (*loyalty.MemberLevel, error) {
	level, err := s.loyaltyRepo.FindMemberLevelByID(id)
	if err != nil {
		return nil, normalizeMarketingError(err)
	}
	return level, nil
}

func (s *MarketingService) CreateMemberLevelAdmin(input MemberLevelCreateInput) (*loyalty.MemberLevel, error) {
	level := &loyalty.MemberLevel{
		Name:             input.Name,
		MinPoints:        input.MinPoints,
		MaxPoints:        input.MaxPoints,
		DiscountRate:     input.DiscountRate,
		PointsMultiplier: input.PointsMultiplier,
		Benefits:         input.Benefits,
		Icon:             input.Icon,
		Color:            input.Color,
		SortOrder:        input.SortOrder,
	}
	if err := s.loyaltyRepo.CreateMemberLevel(level); err != nil {
		return nil, err
	}
	return level, nil
}

func (s *MarketingService) UpdateMemberLevelAdmin(id uint, input MemberLevelUpdateInput) (*loyalty.MemberLevel, error) {
	level, err := s.GetMemberLevel(id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		level.Name = *input.Name
	}
	if input.MinPoints != nil {
		level.MinPoints = *input.MinPoints
	}
	if input.MaxPoints != nil {
		level.MaxPoints = *input.MaxPoints
	}
	if input.DiscountRate != nil {
		level.DiscountRate = *input.DiscountRate
	}
	if input.PointsMultiplier != nil {
		level.PointsMultiplier = *input.PointsMultiplier
	}
	if input.Benefits != nil {
		level.Benefits = *input.Benefits
	}
	if input.Icon != nil {
		level.Icon = *input.Icon
	}
	if input.Color != nil {
		level.Color = *input.Color
	}
	if input.SortOrder != nil {
		level.SortOrder = *input.SortOrder
	}

	if err := s.loyaltyRepo.UpdateMemberLevel(level); err != nil {
		return nil, err
	}

	return level, nil
}

func (s *MarketingService) DeleteMemberLevelAdmin(id uint) error {
	if _, err := s.GetMemberLevel(id); err != nil {
		return err
	}
	return s.loyaltyRepo.DeleteMemberLevel(id)
}

func (s *MarketingService) GetMarketingStats() (map[string]interface{}, error) {
	couponStats, err := s.GetCouponStats()
	if err != nil {
		return nil, err
	}

	loyaltyStats, err := s.loyaltyRepo.GetLoyaltyStats()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"coupons": couponStats,
		"loyalty": loyaltyStats,
	}, nil
}

// UpdateCoupon 更新优惠券
func (s *MarketingService) UpdateCoupon(c *coupon.Coupon) error {
	return s.couponRepo.UpdateCoupon(c)
}

// DeleteCoupon 删除优惠券
func (s *MarketingService) DeleteCoupon(id uint) error {
	return s.couponRepo.DeleteCoupon(id)
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

// GetGiftCardsByUserID 获取用户的礼品卡
func (s *MarketingService) GetGiftCardsByUserID(userID uint) ([]coupon.GiftCard, error) {
	return s.couponRepo.FindGiftCardsByUserID(userID)
}

// UseGiftCard 使用礼品卡
func (s *MarketingService) UseGiftCard(code string, amount float64, orderID uint) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	return s.couponRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		txCouponRepo := s.couponRepo.WithTx(tx)

		card, err := txCouponRepo.FindGiftCardByCodeForUpdate(code)
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

		remainingBalance := card.Balance - amount

		// 扣除余额
		if err := txCouponRepo.UpdateGiftCardBalance(card.ID, -amount); err != nil {
			return err
		}

		// 创建交易记录
		transaction := &coupon.GiftCardTransaction{
			GiftCardID: card.ID,
			OrderID:    orderID,
			Amount:     -amount,
			Type:       "use",
			Balance:    remainingBalance,
		}

		return txCouponRepo.CreateGiftCardTransaction(transaction)
	})
}

// Loyalty 相关方法

// EarnPoints 赚取积分
func (s *MarketingService) EarnPoints(userID uint, points int, source string, sourceID uint, description string) error {
	if points <= 0 {
		return errors.New("points must be positive")
	}
	_, err := s.loyaltyRepo.AdjustUserPoints(userID, points, "earn", source, sourceID, description)
	return err
}

// SpendPoints 消费积分
func (s *MarketingService) SpendPoints(userID uint, points int, orderID uint) error {
	if points <= 0 {
		return errors.New("points must be positive")
	}
	_, err := s.loyaltyRepo.AdjustUserPoints(userID, -points, "spend", "order", orderID, fmt.Sprintf("Spent %d points on order", points))
	return err
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

func (s *MarketingService) ensureCouponCodeAvailable(code string, excludeID uint) error {
	existing, err := s.couponRepo.FindCouponByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return ErrCouponCodeExists
	}
	return nil
}

func ensureGiftCardCodeAvailable(repo *repository.CouponRepository, code string, excludeID uint) error {
	existing, err := repo.FindGiftCardByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return ErrGiftCardCodeExists
	}
	return nil
}

func normalizeMarketingError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrMarketingNotFound
	}
	return err
}

// ==========================================
// B端 (Admin) 会员与积分管理方法
// ==========================================

// CreateMemberLevel 创建会员等级
func (s *MarketingService) CreateMemberLevel(level *loyalty.MemberLevel) error {
	return s.loyaltyRepo.CreateMemberLevel(level)
}

// UpdateMemberLevel 更新会员等级
func (s *MarketingService) UpdateMemberLevel(level *loyalty.MemberLevel) error {
	return s.loyaltyRepo.UpdateMemberLevel(level)
}

// DeleteMemberLevel 删除会员等级
func (s *MarketingService) DeleteMemberLevel(id uint) error {
	return s.loyaltyRepo.DeleteMemberLevel(id)
}

// ListMemberLevels 获取所有会员等级
func (s *MarketingService) ListMemberLevels() ([]loyalty.MemberLevel, error) {
	return s.loyaltyRepo.FindAllMemberLevels()
}

// AdminAdjustPoints 管理员手动调整用户积分
func (s *MarketingService) AdminAdjustPoints(userID uint, points int, reason string) error {
	_, err := s.loyaltyRepo.AdjustUserPoints(userID, points, "adjust", "admin", 0, fmt.Sprintf("Admin Adjustment: %s", reason))
	return err
}

// ==========================================
// 积分兑换礼品卡核心业务方法
// ==========================================

// RedeemResult 兑换结果
type RedeemResult struct {
	GiftCardID      uint       `json:"giftcard_id"`
	CardCode        string     `json:"card_code"`
	Balance         float64    `json:"balance"`
	PointsSpent     int        `json:"points_spent"`
	PointsRemaining int        `json:"points_remaining"`
	ExpiresAt       *time.Time `json:"expires_at"`
}

// RedeemPointsForGiftCard 积分兑换礼品卡（原子事务）
// 将积分扣减、礼品卡生成、交易历史写入封装为统一的事务方法。
// 所有校验失败均返回明确错误信息（Fail Loudly 原则）。
func (s *MarketingService) RedeemPointsForGiftCard(
	userID uint,
	pointsToSpend int,
	giftCardValue float64,
	redeemCfg *setting.RedeemSettings,
) (*RedeemResult, error) {
	// 1. 校验配置是否开启
	if !redeemCfg.Enabled {
		return nil, errors.New("[CRITICAL] Point redemption is disabled")
	}

	// 2. 严格校验兑换率
	expectedPoints := int(giftCardValue * float64(redeemCfg.ExchangeRate))
	if expectedPoints != pointsToSpend {
		return nil, fmt.Errorf("[CRITICAL] Points mismatch: value %.2f requires %d points, got %d", giftCardValue, expectedPoints, pointsToSpend)
	}

	// 3. 校验最小起兑点
	if pointsToSpend < redeemCfg.MinPoints {
		return nil, fmt.Errorf("[CRITICAL] Minimum points required to redeem is %d", redeemCfg.MinPoints)
	}

	// 开启数据库事务
	var giftcard coupon.GiftCard
	var transaction *loyalty.LoyaltyTransaction

	err := s.loyaltyRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		txLoyaltyRepo := s.loyaltyRepo.WithTx(tx)
		txCouponRepo := s.couponRepo.WithTx(tx)
		// 4. 校验用户可用积分余额
		userLoyalty, err := txLoyaltyRepo.FindOrCreateUserLoyaltyForUpdate(userID)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to retrieve user loyalty data: %v", err)
		}

		if userLoyalty.AvailablePoints < pointsToSpend {
			return fmt.Errorf("[CRITICAL] Insufficient points: available %d, required %d", userLoyalty.AvailablePoints, pointsToSpend)
		}

		// 5. 校验今日兑换额度
		todayStart := time.Now().Truncate(24 * time.Hour)
		todayEnd := todayStart.Add(24 * time.Hour)

		sumPoints, err := txLoyaltyRepo.SumTransactionPointsByUser(userID, "spend", "giftcard", todayStart, todayEnd)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to verify daily limit: %v", err)
		}

		todayRedeemedValue := math.Abs(float64(sumPoints)) / float64(redeemCfg.ExchangeRate)
		if redeemCfg.MaxValuePerDay > 0 && todayRedeemedValue+giftCardValue > redeemCfg.MaxValuePerDay {
			return fmt.Errorf("[CRITICAL] Daily limit exceeded. Limit: %.2f, Redeemed: %.2f, Attempted: %.2f", redeemCfg.MaxValuePerDay, todayRedeemedValue, giftCardValue)
		}

		// 6. 生成礼品卡
		cardCode := "REDEEM-" + generateRedeemCode(12)
		var expiresAt *time.Time
		if redeemCfg.CardExpiryDays > 0 {
			t := time.Now().AddDate(0, 0, redeemCfg.CardExpiryDays)
			expiresAt = &t
		}

		giftcard = coupon.GiftCard{
			Code:         cardCode,
			InitialValue: giftCardValue,
			Balance:      giftCardValue,
			Currency:     "USD",
			Status:       "active",
			ExpiresAt:    expiresAt,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := txCouponRepo.CreateGiftCard(&giftcard); err != nil {
			return fmt.Errorf("[CRITICAL] Failed to create gift card: %v", err)
		}

		// 7. 原子扣减积分并写入交易历史
		transaction, err = txLoyaltyRepo.AdjustUserPoints(
			userID,
			-pointsToSpend,
			"spend",
			"giftcard",
			giftcard.ID,
			fmt.Sprintf("Redeemed gift card %s with %d points", cardCode, pointsToSpend),
		)
		if err != nil {
			return fmt.Errorf("[CRITICAL] Failed to deduct points: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RedeemResult{
		GiftCardID:      giftcard.ID,
		CardCode:        giftcard.Code,
		Balance:         giftcard.Balance,
		PointsSpent:     pointsToSpend,
		PointsRemaining: transaction.Balance,
		ExpiresAt:       giftcard.ExpiresAt,
	}, nil
}

// generateRedeemCode 生成指定长度的随机大写字母+数字字符串
func generateRedeemCode(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
