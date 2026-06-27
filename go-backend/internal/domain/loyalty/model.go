package loyalty

import (
	"time"

	"gorm.io/gorm"
)

// LoyaltyTransaction 积分交易记录
type LoyaltyTransaction struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Type        string         `gorm:"not null;index" json:"type"` // earn, spend, expire, adjust
	Points      int            `gorm:"not null" json:"points"`     // 正数为获得，负数为消费
	Balance     int            `gorm:"not null" json:"balance"`    // 交易后余额
	Source      string         `gorm:"index" json:"source"`        // order, checkin, referral, admin, expire
	SourceID    uint           `json:"source_id"`                  // 关联ID（订单ID、推荐ID等）
	Description string         `gorm:"type:text" json:"description"`
	ExpiresAt   *time.Time     `json:"expires_at"` // 积分过期时间
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (LoyaltyTransaction) TableName() string {
	return "loyalty_transactions"
}

// CheckIn 签到记录
type CheckIn struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	UserID          uint      `gorm:"not null;index" json:"user_id"`
	CheckInDate     string    `gorm:"not null;index" json:"check_in_date"` // YYYY-MM-DD
	PointsEarned    int       `gorm:"not null" json:"points_earned"`
	ConsecutiveDays int       `gorm:"default:1" json:"consecutive_days"`
	CreatedAt       time.Time `json:"created_at"`
}

// TableName 指定表名
func (CheckIn) TableName() string {
	return "check_ins"
}

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

// MemberLevel 会员等级
type MemberLevel struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	Name             string         `gorm:"not null" json:"name"`
	MinPoints        int            `gorm:"not null" json:"min_points"`
	MaxPoints        int            `gorm:"not null" json:"max_points"`
	DiscountRate     float64        `gorm:"default:0" json:"discount_rate"`     // 折扣率，如 5 表示 95折
	PointsMultiplier float64        `gorm:"default:1" json:"points_multiplier"` // 积分倍数
	Benefits         string         `gorm:"type:text" json:"benefits"`          // JSON格式的权益说明
	Icon             string         `json:"icon"`
	Color            string         `json:"color"`
	SortOrder        int            `gorm:"default:0" json:"sort_order"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MemberLevel) TableName() string {
	return "member_levels"
}

// UserLoyalty 用户积分汇总
type UserLoyalty struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	UserID          uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	TotalPoints     int       `gorm:"default:0;check:total_points_non_negative,total_points >= 0" json:"total_points"`             // 累计获得积分
	AvailablePoints int       `gorm:"default:0;check:available_points_non_negative,available_points >= 0" json:"available_points"` // 可用积分
	UsedPoints      int       `gorm:"default:0;check:used_points_non_negative,used_points >= 0" json:"used_points"`                // 已使用积分
	ExpiredPoints   int       `gorm:"default:0;check:expired_points_non_negative,expired_points >= 0" json:"expired_points"`       // 已过期积分
	MemberLevelID   uint      `json:"member_level_id"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (UserLoyalty) TableName() string {
	return "user_loyalty"
}
