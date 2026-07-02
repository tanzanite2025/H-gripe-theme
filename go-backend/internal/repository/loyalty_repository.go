package repository

import (
	"errors"
	"fmt"
	"tanzanite/internal/domain/loyalty"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LoyaltyRepository struct {
	db *gorm.DB
}

var (
	ErrInvalidUserID      = errors.New("user ID is required")
	ErrInvalidPoints      = errors.New("points must not be zero")
	ErrInsufficientPoints = errors.New("insufficient points")
)

func NewLoyaltyRepository(db *gorm.DB) *LoyaltyRepository {
	return &LoyaltyRepository{db: db}
}

// WithTx 使用指定的事务创建新的 repository 实例
func (r *LoyaltyRepository) WithTx(tx *gorm.DB) *LoyaltyRepository {
	return &LoyaltyRepository{db: tx}
}

// GetDB 获取底层 GORM DB 实例
// LoyaltyTransaction 相关方法

// CreateTransaction 创建积分交易
func (r *LoyaltyRepository) CreateTransaction(t *loyalty.LoyaltyTransaction) error {
	return r.db.Create(t).Error
}

// FindTransactionByID 根据ID查找交易
func (r *LoyaltyRepository) FindTransactionByID(id uint) (*loyalty.LoyaltyTransaction, error) {
	var t loyalty.LoyaltyTransaction
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindTransactionsByUserID 查找用户的积分交易记录
func (r *LoyaltyRepository) FindTransactionsByUserID(userID uint, page, pageSize int) ([]loyalty.LoyaltyTransaction, int64, error) {
	var transactions []loyalty.LoyaltyTransaction
	var total int64

	query := r.db.Model(&loyalty.LoyaltyTransaction{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&transactions).Error

	return transactions, total, err
}

// GetUserPointsBalance 获取用户积分余额
func (r *LoyaltyRepository) GetUserPointsBalance(userID uint) (int, error) {
	var userLoyalty loyalty.UserLoyalty
	err := r.db.Where("user_id = ?", userID).First(&userLoyalty).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return userLoyalty.AvailablePoints, nil
}

func (r *LoyaltyRepository) FindOrCreateUserLoyaltyForUpdate(userID uint) (*loyalty.UserLoyalty, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return findOrCreateUserLoyaltyForUpdate(r.db, userID)
}

func (r *LoyaltyRepository) SumTransactionPointsByUser(userID uint, transactionType, source string, start, end time.Time) (int, error) {
	var sumPoints int
	err := r.db.Model(&loyalty.LoyaltyTransaction{}).
		Where("user_id = ? AND type = ? AND source = ? AND created_at BETWEEN ? AND ?",
			userID, transactionType, source, start, end).
		Select("COALESCE(SUM(points), 0)").
		Scan(&sumPoints).Error
	return sumPoints, err
}

func (r *LoyaltyRepository) CountTransactionsByUserAndSource(userID uint, transactionType, source string) (int64, error) {
	var count int64
	err := r.db.Model(&loyalty.LoyaltyTransaction{}).
		Where("user_id = ? AND type = ? AND source = ?", userID, transactionType, source).
		Count(&count).Error
	return count, err
}

// AdjustUserPoints atomically updates a user's points summary and creates the matching ledger entry.
func (r *LoyaltyRepository) AdjustUserPoints(userID uint, points int, transactionType, source string, sourceID uint, description string) (*loyalty.LoyaltyTransaction, error) {
	var transaction *loyalty.LoyaltyTransaction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var err error
		transaction, err = r.WithTx(tx).AdjustUserPointsInCurrentTx(userID, points, transactionType, source, sourceID, description)
		return err
	})
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (r *LoyaltyRepository) AdjustUserPointsInCurrentTx(userID uint, points int, transactionType, source string, sourceID uint, description string) (*loyalty.LoyaltyTransaction, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if points == 0 {
		return nil, ErrInvalidPoints
	}

	if transactionType == "" {
		transactionType = "adjust"
	}

	if source == "" {
		source = transactionType
	}

	userLoyalty, err := findOrCreateUserLoyaltyForUpdate(r.db, userID)
	if err != nil {
		return nil, err
	}

	if points < 0 && userLoyalty.AvailablePoints+points < 0 {
		return nil, ErrInsufficientPoints
	}

	applyPointsDelta(userLoyalty, points, transactionType)

	if err := r.db.Save(userLoyalty).Error; err != nil {
		return nil, fmt.Errorf("failed to update user loyalty: %w", err)
	}

	transaction := &loyalty.LoyaltyTransaction{
		UserID:      userID,
		Type:        transactionType,
		Points:      points,
		Balance:     userLoyalty.AvailablePoints,
		Source:      source,
		SourceID:    sourceID,
		Description: description,
	}

	if err := r.db.Create(transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to create loyalty transaction: %w", err)
	}

	return transaction, nil
}

func findOrCreateUserLoyaltyForUpdate(tx *gorm.DB, userID uint) (*loyalty.UserLoyalty, error) {
	var userLoyalty loyalty.UserLoyalty
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&userLoyalty).Error
	if err == nil {
		return &userLoyalty, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoNothing: true,
	}).Create(&loyalty.UserLoyalty{UserID: userID}).Error; err != nil {
		return nil, fmt.Errorf("failed to initialize user loyalty: %w", err)
	}

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&userLoyalty).Error; err != nil {
		return nil, err
	}

	return &userLoyalty, nil
}

