package product

import "time"

type ProductType struct {
	ID                uint                     `gorm:"primarykey" json:"id"`
	Name              string                   `gorm:"type:varchar(120);not null" json:"name"`
	Slug              string                   `gorm:"type:varchar(120);uniqueIndex;not null" json:"slug"`
	Description       string                   `gorm:"type:text" json:"description"`
	ImageMediaAssetID *uint                    `gorm:"index" json:"image_media_asset_id,omitempty"`
	ImageURL          string                   `gorm:"type:text;not null;default:''" json:"image_url,omitempty"`
	SortOrder         int                      `gorm:"default:0;not null" json:"sort_order"`
	IsEnabled         bool                     `gorm:"default:true;not null" json:"is_enabled"`
	IsSystemManaged   bool                     `gorm:"default:false;not null;index" json:"is_system_managed"`
	Translations      []ProductTypeTranslation `gorm:"foreignKey:ProductTypeID;constraint:OnDelete:CASCADE" json:"translations,omitempty"`
	SpecDefinitions   []SpecDefinition         `gorm:"foreignKey:ProductTypeID" json:"spec_definitions,omitempty"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}

func (ProductType) TableName() string {
	return "product_types"
}

func (p ProductType) NameForLocale(locale string) string {
	for _, translation := range p.Translations {
		if translation.Locale == locale && translation.Name != "" {
			return translation.Name
		}
	}

	if locale != "en" {
		for _, translation := range p.Translations {
			if translation.Locale == "en" && translation.Name != "" {
				return translation.Name
			}
		}
	}

	return p.Name
}
