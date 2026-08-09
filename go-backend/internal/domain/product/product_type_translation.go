package product

import "time"

type ProductTypeTranslation struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProductTypeID uint      `gorm:"not null;uniqueIndex:idx_product_type_translations_type_locale" json:"product_type_id"`
	Locale        string    `gorm:"size:32;not null;uniqueIndex:idx_product_type_translations_type_locale" json:"locale"`
	Name          string    `gorm:"size:120;not null" json:"name"`
	Description   string    `gorm:"type:text;not null;default:''" json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (ProductTypeTranslation) TableName() string {
	return "product_type_translations"
}
