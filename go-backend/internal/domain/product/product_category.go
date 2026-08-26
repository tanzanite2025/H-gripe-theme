package product

import "time"

type ProductCategory struct {
	ID                uint                         `gorm:"primarykey" json:"id"`
	ParentID          *uint                        `gorm:"index" json:"parent_id"`
	Parent            *ProductCategory             `gorm:"foreignKey:ParentID;constraint:OnDelete:RESTRICT" json:"parent,omitempty"`
	Children          []ProductCategory            `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Translations      []ProductCategoryTranslation `gorm:"foreignKey:ProductCategoryID" json:"-"`
	Name              string                       `gorm:"type:varchar(120);not null" json:"name"`
	Slug              string                       `gorm:"type:varchar(120);uniqueIndex;not null" json:"slug"`
	Description       string                       `gorm:"type:text" json:"description"`
	MetaTitle         string                       `gorm:"type:varchar(160);not null;default:''" json:"meta_title"`
	MetaDesc          string                       `gorm:"column:meta_description;type:text;not null;default:''" json:"meta_description"`
	SEOIntro          string                       `gorm:"column:seo_intro;type:text;not null;default:''" json:"intro"`
	ImageMediaAssetID *uint                        `gorm:"index" json:"image_media_asset_id,omitempty"`
	ImageURL          string                       `gorm:"type:text;not null;default:''" json:"image_url,omitempty"`
	Depth             int                          `gorm:"not null;default:1;index" json:"depth"`
	SortOrder         int                          `gorm:"not null;default:0" json:"sort_order"`
	IsEnabled         bool                         `gorm:"not null;default:true;index" json:"is_enabled"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

func (ProductCategory) TableName() string {
	return "product_categories"
}
