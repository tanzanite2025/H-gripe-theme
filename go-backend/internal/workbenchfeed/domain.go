package workbenchfeed

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"

	DirectActionDetailDrawer = "detail_drawer"
	DefaultLocale            = "en"
	MaxContentRunes          = 300
	MaxTags                  = 12
	MaxMedia                 = 4
	MaxTaggedProducts        = 12
)

type FeedEntry struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	EntryNumber    string          `gorm:"size:32;not null" json:"entry_number"`
	PublishedAt    time.Time       `gorm:"not null" json:"published_at"`
	Locale         string          `gorm:"size:16;not null;default:'en';index" json:"locale"`
	Content        string          `gorm:"type:text;not null" json:"content"`
	Tags           datatypes.JSON  `gorm:"type:jsonb;not null;default:'[]'" json:"tags"`
	Status         string          `gorm:"size:16;not null;default:'draft';index" json:"status"`
	CreatedBy      uint            `gorm:"not null" json:"created_by"`
	UpdatedBy      uint            `gorm:"not null" json:"updated_by"`
	Media          []FeedMedia     `gorm:"foreignKey:EntryID" json:"media"`
	TaggedProducts []TaggedProduct `gorm:"foreignKey:EntryID" json:"tagged_products"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (FeedEntry) TableName() string {
	return "workbench_feed_entries"
}

type FeedMedia struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	EntryID    uint           `gorm:"not null;index" json:"entry_id"`
	FilePath   string         `gorm:"type:text;not null" json:"file_path"`
	Width      int            `gorm:"not null" json:"width"`
	Height     int            `gorm:"not null" json:"height"`
	FileSizeKB int            `gorm:"not null;default:0" json:"file_size_kb"`
	Caption    string         `gorm:"type:text;not null;default:''" json:"caption"`
	SortOrder  int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FeedMedia) TableName() string {
	return "workbench_feed_media"
}

type TaggedProduct struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	EntryID      uint           `gorm:"not null;index" json:"entry_id"`
	ProductID    uint           `gorm:"not null;index" json:"product_id"`
	VariantID    *uint          `gorm:"index" json:"variant_id,omitempty"`
	ProductSlug  string         `gorm:"size:255;not null;default:''" json:"product_slug"`
	DisplayTitle string         `gorm:"size:255;not null" json:"display_title"`
	Price        float64        `gorm:"type:numeric(12,2);not null;default:0" json:"price"`
	Currency     string         `gorm:"size:3;not null;default:'USD'" json:"currency"`
	DirectAction string         `gorm:"size:32;not null;default:'detail_drawer'" json:"direct_action"`
	Available    bool           `gorm:"not null;default:true" json:"available"`
	SortOrder    int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TaggedProduct) TableName() string {
	return "workbench_feed_tagged_products"
}

type ProductCatalogOption struct {
	ProductID    uint    `json:"product_id"`
	VariantID    *uint   `json:"variant_id,omitempty"`
	ProductName  string  `json:"product_name"`
	ProductSlug  string  `json:"product_slug"`
	VariantTitle string  `json:"variant_title"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	Available    bool    `json:"available"`
}

type ListInput struct {
	Page     int
	PageSize int
	Locale   string
	Tag      string
	Admin    bool
	Status   string
	Search   string
}

type MediaInput struct {
	FilePath   string
	Width      int
	Height     int
	FileSizeKB int
	Caption    string
	SortOrder  int
}

type TaggedProductInput struct {
	ProductID    uint
	VariantID    *uint
	DisplayTitle string
	Price        float64
	Currency     string
	DirectAction string
	Available    bool
	SortOrder    int
}

type EntryInput struct {
	PublishedAt    *time.Time
	Locale         string
	Content        string
	Tags           []string
	Status         string
	Media          []MediaInput
	TaggedProducts []TaggedProductInput
	ActorID        uint
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ListResult struct {
	Entries    []FeedEntry `json:"entries"`
	Pagination Pagination  `json:"pagination"`
}
