package review

import "time"

// ReviewHelpful 评价有用标记
type ReviewHelpful struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ReviewID  uint      `gorm:"not null;index" json:"review_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Helpful   bool      `gorm:"not null" json:"helpful"` // true=有用, false=无用
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (ReviewHelpful) TableName() string {
	return "review_helpful"
}
