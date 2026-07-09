package service

import (
	"errors"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/repository"
	"time"
)

type MarketingService struct {
	txManager   *repository.TxManager
	couponRepo  *repository.CouponRepository
	loyaltyRepo *repository.LoyaltyRepository
}

var (
	ErrMarketingNotFound  = errors.New("marketing resource not found")
	ErrCouponCodeExists   = errors.New("coupon code already exists")
	ErrGiftCardCodeExists = errors.New("gift card code already exists")
)

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
	txManager *repository.TxManager,
	couponRepo *repository.CouponRepository,
	loyaltyRepo *repository.LoyaltyRepository,
) *MarketingService {
	s := &MarketingService{
		txManager:   txManager,
		couponRepo:  couponRepo,
		loyaltyRepo: loyaltyRepo,
	}
	return s
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

// Loyalty 相关方法

// EarnPoints 赚取积分
func (s *MarketingService) EarnPoints(userID uint, points int, source string, sourceID uint, description string) error {
	if points <= 0 {
		return errors.New("points must be positive")
	}
	_, err := s.loyaltyRepo.AdjustUserPoints(userID, points, "earn", source, sourceID, description)
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

func (s *MarketingService) CountRedeemedGiftCards(userID uint) (int64, error) {
	return s.loyaltyRepo.CountTransactionsByUserAndSource(userID, "spend", "giftcard")
}

// 辅助方法

func normalizeMarketingError(err error) error {
	if repository.IsRecordNotFound(err) {
		return ErrMarketingNotFound
	}
	return err
}

// ==========================================
// B端 (Admin) 会员与积分管理方法
// ==========================================

// ListMemberLevels 获取所有会员等级
func (s *MarketingService) ListMemberLevels() ([]loyalty.MemberLevel, error) {
	return s.loyaltyRepo.FindAllMemberLevels()
}
