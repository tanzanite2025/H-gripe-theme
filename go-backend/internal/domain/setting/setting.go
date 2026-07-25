package setting

import (
	"time"

	"gorm.io/gorm"
)

type Setting struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Key         string         `gorm:"uniqueIndex:idx_setting_key_locale;not null" json:"key"`
	Value       string         `gorm:"type:text" json:"value"`
	Type        string         `gorm:"default:'string'" json:"type"` // string, json, boolean, number
	Locale      string         `gorm:"uniqueIndex:idx_setting_key_locale;default:'en'" json:"locale"`
	Group       string         `gorm:"index" json:"group"`            // site, email, seo, social, quick-buy, etc.
	IsPublic    bool           `gorm:"default:true" json:"is_public"` // 是否公开给前端（非敏感信息）
	Description string         `gorm:"type:text" json:"description"`  // 设置说明
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}
