package loyalty

import (
	"time"

	"gorm.io/gorm"
)

// LoyaltyTransaction 积分交易记录
type LoyaltyTransaction struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	UserID          uint           `gorm:"not null;index" json:"user_id"`
	Type            string         `gorm:"not null;index" json:"type"` // earn, spend, expire, adjust
	Points          int            `gorm:"not null" json:"points"`     // 正数为获得，负数为消费
	Balance         int            `gorm:"not null" json:"balance"`    // 交易后余额
	Source          string         `gorm:"index" json:"source"`        // order, checkin, referral, admin, expire
	SourceID        uint           `json:"source_id"`                  // 关联ID（订单ID、推荐ID等）
	ProgramConfigID *uint          `gorm:"index" json:"program_config_id,omitempty"`
	Description     string         `gorm:"type:text" json:"description"`
	ExpiresAt       *time.Time     `json:"expires_at"` // 积分过期时间
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (LoyaltyTransaction) TableName() string {
	return "loyalty_transactions"
}
