package product

import "time"

type ProductCategoryTranslation struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ProductCategoryID uint      `gorm:"not null;uniqueIndex:idx_product_category_translations_category_locale" json:"product_category_id"`
	Locale            string    `gorm:"size:32;not null;uniqueIndex:idx_product_category_translations_category_locale" json:"locale"`
	Name              string    `gorm:"size:120;not null" json:"name"`
	Description       string    `gorm:"type:text;not null;default:''" json:"description"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (ProductCategoryTranslation) TableName() string {
	return "product_category_translations"
}
