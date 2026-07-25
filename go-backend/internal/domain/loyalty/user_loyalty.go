package loyalty

import "time"

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
