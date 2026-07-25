package loyalty

import "time"

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
