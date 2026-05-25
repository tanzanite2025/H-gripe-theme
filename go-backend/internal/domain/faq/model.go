package faq

import (
	"time"

	"gorm.io/gorm"
)

type FAQ struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Question  string         `gorm:"not null;type:text" json:"question"`
	Answer    string         `gorm:"not null;type:text" json:"answer"`
	Category  string         `gorm:"index" json:"category"`
	Locale    string         `gorm:"default:'en';index" json:"locale"`
	ParentID  *uint          `gorm:"index" json:"parent_id"` // 翻译关联
	Order     int            `gorm:"default:0" json:"order"`
	Status    string         `gorm:"default:'published'" json:"status"` // published, draft
	ViewCount int            `gorm:"default:0" json:"view_count"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (FAQ) TableName() string {
	return "faqs"
}
