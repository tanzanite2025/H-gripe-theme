package service

import (
	"errors"
	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/repository"
	"time"
)

type MarketingService struct {
	txManager      *repository.TxManager
	couponRepo     *repository.CouponRepository
	loyaltyRepo    *repository.LoyaltyRepository
	redemptionRepo *repository.GiftCardRedemptionRepository
	setting        *SettingService
	program        *LoyaltyProgramService
}

func isDefaultMemberLevelName(name string) bool {
	switch name {
	case "Ordinary", "Bronze", "Silver", "Gold", "Platinum", "Diamond":
		return true
	default:
		return false
	}
}

var (
	ErrMarketingNotFound  = errors.New("marketing resource not found")
	ErrCouponCodeExists   = errors.New("coupon code already exists")
	ErrInvalidMemberLevel = errors.New("invalid member level")
)

type MemberLevelCreateInput struct {
	Name         string
	MinPoints    int
	MaxPoints    int
	DiscountRate float64
	Benefits     string
	Icon         string
	Color        string
	SortOrder    int
}

type MemberLevelUpdateInput struct {
	Name         *string
	MinPoints    *int
	MaxPoints    *int
	DiscountRate *float64
	Benefits     *string
	Icon         *string
	Color        *string
	SortOrder    *int
}

func NewMarketingService(
	txManager *repository.TxManager,
	couponRepo *repository.CouponRepository,
	loyaltyRepo *repository.LoyaltyRepository,
	settingServices ...*SettingService,
) *MarketingService {
	s := &MarketingService{
		txManager:   txManager,
		couponRepo:  couponRepo,
		loyaltyRepo: loyaltyRepo,
	}
	if len(settingServices) > 0 {
		s.setting = settingServices[0]
	}
	return s
}

func (s *MarketingService) ConfigureLoyaltyProgram(program *LoyaltyProgramService) {
	s.program = program
}

func (s *MarketingService) ConfigureGiftCardRedemptions(repo *repository.GiftCardRedemptionRepository) {
	s.redemptionRepo = repo
}

func (s *MarketingService) ConfigureCurrencyPolicy(policy *CurrencyPolicyService) {
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
	if err := s.validateMemberLevelInput(0, input.MinPoints, input.MaxPoints, input.DiscountRate); err != nil {
		return nil, err
	}

	level := &loyalty.MemberLevel{
		Name:         input.Name,
		MinPoints:    input.MinPoints,
		MaxPoints:    input.MaxPoints,
		DiscountRate: input.DiscountRate,
		Benefits:     input.Benefits,
		Icon:         input.Icon,
		Color:        input.Color,
		SortOrder:    input.SortOrder,
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

	if input.MinPoints != nil {
		level.MinPoints = *input.MinPoints
	}
	if input.MaxPoints != nil {
		level.MaxPoints = *input.MaxPoints
	}
	if input.DiscountRate != nil {
		level.DiscountRate = *input.DiscountRate
	}
	if input.Benefits != nil {
		level.Benefits = *input.Benefits
	}
	if err := s.validateMemberLevelInput(level.ID, level.MinPoints, level.MaxPoints, level.DiscountRate); err != nil {
		return nil, err
	}

	if err := s.loyaltyRepo.UpdateMemberLevel(level); err != nil {
		return nil, err
	}

	return level, nil
}

func (s *MarketingService) DeleteMemberLevelAdmin(id uint) error {
	level, err := s.GetMemberLevel(id)
	if err != nil {
		return err
	}
	if isDefaultMemberLevelName(level.Name) {
		return ErrInvalidMemberLevel
	}
	return s.loyaltyRepo.DeleteMemberLevel(id)
}

func (s *MarketingService) validateMemberLevelInput(excludeID uint, minPoints, maxPoints int, discountRate float64) error {
	if minPoints < 0 || maxPoints < minPoints {
		return ErrInvalidMemberLevel
	}
	if discountRate < 0 || discountRate > 100 {
		return ErrInvalidMemberLevel
	}
	overlaps, err := s.loyaltyRepo.CountOverlappingMemberLevels(excludeID, minPoints, maxPoints)
	if err != nil {
		return err
	}
	if overlaps > 0 {
		return ErrInvalidMemberLevel
	}
	return nil
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
	config, err := s.getCurrentProgramConfig()
	if err != nil {
		return 0, err
	}

	var points int
	err = s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		today := time.Now()
		if _, err := repos.Loyalty.FindOrCreateUserLoyaltyForUpdate(userID); err != nil {
			return err
		}
		existing, findErr := repos.Loyalty.FindCheckInByUserAndDate(userID, today)
		if findErr == nil && existing != nil {
			return errors.New("already checked in today")
		}
		if findErr != nil && !repository.IsRecordNotFound(findErr) {
			return findErr
		}

		streak, err := repos.Loyalty.GetUserCheckInStreak(userID)
		if err != nil {
			return err
		}

		points = config.CheckInBasePoints +
			(streak / config.CheckInStreakIntervalDays * config.CheckInStreakBonusPoints)
		if points > config.CheckInMaxPoints {
			points = config.CheckInMaxPoints
		}

		checkIn := &loyalty.CheckIn{
			UserID:          userID,
			CheckInDate:     today.Format("2006-01-02"),
			PointsEarned:    points,
			ConsecutiveDays: streak + 1,
		}
		if err := repos.Loyalty.CreateCheckIn(checkIn); err != nil {
			return err
		}

		_, err = repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
			userID,
			points,
			"earn",
			"checkin",
			checkIn.ID,
			"Daily check-in reward",
			programConfigID(config),
		)
		return err
	})
	if err != nil {
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
	config, err := s.getCurrentProgramConfig()
	if err != nil {
		return err
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		referral, err := repos.Loyalty.FindReferralByRefereeIDForUpdate(refereeID)
		if err != nil {
			return err
		}
		if referral.Status != "pending" {
			return errors.New("referral already completed")
		}

		now := time.Now()
		referral.Status = "completed"
		referral.CompletedAt = &now
		referral.CompletedOrderID = orderID
		referral.ReferrerPoints = config.ReferralReferrerPoints
		referral.ReferredPoints = config.ReferralRefereePoints
		referral.PointsEarned = config.ReferralReferrerPoints + config.ReferralRefereePoints

		if err := repos.Loyalty.UpdateReferral(referral); err != nil {
			return err
		}
		if _, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
			referral.ReferrerID,
			config.ReferralReferrerPoints,
			"earn",
			"referral",
			referral.ID,
			"Referral reward",
			programConfigID(config),
		); err != nil {
			return err
		}
		_, err = repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
			refereeID,
			config.ReferralRefereePoints,
			"earn",
			"referral",
			referral.ID,
			"New user referral bonus",
			programConfigID(config),
		)
		return err
	})
}

