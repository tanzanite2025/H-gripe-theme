package homevisualtile

import (
	"time"

	"gorm.io/gorm"
)

// Tile is one operator-configured homepage visual tile.
type Tile struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	TileSetKey      string         `gorm:"column:showcase_key;not null;index:idx_visual_showcase_scope,priority:1" json:"showcase_key"`
	Locale          string         `gorm:"not null;index:idx_visual_showcase_scope,priority:2" json:"locale"`
	ImageURL        string         `gorm:"not null" json:"image_url"`
	ThumbnailURL    string         `json:"thumbnail_url"`
	StorageKey      string         `gorm:"not null;default:'';index" json:"storage_key"`
	Title           string         `gorm:"not null" json:"title"`
	Caption         string         `gorm:"type:text" json:"caption"`
	AltText         string         `gorm:"not null" json:"alt_text"`
	DesktopOrder    int            `gorm:"not null;default:0" json:"desktop_order"`
	MobilePairIndex int            `gorm:"not null;default:0" json:"mobile_pair_index"`
	TargetURL       string         `json:"target_url"`
	TargetLabel     string         `json:"target_label"`
	LayoutVariant   string         `gorm:"not null;default:'standard'" json:"layout_variant"`
	IsPublished     bool           `gorm:"not null;default:true" json:"is_published"`
	PublishedFrom   *time.Time     `json:"published_from,omitempty"`
	PublishedUntil  *time.Time     `json:"published_until,omitempty"`
	Width           int            `gorm:"not null;default:900" json:"width"`
	Height          int            `gorm:"not null;default:600" json:"height"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Tile) TableName() string {
	return "visual_showcase_items"
}
