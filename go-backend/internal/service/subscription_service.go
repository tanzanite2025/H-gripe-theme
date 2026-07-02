package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"tanzanite/internal/domain/subscription"
	"tanzanite/internal/repository"
	"time"
)

type SubscriptionService struct {
	subscriptionRepo *repository.SubscriptionRepository
}

var ErrInvalidSubscriptionStatus = errors.New("invalid subscription status")

func NewSubscriptionService(subscriptionRepo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo: subscriptionRepo,
	}
}

// Subscribe 订阅
func (s *SubscriptionService) Subscribe(email, source, locale string, tags []string) (*subscription.Subscription, error) {
	// 检查是否已订阅
	exists, err := s.subscriptionRepo.CheckEmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already subscribed")
	}

	// 生成退订令牌
	token, err := generateUnsubToken()
	if err != nil {
		return nil, err
	}

	// 创建订阅
	sub := &subscription.Subscription{
		Email:        email,
		Status:       "active",
		Locale:       locale,
		Source:       source,
		Tags:         joinTags(tags),
		UnsubToken:   token,
		SubscribedAt: time.Now(),
	}

	if err := s.subscriptionRepo.Create(sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// Unsubscribe 退订
func (s *SubscriptionService) Unsubscribe(token string) error {
	// 查找订阅
	sub, err := s.subscriptionRepo.FindByToken(token)
	if err != nil {
		return errors.New("invalid unsubscribe token")
	}

	// 更新状态
	return s.subscriptionRepo.UpdateStatus(sub.Email, "unsubscribed")
}

// UnsubscribeByEmail 通过邮箱退订
func (s *SubscriptionService) UnsubscribeByEmail(email string) error {
	return s.subscriptionRepo.UpdateStatus(email, "unsubscribed")
}

func (s *SubscriptionService) UpdateStatus(email, status string) error {
	switch status {
	case "active":
		return s.Resubscribe(email)
	case "unsubscribed":
		return s.UnsubscribeByEmail(email)
	default:
		return ErrInvalidSubscriptionStatus
	}
}

// Resubscribe 重新订阅
func (s *SubscriptionService) Resubscribe(email string) error {
	// 查找订阅
	sub, err := s.subscriptionRepo.FindByEmail(email)
	if err != nil {
		return errors.New("subscription not found")
	}

	// 更新状态
	sub.Status = "active"
	now := time.Now()
	sub.SubscribedAt = now
	sub.UnsubscribedAt = nil

	return s.subscriptionRepo.Update(sub)
}

// GetSubscription 获取订阅
func (s *SubscriptionService) GetSubscription(email string) (*subscription.Subscription, error) {
	return s.subscriptionRepo.FindByEmail(email)
}

// GetAllSubscriptions 获取所有订阅（管理员）
func (s *SubscriptionService) GetAllSubscriptions(page, pageSize int, status string) ([]subscription.Subscription, int64, error) {
	return s.subscriptionRepo.FindAll(page, pageSize, status)
}

// GetSubscriptionsByTags 根据标签获取订阅
func (s *SubscriptionService) GetSubscriptionsByTags(tags []string, page, pageSize int) ([]subscription.Subscription, int64, error) {
	return s.subscriptionRepo.FindByTags(tags, page, pageSize)
}

// UpdateSubscription 更新订阅
func (s *SubscriptionService) UpdateSubscription(sub *subscription.Subscription) error {
	return s.subscriptionRepo.Update(sub)
}

// DeleteSubscription 删除订阅
func (s *SubscriptionService) DeleteSubscription(email string) error {
	return s.subscriptionRepo.Delete(email)
}

func (s *SubscriptionService) BatchDelete(emails []string) (int, error) {
	deleted := 0
	for _, email := range emails {
		if err := s.DeleteSubscription(email); err == nil {
			deleted++
		}
	}
	return deleted, nil
}

// GetStats 获取订阅统计
func (s *SubscriptionService) GetStats() (map[string]interface{}, error) {
	return s.subscriptionRepo.GetStats()
}

// GetActiveEmails 获取活跃订阅邮箱（用于群发）
func (s *SubscriptionService) GetActiveEmails() ([]string, error) {
	return s.subscriptionRepo.GetActiveEmails()
}

// GetActiveEmailsByTags 根据标签获取活跃订阅邮箱
func (s *SubscriptionService) GetActiveEmailsByTags(tags []string) ([]string, error) {
	subscriptions, _, err := s.subscriptionRepo.FindByTags(tags, 1, 10000)
	if err != nil {
		return nil, err
	}

	emails := make([]string, len(subscriptions))
	for i, sub := range subscriptions {
		emails[i] = sub.Email
	}

	return emails, nil
}

// generateUnsubToken 生成退订令牌
func generateUnsubToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// joinTags 连接标签
func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ","
		}
		result += tag
	}
	return result
}
