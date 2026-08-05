package loyalty

import (
	"time"

	"gorm.io/gorm"
)

// MemberLevel 会员等级
type MemberLevel struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Name         string         `gorm:"not null" json:"name"`
	MinPoints    int            `gorm:"not null" json:"min_points"`
	MaxPoints    int            `gorm:"not null" json:"max_points"`
	DiscountRate float64        `gorm:"default:0" json:"discount_rate"` // 折扣率，如 5 表示 95折
	Benefits     string         `gorm:"type:text" json:"benefits"`      // JSON格式的权益说明
	Icon         string         `json:"icon"`
	Color        string         `json:"color"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MemberLevel) TableName() string {
	return "member_levels"
}
