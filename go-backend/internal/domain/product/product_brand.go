package product

import "time"

// ProductBrand is a global product master-data record shared by every
// product template, including Rim, Frame, Wheelset, Handlebar, Hub, and Spoke.
type ProductBrand struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:160;not null" json:"name"`
	Slug        string    `gorm:"size:160;uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"type:text;not null;default:''" json:"description"`
	LogoURL     string    `gorm:"type:text;not null;default:''" json:"logo_url,omitempty"`
	WebsiteURL  string    `gorm:"type:text;not null;default:''" json:"website_url,omitempty"`
	IsEnabled   bool      `gorm:"not null;default:true;index" json:"is_enabled"`
	SortOrder   int       `gorm:"not null;default:0;index" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ProductBrand) TableName() string {
	return "product_brands"
}
