package post

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Title    string `gorm:"not null;index" json:"title"`
	Slug     string `gorm:"uniqueIndex:idx_slug_locale;not null" json:"slug"`
	Content  string `gorm:"type:text" json:"content"`
	Excerpt  string `gorm:"type:text" json:"excerpt"`
	Status   string `gorm:"default:'draft';index" json:"status"` // draft, published, archived
	AuthorID uint   `gorm:"not null;index" json:"author_id"`
	Locale   string `gorm:"uniqueIndex:idx_slug_locale;default:'en';index" json:"locale"`

	// 翻译关联 (增强)
	TranslationGroupID *uint `gorm:"index" json:"translation_group_id"` // 翻译组ID，同一组的文章是不同语言版本
	ParentID           *uint `gorm:"index" json:"parent_id"`            // 已废弃，保留用于向后兼容

	// 媒体
	FeaturedImg string `json:"featured_image"`

	// 统计
	ViewCount int `gorm:"default:0" json:"view_count"`

	// SEO 元数据
	MetaTitle    string `json:"meta_title"`
	MetaDesc     string `gorm:"type:text" json:"meta_description"`
	MetaKeywords string `json:"meta_keywords"`
	CanonicalURL string `json:"canonical_url"`

	// 标签和分类
	Tags string `json:"tags"` // 逗号分隔的标签

	// 时间戳
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	PublishedAt *time.Time     `json:"published_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联 (不存储在数据库)
	Translations []Post `gorm:"-" json:"translations,omitempty"` // 其他语言版本
}

// TableName 指定表名
func (Post) TableName() string {
	return "posts"
}

// BeforeCreate GORM钩子：创建前
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.Locale == "" {
		p.Locale = "en"
	}
	if p.Status == "" {
		p.Status = "draft"
	}
	return nil
}
