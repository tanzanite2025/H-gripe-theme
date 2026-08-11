package merchant

import (
	"time"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

// GoogleMerchantOffer is a channel-specific SKU record. It never mutates the
// storefront product or variant that it references.
type GoogleMerchantOffer struct {
	ID                    uint                    `gorm:"primarykey" json:"id"`
	ProductID             uint                    `gorm:"not null;index" json:"product_id"`
	VariantID             uint                    `gorm:"not null;uniqueIndex" json:"variant_id"`
	OfferID               string                  `gorm:"type:varchar(160);uniqueIndex;not null" json:"offer_id"`
	Title                 string                  `gorm:"type:varchar(150)" json:"title"`
	Description           string                  `gorm:"type:text" json:"description"`
	Brand                 string                  `gorm:"type:varchar(140)" json:"brand"`
	Condition             string                  `gorm:"type:varchar(24);not null;default:'new'" json:"condition"`
	GoogleProductCategory string                  `gorm:"type:text" json:"google_product_category"`
	GTIN                  string                  `gorm:"type:varchar(20)" json:"gtin"`
	MPN                   string                  `gorm:"type:varchar(70)" json:"mpn"`
	IdentifierExists      *bool                   `json:"identifier_exists"`
	TargetCountry         string                  `gorm:"type:varchar(2)" json:"target_country"`
	ContentLanguage       string                  `gorm:"type:varchar(8)" json:"content_language"`
	CurrencyCode          string                  `gorm:"type:varchar(3)" json:"currency_code"`
	FeedLabel             string                  `gorm:"type:varchar(64)" json:"feed_label"`
	PriceOverride         *float64                `gorm:"type:decimal(12,2)" json:"price_override"`
	SalePriceOverride     *float64                `gorm:"type:decimal(12,2)" json:"sale_price_override"`
	PublicationStatus     string                  `gorm:"type:varchar(24);not null;default:'draft';index" json:"publication_status"`
	SyncStatus            string                  `gorm:"type:varchar(24);not null;default:'not_synced';index" json:"sync_status"`
	LastValidatedAt       *time.Time              `json:"last_validated_at"`
	LastSyncAt            *time.Time              `json:"last_sync_at"`
	LastError             string                  `gorm:"type:text" json:"last_error"`
	Product               *product.Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Variant               *product.ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	CreatedAt             time.Time               `json:"created_at"`
	UpdatedAt             time.Time               `json:"updated_at"`
	DeletedAt             gorm.DeletedAt          `gorm:"index" json:"-"`
}

func (GoogleMerchantOffer) TableName() string {
	return "google_merchant_offers"
}
