package repository

import (
	"time"

	"tanzanite/internal/domain/coupon"

	"gorm.io/gorm"
)

type GiftCardRedemptionRepository struct {
	db *gorm.DB
}

func NewGiftCardRedemptionRepository(db *gorm.DB) *GiftCardRedemptionRepository {
	return &GiftCardRedemptionRepository{db: db}
}

func (r *GiftCardRedemptionRepository) WithTx(tx *gorm.DB) *GiftCardRedemptionRepository {
	return &GiftCardRedemptionRepository{db: tx}
}

func (r *GiftCardRedemptionRepository) FindByUserAndIdempotencyKey(userID uint, key string) (*coupon.GiftCardRedemption, error) {
	var redemption coupon.GiftCardRedemption
	err := r.db.
		Preload("GiftCard").
		Where("user_id = ? AND idempotency_key = ?", userID, key).
		First(&redemption).Error
	if err != nil {
		return nil, err
	}
	return &redemption, nil
}

func (r *GiftCardRedemptionRepository) Create(redemption *coupon.GiftCardRedemption) error {
	return r.db.Create(redemption).Error
}

func (r *GiftCardRedemptionRepository) Update(redemption *coupon.GiftCardRedemption) error {
	return r.db.Save(redemption).Error
}

func (r *GiftCardRedemptionRepository) FindByUserID(userID uint, page, pageSize int) ([]coupon.GiftCardRedemption, int64, error) {
	var redemptions []coupon.GiftCardRedemption
	var total int64
	query := r.db.Model(&coupon.GiftCardRedemption{}).
		Preload("GiftCard").
		Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&redemptions).Error
	return redemptions, total, err
}

func (r *GiftCardRedemptionRepository) SumValueCentsByUser(userID uint, start, end time.Time) (int64, error) {
	var total int64
	err := r.db.Model(&coupon.GiftCardRedemption{}).
		Where("user_id = ? AND status = ? AND created_at >= ? AND created_at < ?", userID, "completed", start, end).
		Select("COALESCE(SUM(gift_card_value_cents), 0)").
		Scan(&total).Error
	return total, err
}
