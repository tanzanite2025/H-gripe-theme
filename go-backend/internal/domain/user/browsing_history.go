package user

import "time"

// BrowsingHistory 用户浏览历史
type BrowsingHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"` // 用户ID
	ProductID uint      `json:"product_id" gorm:"index;not null"` // 产品ID
	ViewCount int       `json:"view_count" gorm:"default:1"` // 浏览次数
	LastViewedAt time.Time `json:"last_viewed_at"` // 最后浏览时间
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// 关联
	// Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

// TableName 指定表名
func (BrowsingHistory) TableName() string {
	return "browsing_history"
}
