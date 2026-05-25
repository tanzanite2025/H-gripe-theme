package repository

import (
	"tanzanite/internal/domain/subscription"
	"time"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create 创建订阅
func (r *SubscriptionRepository) Create(sub *subscription.Subscription) error {
	return r.db.Create(sub).Error
}

// FindByEmail 根据邮箱查找订阅
func (r *SubscriptionRepository) FindByEmail(email string) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	err := r.db.Where("email = ?", email).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// FindByToken 根据退订令牌查找订阅
func (r *SubscriptionRepository) FindByToken(token string) (*subscription.Subscription, error) {
	var sub subscription.Subscription
	err := r.db.Where("unsub_token = ?", token).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// FindAll 查找所有订阅
func (r *SubscriptionRepository) FindAll(page, pageSize int, status string) ([]subscription.Subscription, int64, error) {
	var subscriptions []subscription.Subscription
	var total int64

	query := r.db.Model(&subscription.Subscription{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&subscriptions).Error
	
	return subscriptions, total, err
}

// FindByTags 根据标签查找订阅
func (r *SubscriptionRepository) FindByTags(tags []string, page, pageSize int) ([]subscription.Subscription, int64, error) {
	var subscriptions []subscription.Subscription
	var total int64

	query := r.db.Model(&subscription.Subscription{}).Where("status = ?", "active")
	
	// 查找包含任一标签的订阅
	for i, tag := range tags {
		if i == 0 {
			query = query.Where("tags LIKE ?", "%"+tag+"%")
		} else {
			query = query.Or("tags LIKE ?", "%"+tag+"%")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&subscriptions).Error
	
	return subscriptions, total, err
}

// Update 更新订阅
func (r *SubscriptionRepository) Update(sub *subscription.Subscription) error {
	return r.db.Save(sub).Error
}

// UpdateStatus 更新订阅状态
func (r *SubscriptionRepository) UpdateStatus(email, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	
	if status == "unsubscribed" {
		now := time.Now()
		updates["unsubscribed_at"] = &now
	}
	
	return r.db.Model(&subscription.Subscription{}).Where("email = ?", email).Updates(updates).Error
}

// Delete 删除订阅
func (r *SubscriptionRepository) Delete(email string) error {
	return r.db.Where("email = ?", email).Delete(&subscription.Subscription{}).Error
}

// CheckEmailExists 检查邮箱是否已订阅
func (r *SubscriptionRepository) CheckEmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&subscription.Subscription{}).
		Where("email = ? AND status = ?", email, "active").Count(&count).Error
	return count > 0, err
}

// GetStats 获取订阅统计
func (r *SubscriptionRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 总订阅数
	var totalCount int64
	if err := r.db.Model(&subscription.Subscription{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount
	
	// 活跃订阅数
	var activeCount int64
	if err := r.db.Model(&subscription.Subscription{}).
		Where("status = ?", "active").Count(&activeCount).Error; err != nil {
		return nil, err
	}
	stats["active_count"] = activeCount
	
	// 已退订数
	var unsubscribedCount int64
	if err := r.db.Model(&subscription.Subscription{}).
		Where("status = ?", "unsubscribed").Count(&unsubscribedCount).Error; err != nil {
		return nil, err
	}
	stats["unsubscribed_count"] = unsubscribedCount
	
	// 本月新增
	var monthlyCount int64
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	if err := r.db.Model(&subscription.Subscription{}).
		Where("created_at >= ?", startOfMonth).Count(&monthlyCount).Error; err != nil {
		return nil, err
	}
	stats["monthly_count"] = monthlyCount
	
	return stats, nil
}

// GetActiveEmails 获取所有活跃订阅的邮箱
func (r *SubscriptionRepository) GetActiveEmails() ([]string, error) {
	var emails []string
	err := r.db.Model(&subscription.Subscription{}).
		Where("status = ?", "active").
		Pluck("email", &emails).Error
	return emails, err
}