func (s *MarketingService) getCurrentProgramConfig() (*loyalty.ProgramConfig, error) {
	if s.program != nil {
		return s.program.GetActive()
	}
	return nil, ErrLoyaltyProgramConfigNotFound
}

func programConfigID(config *loyalty.ProgramConfig) *uint {
	if config == nil || config.ID == 0 {
		return nil
	}
	return &config.ID
}

// GetUserLoyalty 获取用户会员信息
func (s *MarketingService) GetUserLoyalty(userID uint) (*loyalty.UserLoyalty, error) {
	return s.loyaltyRepo.FindUserLoyaltyByUserID(userID)
}

func (s *MarketingService) CountRedeemedGiftCards(userID uint) (int64, error) {
	return s.couponRepo.CountGiftCardsByOwnerID(userID)
}

func (s *MarketingService) ListUserGiftCards(userID uint, page, pageSize int) ([]coupon.GiftCard, int64, error) {
	return s.couponRepo.FindGiftCardsByOwnerID(userID, page, pageSize)
}

func (s *MarketingService) ListGiftCardRedemptionsAdmin(userID uint, page, pageSize int) ([]coupon.GiftCardRedemption, int64, error) {
	if s.redemptionRepo == nil {
		return nil, 0, errors.New("gift card redemption repository is unavailable")
	}
	return s.redemptionRepo.FindByUserID(userID, page, pageSize)
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
	levels, err := s.loyaltyRepo.FindAllMemberLevels()
	if err != nil {
		return nil, err
	}
	if len(levels) > 0 {
		return levels, nil
	}

	for _, level := range defaultMemberLevels() {
		candidate := level
		if err := s.loyaltyRepo.CreateMemberLevel(&candidate); err != nil {
			return nil, err
		}
	}
	return s.loyaltyRepo.FindAllMemberLevels()
}

func defaultMemberLevels() []loyalty.MemberLevel {
	return []loyalty.MemberLevel{
		{Name: "Ordinary", MinPoints: 0, MaxPoints: 499, DiscountRate: 0, Benefits: "[]", Icon: "circle", Color: "#f8fafc", SortOrder: 0},
		{Name: "Bronze", MinPoints: 500, MaxPoints: 1999, DiscountRate: 0, Benefits: "[]", Icon: "medal", Color: "#b87333", SortOrder: 10},
		{Name: "Silver", MinPoints: 2000, MaxPoints: 4999, DiscountRate: 0, Benefits: "[]", Icon: "medal", Color: "#c0c0c0", SortOrder: 20},
		{Name: "Gold", MinPoints: 5000, MaxPoints: 9999, DiscountRate: 0, Benefits: "[]", Icon: "medal", Color: "#d4af37", SortOrder: 30},
		{Name: "Platinum", MinPoints: 10000, MaxPoints: 19999, DiscountRate: 0, Benefits: "[]", Icon: "gem", Color: "#e5e4e2", SortOrder: 40},
		{Name: "Diamond", MinPoints: 20000, MaxPoints: 999999999, DiscountRate: 0, Benefits: "[]", Icon: "gem", Color: "#b9f2ff", SortOrder: 50},
	}
}
