package repository

import "tanzanite/internal/domain/loyalty"

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