func applyPointsDelta(userLoyalty *loyalty.UserLoyalty, points int, transactionType string) {
	if points > 0 {
		userLoyalty.AvailablePoints += points
		if transactionType == "refund" {
			userLoyalty.UsedPoints -= points
			if userLoyalty.UsedPoints < 0 {
				userLoyalty.UsedPoints = 0
			}
			return
		}
		userLoyalty.TotalPoints += points
		return
	}

	userLoyalty.AvailablePoints += points
	userLoyalty.UsedPoints += -points
}

// CheckIn 相关方法

// CreateCheckIn 创建签到记录
func (r *LoyaltyRepository) CreateCheckIn(c *loyalty.CheckIn) error {
	return r.db.Create(c).Error
}

// FindCheckInByUserAndDate 查找用户某天的签到记录
func (r *LoyaltyRepository) FindCheckInByUserAndDate(userID uint, date time.Time) (*loyalty.CheckIn, error) {
	var checkIn loyalty.CheckIn
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Where("user_id = ? AND check_in_date >= ? AND check_in_date < ?",
		userID, startOfDay, endOfDay).First(&checkIn).Error
	if err != nil {
		return nil, err
	}
	return &checkIn, nil
}

// GetUserCheckInStreak 获取用户连续签到天数
func (r *LoyaltyRepository) GetUserCheckInStreak(userID uint) (int, error) {
	var checkIn loyalty.CheckIn
	err := r.db.Where("user_id = ?", userID).Order("check_in_date DESC").First(&checkIn).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return checkIn.ConsecutiveDays, nil
}

