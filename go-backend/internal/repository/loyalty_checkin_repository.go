package repository

import (
	"tanzanite/internal/domain/loyalty"
	"time"

	"gorm.io/gorm"
)

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
