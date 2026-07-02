package repository

import (
	"errors"
	"tanzanite/internal/domain/coupon"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CouponRepository struct {
	db *gorm.DB
}

var (
	ErrCouponUsageLimitReached     = errors.New("coupon usage limit reached")
	ErrGiftCardInsufficientBalance = errors.New("insufficient gift card balance")
)

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

// WithTx 复用事务 db 实例
func (r *CouponRepository) WithTx(tx *gorm.DB) *CouponRepository {
	return &CouponRepository{db: tx}
}

// Coupon 相关方法

// CreateCoupon 创建优惠券
func (r *CouponRepository) CreateCoupon(c *coupon.Coupon) error {
	return r.db.Create(c).Error
}

// FindCouponByID 根据ID查找优惠券
func (r *CouponRepository) FindCouponByID(id uint) (*coupon.Coupon, error) {
	var c coupon.Coupon
	err := r.db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindCouponByCode 根据代码查找优惠券
func (r *CouponRepository) FindCouponByCode(code string) (*coupon.Coupon, error) {
	var c coupon.Coupon
	err := r.db.Where("code = ?", code).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindAllCoupons 查找所有优惠券
func (r *CouponRepository) FindAllCoupons(page, pageSize int) ([]coupon.Coupon, int64, error) {
	var coupons []coupon.Coupon
	var total int64

	if err := r.db.Model(&coupon.Coupon{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&coupons).Error

	return coupons, total, err
}

// FindActiveCoupons 查找有效的优惠券
func (r *CouponRepository) FindActiveCoupons() ([]coupon.Coupon, error) {
	var coupons []coupon.Coupon
	now := time.Now()

	err := r.db.Where("enabled = ? AND start_date <= ? AND end_date >= ?", true, now, now).
		Where("used_count < usage_limit OR usage_limit = 0").
		Find(&coupons).Error

	return coupons, err
}

// UpdateCoupon 更新优惠券
func (r *CouponRepository) UpdateCoupon(c *coupon.Coupon) error {
	return r.db.Save(c).Error
}

// IncrementUsedCount 增加使用次数
func (r *CouponRepository) IncrementUsedCount(id uint) error {
	tx := r.db.Model(&coupon.Coupon{}).
		Where("id = ? AND (usage_limit = 0 OR used_count < usage_limit)", id).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", 1))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrCouponUsageLimitReached
	}
	return nil
}

func (r *CouponRepository) DecrementUsedCount(id uint) error {
	return r.db.Model(&coupon.Coupon{}).Where("id = ? AND used_count > 0", id).
		UpdateColumn("used_count", gorm.Expr("used_count - ?", 1)).Error
}

// DeleteCoupon 删除优惠券
func (r *CouponRepository) DeleteCoupon(id uint) error {
	return r.db.Delete(&coupon.Coupon{}, id).Error
}

// CouponUsage 相关方法

// CreateCouponUsage 创建优惠券使用记录
func (r *CouponRepository) CreateCouponUsage(u *coupon.CouponUsage) error {
	return r.db.Create(u).Error
}

// FindCouponUsageByUserAndCoupon 查找用户的优惠券使用记录
func (r *CouponRepository) FindCouponUsageByUserAndCoupon(userID, couponID uint) ([]coupon.CouponUsage, error) {
	var usages []coupon.CouponUsage
	err := r.db.Where("user_id = ? AND coupon_id = ?", userID, couponID).Find(&usages).Error
	return usages, err
}

// FindCouponUsageByOrderID 根据订单ID查找使用记录
func (r *CouponRepository) FindCouponUsageByOrderID(orderID uint) (*coupon.CouponUsage, error) {
	var usage coupon.CouponUsage
	err := r.db.Where("order_id = ?", orderID).First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *CouponRepository) DeleteCouponUsageByOrderID(orderID uint) error {
	return r.db.Where("order_id = ?", orderID).Delete(&coupon.CouponUsage{}).Error
}

// CountUserCouponUsage 统计用户使用某优惠券的次数
func (r *CouponRepository) CountUserCouponUsage(userID, couponID uint) (int64, error) {
	var count int64
	err := r.db.Model(&coupon.CouponUsage{}).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).
		Count(&count).Error
	return count, err
}

// GiftCard 相关方法

// CreateGiftCard 创建礼品卡
func (r *CouponRepository) CreateGiftCard(g *coupon.GiftCard) error {
	return r.db.Create(g).Error
}

// FindGiftCardByID 根据ID查找礼品卡
func (r *CouponRepository) FindGiftCardByID(id uint) (*coupon.GiftCard, error) {
	var g coupon.GiftCard
	err := r.db.First(&g, id).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// FindGiftCardByCode 根据代码查找礼品卡
func (r *CouponRepository) FindGiftCardByCode(code string) (*coupon.GiftCard, error) {
	var g coupon.GiftCard
	err := r.db.Where("code = ?", code).First(&g).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *CouponRepository) FindGiftCardByCodeForUpdate(code string) (*coupon.GiftCard, error) {
	var g coupon.GiftCard
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ?", code).First(&g).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// FindGiftCardsByUserID 查找用户的礼品卡
func (r *CouponRepository) FindGiftCardsByUserID(userID uint) ([]coupon.GiftCard, error) {
	var cards []coupon.GiftCard
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&cards).Error
	return cards, err
}

func (r *CouponRepository) FindAllGiftCards(page, pageSize int, status string) ([]coupon.GiftCard, int64, error) {
	var cards []coupon.GiftCard
	var total int64

	query := r.db.Model(&coupon.GiftCard{})
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&cards).Error

	return cards, total, err
}

// UpdateGiftCard 更新礼品卡
func (r *CouponRepository) UpdateGiftCard(g *coupon.GiftCard) error {
	return r.db.Save(g).Error
}

// UpdateGiftCardBalance 更新礼品卡余额
func (r *CouponRepository) UpdateGiftCardBalance(id uint, amount float64) error {
	query := r.db.Model(&coupon.GiftCard{}).Where("id = ?", id)
	if amount < 0 {
		query = query.Where("balance >= ?", -amount)
	}

	tx := query.UpdateColumn("balance", gorm.Expr("balance + ?", amount))
	if tx.Error != nil {
		return tx.Error
	}
	if amount < 0 && tx.RowsAffected == 0 {
		return ErrGiftCardInsufficientBalance
	}
	return nil
}

// GiftCardTransaction 相关方法

// CreateGiftCardTransaction 创建礼品卡交易记录
func (r *CouponRepository) CreateGiftCardTransaction(t *coupon.GiftCardTransaction) error {
	return r.db.Create(t).Error
}

// FindGiftCardTransactionsByCardID 查找礼品卡的交易记录
func (r *CouponRepository) FindGiftCardTransactionsByCardID(cardID uint) ([]coupon.GiftCardTransaction, error) {
	var transactions []coupon.GiftCardTransaction
	err := r.db.Where("gift_card_id = ?", cardID).Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

// FindGiftCardTransactionByOrderID 根据订单ID查找交易
func (r *CouponRepository) FindGiftCardTransactionByOrderID(orderID uint) ([]coupon.GiftCardTransaction, error) {
	var transactions []coupon.GiftCardTransaction
	err := r.db.Where("order_id = ?", orderID).Find(&transactions).Error
	return transactions, err
}
