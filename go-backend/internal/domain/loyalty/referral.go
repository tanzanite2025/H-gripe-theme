package loyalty

import (
	"time"

	"gorm.io/gorm"
)

// Referral 推荐记录
type Referral struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ReferrerID   uint           `gorm:"not null;index" json:"referrer_id"` // 推荐人
	ReferredID   uint           `gorm:"index" json:"referred_id"`          // 被推荐人
	ReferralCode string         `gorm:"uniqueIndex" json:"referral_code"`
	Status       string         `gorm:"index" json:"status"` // pending, completed, expired
	PointsEarned int            `gorm:"default:0" json:"points_earned"`
	CompletedAt  *time.Time     `json:"completed_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Referral) TableName() string {
	return "referrals"
}
