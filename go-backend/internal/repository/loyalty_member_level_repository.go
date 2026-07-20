package repository

import "tanzanite/internal/domain/loyalty"

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
	err := r.db.Order("sort_order ASC, min_points ASC, id ASC").Find(&levels).Error
	return levels, err
}

// FindMemberLevelByPoints 根据积分查找对应等级
func (r *LoyaltyRepository) FindMemberLevelByPoints(points int) (*loyalty.MemberLevel, error) {
	var level loyalty.MemberLevel
	err := r.db.Where("min_points <= ?", points).
		Order("min_points DESC").First(&level).Error
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
