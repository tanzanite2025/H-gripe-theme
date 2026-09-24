package repository

import (
	"commerce-platform/internal/domain/coupon"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CouponRepository struct {
	db *gorm.DB
}

var (
	ErrCouponUsageLimitReached = errors.New("coupon usage limit reached")
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
	enabled := c.Enabled
	if err := r.db.Create(c).Error; err != nil {
		return err
	}
	if !enabled {
		if err := r.db.Model(c).Update("enabled", false).Error; err != nil {
			return err
		}
		c.Enabled = false
	}
	return nil
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

func (r *CouponRepository) FindCouponByIDForUpdate(id uint) (*coupon.Coupon, error) {
	var c coupon.Coupon
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CouponRepository) FindCouponByIDIncludingDeleted(id uint) (*coupon.Coupon, error) {
	var c coupon.Coupon
	err := r.db.Unscoped().First(&c, id).Error
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

func (r *CouponRepository) FindCouponByCodeForUpdate(code string) (*coupon.Coupon, error) {
	var c coupon.Coupon
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ?", code).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindAllCoupons 查找所有优惠券
func (r *CouponRepository) FindAllCoupons(page, pageSize int) ([]coupon.Coupon, int64, error) {
	return r.findAllCoupons(page, pageSize, "")
}

// FindAllCouponsByStatus applies status predicates in SQL before counting and
// paginating. Filtering a paginated slice in memory under-reports totals and
// makes later pages inaccessible in the admin UI.
func (r *CouponRepository) FindAllCouponsByStatus(page, pageSize int, status string) ([]coupon.Coupon, int64, error) {
	return r.findAllCoupons(page, pageSize, status)
}

func (r *CouponRepository) findAllCoupons(page, pageSize int, status string) ([]coupon.Coupon, int64, error) {
	var coupons []coupon.Coupon
	var total int64

	query := r.db.Model(&coupon.Coupon{})
	now := time.Now()
	switch status {
	case "", "all":
	case "active":
		query = query.Where("enabled = ? AND start_date < ? AND end_date > ?", true, now, now)
	case "expired":
		query = query.Where("end_date < ?", now)
	case "disabled":
		query = query.Where("enabled = ?", false)
	default:
		return nil, 0, errors.New("unsupported coupon status filter " + status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&coupons).Error

	return coupons, total, err
}

// FindActiveCoupons 查找有效的优惠券
func (r *CouponRepository) FindActiveCoupons() ([]coupon.Coupon, error) {
	var coupons []coupon.Coupon
	now := time.Now()

	err := r.db.Where("enabled = ? AND start_date <= ? AND end_date >= ?", true, now, now).
		Where("referral_recipient_user_id IS NULL").
		Where("used_count < usage_limit OR usage_limit = 0").
		Find(&coupons).Error

	return coupons, err
}

func (r *CouponRepository) FindCouponsForRiskAnalysis(limit int) ([]coupon.Coupon, error) {
	var coupons []coupon.Coupon
	now := time.Now()
	if limit <= 0 {
		limit = 1000
	}

	err := r.db.Where("enabled = ? AND end_date >= ?", true, now).
		Where("used_count < usage_limit OR usage_limit = 0").
		Order("start_date ASC, id ASC").
		Limit(limit).
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
	if u.Status == "" {
		u.Status = coupon.CouponUsageStatusApplied
	}
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

func (r *CouponRepository) ReverseCouponUsageByOrderID(orderID uint, reversedAt time.Time, reason string) error {
	if reversedAt.IsZero() {
		reversedAt = time.Now().UTC()
	}

	return r.db.Model(&coupon.CouponUsage{}).
		Where("order_id = ? AND status = ?", orderID, coupon.CouponUsageStatusApplied).
		Updates(map[string]interface{}{
			"status":          coupon.CouponUsageStatusReversed,
			"reversed_at":     reversedAt,
			"reversal_reason": reason,
		}).Error
}

// CountUserCouponUsage 统计用户使用某优惠券的次数
func (r *CouponRepository) CountUserCouponUsage(userID, couponID uint) (int64, error) {
	var count int64
	err := r.db.Model(&coupon.CouponUsage{}).
		Where("user_id = ? AND coupon_id = ? AND status = ?", userID, couponID, coupon.CouponUsageStatusApplied).
		Count(&count).Error
	return count, err
}

// CountEmailCouponUsage 统计邮箱使用某优惠券的次数
func (r *CouponRepository) CountEmailCouponUsage(email string, couponID uint) (int64, error) {
	var count int64
	err := r.db.Model(&coupon.CouponUsage{}).
		Where("email = ? AND coupon_id = ? AND status = ?", coupon.NormalizeEmail(email), couponID, coupon.CouponUsageStatusApplied).
		Count(&count).Error
	return count, err
}
