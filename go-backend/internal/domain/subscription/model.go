package subscription

import (
	"time"

	"gorm.io/gorm"
)

type Subscription struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	Email          string         `gorm:"uniqueIndex;not null" json:"email"`
	Status         string         `gorm:"default:'active';index" json:"status"` // active, unsubscribed
	Locale         string         `gorm:"default:'en'" json:"locale"`
	Source         string         `json:"source"` // website, popup, checkout
	Tags           string         `json:"tags"` // 订阅标签（逗号分隔）
	UnsubToken     string         `gorm:"uniqueIndex" json:"unsub_token"` // 退订令牌
	SubscribedAt   time.Time      `json:"subscribed_at"`
	UnsubscribedAt *time.Time     `json:"unsubscribed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Subscription) TableName() string {
	return "subscriptions"
}
