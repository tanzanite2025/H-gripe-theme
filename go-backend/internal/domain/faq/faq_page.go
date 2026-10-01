package faq

import (
	"time"

	"gorm.io/gorm"
)

type FAQPage struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	PageID          string         `gorm:"size:120;not null;uniqueIndex:idx_faq_pages_page_locale" json:"page_id"`
	RoutePath       string         `gorm:"size:255;not null;default:''" json:"route_path"`
	RouteKey        string         `gorm:"size:255;index" json:"route_key"`
	ManifestVersion string         `gorm:"size:64;index" json:"manifest_version"`
	RouteStatus     string         `gorm:"size:20;not null;default:'current';index" json:"route_status"` // current, stale, missing, alias
	LastSeenAt      *time.Time     `json:"last_seen_at,omitempty"`
	Domain          string         `gorm:"size:80;not null;default:'';index" json:"domain"`
	Locale          string         `gorm:"size:10;not null;default:'en';uniqueIndex:idx_faq_pages_page_locale;index" json:"locale"`
	Title           string         `gorm:"size:255;not null;default:''" json:"title"`
	Subtitle        string         `gorm:"type:text" json:"subtitle"`
	SortOrder       int            `gorm:"not null;default:0;index" json:"sort_order"`
	Status          string         `gorm:"size:20;not null;default:'active';index" json:"status"` // active, hidden
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// FAQRouteSyncStats describes the non-destructive route metadata reconciliation
// performed from the storefront route manifest. FAQ answers are never changed
// by this operation.
type FAQRouteSyncStats struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Stale   int `json:"stale"`
	Total   int `json:"total"`
}

func (FAQPage) TableName() string {
	return "faq_pages"
}