// FindCheckInsByUserID 查找用户的签到记录
func (r *LoyaltyRepository) FindCheckInsByUserID(userID uint, page, pageSize int) ([]loyalty.CheckIn, int64, error) {
	var checkIns []loyalty.CheckIn
	var total int64

	query := r.db.Model(&loyalty.CheckIn{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("check_in_date DESC").Offset(offset).Limit(pageSize).Find(&checkIns).Error

	return checkIns, total, err
}

// Referral 相关方法

// CreateReferral 创建推荐记录
func (r *LoyaltyRepository) CreateReferral(ref *loyalty.Referral) error {
	return r.db.Create(ref).Error
}

// FindReferralByID 根据ID查找推荐记录
func (r *LoyaltyRepository) FindReferralByID(id uint) (*loyalty.Referral, error) {
	var ref loyalty.Referral
	err := r.db.First(&ref, id).Error
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

// FindReferralsByReferrerID 查找推荐人的推荐记录
func (r *LoyaltyRepository) FindReferralsByReferrerID(referrerID uint) ([]loyalty.Referral, error) {
	var referrals []loyalty.Referral
	err := r.db.Where("referrer_id = ?", referrerID).Order("created_at DESC").Find(&referrals).Error
	return referrals, err
}

// FindReferralByRefereeID 根据被推荐人ID查找记录
func (r *LoyaltyRepository) FindReferralByRefereeID(refereeID uint) (*loyalty.Referral, error) {
	var ref loyalty.Referral
	err := r.db.Where("referee_id = ?", refereeID).First(&ref).Error
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

// UpdateReferral 更新推荐记录
func (r *LoyaltyRepository) UpdateReferral(ref *loyalty.Referral) error {
	return r.db.Save(ref).Error
}

// CountSuccessfulReferrals 统计成功推荐数
func (r *LoyaltyRepository) CountSuccessfulReferrals(referrerID uint) (int64, error) {
	var count int64
	err := r.db.Model(&loyalty.Referral{}).
		Where("referrer_id = ? AND status = ?", referrerID, "completed").
		Count(&count).Error
	return count, err
}

// MemberLevel 相关方法

// CreateMemberLevel 创建会员等级
func (r *LoyaltyRepository) CreateMemberLevel(l *loyalty.MemberLevel) error {
	return r.db.Create(l).Error
}

// FindMemberLevelByID 根据ID查找会员等级
func (r *LoyaltyRepository) FindMemberLevelByID(id uint) (*loyalty.MemberLevel, error) {
	var l loyalty.MemberLevel
	err := r.db.First(&l, id).Error
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// FindAllMemberLevels 查找所有会员等级
func (r *LoyaltyRepository) FindAllMemberLevels() ([]loyalty.MemberLevel, error) {
	var levels []loyalty.MemberLevel
	err := r.db.Order("level ASC").Find(&levels).Error
	return levels, err
}

// FindMemberLevelByPoints 根据积分查找对应等级
func (r *LoyaltyRepository) FindMemberLevelByPoints(points int) (*loyalty.MemberLevel, error) {
	var level loyalty.MemberLevel
	err := r.db.Where("required_points <= ?", points).
		Order("required_points DESC").First(&level).Error
	if err != nil {
		return nil, err
	}
	return &level, nil
}

// UpdateMemberLevel 更新会员等级
func (r *LoyaltyRepository) UpdateMemberLevel(l *loyalty.MemberLevel) error {
	return r.db.Save(l).Error
}

// DeleteMemberLevel 删除会员等级
func (r *LoyaltyRepository) DeleteMemberLevel(id uint) error {
	return r.db.Delete(&loyalty.MemberLevel{}, id).Error
}

// UserLoyalty 相关方法

// CreateUserLoyalty 创建用户会员信息
func (r *LoyaltyRepository) CreateUserLoyalty(u *loyalty.UserLoyalty) error {
	return r.db.Create(u).Error
}

// FindUserLoyaltyByUserID 根据用户ID查找会员信息
func (r *LoyaltyRepository) FindUserLoyaltyByUserID(userID uint) (*loyalty.UserLoyalty, error) {
	var ul loyalty.UserLoyalty
	err := r.db.Where("user_id = ?", userID).First(&ul).Error
	if err != nil {
		return nil, err
	}
	return &ul, nil
}

// UpdateUserLoyalty 更新用户会员信息
func (r *LoyaltyRepository) UpdateUserLoyalty(u *loyalty.UserLoyalty) error {
	return r.db.Save(u).Error
}

// UpdateUserPoints 更新用户积分
func (r *LoyaltyRepository) UpdateUserPoints(userID uint, points int) error {
	_, err := r.AdjustUserPoints(userID, points, "adjust", "system", 0, "System points update")
	return err
}

// GetLoyaltyStats 获取会员统计信息
func (r *LoyaltyRepository) GetLoyaltyStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总会员数
	var totalMembers int64
	if err := r.db.Model(&loyalty.UserLoyalty{}).Count(&totalMembers).Error; err != nil {
		return nil, err
	}
	stats["total_members"] = totalMembers

	// 各等级会员数
	var levelStats []struct {
		LevelID uint
		Count   int64
	}
	if err := r.db.Model(&loyalty.UserLoyalty{}).
		Select("level_id, count(*) as count").
		Group("level_id").
		Scan(&levelStats).Error; err != nil {
		return nil, err
	}
	stats["level_distribution"] = levelStats

	return stats, nil
}
