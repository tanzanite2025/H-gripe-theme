package post

import (
	"time"

	"gorm.io/gorm"
)

// Category 分类模型
type Category struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Slug        string         `gorm:"uniqueIndex:idx_category_slug_locale;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	Locale      string         `gorm:"uniqueIndex:idx_category_slug_locale;default:'en'" json:"locale"`
	ParentID    *uint          `json:"parent_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}

// PostCategory 文章分类关联表
type PostCategory struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	PostID     uint      `gorm:"not null;index" json:"post_id"`
	CategoryID uint      `gorm:"not null;index" json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (PostCategory) TableName() string {
	return "post_categories"
}
