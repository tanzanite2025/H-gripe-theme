package repository

import "tanzanite/internal/domain/loyalty"

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
