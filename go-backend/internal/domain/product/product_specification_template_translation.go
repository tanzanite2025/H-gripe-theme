package product

import "time"

type ProductSpecificationTemplateTranslation struct {
	ID                             uint      `gorm:"primaryKey" json:"id"`
	ProductSpecificationTemplateID uint      `gorm:"not null;uniqueIndex:idx_product_specification_template_translations_type_locale" json:"product_specification_template_id"`
	Locale                         string    `gorm:"size:32;not null;uniqueIndex:idx_product_specification_template_translations_type_locale" json:"locale"`
	Name                           string    `gorm:"size:120;not null" json:"name"`
	Description                    string    `gorm:"type:text;not null;default:''" json:"description"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
}

func (ProductSpecificationTemplateTranslation) TableName() string {
	return "product_specification_template_translations"
}
