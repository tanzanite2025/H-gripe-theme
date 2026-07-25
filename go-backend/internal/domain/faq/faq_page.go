package faq

import (
	"time"

	"gorm.io/gorm"
)

type FAQPage struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	PageID    string         `gorm:"size:120;not null;uniqueIndex:idx_faq_pages_page_locale" json:"page_id"`
	RoutePath string         `gorm:"size:255;not null;default:''" json:"route_path"`
	Domain    string         `gorm:"size:80;not null;default:'';index" json:"domain"`
	Locale    string         `gorm:"size:10;not null;default:'en';uniqueIndex:idx_faq_pages_page_locale;index" json:"locale"`
	Title     string         `gorm:"size:255;not null;default:''" json:"title"`
	Subtitle  string         `gorm:"type:text" json:"subtitle"`
	SortOrder int            `gorm:"not null;default:0;index" json:"sort_order"`
	Status    string         `gorm:"size:20;not null;default:'active';index" json:"status"` // active, hidden
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FAQPage) TableName() string {
	return "faq_pages"
}

type FAQCategory struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	PageID      string         `gorm:"size:120;not null;uniqueIndex:idx_faq_categories_page_key_locale;index" json:"page_id"`
	CategoryKey string         `gorm:"size:120;not null;uniqueIndex:idx_faq_categories_page_key_locale" json:"category_key"`
	Name        string         `gorm:"size:180;not null;default:''" json:"name"`
	Icon        string         `gorm:"size:40;not null;default:''" json:"icon"`
	Locale      string         `gorm:"size:10;not null;default:'en';uniqueIndex:idx_faq_categories_page_key_locale;index" json:"locale"`
	SortOrder   int            `gorm:"not null;default:0;index" json:"sort_order"`
	Status      string         `gorm:"size:20;not null;default:'active';index" json:"status"` // active, hidden
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FAQCategory) TableName() string {
	return "faq_categories"
}
