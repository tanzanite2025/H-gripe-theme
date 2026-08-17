package product

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProductMedia stores product-owned media placement.
// Galleries are intentionally separate showcase collections and should not be
// used as the source of truth for product images or videos.
type ProductMedia struct {
	ID                   uint           `gorm:"primarykey" json:"id"`
	ProductID            uint           `gorm:"not null;index" json:"product_id"`
	VariantID            *uint          `gorm:"index" json:"variant_id,omitempty"`
	VariantOptionValueID *uint          `gorm:"index" json:"variant_option_value_id,omitempty"`
	MediaAssetID         *uint          `gorm:"index" json:"media_asset_id,omitempty"`
	MediaType            string         `gorm:"default:'image';not null;index" json:"media_type"`
	Role                 string         `gorm:"default:'gallery';not null;index" json:"role"`
	URL                  string         `gorm:"not null" json:"url"`
	ThumbnailURL         string         `json:"thumbnail_url"`
	PosterURL            string         `json:"poster_url"`
	ImageVariantData     datatypes.JSON `gorm:"column:image_variants;type:jsonb;not null;default:'{}'" json:"image_variants,omitempty"`
	Alt                  string         `json:"alt"`
	Title                string         `json:"title"`
	Locale               string         `gorm:"index" json:"locale"`
	SortOrder            int            `gorm:"default:0;not null" json:"sort_order"`
	IsPrimary            bool           `gorm:"default:false;not null" json:"is_primary"`
	IsVisible            bool           `gorm:"default:true;not null" json:"is_visible"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductMedia) TableName() string {
	return "product_media"
}

func (m *ProductMedia) BeforeCreate(tx *gorm.DB) error {
	m.ensureImageVariantData()
	return nil
}

func (m *ProductMedia) BeforeSave(tx *gorm.DB) error {
	m.ensureImageVariantData()
	return nil
}

func (m *ProductMedia) ensureImageVariantData() {
	if len(m.ImageVariantData) == 0 {
		m.ImageVariantData = datatypes.JSON([]byte("{}"))
	}
}

type ProductMediaImageVariant struct {
	URL      string `json:"url"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

func ProductMediaImageVariantsJSON(values map[string]ProductMediaImageVariant) datatypes.JSON {
	if len(values) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(encoded)
}

func ParseProductMediaImageVariants(raw datatypes.JSON) map[string]ProductMediaImageVariant {
	if len(raw) == 0 {
		return map[string]ProductMediaImageVariant{}
	}
	var values map[string]ProductMediaImageVariant
	if err := json.Unmarshal(raw, &values); err != nil || len(values) == 0 {
		return map[string]ProductMediaImageVariant{}
	}

	normalized := make(map[string]ProductMediaImageVariant, len(values))
	for preset, item := range values {
		if preset == "" || item.URL == "" {
			continue
		}
		normalized[preset] = item
	}
	return normalized
}
